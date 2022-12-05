package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"tfectl/resources"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type User struct {
	ID     string                           `json:"user_id"`
	Email  string                           `json:"email"`
	Status tfe.OrganizationMembershipStatus `json:"status"`
}

type TeamDetail struct {
	Team  Team   `json:"team"`
	Users []User `json:"user_list"`
}

type Team struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	UserCount int    `json:"user_count"`
}

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage TFE teams",
	Long:  `Manage TFE teams.`,
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE teams",
	Long:  `List TFE teams.`,
	Run: func(cmd *cobra.Command, args []string) {
		// setup
		organization, client, err := resources.Setup(cmd)
		check(err)

		query, _ := cmd.Flags().GetString("query")

		// List teams.
		teams, err := listTeams(client, organization, []string{})
		check(err)

		var teamJson []byte
		var teamList []Team

		for _, team := range teams {
			var tmpTeam Team

			tmpTeam.ID = team.ID
			tmpTeam.Name = team.Name
			tmpTeam.UserCount = team.UserCount

			log.Debugf("Adding team %v", tmpTeam)
			teamList = append(teamList, tmpTeam)
		}

		teamJson, _ = json.MarshalIndent(teamList, "", "  ")

		if query != "" {
			resources.JqRun(teamJson, query)
		} else {
			fmt.Println(string(teamJson))
		}

	},
}

var teamGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get TFE team details",
	Long:  `Get TFE team details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// setup
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		names, _ := cmd.Flags().GetString("names")

		if names != "" && ids != "" {
			log.Fatal("names and ids are mutually exclusive, use one or the other!")
		}

		if names == "" && ids == "" {
			log.Fatal("please provide one of ids or names to perform this operation!")
		}

		query, _ := cmd.Flags().GetString("query")

		var teamJson []byte
		var teamList []TeamDetail

		if ids != "" {
			idList := strings.Split(ids, ",")

			for _, id := range idList {
				var tmpTeam TeamDetail

				team, _ := readTeam(client, id)
				tmpTeam = genTeamDetail(client, team)

				log.Debugf("Adding team %v", tmpTeam)
				teamList = append(teamList, tmpTeam)
			}
		}

		if names != "" {
			namesList := strings.Split(names, ",")
			teams, err := listTeams(client, organization, namesList)
			check(err)

			for _, team := range teams {
				var tmpTeam TeamDetail

				tmpTeam = genTeamDetail(client, team)

				log.Debugf("Adding team %v", tmpTeam)
				teamList = append(teamList, tmpTeam)
			}
		}

		teamJson, _ = json.MarshalIndent(teamList, "", "  ")

		if query != "" {
			resources.JqRun(teamJson, query)
		} else {
			fmt.Println(string(teamJson))
		}

	},
}

func init() {
	rootCmd.AddCommand(teamCmd)

	// List sub-command
	teamCmd.AddCommand(teamListCmd)
	teamListCmd.Flags().Bool("detail", false, "Provide team membership details: userID, email and status")

	// Get sub-command
	teamCmd.AddCommand(teamGetCmd)
	// Begin mutually-exclusive flags
	teamGetCmd.Flags().String("ids", "", "Comma separated string of team ids")
	teamGetCmd.Flags().String("names", "", "comma separated string of Team names to filter")
	// End mutually-exclusive flags
}

func listTeams(client *tfe.Client, organization string, filters []string) ([]*tfe.Team, error) {
	results := []*tfe.Team{}
	currentPage := 1

	// Go through the pages of results until there is no more pages.
	for {
		log.Debugf("Processing page %d.\n", currentPage)

		options := &tfe.TeamListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
		}

		if len(filters) != 0 {
			options.Names = filters
		}

		log.Debugf("options: %v", options)

		t, err := client.Teams.List(context.Background(), organization, options)
		check(err)

		log.Debugf("%v", t.Pagination.TotalPages)
		log.Debugf("%v", t.Pagination.NextPage)

		results = append(results, t.Items...)

		// Check if there is another page to retrieve.
		if t.Pagination.NextPage == 0 {
			break
		}

		// Increment the page number.
		currentPage++
	}

	return results, nil
}

func getOrgMember(client *tfe.Client, orgMemID string) (User, error) {
	result := User{}

	o, err := client.OrganizationMemberships.Read(context.Background(), orgMemID)
	check(err)

	result.ID = o.User.ID
	result.Email = o.Email
	result.Status = o.Status

	return result, nil
}

func readTeam(client *tfe.Client, teamID string) (*tfe.Team, error) {
	result, err := client.Teams.Read(context.Background(), teamID)
	check(err)

	return result, nil
}

func genTeamDetail(client *tfe.Client, team *tfe.Team) TeamDetail {
	result := TeamDetail{}

	result.Team.ID = team.ID
	result.Team.Name = team.Name
	result.Team.UserCount = team.UserCount

	for _, orgMem := range team.OrganizationMemberships {
		var tmpUser User
		log.Debugf("Org mem: %v", &orgMem)
		tmpUser, _ = getOrgMember(client, orgMem.ID)

		log.Debugf("Adding User %v", tmpUser)
		result.Users = append(result.Users, tmpUser)
	}

	return result
}
