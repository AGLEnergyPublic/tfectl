package cmd

import (
	"testing"
)

func TestTeamListCmd(t *testing.T) {

	tt := []struct {
		args []string
		err  error
	}{
		{
			args: []string{"team", "list"},
			err:  nil,
		},
	}

	r := rootCmd
	c1 := teamCmd
	c2 := teamListCmd
	r.AddCommand(c1, c2)

	runTestCasesNoOutput(t, r, tt)
}
