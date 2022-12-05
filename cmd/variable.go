package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"tfectl/resources"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Variable struct {
	ID          string           `json:"id"`
	Key         string           `json:"key"`
	Value       string           `json:"value"`
	Description string           `json:"description"`
	Category    tfe.CategoryType `json:"category"`
	HCL         bool             `json:"hcl"`
	Sensitive   bool             `json:"sensitive"`
}

type Variables struct {
	Variables []Variable `json:"variables"`
}

type WorkspaceLite struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
}

type WorkspaceVar struct {
	WorkspaceID   string   `json:"workspace_id"`
	WorkspaceName string   `json:"workspace_name"`
	Variable      Variable `json:"variable"`
}

type WorkspaceVars struct {
	WorkspaceID   string     `json:"workspace_id"`
	WorkspaceName string     `json:"workspace_name"`
	Variables     []Variable `json:"variables"`
}

// variableCmd represents the variable command.
var variableCmd = &cobra.Command{
	Use:   "variable",
	Short: "Manage TFE workspace variables",
	Long:  `Manage TFE workspace variables.`,
}

var variableListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TFE workspace variables",
	Long:  `List TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//list operations
		organization, client, err := resources.Setup(cmd)
		check(err)

		workspaceIds, _ := cmd.Flags().GetString("workspace-ids")
		workspaceFilter, _ := cmd.Flags().GetString("workspace-filter")

		if workspaceFilter != "" && workspaceIds != "" {
			log.Fatal("workspace-filter and workspace-ids are mutually exclusive, use one or the other!")
		}

		if workspaceFilter == "" && workspaceIds == "" {
			log.Fatal("please provide one of workspace-ids or workspace-filter to perform this operation!")
		}

		query, _ := cmd.Flags().GetString("query")

		var workspaceList []WorkspaceLite
		var tmpWorkspace WorkspaceLite

		if workspaceFilter != "" {
			workspaces, err := listWorkspaces(client, organization, workspaceFilter)
			check(err)

			for _, workspace := range workspaces {
				tmpWorkspace.WorkspaceID = workspace.ID
				tmpWorkspace.WorkspaceName = workspace.Name

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		if workspaceIds != "" {
			workspaceIdList := strings.Split(workspaceIds, ",")
			for _, id := range workspaceIdList {
				workspaceName, err := getWorkspaceNameByID(client, organization, id)
				check(err)
				tmpWorkspace.WorkspaceID = id
				tmpWorkspace.WorkspaceName = workspaceName

				workspaceList = append(workspaceList, tmpWorkspace)
			}
		}

		var workspaceVarsListJson []byte
		var workspaceVarsList []WorkspaceVars

		for _, wrk := range workspaceList {
			w, err := listVariables(client, wrk)
			check(err)
			workspaceVarsList = append(workspaceVarsList, w)
		}

		workspaceVarsListJson, _ = json.MarshalIndent(workspaceVarsList, "", "  ")
		if query != "" {
			resources.JqRun(workspaceVarsListJson, query)
		} else {
			fmt.Println(string(workspaceVarsListJson))
		}
	},
}

var variableReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Read TFE workspace variables",
	Long:  `Read TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//read operations
		organization, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		variableID, _ := cmd.Flags().GetString("variable-id")
		query, _ := cmd.Flags().GetString("query")

		var tmpWorkspace WorkspaceLite

		workspaceName, err := getWorkspaceNameByID(client, organization, workspaceID)
		check(err)
		tmpWorkspace.WorkspaceID = workspaceID
		tmpWorkspace.WorkspaceName = workspaceName

		v, err := readVariable(client, tmpWorkspace, variableID)
		check(err)

		variableJson, _ := json.MarshalIndent(v, "", "  ")
		if query != "" {
			resources.JqRun(variableJson, query)
		} else {
			fmt.Println(string(variableJson))
		}
	},
}

var variableCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create TFE workspace variables",
	Long:  `Create TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//create operations
		_, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		description, _ := cmd.Flags().GetString("description")
		categoryTypeStr, _ := cmd.Flags().GetString("type")
		hcl, _ := cmd.Flags().GetBool("hcl")
		sensitive, _ := cmd.Flags().GetBool("sensitive")
		query, _ := cmd.Flags().GetString("query")

		categoryType := tfe.CategoryType(categoryTypeStr)

		v, err := createVariable(client, workspaceID, &key, &value, &description, &categoryType, &hcl, &sensitive)
		check(err)

		variableJson, _ := json.MarshalIndent(v, "", "  ")
		if query != "" {
			resources.JqRun(variableJson, query)
		} else {
			fmt.Println(string(variableJson))
		}
	},
}

var variableCreateFromFileCmd = &cobra.Command{
	Use:   "from-file",
	Short: "Create variables using JSON file",
	Long:  `Create variables using JSON file`,
	Run: func(cmd *cobra.Command, args []string) {
		_, client, err := resources.Setup(cmd)
		check(err)

		file, _ := cmd.Flags().GetString("file")
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		query, _ := cmd.Flags().GetString("query")

		byteVarJson := readJsonFile(file)

		var variables Variables
		var outputVariablesList []Variable
		var ouputVariablesListJson []byte

		json.Unmarshal([]byte(byteVarJson), &variables)
		for _, newVar := range variables.Variables {
			v, err := createVariable(client, workspaceID, &newVar.Key, &newVar.Value, &newVar.Description, &newVar.Category, &newVar.HCL, &newVar.Sensitive)
			check(err)
			outputVariablesList = append(outputVariablesList, v)
		}
		ouputVariablesListJson, _ = json.MarshalIndent(outputVariablesList, "", "  ")
		if query != "" {
			resources.JqRun(ouputVariablesListJson, query)
		} else {
			fmt.Println(string(ouputVariablesListJson))
		}
	},
}

var variableUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update TFE workspace variables",
	Long:  `Update TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//update operations
		_, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		variableID, _ := cmd.Flags().GetString("variable-id")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		description, _ := cmd.Flags().GetString("description")
		hcl, _ := cmd.Flags().GetBool("hcl")
		sensitive, _ := cmd.Flags().GetBool("sensitive")
		query, _ := cmd.Flags().GetString("query")

		v, err := updateVariable(client, workspaceID, variableID, &key, &value, &description, &hcl, &sensitive)
		check(err)

		variableJson, _ := json.MarshalIndent(v, "", "  ")
		if query != "" {
			resources.JqRun(variableJson, query)
		} else {
			fmt.Println(string(variableJson))
		}
	},
}

var variableUpdateFromFileCmd = &cobra.Command{
	Use:   "from-file",
	Short: "Update variables using JSON file",
	Long:  `Update variables using JSON file`,
	Run: func(cmd *cobra.Command, args []string) {
		_, client, err := resources.Setup(cmd)
		check(err)

		file, _ := cmd.Flags().GetString("file")
		workspaceID, _ := cmd.Flags().GetString("workspace-id")

		query, _ := cmd.Flags().GetString("query")

		byteVarJson := readJsonFile(file)

		var variables Variables
		var outputVariablesList []Variable
		var ouputVariablesListJson []byte

		json.Unmarshal([]byte(byteVarJson), &variables)
		for _, newVar := range variables.Variables {
			v, err := updateVariable(client, workspaceID, newVar.ID, &newVar.Key, &newVar.Value, &newVar.Description, &newVar.HCL, &newVar.Sensitive)
			check(err)
			outputVariablesList = append(outputVariablesList, v)
		}
		ouputVariablesListJson, _ = json.MarshalIndent(outputVariablesList, "", "  ")
		if query != "" {
			resources.JqRun(ouputVariablesListJson, query)
		} else {
			fmt.Println(string(ouputVariablesListJson))
		}
	},
}

var variableDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete TFE workspace variables",
	Long:  `Delete TFE workspace variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		//delete operations
		organization, client, err := resources.Setup(cmd)
		check(err)

		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		variableID, _ := cmd.Flags().GetString("variable-id")
		query, _ := cmd.Flags().GetString("query")

		err = deleteVariable(client, workspaceID, variableID)
		check(err)

		var workspaceVarsListJson []byte
		var workspaceVarsList []WorkspaceVars
		var tmpWorkspace WorkspaceLite

		workspaceName, err := getWorkspaceNameByID(client, organization, workspaceID)
		check(err)
		tmpWorkspace.WorkspaceID = workspaceID
		tmpWorkspace.WorkspaceName = workspaceName

		w, err := listVariables(client, tmpWorkspace)
		check(err)

		workspaceVarsList = append(workspaceVarsList, w)
		workspaceVarsListJson, _ = json.MarshalIndent(workspaceVarsList, "", "  ")
		if query != "" {
			resources.JqRun(workspaceVarsListJson, query)
		} else {
			fmt.Println(string(workspaceVarsListJson))
		}
	},
}

func init() {
	rootCmd.AddCommand(variableCmd)

	// List sub-command
	variableCmd.AddCommand(variableListCmd)
	variableListCmd.Flags().String("workspace-ids", "", "Comma separated list of workspaceIDs")
	variableListCmd.Flags().String("workspace-filter", "", "Search filter for workspace")

	// Read sub-command
	variableCmd.AddCommand(variableReadCmd)
	variableReadCmd.Flags().String("workspace-id", "", "workspaceID")
	variableReadCmd.Flags().String("variable-id", "", "variableID of the variable")

	// Create sub-command
	variableCmd.AddCommand(variableCreateCmd)
	variableCreateCmd.Flags().String("workspace-id", "", "workspaceID")
	variableCreateCmd.Flags().String("key", "", "Variable Name")
	variableCreateCmd.Flags().String("value", "", "Variable Value")
	variableCreateCmd.Flags().Bool("sensitive", false, "Set sensitive flag for variable")
	variableCreateCmd.Flags().Bool("hcl", false, "Set if variable has HCL syntax")
	variableCreateCmd.Flags().String("type", "env", "Variable type")
	variableCreateCmd.Flags().String("description", "Variable Created by tfectl", "Description for the variable")
	// Create from file sub-command
	variableCreateCmd.AddCommand(variableCreateFromFileCmd)
	variableCreateFromFileCmd.Flags().String("file", "", "File containing workspace variables")
	variableCreateFromFileCmd.Flags().String("workspace-id", "", "workspaceID")

	// Update sub-command
	variableCmd.AddCommand(variableUpdateCmd)
	variableUpdateCmd.Flags().String("workspace-id", "", "workspaceID")
	variableUpdateCmd.Flags().String("variable-id", "", "variableID")
	variableUpdateCmd.Flags().String("key", "", "Variable Name")
	variableUpdateCmd.Flags().String("value", "", "Variable Value")
	variableUpdateCmd.Flags().Bool("sensitive", false, "Set sensitive flag for variable")
	variableUpdateCmd.Flags().Bool("hcl", false, "Set if variable has HCL syntax")
	variableUpdateCmd.Flags().String("description", "Variable Updated by tfectl", "Description for the variable")
	// Update from file sub-command
	variableUpdateCmd.AddCommand(variableUpdateFromFileCmd)
	variableUpdateFromFileCmd.Flags().String("file", "", "File containing workspace variables")
	variableUpdateFromFileCmd.Flags().String("workspace-id", "", "workspaceID")

	// Delete sub-command
	variableCmd.AddCommand(variableDeleteCmd)
	variableDeleteCmd.Flags().String("workspace-id", "", "workspaceID")
	variableDeleteCmd.Flags().String("variable-id", "", "variableID of the variable")

}

func listVariables(client *tfe.Client, workspace WorkspaceLite) (WorkspaceVars, error) {
	result := WorkspaceVars{}
	currentPage := 1

	for {
		log.Debugf("Processing page %d.\n", currentPage)
		options := &tfe.VariableListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: currentPage,
				PageSize:   50,
			},
		}

		varList, err := client.Variables.List(context.Background(), workspace.WorkspaceID, options)
		check(err)

		var tmpVarList []Variable
		for _, v := range varList.Items {
			var tmpVar = Variable{
				ID:          v.ID,
				Key:         v.Key,
				Value:       v.Value,
				Description: v.Description,
				Category:    v.Category,
				HCL:         v.HCL,
				Sensitive:   v.Sensitive,
			}

			tmpVarList = append(tmpVarList, tmpVar)
		}

		result = WorkspaceVars{
			WorkspaceID:   workspace.WorkspaceID,
			WorkspaceName: workspace.WorkspaceName,
			Variables:     tmpVarList,
		}

		if varList.Pagination.NextPage == 0 {
			break
		}

		currentPage++
	}

	return result, nil
}

func readVariable(client *tfe.Client, workspace WorkspaceLite, variableID string) (WorkspaceVar, error) {
	result := WorkspaceVar{}

	v, err := client.Variables.Read(context.Background(), workspace.WorkspaceID, variableID)
	check(err)

	result.WorkspaceID = workspace.WorkspaceID
	result.WorkspaceName = workspace.WorkspaceName
	result.Variable.ID = v.ID
	result.Variable.Key = v.Key
	result.Variable.Value = v.Value
	result.Variable.Description = v.Description
	result.Variable.Category = v.Category
	result.Variable.HCL = v.HCL
	result.Variable.Sensitive = v.Sensitive

	return result, nil
}

func createVariable(client *tfe.Client, workspaceID string, key *string, value *string, description *string, category *tfe.CategoryType, hcl *bool, sensitive *bool) (Variable, error) {
	result := Variable{}

	options := tfe.VariableCreateOptions{
		Key:         key,
		Value:       value,
		Description: description,
		Category:    category,
		HCL:         hcl,
		Sensitive:   sensitive,
	}

	v, err := client.Variables.Create(context.Background(), workspaceID, options)
	check(err)

	result = Variable{
		ID:          v.ID,
		Key:         v.Key,
		Value:       v.Value,
		Description: v.Description,
		Category:    v.Category,
		HCL:         v.HCL,
		Sensitive:   v.Sensitive,
	}

	return result, nil
}

func updateVariable(client *tfe.Client, workspaceID string, variableID string, key *string, value *string, description *string, hcl *bool, sensitive *bool) (Variable, error) {
	result := Variable{}

	options := tfe.VariableUpdateOptions{
		Key:         key,
		Value:       value,
		Description: description,
		HCL:         hcl,
		Sensitive:   sensitive,
	}

	v, err := client.Variables.Update(context.Background(), workspaceID, variableID, options)
	check(err)

	result = Variable{
		ID:          v.ID,
		Key:         v.Key,
		Value:       v.Value,
		Description: v.Description,
		Category:    v.Category,
		HCL:         v.HCL,
		Sensitive:   v.Sensitive,
	}

	return result, nil
}

func deleteVariable(client *tfe.Client, workspaceID string, variableID string) error {

	err := client.Variables.Delete(context.Background(), workspaceID, variableID)

	return err
}

func readJsonFile(file string) []byte {
	jsonFile, err := os.Open(file)
	check(err)

	defer jsonFile.Close()

	byteJson, err := ioutil.ReadAll(jsonFile)
	check(err)

	return []byte(byteJson)
}
