//go:build all
// +build all

package cmd

import (
	"testing"
)

func TestPolicyListCmd(t *testing.T) {

	tt := []struct {
		args []string
		err  error
	}{
		{
			args: []string{"policy", "list"},
			err:  nil,
		},
	}

	r := rootCmd
	c1 := policyCmd
	c2 := policyListCmd
	r.AddCommand(c1, c2)

	runTestCasesNoOutput(t, r, tt)

	r.RemoveCommand(c1, c2)
	r.AddCommand(c1)
}
