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

type RegistryModule struct {
	ID                  string                   `json:"id"`
	Name                string                   `json:"name"`
	Provider            string                   `json:"provider"`
	RegistryName        tfe.RegistryName         `json:"registry_name"`
	Namespace           string                   `json:"namespace"`
	PublishingMechanism tfe.PublishingMechanism  `json:"publishing_mechanism"`
	Status              tfe.RegistryModuleStatus `json:"status"`
	TestConfig          bool                     `json:"test_config"`
	VCSRepo             string                   `json:"vcs_repo"`
	ModuleLatestVersion string                   `json:"module_latest_version"`
}

var registryModuleCmd = &cobra.Command{
	Use:   "registry-module",
	Short: "Query/Manage TFE private module registry",
	Long:  `Query/Manage TFE private module regsistry.`,
}

var registryModuleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List private modules in TFE Organization",
	Long:  `List private modules in TFE Organization.`,
	Run: func(cmd *cobra.Command, args []string) {

		organization, client, err := resources.Setup(cmd)
		check(err)

		query, _ := cmd.Flags().GetString("query")

		moduleList, err := listPrivateModules(client, organization)
		check(err)

		moduleListJson, _ := json.MarshalIndent(moduleList, "", "  ")

		if query != "" {
			resources.JqRun(moduleListJson, query)
		} else {
			fmt.Println(string(moduleListJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(registryModuleCmd)
	registryModuleCmd.AddCommand(registryModuleListCmd)
}

func listPrivateModules(client *tfe.Client, organization string) ([]RegistryModule, error) {
	results := []RegistryModule{}
	result := RegistryModule{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d\n", currentPage)
		options := &tfe.RegistryModuleListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
		}

		rms, err := client.RegistryModules.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}

		for _, rmItem := range rms.Items {
			result.RegistryName = rmItem.RegistryName
			result.ID = rmItem.ID
			result.Name = rmItem.Name
			result.Namespace = rmItem.Namespace
			result.VCSRepo = rmItem.VCSRepo.DisplayIdentifier
			result.PublishingMechanism = rmItem.PublishingMechanism
			result.Provider = rmItem.Provider
			result.Status = rmItem.Status

			if rmItem.TestConfig != nil {
				result.TestConfig = rmItem.TestConfig.TestsEnabled
			}

			for _, rmvs := range rmItem.VersionStatuses {
				if rmvs.Status == "ok" {
					// get latest module version which published without errors
					result.ModuleLatestVersion = rmvs.Version
					break
				}
			}

			results = append(results, result)
		}

		if rms.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}
