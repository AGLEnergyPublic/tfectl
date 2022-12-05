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

type Run struct {
	ID            string `json:"id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	Status        string `json:"status"`
	//HasChanges    bool `json:"has_changes"`
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Manage TFE runs",
	Long:  `Manage TFE runs.`,
}

var runQueueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Queue TFE runs",
	Long:  `Queue TFE runs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// bulk queue function
		organization, client, err := resources.Setup(cmd)
		check(err)

		filter, _ := cmd.Flags().GetString("filter")
		query, _ := cmd.Flags().GetString("query")

		var runListJson []byte
		var runList []Run

		workspaces, err := listWorkspaces(client, organization, filter)
		check(err)

		for _, workspace := range workspaces {
			var tmpRun Run

			log.Debugf("Queuing run on %s", workspace.Name)
			run, err := queueRun(client, organization, workspace)
			check(err)

			entry := fmt.Sprintf(`{"id":"%s","workspace_id":"%s","workspace_name":"%s","status":"%s"}`, run.ID, run.Workspace.ID, workspace.Name, run.Status)

			err = json.Unmarshal([]byte(entry), &tmpRun)
			check(err)

			runList = append(runList, tmpRun)
		}

		runListJson, _ = json.MarshalIndent(runList, "", "  ")
		if query != "" {
			resources.JqRun(runListJson, query)
		} else {
			fmt.Println(string(runListJson))
		}
	},
}

var runApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply Runs with given runIDs",
	Long:  `Apply Runs with given runIDs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// apply run function
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		query, _ := cmd.Flags().GetString("query")

		var runApplyListJson []byte
		var runApplyList []Run

		idList := strings.Split(ids, ",")
		for _, id := range idList {
			var tmpRun Run

			// get workspaceID from run
			run, _ := getRun(client, id)
			workspaceID := run.Workspace.ID

			// get workspaceName from run
			workspaceName, _ := getWorkspaceNameByID(client, organization, workspaceID)

			log.Debugf("Applying run with id: %s", id)
			applyRun(client, id)

			entry := fmt.Sprintf(`{"id":"%s","workspace_id":"%s","workspace_name":"%s","status":"%s"}`, id, workspaceID, workspaceName, "applying")
			err = json.Unmarshal([]byte(entry), &tmpRun)
			check(err)
			runApplyList = append(runApplyList, tmpRun)
		}

		runApplyListJson, _ = json.MarshalIndent(runApplyList, "", "  ")
		if query != "" {
			resources.JqRun(runApplyListJson, query)
		} else {
			fmt.Println(string(runApplyListJson))
		}
	},
}

var runGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Runs with given runIDs",
	Long:  `Get Runs with given runIDs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get run function
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		query, _ := cmd.Flags().GetString("query")

		var runGetListJson []byte
		var runGetList []Run

		idList := strings.Split(ids, ",")
		for _, id := range idList {
			var tmpRun Run

			log.Debugf("Querying run with id: %s", id)
			run, _ := getRun(client, id)
			workspaceID := run.Workspace.ID

			// get workspaceName from run
			workspaceName, _ := getWorkspaceNameByID(client, organization, workspaceID)

			entry := fmt.Sprintf(`{"id":"%s","workspace_id":"%s","workspace_name":"%s","status":"%s"}`, run.ID, workspaceID, workspaceName, run.Status)
			err = json.Unmarshal([]byte(entry), &tmpRun)
			check(err)

			runGetList = append(runGetList, tmpRun)
		}

		runGetListJson, _ = json.MarshalIndent(runGetList, "", "  ")
		if query != "" {
			resources.JqRun(runGetListJson, query)
		} else {
			fmt.Println(string(runGetListJson))
		}
	},
}

var runCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel Runs with given runIDs",
	Long:  `Cancel Runs with given runIDs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// cancel run function
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		filter, _ := cmd.Flags().GetString("filter")

		if filter != "" && ids != "" {
			log.Fatal("filter and ids are mutually exclusive, use one or the other!")
		}

		if filter == "" && ids == "" {
			log.Fatal("please provide one of ids or filter to perform this operation!")
		}

		force, _ := cmd.Flags().GetBool("force")
		query, _ := cmd.Flags().GetString("query")

		var runCancelListJson []byte
		var runCancelList []Run
		var idList []string

		if filter != "" {
			workspaces, err := listWorkspaces(client, organization, filter)
			check(err)

			for _, workspace := range workspaces {
				// get runIds
				idList = append(idList, workspace.CurrentRun.ID)
			}
		}

		if ids != "" {
			idList = strings.Split(ids, ",")
		}

		for _, id := range idList {
			var tmpRun Run

			// get workspaceid from run
			run, _ := getRun(client, id)
			workspaceID := run.Workspace.ID

			// get workspacename from run
			workspaceName, _ := getWorkspaceNameByID(client, organization, workspaceID)

			log.Debugf("Cancelling run with id: %s", id)
			if force {
				forceCancelRun(client, id)
			} else {
				cancelRun(client, id)
			}

			entry := fmt.Sprintf(`{"id":"%s","workspace_id":"%s","workspace_name":"%s","status":"%s"}`, id, workspaceID, workspaceName, "cancelling")
			err = json.Unmarshal([]byte(entry), &tmpRun)
			check(err)

			runCancelList = append(runCancelList, tmpRun)
		}

		runCancelListJson, _ = json.MarshalIndent(runCancelList, "", "  ")
		if query != "" {
			resources.JqRun(runCancelListJson, query)
		} else {
			fmt.Println(string(runCancelListJson))
		}
	},
}

var runDiscardCmd = &cobra.Command{
	Use:   "discard",
	Short: "Discard Runs with given runIDs",
	Long:  `Discard Runs with given runIDs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// discard run function
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		filter, _ := cmd.Flags().GetString("filter")

		if filter != "" && ids != "" {
			log.Fatal("filter and ids are mutually exclusive, use one or the other!")
		}

		if filter == "" && ids == "" {
			log.Fatal("please provide one of ids or filter to perform this operation!")
		}

		query, _ := cmd.Flags().GetString("query")

		var runDiscardListJson []byte
		var runDiscardList []Run
		var idList []string

		if filter != "" {
			workspaces, err := listWorkspaces(client, organization, filter)
			check(err)

			for _, workspace := range workspaces {
				// get runIds
				idList = append(idList, workspace.CurrentRun.ID)
			}
		}

		if ids != "" {
			idList = strings.Split(ids, ",")
		}

		for _, id := range idList {
			var tmpRun Run

			// get workspaceid from run
			run, _ := getRun(client, id)
			workspaceID := run.Workspace.ID

			// get workspacename from run
			workspaceName, _ := getWorkspaceNameByID(client, organization, workspaceID)

			log.Debugf("Discarding run with id: %s", id)
			discardRun(client, id)

			entry := fmt.Sprintf(`{"id":"%s","workspace_id":"%s","workspace_name":"%s","status":"%s"}`, id, workspaceID, workspaceName, "discarding")
			err = json.Unmarshal([]byte(entry), &tmpRun)
			check(err)

			runDiscardList = append(runDiscardList, tmpRun)
		}

		runDiscardListJson, _ = json.MarshalIndent(runDiscardList, "", "  ")
		if query != "" {
			resources.JqRun(runDiscardListJson, query)
		} else {
			fmt.Println(string(runDiscardListJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.AddCommand(runQueueCmd)
	runCmd.AddCommand(runApplyCmd)
	runCmd.AddCommand(runGetCmd)
	runCmd.AddCommand(runCancelCmd)
	runCmd.AddCommand(runDiscardCmd)

	// Queue sub-command
	runQueueCmd.Flags().String("filter", "", "Queue plans on workspaces matching filter")

	// Apply sub-command
	runApplyCmd.Flags().String("ids", "", "Apply comma-separated string of runIDs")

	// Get sub-command
	runGetCmd.Flags().String("ids", "", "Query comma-separated string of runIDs")

	// Cancel sub-command
	runCancelCmd.Flags().String("ids", "", "Cancel comma-separated string of runIDs")     // Mutually exclusive with `filter`
	runCancelCmd.Flags().String("filter", "", "Cancel run on workspaces matching filter") // Mutually exclusive with `ids`

	runCancelCmd.Flags().Bool("force", false, "Force cancel comma-separated string of runIDs")

	// Discard sub-command
	runDiscardCmd.Flags().String("ids", "", "Discard comma-separated string of runIDs")     // Mutually exclusive with `filter`
	runDiscardCmd.Flags().String("filter", "", "Discard run on workspaces matching filter") // Mutually exclusive with `ids`

}

func queueRun(client *tfe.Client, organization string, workspace *tfe.Workspace) (*tfe.Run, error) {

	message := fmt.Sprintf("Queue plan on %s", workspace.Name)
	options := tfe.RunCreateOptions{
		Message:   &message,
		Workspace: workspace,
	}

	result, err := client.Runs.Create(context.Background(), options)
	check(err)

	return result, nil
}

func applyRun(client *tfe.Client, runID string) {

	comment := fmt.Sprintf("Apply run %s", runID)
	options := tfe.RunApplyOptions{
		Comment: &comment,
	}

	err := client.Runs.Apply(context.Background(), runID, options)

	check(err)
}

func getRun(client *tfe.Client, runID string) (*tfe.Run, error) {

	result, err := client.Runs.Read(context.Background(), runID)

	check(err)

	return result, nil
}

func cancelRun(client *tfe.Client, runID string) {
	comment := fmt.Sprintf("Cancel run %s", runID)

	options := tfe.RunCancelOptions{
		Comment: &comment,
	}

	err := client.Runs.Cancel(context.Background(), runID, options)

	check(err)
}

func forceCancelRun(client *tfe.Client, runID string) {
	comment := fmt.Sprintf("Force-cancel run %s", runID)

	options := tfe.RunForceCancelOptions{
		Comment: &comment,
	}

	err := client.Runs.ForceCancel(context.Background(), runID, options)

	check(err)
}

func discardRun(client *tfe.Client, runID string) {
	comment := fmt.Sprintf("Discarding run %s", runID)

	options := tfe.RunDiscardOptions{
		Comment: &comment,
	}

	err := client.Runs.Discard(context.Background(), runID, options)

	check(err)
}
