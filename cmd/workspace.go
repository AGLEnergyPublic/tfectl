package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/AGLEnergyPublic/tfectl/resources"
	tfe "github.com/hashicorp/go-tfe"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// WorkspaceDetail for -detail flag with list
type WorkspaceDetail struct {
	Workspace
	CreatedDaysAgo         string `json:"created_days_ago"`
	UpdatedDaysAgo         string `json:"updated_days_ago"`
	LastRemoteRunDaysAgo   string `json:"last_remote_run_days_ago"`
	LastStateUpdateDaysAgo string `json:"last_state_update_days_ago"`
}

type Workspace struct {
	Name             string   `json:"name"`
	ID               string   `json:"id"`
	Locked           bool     `json:"locked"`
	ExecutionMode    string   `json:"execution_mode"`
	TerraformVersion string   `json:"terraform_version"`
	Tags             []string `json:"tags"`
}

type WorkspaceLock struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Locked bool   `json:"locked"`
}

// workspaceCmd represents the workspace command.
var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage TFE workspaces",
	Long:  `Manage TFE workspaces.`,
}

var workspaceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get/Show TFE workspace",
	Long:  `Get/Show TFE workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		organization, client, err := resources.Setup(cmd)
		check(err)

		ids, _ := cmd.Flags().GetString("ids")
		query, _ := cmd.Flags().GetString("query")
		idList := strings.Split(ids, ",")

		var workspaceList []WorkspaceDetail

		// Get workspace
		for _, id := range idList {
			workspace, err := getWorkspace(client, organization, id)
			check(err)

			workspaceList = append(workspaceList, workspace)
		}

		workspaceListJson, _ := json.MarshalIndent(workspaceList, "", "  ")
		if query != "" {
			resources.JqRun(workspaceListJson, query)
		} else {
			fmt.Println(string(workspaceListJson))
		}

	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE workspaces",
	Long:  `List TFE workspaces.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the command.
		organization, client, err := resources.Setup(cmd)
		check(err)

		detail, _ := cmd.Flags().GetBool("detail")
		filter, _ := cmd.Flags().GetString("filter")
		query, _ := cmd.Flags().GetString("query")

		// List workspaces.
		workspaces, err := listWorkspaces(client, organization, filter)
		check(err)

		var workspaceJson []byte

		if !detail {
			var workspaceList []Workspace
			// Print the workspace names.
			for _, workspace := range workspaces {
				var tmpWorkspace Workspace

				log.Debugf("Processing workspace: %s - %s", workspace.Name, workspace.ID)
				entry := fmt.Sprintf(`{"name":"%s","id":"%s","locked":%v,"execution_mode":"%s","terraform_version":"%s"}`, workspace.Name, workspace.ID, workspace.Locked, workspace.ExecutionMode, workspace.TerraformVersion)
				err := json.Unmarshal([]byte(entry), &tmpWorkspace)
				check(err)

				tmpWorkspace.Tags = workspace.TagNames

				workspaceList = append(workspaceList, tmpWorkspace)
			}
			workspaceJson, _ = json.MarshalIndent(workspaceList, "", "  ")

		} else {

			var workspaceList []WorkspaceDetail
			wg := sync.WaitGroup{}

			// Get additional details
			ch := make(chan WorkspaceDetail, len(workspaces))

			// Ratelimit
			var chunkSize int

			if len(workspaces) < 3 {
				chunkSize = len(workspaces)
			} else {
				chunkSize = 3
			}

			log.Debugf("RateLimit: %d", chunkSize)

			for i := 0; i < len(workspaces); i += chunkSize {
				if chunkSize > len(workspaces)-i {
					chunkSize = len(workspaces) - i
				}
				workspacesChunk := workspaces[i : i+chunkSize]
				for _, workspace := range workspacesChunk {
					wg.Add(1)

					id := workspace.ID
					name := workspace.Name

					go func(id string, name string) {
						log.Debugf("Processing workspace: %s - %s", name, id)
						tmpWorkspace, err := getWorkspace(client, organization, id)
						check(err)
						ch <- tmpWorkspace

						wg.Done()
					}(id, name)

					time.Sleep(500 * time.Millisecond)
				}

			}
			wg.Wait()
			for j := 0; j < len(workspaces); j++ {
				comm := <-ch
				workspaceList = append(workspaceList, comm)
			}

			workspaceJson, _ = json.MarshalIndent(workspaceList, "", "  ")
		}

		if query != "" {
			resources.JqRun(workspaceJson, query)
		} else {
			fmt.Println(string(workspaceJson))
		}

	},
}

var workspaceLockAllCmd = &cobra.Command{
	Use:   "lockall",
	Short: "Lock All TFE workspace",
	Long:  `Lock All TFE workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the command.
		organization, client, err := resources.Setup(cmd)
		check(err)

		reason, _ := cmd.Flags().GetString("reason")
		query, _ := cmd.Flags().GetString("query")

		var lockedWorkspaceList []WorkspaceLock

		lockedWorkspaceList, _ = lockAllWorkspaces(client, organization, &reason)

		lockedWorkspaceListJson, _ := json.MarshalIndent(lockedWorkspaceList, "", " ")
		if query != "" {
			resources.JqRun(lockedWorkspaceListJson, query)
		} else {
			fmt.Println(string(lockedWorkspaceListJson))
		}
	},
}

var workspaceLockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock specific TFE workspace",
	Long:  `Lock specific TFE workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the command.
		organization, client, err := resources.Setup(cmd)
		check(err)

		// Lock based on workspace ids
		ids, _ := cmd.Flags().GetString("ids")
		// Lock based on workspace name filters
		filter, _ := cmd.Flags().GetString("filter")

		if filter != "" && ids != "" {
			log.Fatal("filter and ids are mutually exclusive, use one or the other!")
		}

		if filter == "" && ids == "" {
			log.Fatal("please provide one of ids or filter to perform this operation!")
		}

		reason, _ := cmd.Flags().GetString("reason")
		// Parse JMESPath query from CLI
		query, _ := cmd.Flags().GetString("query")

		var lockedWorkspaceList []WorkspaceLock
		var lockedWorkspace WorkspaceLock
		var tmpWorkspace WorkspaceLite
		var workspaceList []WorkspaceLite

		if filter != "" {
			// get workspace Ids from filter
			workspaces, err := listWorkspaces(client, organization, filter)
			check(err)

			for _, workspace := range workspaces {
				tmpWorkspace.WorkspaceID = workspace.ID
				tmpWorkspace.WorkspaceName = workspace.Name

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		if ids != "" {
			workspaceIdList := strings.Split(ids, ",")
			for _, id := range workspaceIdList {
				workspaceName, err := getWorkspaceNameByID(client, organization, id)
				check(err)
				tmpWorkspace.WorkspaceID = id
				tmpWorkspace.WorkspaceName = workspaceName

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		for _, wrk := range workspaceList {
			var entry string

			workspace, err := lockWorkspace(client, organization, wrk.WorkspaceID, &reason)
			if err != nil {
				entry = fmt.Sprintf(`{"name":"%s","id":"%s","locked":true}`, wrk.WorkspaceName, wrk.WorkspaceID)
			} else {
				entry = fmt.Sprintf(`{"name":"%s","id":"%s","locked":%v}`, workspace.Name, workspace.ID, workspace.Locked)
			}
			_ = json.Unmarshal([]byte(entry), &lockedWorkspace)
			lockedWorkspaceList = append(lockedWorkspaceList, lockedWorkspace)
		}
		lockedWorkspaceListJson, _ := json.MarshalIndent(lockedWorkspaceList, "", "  ")
		if query != "" {
			resources.JqRun(lockedWorkspaceListJson, query)
		} else {
			fmt.Println(string(lockedWorkspaceListJson))
		}
	},
}

var workspaceUnlockAllCmd = &cobra.Command{
	Use:   "unlockall",
	Short: "Unlock All TFE workspace",
	Long:  `Unock All TFE workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the command.
		organization, client, err := resources.Setup(cmd)
		check(err)

		query, _ := cmd.Flags().GetString("query")

		var unlockedWorkspaceList []WorkspaceLock

		unlockedWorkspaceList, _ = unlockAllWorkspaces(client, organization)

		unlockedWorkspaceListJson, _ := json.MarshalIndent(unlockedWorkspaceList, "", "  ")
		if query != "" {
			resources.JqRun(unlockedWorkspaceListJson, query)
		} else {
			fmt.Println(string(unlockedWorkspaceListJson))
		}
	},
}

var workspaceUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock specific TFE workspace",
	Long:  `Unlock specific TFE workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the command.
		organization, client, err := resources.Setup(cmd)
		check(err)

		// Lock based on workspace ids
		ids, _ := cmd.Flags().GetString("ids")
		// Lock based on workspace name filters
		filter, _ := cmd.Flags().GetString("filter")

		if filter != "" && ids != "" {
			log.Fatal("filter and ids are mutually exclusive, use one or the other!")
		}

		if filter == "" && ids == "" {
			log.Fatal("please provide one of ids or filter to perform this operation!")
		}

		// Parse JMESPath query from CLI
		query, _ := cmd.Flags().GetString("query")

		var unlockedWorkspaceList []WorkspaceLock
		var unlockedWorkspace WorkspaceLock
		var tmpWorkspace WorkspaceLite
		var workspaceList []WorkspaceLite

		if filter != "" {
			// get workspace Ids from filter
			workspaces, err := listWorkspaces(client, organization, filter)
			check(err)

			for _, workspace := range workspaces {
				tmpWorkspace.WorkspaceID = workspace.ID
				tmpWorkspace.WorkspaceName = workspace.Name

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		if ids != "" {
			workspaceIdList := strings.Split(ids, ",")
			for _, id := range workspaceIdList {
				workspaceName, err := getWorkspaceNameByID(client, organization, id)
				check(err)
				tmpWorkspace.WorkspaceID = id
				tmpWorkspace.WorkspaceName = workspaceName

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		for _, wrk := range workspaceList {
			var entry string

			workspace, err := unlockWorkspace(client, organization, wrk.WorkspaceID)
			if err != nil {
				entry = fmt.Sprintf(`{"name":"%s","id":"%s","locked":false}`, workspace.Name, workspace.ID)
			} else {
				entry = fmt.Sprintf(`{"name":"%s","id":"%s","locked":%v}`, workspace.Name, workspace.ID, workspace.Locked)
			}

			_ = json.Unmarshal([]byte(entry), &unlockedWorkspace)
			unlockedWorkspaceList = append(unlockedWorkspaceList, unlockedWorkspace)
		}
		unlockedWorkspaceListJson, _ := json.MarshalIndent(unlockedWorkspaceList, "", "  ")
		if query != "" {
			resources.JqRun(unlockedWorkspaceListJson, query)
		} else {
			fmt.Println(string(unlockedWorkspaceListJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)

	// List sub-command
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceListCmd.Flags().Bool("detail", false, "Provide details about workspace")
	workspaceListCmd.Flags().String("filter", "", "Filter workspaces by name or by tag\nTo filter by tag, prefix filter with \"tags|\"\ne.g. \"tags|tagName,tag:Name\"")

	// Get sub-command
	workspaceCmd.AddCommand(workspaceGetCmd)
	workspaceGetCmd.Flags().String("ids", "", "Comma separated list of workspaceIDs")

	// Lock sub-command
	workspaceCmd.AddCommand(workspaceLockCmd)
	// Begin Mutually exclusive flags //
	workspaceLockCmd.Flags().String("ids", "", "Comma separated list of workspaceIDs to lock")
	workspaceLockCmd.Flags().String("filter", "", "Lock workspaces identified by a filter")
	// End Mutually exclusive flags //
	workspaceLockCmd.Flags().String("reason", "Locking", "Reason why workspace is locked")

	// LockAll sub-command
	workspaceCmd.AddCommand(workspaceLockAllCmd)
	workspaceLockAllCmd.Flags().String("reason", "Locking", "Reason why workspaces are locked")

	// Unlock sub-command
	workspaceCmd.AddCommand(workspaceUnlockCmd)
	workspaceUnlockCmd.Flags().String("ids", "", "Comma separated list of workspaceIDs to unlock")
	workspaceUnlockCmd.Flags().String("filter", "", "Unlock workspaces identified by a filter")

	// UnlockAll sub-command
	workspaceCmd.AddCommand(workspaceUnlockAllCmd)
}

func listWorkspaces(client *tfe.Client, organization string, filter string) ([]*tfe.Workspace, error) {
	results := []*tfe.Workspace{}
	currentPage := 1
	listOptions := &tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 50,
		},
	}

	// Parse filter to determine if it is a filter by workspace name or tag
	// Default behaviour is to filter by name
	if strings.Contains(filter, "tags|") {
		re := regexp.MustCompile(`tags\|(.*)`)
		match := re.FindStringSubmatch(filter)

		listOptions.Tags = match[1]
	} else {
		listOptions.Search = filter
	}

	// Go through the pages of results until there is no more pages.
	for {
		log.Debugf("Processing page %d.\n", currentPage)
		listOptions.PageNumber = currentPage

		options := listOptions

		w, err := client.Workspaces.List(context.Background(), organization, options)
		if err != nil {
			return nil, err
		}
		results = append(results, w.Items...)

		// Check if there is another poage to retrieve.
		if w.Pagination.NextPage == 0 {
			break
		}

		// Increment the page number.
		currentPage++
	}

	return results, nil
}

func getWorkspaceNameByID(client *tfe.Client, organization string, workspaceID string) (string, error) {
	workspaceRead, err := client.Workspaces.ReadByID(context.Background(), workspaceID)
	check(err)

	return workspaceRead.Name, nil
}

func getWorkspace(client *tfe.Client, organization string, workspaceID string) (WorkspaceDetail, error) {
	result := WorkspaceDetail{}

	workspaceRead, err := client.Workspaces.ReadByID(context.Background(), workspaceID)
	check(err)

	workspaceDetails, err := getWorkspaceDetails(client, organization, workspaceID)
	check(err)

	result.ID = workspaceRead.ID
	result.Name = workspaceRead.Name
	result.Locked = workspaceRead.Locked
	result.ExecutionMode = workspaceRead.ExecutionMode
	result.TerraformVersion = workspaceRead.TerraformVersion
	result.Tags = workspaceRead.TagNames
	result.CreatedDaysAgo = fmt.Sprintf("%f", time.Since(workspaceRead.CreatedAt).Hours()/24)
	result.UpdatedDaysAgo = fmt.Sprintf("%f", time.Since(workspaceRead.UpdatedAt).Hours()/24)
	result.LastRemoteRunDaysAgo = workspaceDetails.LastRemoteRunDaysAgo
	result.LastStateUpdateDaysAgo = workspaceDetails.LastStateUpdateDaysAgo

	return result, nil
}

func getWorkspaceDetails(client *tfe.Client, organization string, workspaceID string) (WorkspaceDetail, error) {
	results := WorkspaceDetail{}

	rList, err := client.Runs.List(context.Background(), workspaceID, &tfe.RunListOptions{
		ListOptions: tfe.ListOptions{
			PageSize: 1,
		},
	})
	check(err)

	lastRemoteRunDaysAgo := "NA"
	if len(rList.Items) > 0 {
		lastRemoteRunDaysAgo = fmt.Sprintf("%f", time.Since(rList.Items[0].CreatedAt).Hours()/24)
	}

	// Determine when current state-version was created
	lastStateUpdateDaysAgo := "NA"
	stateVersion, err := client.StateVersions.ReadCurrent(context.Background(), workspaceID)
	if err != nil {
		// Verify workspace has states
		if !strings.Contains(err.Error(), "resource not found") {
			log.Fatal(err)
		}
	}

	if stateVersion != nil {
		lastStateUpdateDaysAgo = fmt.Sprintf("%f", time.Since(stateVersion.CreatedAt).Hours()/24)
	}

	results.LastRemoteRunDaysAgo = lastRemoteRunDaysAgo
	results.LastStateUpdateDaysAgo = lastStateUpdateDaysAgo

	return results, nil
}

func lockAllWorkspaces(client *tfe.Client, organization string, lockReason *string) ([]WorkspaceLock, error) {
	result := []WorkspaceLock{}
	wg := sync.WaitGroup{}
	var lockedErr error
	var lockedWorkspace WorkspaceLock

	// Get IDs of all workspaces
	allWorkspaces, err := listWorkspaces(client, organization, "")
	check(err)

	ch := make(chan WorkspaceLock, len(allWorkspaces))

	for _, workspace := range allWorkspaces {
		wg.Add(1)

		id := workspace.ID
		name := workspace.Name
		locked := false

		go func(id string, name string) {
			tmpWorkspace, lockedErr := lockWorkspace(client, organization, id, lockReason)
			if lockedErr != nil && strings.Contains(lockedErr.Error(), "workspace already locked") {
				// workspace already locked
				locked = true
			} else if lockedErr != nil {
				log.Fatal(err)
			} else {
				locked = tmpWorkspace.Locked
			}

			lockedWorkspace.ID = id
			lockedWorkspace.Name = name
			lockedWorkspace.Locked = locked

			ch <- lockedWorkspace

			wg.Done()
		}(id, name)
	}

	wg.Wait()
	for i := 0; i < len(allWorkspaces); i++ {
		comm := <-ch
		result = append(result, comm)
	}

	return result, lockedErr

}

func lockWorkspace(client *tfe.Client, organization string, workspaceID string, lockReason *string) (*tfe.Workspace, error) {
	var lockedErr error

	result, err := client.Workspaces.Lock(context.Background(), workspaceID, tfe.WorkspaceLockOptions{
		Reason: lockReason,
	})

	// handle err if workspace already locked
	if err != nil && strings.Contains(err.Error(), "workspace already locked") {
		lockedErr = err
	} else if err != nil {
		log.Fatal(err)
	}

	return result, lockedErr
}

func unlockAllWorkspaces(client *tfe.Client, organization string) ([]WorkspaceLock, error) {
	result := []WorkspaceLock{}
	wg := sync.WaitGroup{}
	var unlockedErr error
	var unlockedWorkspace WorkspaceLock

	// Get IDs of all workspaces
	allWorkspaces, err := listWorkspaces(client, organization, "")
	check(err)

	ch := make(chan WorkspaceLock, len(allWorkspaces))

	for _, workspace := range allWorkspaces {
		wg.Add(1)

		id := workspace.ID
		name := workspace.Name
		locked := true

		go func(id string, name string) {
			log.Debugf("Unlocking workspace: %s", id)
			tmpWorkspace, unlockedErr := unlockWorkspace(client, organization, id)
			if unlockedErr != nil && strings.Contains(unlockedErr.Error(), "workspace already unlocked") {
				// workspace already unlocked
				locked = false
			} else if unlockedErr != nil {
				log.Fatal(err)
			} else {
				locked = tmpWorkspace.Locked
			}

			unlockedWorkspace.ID = id
			unlockedWorkspace.Name = name
			unlockedWorkspace.Locked = locked

			ch <- unlockedWorkspace

			wg.Done()
		}(id, name)
	}

	wg.Wait()
	for i := 0; i < len(allWorkspaces); i++ {
		comm := <-ch
		result = append(result, comm)
	}

	return result, unlockedErr

}

func unlockWorkspace(client *tfe.Client, organization string, workspaceID string) (*tfe.Workspace, error) {
	var unlockedErr error

	result, err := client.Workspaces.Unlock(context.Background(), workspaceID)

	if err != nil && strings.Contains(err.Error(), "workspace already unlocked") {
		unlockedErr = err
	} else if err != nil {
		log.Fatal(err)
	}

	return result, unlockedErr
}
