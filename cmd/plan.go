package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/AGLEnergyPublic/tfectl/resources"
	tfe "github.com/hashicorp/go-tfe"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Plan struct {
	ID                        string `json:"id"`
	HasChanges                bool   `json:"has_changes"`
	Status                    string `json:"status"`
	ResourceAdditions         int    `json:"resource_additions"`
	ResourceChanges           int    `json:"resource_changes"`
	ResourceDestructions      int    `json:"resource_destructions"`
	ResourceImports           int    `json:"resource_imports"`
	ChangedResourceProperties []any  `json:"changed_resource_properties"`
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Query TFE Plans",
	Long:  `Query TFE Plans.`,
}

var planShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of a plan with given planID",
	Long:  `Show details of a plan with give planID.`,
	Run: func(cmd *cobra.Command, args []string) {
		var buffer bytes.Buffer
		var jsonEnc = json.NewEncoder(&buffer)

		jsonEnc.SetEscapeHTML(false)
		jsonEnc.SetIndent("", "  ")

		_, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		detailedChanges, _ := cmd.Flags().GetBool("detailed-changes")
		query, _ := cmd.Flags().GetString("query")

		var planShowJson []byte
		var planShowList []Plan

		idList := strings.Split(ids, ",")
		for _, id := range idList {

			log.Debugf("Querying plan with id: %s", id)
			plan, _ := showPlan(client, id, detailedChanges)

			planShowList = append(planShowList, plan)
		}

		_ = jsonEnc.Encode(planShowList)
		planShowJson = buffer.Bytes()

		if query != "" {
			outputJsonStr, err := resources.JqRun(planShowJson, query)
			check(err)
			cmd.Println(string(outputJsonStr))
		} else {
			cmd.Println(string(planShowJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.AddCommand(planShowCmd)

	planShowCmd.Flags().String("ids", "", "Query comma-separated string of planIDs")
	planShowCmd.Flags().Bool("detailed-changes", false, "Returns a map describing the changed resource attributes")
}

func showPlan(client *tfe.Client, planID string, detailedChanges bool) (Plan, error) {
	result := Plan{}
	pl, err := client.Plans.Read(context.Background(), planID)
	check(err)

	result.ID = pl.ID
	result.Status = string(pl.Status)
	result.ResourceChanges = pl.ResourceChanges
	result.ResourceAdditions = pl.ResourceAdditions
	result.ResourceDestructions = pl.ResourceDestructions
	result.ResourceImports = pl.ResourceImports
	result.HasChanges = pl.HasChanges

	if string(pl.Status) == "finished" && detailedChanges {
		planJsonOut, err := client.Plans.ReadJSONOutput(context.Background(), planID)
		check(err)
		// Generate Query string
		// This query parses the Output JSON and extracts the resources that are changing
		// <address_of_changing_resource>: {
		//    <attribute.0>: <current_value> -> <planned_value>,
		//    <attribute.1>: <current_value> -> <planned_value>,
		// }
		queryChangeString := `.resource_changes[] | select(.change.actions | inside(["create", "read", "update", "delete"])) | { (.address): (.change.after | with_entries(select(.value != .before)) | with_entries(.value = "\(.before) -> \(.value)")) }`
		out, err := resources.JqRun(planJsonOut, queryChangeString)
		check(err)
		json.Unmarshal(out, &result.ChangedResourceProperties)
	}

	return result, nil
}
