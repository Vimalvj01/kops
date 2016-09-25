package vfs

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/util/pkg/hashing"
	"strings"
)

// Yet another VFS package
// If there's a "winning" VFS implementation in go, we should switch to it!

type VFS interface {
}

func IsDirectory(p Path) bool {
	_, err := p.ReadDir()
	return err == nil
}

type Path interface {
	Join(relativePath ...string) Path
	ReadFile() ([]byte, error)

	WriteFile(data []byte) error
	// CreateFile writes the file contents, but only if the file does not already exist
	CreateFile(data []byte) error

	// Remove deletes the file
	Remove() error

	// Base returns the base name (last element)
	Base() string

	// Path returns a string representing the full path
	Path() string

	// ReadDir lists the files in a particular Pathss
	ReadDir() ([]Path, error)

	// ReadTree lists all files in the subtree rooted at the current Path
	ReadTree() ([]Path, error)
}

type HasHash interface {
	// Returns the hash of the file contents, with the preferred hash algorithm
	PreferredHash() (*hashing.Hash, error)

	// Gets the hash, or nil if the hash cannot be (easily) computed
	Hash(algorithm hashing.HashAlgorithm) (*hashing.Hash, error)
}

func RelativePath(base Path, child Path) (string, error) {
	basePath := base.Path()
	childPath := child.Path()
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	if !strings.HasPrefix(childPath, basePath) {
		return "", fmt.Errorf("Path %q is not a child of %q", child, base)
	}

	relativePath := childPath[len(basePath):]
	return relativePath, nil
}

func IsClusterReadable(p Path) bool {
	if hcr, ok := p.(HasClusterReadable); ok {
		return hcr.IsClusterReadable()
	}

	switch p.(type) {
	case *S3Path:
		return true

	case *SSHPath:
		return false

	case *FSPath:
		return false

	default:
		glog.Fatalf("IsClusterReadable not implemented for type %T", p)
		return false
	}
}

type HasClusterReadable interface {
	IsClusterReadable() bool
}
