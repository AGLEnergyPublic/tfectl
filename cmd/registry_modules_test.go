package cmd

import (
	"testing"
)

func TestRegistryModulesListCmd(t *testing.T) {

	tt := []struct {
		args []string
		err  error
	}{
		{
			args: []string{"registry-module", "list"},
			err:  nil,
		},
	}

	r := rootCmd
	c1 := registryModuleCmd
	c2 := registryModuleListCmd
	r.AddCommand(c1, c2)

	runTestCasesNoOutput(t, r, tt)

	r.RemoveCommand(c1, c2)
	r.AddCommand(c1)
}
