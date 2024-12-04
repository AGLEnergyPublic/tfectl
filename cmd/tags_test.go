//go:build all
// +build all

package cmd

import (
	"testing"
)

func TestTagsListCmd(t *testing.T) {

	tt := []struct {
		args []string
		err  error
	}{
		{
			args: []string{"tag", "list"},
			err:  nil,
		},
	}

	r := rootCmd
	c1 := tagCmd
	c2 := tagListCmd
	r.AddCommand(c1, c2)

	runTestCasesNoOutput(t, r, tt)

	r.RemoveCommand(c1, c2)
	r.AddCommand(c1)
}
