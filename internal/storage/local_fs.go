package storage

import (
	"os"
	"path/filepath"
	"time"

	"github.com/mockctl-hq/mockctl/internal/shared"
)

// LocalFileSystem implements shared.FileSystem using the standard os package.
type LocalFileSystem struct{}

var _ shared.FileSystem = (*LocalFileSystem)(nil)

// NewLocalFileSystem creates a new LocalFileSystem.
func NewLocalFileSystem() *LocalFileSystem {
	return &LocalFileSystem{}
}

// ReadFile reads the named file and returns the contents.
func (fs *LocalFileSystem) ReadFile(name string) ([]byte, error) {
	// #nosec G304 -- LocalFileSystem wraps disk access for application core
	return os.ReadFile(filepath.Clean(name))
}

// WriteFile writes data to a file named by filename.
func (fs *LocalFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	// #nosec G304 -- LocalFileSystem wraps disk access for application core
	return os.WriteFile(filepath.Clean(name), data, perm)
}

// MkdirAll creates a directory named path, along with any necessary parents.
func (fs *LocalFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// ModTime returns the modification time of a file.
func (fs *LocalFileSystem) ModTime(name string) (time.Time, error) {
	info, err := os.Stat(name)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
