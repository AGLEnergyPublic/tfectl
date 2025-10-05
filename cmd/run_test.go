//go:build all
// +build all

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
)

func TestRunListCmd(t *testing.T) {
	buf := new(bytes.Buffer)

	wr := rootCmd
	wc := workspaceCmd
	wlc := workspaceListCmd

	wr.AddCommand(wc, wlc)
	wr.SetOut(buf)
	wr.SetErr(buf)
	wr.SetArgs([]string{"workspace", "list", "--query", ".[0].id"})

	err := wr.Execute()
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	out := strings.TrimSpace(buf.String())

	re := regexp.MustCompile(`"([^"]*)"`)
	matches := re.FindStringSubmatch(out)

	workspaceId := "NA"

	if len(matches) >= 2 {
		workspaceId = matches[1]
	}

	if workspaceId != "NA" {
		rr := rootCmd
		rc := runCmd
		rlc := runListCmd

		rr.AddCommand(rc, rlc)

		rr.SetArgs([]string{"run", "list", "--workspace-id", workspaceId})

		rbuf := bytes.NewBufferString("")
		rr.SetOut(rbuf)

		err := rr.Execute()

		if err != nil {
			t.Fatalf("Error executing command: %v", err)
		}
	}

	require.Nil(t, err)
}
