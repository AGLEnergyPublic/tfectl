package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AGLEnergyPublic/tfectl/resources"
	tfe "github.com/hashicorp/go-tfe"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type PolicySet struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	Workspaces []string `json:"workspaces"`
}

var policySetCmd = &cobra.Command{
	Use:   "policy-set",
	Short: "Query TFE policy sets",
	Long:  `Query TFE policy sets.`,
}

var policySetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List policy sets in a TFE Organization",
	Long:  `List policy sets in a TFE Organization.`,
	Run: func(cmd *cobra.Command, args []string) {

		organization, client, err := resources.Setup(cmd)
		check(err)

		filter, _ := cmd.Flags().GetString("filter")
		query, _ := cmd.Flags().GetString("query")

		policySets, err := listPolicySets(client, organization, filter)
		check(err)

		var policySetList []PolicySet
		var policySetListJson []byte

		for _, policySet := range policySets {
			var tmpPolicySet PolicySet
			var tmpPolicySetWorkspaceList []string
			log.Debugf("Processing policySet: %s - %s", policySet.Name, policySet.ID)
			for _, workspace := range policySet.Workspaces {
				log.Debugf("Processing workspaces in policySet: %s - %s - %s", policySet.Name, workspace.Name, workspace.ID)
				tmpPolicySetWorkspaceList = append(tmpPolicySetWorkspaceList, workspace.ID)
			}
			entry := fmt.Sprintf(`{"name":"%s","id":"%s","kind":"%s","workspaces":"%v"}`, policySet.Name, policySet.ID, policySet.Kind, tmpPolicySetWorkspaceList)
			err := json.Unmarshal([]byte(entry), &tmpPolicySet)
			check(err)

			policySetList = append(policySetList, tmpPolicySet)
		}
		policySetListJson, _ = json.MarshalIndent(policySetList, "", "  ")

		if query != "" {
			resources.JqRun(policySetListJson, query)
		} else {
			fmt.Println(string(policySetListJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(policySetCmd)
	policySetCmd.AddCommand(policySetListCmd)

	// List sub-command
	policySetListCmd.Flags().String("fitler", "", "Search for policy sets by name")
}

func listPolicySets(client *tfe.Client, organization string, filter string) ([]*tfe.PolicySet, error) {
	results := []*tfe.PolicySet{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d\n", currentPage)
		options := &tfe.PolicySetListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			Search: filter,
		}

		ps, err := client.PolicySets.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}
		results = append(results, ps.Items...)

		if ps.Pagination.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}
