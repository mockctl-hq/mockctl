package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRootCmd(t *testing.T) {
	// Override output to capture it
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	// Execute command with --help
	rootCmd.SetArgs([]string{"--help"})
	err := ExecuteContext(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := b.String()
	if !strings.Contains(output, "Mock:ctl is an advanced API simulation engine") {
		t.Errorf("expected help output to contain application description, got: %s", output)
	}
}
