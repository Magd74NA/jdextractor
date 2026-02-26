package main

import (
	"fmt"
	"jdextract/jdextract"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Hello World!")
	App, err := jdextract.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errors during setup: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Initial setup complete! app data available at: %s\n", App.Paths.Root)

	configPath := filepath.Join(App.Paths.Config, "config.json")
	conf, err := os.Open(configPath)
	if err != nil {
		err = jdextract.CreateEmptyConfig(configPath)
		if err != nil {
			fmt.Printf("Error creating config file: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Please fill out the generated config file as %s\n", configPath)
		os.Exit(0)
	}
	config, err := jdextract.LoadConfig(conf)
	if err != nil {
		fmt.Printf("Config Loading Error: %s\n", err)
	}
	App.Config = *config

	fmt.Printf("Current config!: \nPORT:%d ", App.Config.Port)
}
