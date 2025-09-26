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

type Policy struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	Enforce        string `json:"enforce"`
	PolicySetCount int    `json:"policy_set_count"`
}

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Query TFE policies",
	Long:  `Query TFE policies.`,
}

var policyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE policies",
	Long:  `List TFE policies.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		filter, _ := cmd.Flags().GetString("filter")

		policies, err := listPolicies(client, organization, filter)
		check(err)

		var policyList []Policy
		var policyListJson []byte

		for _, policy := range policies {
			var tmpPolicy Policy
			log.Debugf("Processing policy: %s - %s", policy.Name, policy.ID)
			entry := fmt.Sprintf(`{
        "name":"%s",
        "id":"%s",
        "kind":"%s",
        "enforce":"%s",
        "policy_set_count":%d
      }`,
				policy.Name,
				policy.ID,
				policy.Kind,
				policy.Enforce[0].Mode,
				policy.PolicySetCount)
			err := json.Unmarshal([]byte(entry), &tmpPolicy)
			check(err)

			policyList = append(policyList, tmpPolicy)
		}
		policyListJson, _ = json.MarshalIndent(policyList, "", "  ")

		outputData(cmd, policyListJson)
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)

	// List sub-command
	policyCmd.AddCommand(policyListCmd)
	policyListCmd.Flags().String("filter", "", "Search for policy by name")
}

func listPolicies(client *tfe.Client, organization string, filter string) ([]*tfe.Policy, error) {
	results := []*tfe.Policy{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d.\n", currentPage)
		options := &tfe.PolicyListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			Search: filter,
		}

		p, err := client.Policies.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}
		results = append(results, p.Items...)

		if p.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}
