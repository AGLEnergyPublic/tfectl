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

type Tag struct {
	Name          string `json:"name"`
	ID            string `json:"id"`
	InstanceCount int    `json:"instance_count"`
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Query TFE tags",
	Long:  `Query TFE tags`,
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE tags",
	Long:  `List TFE tags`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		filter, _ := cmd.Flags().GetString("filter")
		query, _ := cmd.Flags().GetString("query")
		search, _ := cmd.Flags().GetString("search")

		tags, err := listTags(client, organization, filter, search)
		check(err)

		var tagList []Tag
		var tagListJson []byte

		for _, tag := range tags {
			var tmpTag Tag
			log.Debugf("Processing tag: %s - %s", tag.Name, tag.ID)
			entry := fmt.Sprintf(`{"name":"%s","id":"%s","instance_count":%d}`, tag.Name, tag.ID, tag.InstanceCount)
			err := json.Unmarshal([]byte(entry), &tmpTag)
			check(err)

			tagList = append(tagList, tmpTag)
		}
		tagListJson, _ = json.MarshalIndent(tagList, "", "  ")

		if query != "" {
			outputJsonStr, err := resources.JqRun(tagListJson, query)
			check(err)
			cmd.Println(string(outputJsonStr))
		} else {
			cmd.Println(string(tagListJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)

	// List sub-command
	tagCmd.AddCommand(tagListCmd)
	tagListCmd.Flags().String("filter", "", "Filters the list of all Org tags based-on those associated with a given WorkspaceId")
	tagListCmd.Flags().String("search", "", "A search query string. Organization tags are searchable by name likeness, takes precedence over --filter")
}

func listTags(client *tfe.Client, organization string, filter string, search string) ([]*tfe.OrganizationTag, error) {
	log.Debugf("Filter: %s", filter)
	results := []*tfe.OrganizationTag{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d.\n", currentPage)
		options := &tfe.OrganizationTagsListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
			Filter: filter,
			Query:  search,
		}

		p, err := client.OrganizationTags.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}
		results = append(results, p.Items...)

		if p.Pagination.NextPage == 0 {
			break
		}

		currentPage++
	}

	return results, nil
}
