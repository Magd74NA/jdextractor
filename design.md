---

# Design Document: JobAuto

## 1. Project Summary
**JobAuto** is a local-first CLI and embedded web tool designed to eliminate the cognitive fatigue of tailoring job applications. It does not automate *applying*; it automates *context-switching*. 

By providing a target Job URL and a base Markdown resume, the tool uses an LLM to align the resume's experience bullets to the job's requirements, outputting a customized Markdown file for human review, and finally compiling it to a professional PDF via Pandoc. The generated files serve as a natural "CRM" for the user's job hunt.

## 2. Core Constraints (To Prevent Scope Creep)
*   **Zero Databases:** The filesystem is the database. Every job run creates a structured folder.
*   **Zero JS Build Steps:** The web UI is a single embedded HTML file using CDN-hosted Alpine.js and Tailwind.
*   **One API Dependency:** DeepSeek for all text analysis and generation.
*   **One System Dependency:** Pandoc (must be installed by the user) for Markdown → PDF conversion.
*   **Human-in-the-Loop:** The AI generates Markdown. The human edits the Markdown. The tool turns the Markdown into a PDF.

## 3. Directory Structure (Flat & Idiomatic Go)
Keep the Go code in a single reusable `jobauto` package, with a thin `cmd` wrapper.

```text
jobauto/
├── go.mod
├── cmd/
│   └── jobauto/
│       └── main.go           # CLI argument parsing; calls jobauto.App methods
├── jobauto/                  # Core package
│   ├── app.go                # Central App struct (holds config, coordinates flow)
│   ├── config.go             # Resolves ~/.jobauto paths, reads env vars
│   ├── fetch.go              # Simple HTTP GET to pull raw job HTML
│   ├── llm.go                # DeepSeek HTTP client and system prompts
│   ├── generate.go           # Orchestration: fetch -> LLM -> save to disk
│   ├── pandoc.go             # os/exec wrapper for Pandoc
│   └── web.go                # net/http server wrapping App methods
└── static/
    └── index.html            # Embedded UI (HTML/Alpine/Tailwind)
```

## 4. The Data Model (Filesystem as CRM)
Every time a user runs the tool against a job, it generates a "Run Folder" in their output directory. This creates a natural, searchable history of all job applications.

**Base Directory:** `~/.jobauto/`
```text
~/.jobauto/
├── config.yaml               # Optional: Just stores API key
├── templates/
│   ├── resume.md             # The user's master resume
│   └── cover.md              # The user's base cover letter
└── jobs/
    └── 2026-02-24_acme-corp_copywriter/    <-- The "Run Folder"
        ├── job_raw.txt                     # The scraped webpage
        ├── resume_custom.md                # AI-tailored resume (user edits this)
        ├── resume_final.pdf                # Generated via pandoc
        └── cover_letter.md                 # AI-drafted cover letter
```

## 5. System Components (The `jobauto` package)

### `App` (app.go)
The central orchestrator. It holds the configuration and provides the high-level methods used by both the CLI and Web interfaces.
*   `func NewApp() *App`
*   `func (a *App) Setup() error` (Creates directories and example templates)

### `Config` (config.go)
*   Finds the user's home directory.
*   Loads `DEEPSEEK_API_KEY` from the environment.

### `LLM Client` (llm.go)
A simple `net/http` wrapper specifically for `api.deepseek.com`.
*   `func (l *LLM) CustomizeResume(jobText, baseResume string) (string, error)`
    *   *Prompt:* "You are a resume editor. Modify the bullets under 'Experience' in the provided resume to highlight overlaps with the provided job description. Output only Markdown. Do not change the document structure."

### `Generator` (generate.go)
The heavy lifter. 
*   `func (a *App) ProcessJob(url string) (folderPath string, err error)`
    1. Fetches URL HTML.
    2. Extracts Company Name & Role from HTML (via a quick LLM call) to name the folder.
    3. Creates `~/.jobauto/jobs/YYYY-MM-DD_Company_Role/`.
    4. Calls LLM to customize the base `resume.md`.
    5. Writes `resume_custom.md` into the folder.

### `Pandoc Wrapper` (pandoc.go)
*   `func (a *App) MarkdownToPDF(mdPath string) (pdfPath string, err error)`
    *   Executes: `pandoc <mdPath> -o <pdfPath> --pdf-engine=xelatex -V geometry:margin=1in`

### `Web Server` (web.go)
*   `//go:embed static/*`
*   `func (a *App) Serve(port string)`
*   Exposes endpoints `/api/process` (calls `ProcessJob`) and `/api/pdf` (calls `MarkdownToPDF`).

## 6. User Interface (CLI & Web)

**CLI Flow (for power users / scripting):**
```bash
$ jobauto setup
$ jobauto resume https://acme.com/job/123
  > Saved to: ~/.jobauto/jobs/2026-02-24_acme_copywriter/resume_custom.md
$ jobauto pdf ~/.jobauto/jobs/2026-02-24_acme_copywriter/resume_custom.md
  > Generated: resume_final.pdf
```

**Web Flow (for the copywriter friend):**
```bash
$ jobauto serve
  > Starting UI at http://localhost:8080
```
*   User opens browser.
*   Pastes Job URL into an input field. Clicks "Generate".
*   UI hits `/api/process`.
*   UI displays the generated Markdown in a read-only text box and shows the file path.
*   User clicks "Convert to PDF" -> UI hits `/api/pdf`.

## 7. Execution Plan (Weekend Roadmap)

**Phase 1: The Engine (Saturday)**
1. Write `config.go` to establish the `~/.jobauto` folder structure.
2. Write `fetch.go` to do a dumb HTTP GET of a URL.
3. Write `llm.go` to send hardcoded strings to DeepSeek and print the response.
4. Combine them in `generate.go` to fetch a real URL, send it to the LLM alongside a dummy `resume.md`, and write the output to disk.

**Phase 2: CLI & Polish (Sunday)**
1. Wire up `cmd/jobauto/main.go` using standard `os.Args` or `flag` (no need for heavy CLI frameworks like Cobra for 4 commands).
2. Write `pandoc.go` and test generating a PDF from the CLI.
3. Add basic error handling ("API key missing", "Pandoc not installed").

**Phase 3: Visual Mode (Next Weekend)**
1. Create `static/index.html` with Alpine.js.
2. Add `web.go` with standard `net/http` handlers.
3. Wire the web buttons to call the Go functions you already wrote in Phase 1 & 2.

This design gives you a robust, highly functional tool with practically zero boilerplate.
