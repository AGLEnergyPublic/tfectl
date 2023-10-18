package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AGLEnergyPublic/tfectl/resources"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var policyCheckCmd = &cobra.Command{
	Use:   "policy-check",
	Short: "Manage policy check workflows of a TFE run",
	Long:  `Manage policy check workflows of a TFE run.`,
}

var policyCheckShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of the policy check in a TFE run",
	Long:  `Show details of the policy check in a TFE run`,
	Run: func(cmd *cobra.Command, args []string) {
		// policy check show function
		_, _, err := resources.Setup(cmd)
		check(err)

		runIds, _ := cmd.Flags().GetString("run-ids")
		query, _ := cmd.Flags().GetString("query")

		var policyCheck interface{}
		var policyCheckShow []interface{}
		var policyCheckShowJson []byte

		// get policy check from runIDs
		runIdShow := strings.Split(runIds, ",")
		for _, id := range runIdShow {
			var tmpPolicyObj []byte
			e := fmt.Sprintf("runs/%s/policy-checks?include=run.workspace", id)
			log.Debug(e)
			httpReq, err := resources.HttpClientSetup(cmd, "GET", e, nil)
			check(err)
			tmpPolicyObj, err = showPolicyCheckHttpClient(httpReq)
			check(err)

			json.Unmarshal(tmpPolicyObj, &policyCheck)
			policyCheckMap := policyCheck.(map[string]interface{})
			policyCheckShow = append(policyCheckShow, policyCheckMap)
		}

		policyCheckShowJson, _ = json.MarshalIndent(policyCheckShow, "", "  ")
		if query != "" {
			resources.JqRun(policyCheckShowJson, query)
		} else {
			fmt.Println(string(policyCheckShowJson))
		}

	},
}

func init() {
	rootCmd.AddCommand(policyCheckCmd)

	// Show sub-command
	// Returns the detailed policy check results for a given list of RunIDs
	policyCheckCmd.AddCommand(policyCheckShowCmd)
	policyCheckShowCmd.Flags().String("run-ids", "", "Comma-separated list of runIds")
}

func showPolicyCheckHttpClient(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	check(err)

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	check(err)

	return b, nil
}
