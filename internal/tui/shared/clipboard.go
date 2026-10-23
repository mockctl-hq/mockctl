package shared

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

const maxClipboardBytes = 1024 * 1024 // 1MB Limit

// WriteToClipboard normalizes newlines to OS standards and writes to clipboard.
func WriteToClipboard(text string) error {
	if len(text) > maxClipboardBytes {
		return fmt.Errorf("payload too large to copy (max 1MB)")
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Emit OSC 52 sequence for terminal-native clipboard (works natively in Termux, Alacritty, iTerm2, tmux, etc.)
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	_, _ = fmt.Fprintf(os.Stderr, "\x1b]52;c;%s\x07", encoded)

	// Also attempt system clipboard with a safe timeout to prevent blocking on headless/unresponsive environments
	done := make(chan error, 1)
	go func() {
		done <- clipboard.WriteAll(text)
	}()

	select {
	case err := <-done:
		if err != nil {
			// Since OSC 52 was already emitted, system tool error is not fatal
			return nil
		}
		return nil
	case <-time.After(300 * time.Millisecond):
		// Timed out waiting for external process, OSC 52 was already sent
		return nil
	}
}

// ReadFromClipboard reads from OS clipboard.
func ReadFromClipboard() (string, error) {
	type readResult struct {
		text string
		err  error
	}
	done := make(chan readResult, 1)
	go func() {
		text, err := clipboard.ReadAll()
		done <- readResult{text: text, err: err}
	}()

	select {
	case res := <-done:
		if res.err != nil {
			return "", res.err
		}
		if len(res.text) > maxClipboardBytes {
			return "", fmt.Errorf("clipboard payload too large (max 1MB)")
		}
		return res.text, nil
	case <-time.After(500 * time.Millisecond):
		return "", fmt.Errorf("clipboard read timed out")
	}
}
