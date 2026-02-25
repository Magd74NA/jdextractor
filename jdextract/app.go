package jdextract

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type PortablePaths struct {
	Root      string
	Data      string
	Config    string
	Templates string
}

func CheckWiring() {
	fmt.Println("Hello from app.go")
}

func GetPortablePaths() (*PortablePaths, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("error resolving exec: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("error resolving symlink: %w", err)
	}

	root := filepath.Dir(execPath)

	if runtime.GOOS == "darwin" {
		// Walk up until we find the .app bundle root
		for {
			if strings.HasSuffix(root, ".app") {
				// Found it, now go one more level up to get the container
				root = filepath.Dir(root)
				break
			}
			parent := filepath.Dir(root)
			if parent == root {
				// Hit filesystem root without finding .app
				break
			}
			root = parent
		}
	}

	paths := &PortablePaths{
		Root:      root,
		Data:      filepath.Join(root, "data"),
		Config:    filepath.Join(root, "config"),
		Templates: filepath.Join(root, "templates"),
	}

	// Create directories if they don't exist
	for _, dir := range []string{paths.Data, paths.Config} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}
	return paths, nil
}
