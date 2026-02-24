package main

import (
	"fmt"
	"os"
	"path/filepath"

	"jdextractor"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	app := jdextractor.NewApp()

	switch os.Args[1] {
	case "setup":
		handleSetup(app)
	case "resume":
		handleResume(app, os.Args[2:])
	// case "pdf"
	case "serve":
		handleServe(app, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("JobAuto — Context-switching automation for job applications")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  jobauto setup              Initialize ~/.jobauto directory structure")
	fmt.Println("  jobauto resume <url>       Generate tailored resume from job URL")
	fmt.Println("  jobauto pdf <path>         Convert markdown to PDF via pandoc")
	fmt.Println("  jobauto serve [port]       Start web UI (default :8080)")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  DEEPSEEK_API_KEY           Required for LLM operations")
	fmt.Println("  PANDOC_PATH                Optional: path to pandoc binary")
}

func handleSetup(app *jdextractor.App) {
	fmt.Println("Initializing JobAuto...")

	if err := app.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created directory structure at %s\n", app.Config.BaseDir)
	fmt.Printf("✓ Generated example templates in %s\n", app.Config.TemplatesDir)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Set DEEPSEEK_API_KEY environment variable")
	fmt.Println("2. Edit ~/.jobauto/templates/resume.md with your base resume")
	fmt.Println("3. Run: jobauto resume <job-url>")
}

func handleResume(app *jdextractor.App, args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: resume command requires a URL\n")
		fmt.Fprintf(os.Stderr, "Usage: jobauto resume <url>\n")
		os.Exit(1)
	}

	url := args[0]
	fmt.Printf("Processing job: %s\n", url)

	folderPath, err := app.ProcessJob(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated resume: %s\n", filepath.Join(folderPath, "resume_custom.md"))
	fmt.Printf("✓ Run folder: %s\n", folderPath)
}

func handleServe(app *jdextractor.App, args []string) {
	port := "8080"
	if len(args) > 0 {
		port = args[0]
	}

	// Validate port is numeric (basic check)
	if len(port) > 0 && port[0] == ':' {
		port = port[1:]
	}

	fmt.Printf("Starting JobAuto server on http://localhost:%s\n", port)
	fmt.Println("Press Ctrl+C to stop")

	if err := app.Serve(port); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
