package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// PKS-030 Section 4.4: Global Panic Handler & Crash Dumps
	defer func() {
		if r := recover(); r != nil {
			handlePanic(r)
			os.Exit(1)
		}
	}()

	// PKS-030 Section 4.5: Graceful Shutdown Signal Handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Execute root command
	if err := ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// handlePanic generates a mockctl_crash_dump.json file.
func handlePanic(panicData any) {
	dump := map[string]any{
		"error": fmt.Sprintf("%v", panicData),
		"type":  "mockctl_fatal_crash",
	}

	bytes, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: failed to write crash dump: %v\n", err)
		return
	}

	dumpFile := "mockctl_crash_dump.json"
	if err := os.WriteFile(dumpFile, bytes, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: failed to save %s: %v\n", dumpFile, err)
		return
	}

	fmt.Fprintf(os.Stderr, "Mock:ctl crashed! A crash dump was saved to %s\n", dumpFile)
}
