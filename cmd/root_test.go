package cmd

import (
	"os"
	"testing"
)

func TestExecute_Help(t *testing.T) {
	// Save args and restore later
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"ip-2-cloudresource", "--help"}

	if err := Execute(); err != nil {
		// Cobra prints help to stdout and returns nil; some environments return an exit error.
		t.Fatalf("Execute returned error for --help: %v", err)
	}
}
