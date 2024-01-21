package cmd

import (
	"testing"
)

func TestPolicySetListCmd(t *testing.T) {

	tt := []struct {
		args []string
		err  error
	}{
		{
			args: []string{"policy-set", "list"},
			err:  nil,
		},
	}

	r := rootCmd
	c1 := policySetCmd
	c2 := policySetListCmd
	r.AddCommand(c1, c2)

	runTestCasesNoOutput(t, r, tt)

	r.RemoveCommand(c1, c2)
	r.AddCommand(c1)
}
