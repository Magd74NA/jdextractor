package jdextract

import (
	"fmt"
	"net/http"
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
	Paths  PortablePaths
	Config Config
	Client http.Client
}

func getPortablePaths() (PortablePaths, error) {
	execPath, err := os.Executable()

	if err != nil {
		return PortablePaths{}, fmt.Errorf("error resolving exec: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return PortablePaths{}, fmt.Errorf("error resolving symlink: %w", err)
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

	paths := PortablePaths{
		Root:      root,
		Data:      filepath.Join(root, "data"),
		Jobs:      filepath.Join(root, "data", "jobs"),
		Config:    filepath.Join(root, "config"),
		Templates: filepath.Join(root, "config", "templates"),
	}

	return paths, nil
}

func NewApp() (*App, error) {
	paths, err := getPortablePaths()
	if err != nil {
		return nil, err
	}
	app := &App{
		Paths: paths,
	}

	err = app.Setup()

	if err != nil {
		return nil, err
	}
	return app, nil
}
