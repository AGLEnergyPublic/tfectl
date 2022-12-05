package cmd

import (
	"testing"
)

func TestWorkspaceListCmd(t *testing.T) {

	tt := []struct {
		args []string
		err  error
	}{
		{
			args: []string{"workspace", "list"},
			err:  nil,
		},
	}

	r := rootCmd
	c1 := workspaceCmd
	c2 := workspaceListCmd
	r.AddCommand(c1, c2)

	runTestCasesNoOutput(t, r, tt)
}
