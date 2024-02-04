package cmd

import (
	"testing"
)

func TestRootCmd(t *testing.T) {

	helpOutput := `Query TFE and TFC from the command line.

Usage:
  tfectl [command]

Available Commands:
  admin             Manage TFE admin operations
  completion        Generate the autocompletion script for the specified shell
  help              Help about any command
  policy            Query TFE policies
  policy-check      Manage policy check workflows of a TFE run
  policy-set        Query TFE policy sets
  registry-module   Query/Manage TFE private module registry
  registry-provider Manage TFE private provider Registry
  run               Manage TFE runs
  tag               Query TFE tags
  team              Manage TFE teams
  variable          Manage TFE workspace variables
  workspace         Manage TFE workspaces

Flags:
  -h, --help                  help for tfectl
  -l, --log string            log level (debug, info, warn, error, fatal, panic)
  -o, --organization string   terraform organization or set TFE_ORG
  -q, --query string          JQ compatible query to parse JSON output
  -t, --token string          terraform token or set TFE_TOKEN
  -v, --version               version for tfectl

Use "tfectl [command] --help" for more information about a command.`

	tt := []struct {
		args []string
		err  error
		out  string
	}{
		{
			args: nil,
			err:  nil,
			out:  helpOutput,
		},
		{
			args: []string{"-h"},
			err:  nil,
			out:  helpOutput,
		},
		{
			args: []string{"--help"},
			err:  nil,
			out:  helpOutput,
		},
	}

	r := rootCmd

	runTestCasesWithOutput(t, r, tt)
}
