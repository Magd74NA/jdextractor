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
	Jobs      string
	Data      string
	Config    string
	Templates string
}

type App struct {
	paths PortablePaths
}

func CheckWiring() {
	fmt.Println("Hello from app.go")
}

func getPortablePaths() (*PortablePaths, error) {
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
		Jobs:      filepath.Join(root, "data", "jobs"),
		Config:    filepath.Join(root, "config"),
		Templates: filepath.Join(root, "config", "templates"),
	}

	return paths, nil
}

func (A *App) Setup() error {
	// Create directories if they don't exist
	for _, dir := range []string{A.paths.Data, A.paths.Config, A.paths.Jobs, A.paths.Templates} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}
	return nil
}

func NewApp() *App {
	paths, err := getPortablePaths()
	if err != nil {
		fmt.Printf("errors: %s", err)
		os.Exit(1)
	}
	app := &App{
		paths: *paths,
	}

	err = app.Setup()

	if err != nil {
		fmt.Printf("error")
		os.Exit(1)
	}
	return app
}
