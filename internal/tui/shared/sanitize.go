package shared

import (
	"github.com/charmbracelet/x/ansi"
)

// SanitizeANSI strips all ANSI escape codes, OSC sequences, and control characters
// from a raw string. This is CRITICAL for preventing Terminal Injection (ANSI spoofing)
// and OS Clipboard Hijacking (OSC 52) when rendering untrusted telemetry payloads.
func SanitizeANSI(raw string) string {
	// The charmbracelet/x/ansi package provides an optimized, zero-allocation stripper
	return ansi.Strip(raw)
}
