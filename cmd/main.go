package main

import (
	"fmt"
	"jdextract/jdextract"
	"os"
)

func main() {
	fmt.Println("Hello World!")
	App, err := jdextract.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errors during setup: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Initial setup complete! app data available at: %s", App.Paths.Root)
}
