package cmd

import (
	"context"
	"encoding/json"

	"github.com/AGLEnergyPublic/tfectl/resources"
	tfe "github.com/hashicorp/go-tfe"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type AgentPool struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	AgentCount         int      `json:"agent_count"`
	OrganizationScoped bool     `json:"organization_scoped"`
	Organization       string   `json:"organization"`
	Workspaces         []string `json:"workspaces"`
	AllowedWorkspaces  []string `json:"allowed_workspaces"`
}

var agentPoolCmd = &cobra.Command{
	Use:   "agent-pool",
	Short: "Query TFE/TFC Agent Pools",
	Long:  `Query TFE/TFC Agent Pools.`,
}

var agentPoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured Agent Pools in the TFE/TFC Organization",
	Long:  `List all configured Agent Pools in the TFE/TFC Organization.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		query, _ := cmd.Flags().GetString("query")

		agentPools, err := listAgentPools(client, organization)
		check(err)

		agentPoolsJson, err := json.MarshalIndent(agentPools, "", "  ")
		check(err)

		if query != "" {
			outputJsonStr, err := resources.JqRun(agentPoolsJson, query)
			check(err)
			cmd.Println(string(outputJsonStr))
		} else {
			cmd.Println(string(agentPoolsJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(agentPoolCmd)
	agentPoolCmd.AddCommand(agentPoolListCmd)
}

func listAgentPools(client *tfe.Client, organization string) ([]AgentPool, error) {
	results := []AgentPool{}
	result := AgentPool{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d\n", currentPage)
		options := &tfe.AgentPoolListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
		}

		aps, err := client.AgentPools.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}

		for _, apsItem := range aps.Items {
			result.ID = apsItem.ID
			result.Name = apsItem.Name
			result.AgentCount = apsItem.AgentCount
			result.OrganizationScoped = apsItem.OrganizationScoped
			if apsItem.Organization != nil {
				result.Organization = apsItem.Organization.Name
			}

			if len(apsItem.Workspaces) > 0 {
				for _, wk := range apsItem.Workspaces {
					check(err)
					result.Workspaces = append(result.Workspaces, wk.ID)
				}
			}

			if len(apsItem.AllowedWorkspaces) > 0 {
				for _, wk := range apsItem.AllowedWorkspaces {
					check(err)
					result.AllowedWorkspaces = append(result.AllowedWorkspaces, wk.ID)
				}
			}

			results = append(results, result)
		}

		if aps.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}
