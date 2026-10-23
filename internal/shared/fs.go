package shared

import (
	"os"
	"time"
)

// FileSystem provides an abstraction over OS-level file operations.
// This prevents components from bypassing the architecture to hit the disk directly,
// ensuring PKS-024 compliance.
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	ModTime(name string) (time.Time, error)
}
