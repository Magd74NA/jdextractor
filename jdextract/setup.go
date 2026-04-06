package jdextract

import (
	"fmt"
	"os"
	"path/filepath"
)

func (a *App) createExampleTemplates() error {
	templates := []struct {
		name    string
		content string
	}{
		{"resume.txt", defaultResumeTemplate},
		{"cover.txt", defaultCoverTemplate},
	}

	for _, t := range templates {
		path := filepath.Join(a.Paths.Templates, t.name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(t.content), 0644); err != nil {
				return fmt.Errorf("cannot write %s: %w", t.name, err)
			}
		}
	}

	return nil
}

// Setup creates the portable directory structure (data/, config/, data/jobs/,
// config/templates/) and writes example resume.txt and cover.txt templates if
// they do not already exist. It is safe to call Setup on an existing installation;
// it will not overwrite files the user has already customised.
func (a *App) Setup() error {
	for _, dir := range []string{a.Paths.Data, a.Paths.Config, a.Paths.Jobs, a.Paths.Templates, a.Paths.Contacts} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}

	if err := a.createExampleTemplates(); err != nil {
		return fmt.Errorf("cannot create example templates: %w", err)
	}

	return nil
}
