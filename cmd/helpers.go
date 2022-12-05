package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func execute(t *testing.T, c *cobra.Command, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs(args)

	err := c.Execute()

	return strings.TrimSpace(buf.String()), err
}

func runTestCasesNoOutput(t *testing.T, c *cobra.Command, tt []struct {
	args []string
	err  error
}) {
	for _, testCase := range tt {
		_, err := execute(t, c, testCase.args...)
		require.Nil(t, err)
	}
}

func runTestCasesWithOutput(t *testing.T, c *cobra.Command, tt []struct {
	args []string
	err  error
	out  string
}) {
	for _, testCase := range tt {
		out, err := execute(t, c, testCase.args...)
		require.Nil(t, err)

		if testCase.err == nil {
			require.Equal(t, testCase.out, out)
		}
	}
}
