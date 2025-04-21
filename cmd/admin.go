package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AGLEnergyPublic/tfectl/resources"
	tfe "github.com/hashicorp/go-tfe"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage TFE admin operations",
	Long:  `Admin related methods that the Terraform Enterprise API supports`,
}

// Only supports List and ForceCancel operation
var adminRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Manage TFE admin run operation",
	Long:  `Run related methods supported by TFE Admin API`,
}

var adminRunListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all runs",
	Long:  `List all runs`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		filter, _ := cmd.Flags().GetString("filter")
		query, _ := cmd.Flags().GetString("query")

		var adminRunListJson []byte
		// struct Run defined in run.go
		var adminRunList []Run

		runs, err := listAdminRuns(client, filter)
		check(err)

		for _, run := range runs {
			var tmpAdminRun Run

			log.Debugf("Processing run %s", run.ID)

			workspaceName, _ := getWorkspaceNameByID(client, organization, run.Workspace.ID)
			entry := fmt.Sprintf(`{
        "id":"%s",
        "workspace_id":"%s",
        "workspace_name":"%s",
        "status":"%s"
      }`,
				run.ID,
				run.Workspace.ID,
				workspaceName,
				run.Status)

			err = json.Unmarshal([]byte(entry), &tmpAdminRun)
			check(err)

			adminRunList = append(adminRunList, tmpAdminRun)
		}

		adminRunListJson, _ = json.MarshalIndent(adminRunList, "", "  ")
		if query != "" {
			outputJsonStr, err := resources.JqRun(adminRunListJson, query)
			check(err)
			cmd.Println(string(outputJsonStr))
		} else {
			cmd.Println(string(adminRunListJson))
		}
	},
}

var adminRunForceCancelCmd = &cobra.Command{
	Use:   "force-cancel",
	Short: "Force cancel runs",
	Long:  `Force cancel runs`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		query, _ := cmd.Flags().GetString("query")

		idList := strings.Split(ids, ",")

		var adminRunForceCancelListJson []byte
		var adminRunForceCancelList []Run

		for _, id := range idList {
			var tmpRun Run

			// get workspaceID from run
			run, _ := getRun(client, id)
			workspaceID := run.Workspace.ID

			// get workspaceName from run
			workspaceName, _ := getWorkspaceNameByID(client, organization, workspaceID)

			adminForceCancelRun(client, id)

			entry := fmt.Sprintf(`{
        "id":"%s",
        "workspace_id":"%s",
        "workspace_name":"%s",
        "status":"%s"
      }`,
				id,
				workspaceID,
				workspaceName,
				"cancelling")
			err = json.Unmarshal([]byte(entry), &tmpRun)
			check(err)

			adminRunForceCancelList = append(adminRunForceCancelList, tmpRun)
		}
		adminRunForceCancelListJson, _ = json.MarshalIndent(adminRunForceCancelList, "", "  ")
		if query != "" {
			outputJsonStr, err := resources.JqRun(adminRunForceCancelListJson, query)
			check(err)
			cmd.Println(string(outputJsonStr))
		} else {
			cmd.Println(string(adminRunForceCancelListJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminRunCmd)
	adminRunCmd.AddCommand(adminRunListCmd)
	adminRunCmd.AddCommand(adminRunForceCancelCmd)

	// List sub-command
	adminRunListCmd.Flags().String("filter", "pending", "List plans matching filter - default: pending")

	// Force-Cancel sub-command
	adminRunForceCancelCmd.Flags().String("ids", "", "Force cancel comma-separated string of runIDs")
}

func listAdminRuns(client *tfe.Client, filter string) ([]*tfe.AdminRun, error) {
	results := []*tfe.AdminRun{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d of runs\n", currentPage)
		options := &tfe.AdminRunsListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			RunStatus: filter,
		}
		r, err := client.Admin.Runs.List(context.Background(), options)

		if err != nil {
			return nil, err
		}
		results = append(results, r.Items...)

		if r.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}

func adminForceCancelRun(client *tfe.Client, runID string) {
	comment := fmt.Sprintf("Force-cancel run as Admin %s", runID)

	options := tfe.AdminRunForceCancelOptions{
		Comment: &comment,
	}

	err := client.Admin.Runs.ForceCancel(context.Background(), runID, options)
	check(err)
}
