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

type RegistryProvider struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Namespace    string           `json:"namespace"`
	RegistryName tfe.RegistryName `json:"registry_name"`
}

type ProviderPlatform struct {
	ID       string `json:"id"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
}

type PrivateProviderDetail struct {
	RegistryProvider
	ProviderLatestVersion string             `json:"provider_latest_version"`
	ProviderPlatforms     []ProviderPlatform `json:"provider_platforms"`
}

var registryProviderCmd = &cobra.Command{
	Use:   "registry-provider",
	Short: "Manage TFE private provider Registry",
	Long:  `Manage TFE private provider Registry.`,
}

var registryProviderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List private providers in a TFE Organization",
	Long:  `List private providers in a TFE Organization.`,
	Run: func(cmd *cobra.Command, args []string) {

		organization, client, err := resources.Setup(cmd)
		check(err)

		filter, _ := cmd.Flags().GetString("filter")
		query, _ := cmd.Flags().GetString("query")

		providerList, err := listPrivateProviders(client, organization, filter)
		check(err)

		providerListJson, _ := json.MarshalIndent(providerList, "", "  ")

		if query != "" {
			resources.JqRun(providerListJson, query)
		} else {
			fmt.Println(string(providerListJson))
		}
	},
}

var registryProviderGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show details of the TFE/C private provider registry",
	Long:  `Show details of the TFE/C private provider registry`,
	Run: func(cmd *cobra.Command, args []string) {

		organization, client, err := resources.Setup(cmd)
		check(err)

		name, _ := cmd.Flags().GetString("name")
		query, _ := cmd.Flags().GetString("query")

		privateProviderDetail, err := getPrivateProviderDetails(client, organization, name)
		check(err)

		privateProviderDetailJson, _ := json.MarshalIndent(privateProviderDetail, "", "  ")
		if query != "" {
			resources.JqRun(privateProviderDetailJson, query)
		} else {
			fmt.Println(string(privateProviderDetailJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(registryProviderCmd)
	registryProviderCmd.AddCommand(registryProviderListCmd)
	registryProviderCmd.AddCommand(registryProviderGetCmd)

	// List sub-command
	registryProviderListCmd.Flags().String("filter", "", "Search for private provider registries by name")

	// Show sub-command
	registryProviderGetCmd.Flags().String("name", "", "Name of the private provider in the registry")
}

func listPrivateProviders(client *tfe.Client, organization string, filter string) ([]RegistryProvider, error) {
	results := []RegistryProvider{}
	result := RegistryProvider{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d\n", currentPage)
		options := &tfe.RegistryProviderListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			Search:       filter,
			RegistryName: "private",
		}

		rps, err := client.RegistryProviders.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}

		for _, rpItem := range rps.Items {
			result.RegistryName = rpItem.RegistryName
			result.ID = rpItem.ID
			result.Name = rpItem.Name
			result.Namespace = rpItem.Namespace

			results = append(results, result)
		}

		if rps.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}

func getPrivateProviderDetails(client *tfe.Client, organization string, name string) (PrivateProviderDetail, error) {
	var result PrivateProviderDetail

	registryProviderList, err := listPrivateProviders(client, organization, name)
	check(err)

	if len(registryProviderList) > 1 {
		return result, fmt.Errorf("query returns more than one Provider for name: %s", name)
	}

	registryProvider := registryProviderList[0]

	registryProviderID := tfe.RegistryProviderID{
		OrganizationName: organization,
		RegistryName:     "private",
		Namespace:        registryProvider.Namespace,
		Name:             registryProvider.Name,
	}

	pr, err := client.RegistryProviders.Read(context.Background(), registryProviderID, &tfe.RegistryProviderReadOptions{})
	check(err)

	//Get latest provider version
	currentPage := 1
	prv, err := client.RegistryProviderVersions.List(context.Background(), registryProviderID, &tfe.RegistryProviderVersionListOptions{})
	check(err)

	if len(prv.Items) == 0 {
		return result, fmt.Errorf("unable to query Provider with given id: %s", registryProvider.ID)
	}

	log.Debugf("CurrentPage: %d, LastPage: %d", prv.CurrentPage, prv.TotalPages)
	lastPage := prv.TotalPages

	if currentPage != lastPage {
		prv, err = client.RegistryProviderVersions.List(context.Background(), registryProviderID, &tfe.RegistryProviderVersionListOptions{ListOptions: tfe.ListOptions{PageNumber: lastPage}})
		check(err)
	}

	items := prv.Items

	latestProviderVersion := items[len(items)-1]
	latestVersion := latestProviderVersion.Version

	rpv := tfe.RegistryProviderVersionID{
		RegistryProviderID: registryProviderID,
		Version:            latestVersion,
	}

	//Get provider platform details
	prpv, err := client.RegistryProviderPlatforms.List(context.Background(), rpv, &tfe.RegistryProviderPlatformListOptions{ListOptions: tfe.ListOptions{PageSize: 100}})
	check(err)

	if len(prpv.Items) == 0 {
		return result, fmt.Errorf("unable to query Provider Platforms for Provider with given name: %s", name)
	}

	var providerPlatforms []ProviderPlatform

	result.ID = registryProvider.ID
	result.Name = pr.Name
	result.Namespace = pr.Namespace
	result.RegistryName = pr.RegistryName
	result.ProviderLatestVersion = latestVersion

	for _, platform := range prpv.Items {
		var tmpProviderPlatform ProviderPlatform

		tmpProviderPlatform.ID = platform.ID
		tmpProviderPlatform.OS = platform.OS
		tmpProviderPlatform.Arch = platform.Arch
		tmpProviderPlatform.Filename = platform.Filename

		providerPlatforms = append(providerPlatforms, tmpProviderPlatform)
	}

	result.ProviderPlatforms = providerPlatforms

	return result, nil
}
