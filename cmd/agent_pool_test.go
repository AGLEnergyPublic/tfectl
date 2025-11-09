//go:build all
// +build all

package cmd

import (
	"testing"
)

func TestAgentPoolListCmd(t *testing.T) {

	tt := []struct {
		args []string
		err  error
	}{
		{
			args: []string{"agent-pool", "list"},
			err:  nil,
		},
	}

	r := rootCmd
	c1 := agentPoolCmd
	c2 := agentPoolListCmd
	r.AddCommand(c1, c2)

	runTestCasesNoOutput(t, r, tt)

	r.RemoveCommand(c1, c2)
	r.AddCommand(c1)
}
