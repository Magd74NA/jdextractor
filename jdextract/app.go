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
	Contacts  string // Root/data/contacts — one subdirectory per networking contact
}

// App is the central application object. It is initialised by NewApp and shared
// across all operations. The CLI creates one App per invocation; the web server
// holds a single App for its lifetime.
type App struct {
	Paths                  PortablePaths
	Config                 Config
	PromptConfig           PromptConfig
	NetworkingPromptConfig NetworkingPromptConfig
	Client                 http.Client
	Jobs                   Store[ApplicationMeta]
	Contacts               Store[ContactMeta]
}

// LLMBackend holds the resolved invoker functions and credentials for the
// configured LLM backend.
type LLMBackend struct {
	Invoker       LLMInvoker
	StreamInvoker StreamingLLMInvoker
	APIKey        string
	Model         string
}

// Backend returns the LLM invoker functions and credentials for the currently
// configured backend (deepseek or kimi).
func (a *App) Backend() LLMBackend {
	if a.Config.Backend == "kimi" {
		return LLMBackend{
			Invoker:       InvokeKimiApi,
			StreamInvoker: InvokeKimiApiStream,
			APIKey:        a.Config.KimiApiKey,
			Model:         a.Config.KimiModel,
		}
	}
	return LLMBackend{
		Invoker:       InvokeDeepseekApi,
		StreamInvoker: InvokeDeepseekApiStream,
		APIKey:        a.Config.DeepSeekApiKey,
		Model:         a.Config.DeepSeekModel,
	}
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
		Contacts:  filepath.Join(root, "data", "contacts"),
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
		Jobs: Store[ApplicationMeta]{
			BasePath: paths.Jobs,
			SetDir:   func(m *ApplicationMeta, d string) { m.Dir = d },
		},
		Contacts: Store[ContactMeta]{
			BasePath: paths.Contacts,
			SetDir:   func(m *ContactMeta, d string) { m.Dir = d },
			PostRead: func(m *ContactMeta) {
				if m.Conversations == nil {
					m.Conversations = []Conversation{}
				}
			},
		},
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
