package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"jdextract/jdextract"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

var version = "dev"

const usage = `jdextract — tailor resumes to job descriptions

Usage:
  jdextract setup
  jdextract generate <url>
  jdextract generate --local <file>
  jdextract generate --batch <url> [<url>...]
  jdextract generate          (reads from stdin)
  jdextract list
  jdextract status <prefix> <status>
  jdextract serve [--port <port>] [--open]

Subcommands:
  setup     Initialize portable directory structure and example templates.
  generate  Generate tailored resume and cover letter from a job description.
            Pass a URL (fetched via jina.ai), a local file path (--local),
            multiple URLs for concurrent batch processing (--batch),
            or pipe raw text via stdin.
  list      Print a table of processed job applications.
  status    Update the status of a job by directory prefix.
            Valid statuses: draft, applied, interviewing, offer, rejected
  serve     Start the web UI. Defaults to port 8080; --open launches a browser.
`

func main() {
	if len(os.Args) < 2 {
		// Double-clicked or launched outside a terminal: open a terminal window
		// first so the user can see server output, then auto-open the browser.
		if !isTerminal() {
			if relaunchInTerminal() {
				return
			}
		}
		cmdServe([]string{"--open"})
		return
	}

	switch os.Args[1] {
	case "setup":
		cmdSetup()
	case "generate":
		cmdGenerate(os.Args[2:])
	case "list":
		cmdList()
	case "status":
		cmdStatus(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	case "version", "--version", "-version":
		fmt.Println(version)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}

// initApp initializes the App without loading config (for commands that don't need the API).
func initApp() *jdextract.App {
	app, err, _ := jdextract.NewApp(boolPtr(false))
	if err != nil {
		fmt.Fprintf(os.Stderr, "init error: %s\n", err)
		os.Exit(1)
	}
	return app
}

// initAppWithConfig loads config and HTTP client, required for API-calling commands.
func initAppWithConfig() *jdextract.App {
	app := initApp()
	configPath := filepath.Join(app.Paths.Config, "config.json")
	promptConfigPath := filepath.Join(app.Paths.Config, "prompt.json")
	conf, err := os.Open(configPath)
	if err != nil {
		err = jdextract.CreateEmptyConfig(configPath)
		err = jdextract.CreateEmptyPromptConfig(promptConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating config: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created config at %s — fill in your API key and re-run.\n", configPath)
		os.Exit(0)
	}
	config, err := jdextract.LoadConfig(conf)
	conf.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load error: %s\n", err)
		os.Exit(1)
	}
	app.Config = *config

	if pconf, err := os.Open(promptConfigPath); err == nil {
		promptConfig, err := jdextract.LoadPromptConfig(pconf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "prompt config load error: %s\n", err)
			os.Exit(1)
		}
		app.PromptConfig = *promptConfig
	}

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
	return app
}

func cmdSetup() {
	_, _, setupComplete := jdextract.NewApp(boolPtr(true))
	if setupComplete {
		fmt.Println("Setup complete.")
		os.Exit(0)
	}
}

func cmdGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	local := fs.Bool("local", false, "Read job description from a local file instead of fetching via URL.")
	batch := fs.Bool("batch", false, "Process multiple URLs concurrently (pass URLs as arguments).")
	fs.Parse(args)

	app := initAppWithConfig()

	if *batch {
		if *local {
			fmt.Fprintln(os.Stderr, "error: --batch and --local are mutually exclusive")
			os.Exit(1)
		}
		urls := fs.Args()
		if len(urls) == 0 {
			fmt.Fprintln(os.Stderr, "error: --batch requires at least one URL argument")
			os.Exit(1)
		}
		ctx := context.Background()
		total := len(urls)
		done := 0
		failed := 0
		for r := range app.ProcessBatch(ctx, urls) {
			done++
			if r.Err != nil {
				fmt.Fprintf(os.Stderr, "[%d/%d] error %s: %s\n", done, total, r.URL, r.Err)
				failed++
			} else {
				fmt.Printf("[%d/%d] done: %s\n", done, total, r.Dir)
			}
		}
		if failed > 0 {
			fmt.Fprintf(os.Stderr, "%d/%d failed\n", failed, total)
			os.Exit(1)
		}
		return
	}

	var raw string
	var err error

	progress := func(e jdextract.ProgressEvent) {
		if e.Message != "" {
			fmt.Fprintf(os.Stderr, "%s\n", e.Message)
		}
	}

	if *local {
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "error: --local requires a file path argument")
			os.Exit(1)
		}
		raw, err = jdextract.FetchJobDescriptionLocal(fs.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
			os.Exit(1)
		}
	} else if fs.NArg() >= 1 {
		fmt.Fprintf(os.Stderr, "Fetching job description\u2026\n")
		raw, err = jdextract.FetchJobDescription(context.Background(), fs.Arg(0), &app.Client, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch error: %s\n", err)
			os.Exit(1)
		}
	} else {
		// stdin
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeCharDevice != 0 {
			fmt.Fprintln(os.Stderr, "error: no input provided — pass a URL, use --local <file>, or pipe text via stdin")
			os.Exit(1)
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %s\n", err)
			os.Exit(1)
		}
		raw = string(data)
	}

	dir, err := app.ProcessWithProgress(context.Background(), raw, progress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "process error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Done. Output written to: %s\n", dir)
}

func cmdList() {
	app := initApp()
	jobs, err := jdextract.ListJobs(app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list error: %s\n", err)
		os.Exit(1)
	}
	if len(jobs) == 0 {
		fmt.Println("No jobs found.")
		return
	}
	fmt.Print(jdextract.FormatJobs(jobs))
}

func cmdStatus(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: jdextract status <prefix> <status>")
		os.Exit(1)
	}
	app := initApp()
	if err := jdextract.UpdateJobStatus(app, args[0], args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "status error: %s\n", err)
		os.Exit(1)
	}
}

// initAppForServe loads paths and config (if it exists) without failing on a
// missing or unconfigured API key. The HTTP client is initialised best-effort.
// Serve() itself calls Setup() to create dirs and an empty config if needed.
func initAppForServe() *jdextract.App {
	app := initApp()
	configPath := filepath.Join(app.Paths.Config, "config.json")
	promptConfigPath := filepath.Join(app.Paths.Config, "prompt.json")
	if conf, err := os.Open(configPath); err == nil {
		config, err := jdextract.LoadConfig(conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config load error: %s\n", err)
			os.Exit(1)
		}
		app.Config = *config
	}
	if pconf, err := os.Open(promptConfigPath); err == nil {
		promptConfig, err := jdextract.LoadPromptConfig(pconf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "prompt config load error: %s\n", err)
			os.Exit(1)
		}
		app.PromptConfig = *promptConfig
	}
	if client, err := jdextract.InitiateClient(); err == nil {
		app.Client = *client
	}
	return app
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port to listen on.")
	open := fs.Bool("open", false, "Open browser after startup.")
	fs.Parse(args)

	app := initAppForServe()
	// Prefer config port unless --port was explicitly passed on the command line.
	portExplicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "port" {
			portExplicit = true
		}
	})
	if !portExplicit && app.Config.Port != 0 {
		*port = app.Config.Port
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *open {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(fmt.Sprintf("http://localhost:%d", *port))
		}()
	}

	if err := app.Serve(ctx, fmt.Sprintf("%d", *port)); err != nil {
		fmt.Fprintf(os.Stderr, "serve error: %s\n", err)
		os.Exit(1)
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func relaunchInTerminal() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	switch runtime.GOOS {
	case "darwin":
		// Open a new Terminal.app window running this binary.
		script := `tell app "Terminal" to do script "` + exe + `"`
		return exec.Command("osascript", "-e", script).Start() == nil
	case "windows":
		// Windows .exe already opens a console window — nothing to do.
		return false
	default: // linux and other Unix
		terminals := [][]string{
			{"xterm", "-e"},
			{"konsole", "-e"},
			{"gnome-terminal", "--"},
			{"xfce4-terminal", "-e"},
			{"alacritty", "-e"},
			{"kitty"},
			{"foot"},
		}
		for _, t := range terminals {
			path, err := exec.LookPath(t[0])
			if err != nil {
				continue
			}
			args := append(t[1:], exe)
			cmd := exec.Command(path, args...)
			if err := cmd.Start(); err == nil {
				return true
			}
		}
		return false
	}
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}

func boolPtr(b bool) *bool { return &b }
