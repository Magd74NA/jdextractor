package main

import (
	"context"
	"flag"
	"fmt"
	"jdextract/jdextract"
	"os"
	"path/filepath"
)

func main() {

	setup := flag.Bool("setup", false, "This flag is for running the setup for portable mode.")
	local := flag.Bool("local", false, "This flag is for using ")
	flag.Parse()
	app, err, setupComplete := jdextract.NewApp(setup)
	if setupComplete {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup error: %s\n", err)
		os.Exit(1)
	}
	target := flag.Args()[0]
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

	var raw string

	if !*local {
		var err error
		raw, err = jdextract.FetchJobDescription(context.Background(), target, &app.Client, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch error: %s\n", err)
			os.Exit(1)
		}
	} else {
		raw, err = jdextract.FetchJobDescriptionLocal(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading content from file: %s\n", err)
			os.Exit(1)
		}
	}

	dir, err := app.Process(context.Background(), raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "process error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done. Output written to: %s\n", dir)
}
