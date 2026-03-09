package jdextract

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// PortablePaths holds all directory paths resolved relative to the executable.
// On macOS inside a .app bundle, Root is the directory containing the bundle,
// not the bundle itself, so data and config survive app re-installs.
type PortablePaths struct {
	Root      string // directory containing the executable (or .app container on macOS)
	Jobs      string // Root/data/jobs — one subdirectory per processed application
	Data      string // Root/data
	Config    string // Root/config — holds config.json
	Templates string // Root/config/templates — resume.txt and cover.txt
}

// App is the central application object. It is initialised by NewApp and shared
// across all operations. The CLI creates one App per invocation; the web server
// holds a single App for its lifetime.
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

	paths := PortablePaths{
		Root:      root,
		Data:      filepath.Join(root, "data"),
		Jobs:      filepath.Join(root, "data", "jobs"),
		Config:    filepath.Join(root, "config"),
		Templates: filepath.Join(root, "config", "templates"),
	}

	return paths, nil
}

// NewApp initialises an App by resolving portable paths from the executable location.
// If setup is true, it calls Setup() to create the directory structure and example
// templates, then returns (app, nil, true) so callers can exit cleanly after setup.
// On any other path it returns (app, nil, false).
func NewApp(setup *bool) (*App, error, bool) {
	paths, err := getPortablePaths()
	if err != nil {
		return nil, err, false
	}
	app := &App{
		Paths: paths,
	}

	if *setup {
		err = app.Setup()
		if err != nil {
			return nil, err, false
		}
		return app, nil, true
	}
	return app, nil, false
}
