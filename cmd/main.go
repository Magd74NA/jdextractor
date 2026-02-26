package main

import (
	"context"
	"fmt"
	"jdextract/jdextract"
	"os"
	"path/filepath"
)

func main() {
	app, err := jdextract.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup error: %s\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(app.Paths.Config, "config.json")
	conf, err := os.Open(configPath)
	if err != nil {
		err = jdextract.CreateEmptyConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating config: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created config at %s — fill in your API key and re-run.\n", configPath)
		os.Exit(0)
	}
	config, err := jdextract.LoadConfig(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load error: %s\n", err)
		os.Exit(1)
	}
	app.Config = *config

	if app.Config.DeepSeekApiKey == "" || app.Config.DeepSeekApiKey == "example_key" {
		fmt.Fprintf(os.Stderr, "error: set deepseek_api_key in %s\n", configPath)
		os.Exit(1)
	}

	client, err := jdextract.InitiateClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "http client error: %s\n", err)
		os.Exit(1)
	}
	app.Client = *client

	raw, err := os.ReadFile(filepath.Join(app.Paths.Root, "test_jd.md"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read test_jd.md: %s\n", err)
		os.Exit(1)
	}

	dir, err := app.Process(context.Background(), string(raw))
	if err != nil {
		fmt.Fprintf(os.Stderr, "process error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done. Output written to: %s\n", dir)
}
