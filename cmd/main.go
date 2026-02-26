package main

import (
	"context"
	"fmt"
	"jdextract/jdextract"
	"os"
	"path/filepath"
)

func main() {
	App, err := jdextract.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup error: %s\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(App.Paths.Config, "config.json")
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
	App.Config = *config

	if App.Config.DeepSeekApiKey == "" || App.Config.DeepSeekApiKey == "example_key" {
		fmt.Fprintf(os.Stderr, "error: set deepseek_api_key in %s\n", configPath)
		os.Exit(1)
	}

	client, err := jdextract.InitiateClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "http client error: %s\n", err)
		os.Exit(1)
	}
	App.Client = *client

	// --- hardcoded test: parse test_jd.md and call GenerateAll ---

	raw, err := os.ReadFile("test_jd.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read test_jd.md: %s\n", err)
		os.Exit(1)
	}

	resumeBytes, err := os.ReadFile(filepath.Join(App.Paths.Templates, "resume.txt"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read resume template: %s\n", err)
		os.Exit(1)
	}

	var baseCover *string
	if coverBytes, err := os.ReadFile(filepath.Join(App.Paths.Templates, "cover.txt")); err == nil {
		s := string(coverBytes)
		baseCover = &s
	}

	nodes := jdextract.Parse(string(raw))
	fmt.Printf("Parsed %d nodes from test_jd.md\n", len(nodes))

	company, role, resume, cover, score, tokens, err := jdextract.GenerateAll(
		context.Background(),
		App.Config.DeepSeekApiKey,
		App.Config.DeepSeekModel,
		&App.Client,
		nodes,
		string(resumeBytes),
		baseCover,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GenerateAll error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n--- RESULT ---\n")
	fmt.Printf("Company:     %s\n", company)
	fmt.Printf("Role:        %s\n", role)
	fmt.Printf("Match score: %d/10\n", score)
	fmt.Printf("Tokens used: %d\n", tokens)
	fmt.Printf("\n--- RESUME ---\n%s\n", resume)
	if cover != nil {
		fmt.Printf("\n--- COVER LETTER ---\n%s\n", *cover)
	}
}
