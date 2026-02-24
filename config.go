package jdextractor

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all path and environment configuration
type Config struct {
	// Base paths
	BaseDir      string
	TemplatesDir string
	JobsDir      string

	// Files
	ConfigFile     string
	ResumeTemplate string
	CoverTemplate  string

	// External dependencies
	DeepSeekAPIKey string
	PandocBinary   string
}

// NewConfig initializes paths from environment and filesystem
func NewConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".jobauto")

	cfg := &Config{
		BaseDir:        baseDir,
		TemplatesDir:   filepath.Join(baseDir, "templates"),
		JobsDir:        filepath.Join(baseDir, "jobs"),
		ConfigFile:     filepath.Join(baseDir, "config.yaml"),
		ResumeTemplate: filepath.Join(baseDir, "templates", "resume.md"),
		CoverTemplate:  filepath.Join(baseDir, "templates", "cover.md"),
		DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"),
		PandocBinary:   os.Getenv("PANDOC_PATH"),
	}

	// Default pandoc path if not specified
	if cfg.PandocBinary == "" {
		cfg.PandocBinary = "pandoc"
	}

	return cfg, nil
}

// EnsureDirs creates the directory structure if it doesn't exist
func (c *Config) EnsureDirs() error {
	dirs := []string{c.BaseDir, c.TemplatesDir, c.JobsDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// ValidateAPIKey checks if DeepSeek API key is configured
func (c *Config) ValidateAPIKey() error {
	if c.DeepSeekAPIKey == "" {
		return fmt.Errorf("DEEPSEEK_API_KEY environment variable not set")
	}
	return nil
}
