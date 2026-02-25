---

# Design Document: jdextract

## 1. Project Summary
**jdextract** is a local-first CLI and embedded web tool designed to eliminate the cognitive fatigue of tailoring job applications. It does not automate *applying*; it automates *context-switching*. 

By providing a target Job URL and a base resume text file, the tool uses an LLM to align the resume's experience bullets to the job's requirements, outputting a customized plain text file for human review. The generated files serve as a natural "CRM" for the user's job hunt.

## 2. Core Constraints (To Prevent Scope Creep)
*   **Zero Databases:** The filesystem is the database. Every job run creates a structured folder.
*   **Zero JS Build Steps:** The web UI is a single embedded HTML file using CDN-hosted Alpine.js and Tailwind.
*   **One API Dependency:** DeepSeek for all text analysis and generation.
*   **Zero System Dependencies:** No external tools required (no Pandoc for MVP1).
*   **Plain Text Output:** All generated files are `.txt` for simplicity.
*   **Human-in-the-Loop:** The AI generates text. The human reviews and edits the text.

## 3. Directory Structure (Flat & Idiomatic Go)
Keep the Go code in a single reusable `jdextract` package, with a thin `cmd` wrapper.

```text
jdextract/
├── go.mod
├── cmd/
│   └── jdextract/
│       └── main.go           # CLI argument parsing; calls jdextract.App methods
├── jdextract/                # Core package
│   ├── app.go                # Central App struct (holds config, coordinates flow)
│   ├── config.go             # Resolves ~/.jdextract paths, reads env vars
│   ├── fetch.go              # Simple HTTP GET to pull raw job HTML
│   ├── llm.go                # DeepSeek HTTP client and system prompts
│   ├── generate.go           # Orchestration: fetch -> LLM -> save to disk
│   └── web.go                # net/http server wrapping App methods
└── static/
    └── index.html            # Embedded UI (HTML/Alpine/Tailwind)
```

## 4. The Data Model (Filesystem as CRM)
Every time a user runs the tool against a job, it generates a "Run Folder" in their output directory. This creates a natural, searchable history of all job applications.

**Base Directory:** `~/.jdextract/`
```text
~/.jdextract/
├── config.yaml               # Optional: Just stores API key
├── templates/
│   ├── resume.txt            # The user's master resume
│   └── cover.txt             # The user's base cover letter
└── jobs/
    └── 2026-02-24_acme-corp_copywriter/    <-- The "Run Folder"
        ├── job_raw.txt                     # The scraped webpage
        ├── resume_custom.txt               # AI-tailored resume (user edits this)
        └── cover_letter.txt                # AI-drafted cover letter
```

## 5. System Components (The `jdextract` package)

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
    *   *Prompt:* "You are a resume editor. Modify the bullets under 'Experience' in the provided resume to highlight overlaps with the provided job description. Output only plain text. Do not change the document structure."

### `Generator` (generate.go)
The heavy lifter. 
*   `func (a *App) ProcessJob(url string) (folderPath string, err error)`
    1. Fetches URL HTML.
    2. Extracts Company Name & Role from HTML (via a quick LLM call) to name the folder.
    3. Creates `~/.jdextract/jobs/YYYY-MM-DD_Company_Role/`.
    4. Calls LLM to customize the base `resume.txt`.
    5. Writes `resume_custom.txt` into the folder.
    6. Optionally generates `cover_letter.txt`.

### `Web Server` (web.go)
*   `//go:embed static/*`
*   `func (a *App) Serve(port string)`
*   Exposes endpoints `/api/process` (calls `ProcessJob`) and `/api/cover` (calls cover letter generation).

## 6. User Interface (CLI & Web)

**CLI Flow (for power users / scripting):**
```bash
$ jdextract setup
$ jdextract resume https://acme.com/job/123
  > Saved to: ~/.jdextract/jobs/2026-02-24_acme_copywriter/resume_custom.txt
$ jdextract cover https://acme.com/job/123
  > Saved to: ~/.jdextract/jobs/2026-02-24_acme_copywriter/cover_letter.txt
```

**Web Flow (for the copywriter friend):**
```bash
$ jdextract serve
  > Starting UI at http://localhost:8080
```
*   User opens browser.
*   Pastes Job URL into an input field. Clicks "Generate".
*   UI hits `/api/process`.
*   UI displays the generated text in a read-only text box and shows the file path.
*   User can click "Generate Cover Letter" -> UI hits `/api/cover`.

## 7. Execution Plan (Weekend Roadmap)

**Phase 1: The Engine (Saturday)**
1. Write `config.go` to establish the `~/.jdextract` folder structure.
2. Write `fetch.go` to do a dumb HTTP GET of a URL.
3. Write `llm.go` to send hardcoded strings to DeepSeek and print the response.
4. Combine them in `generate.go` to fetch a real URL, send it to the LLM alongside a dummy `resume.txt`, and write the output to disk.

**Phase 2: CLI & Polish (Sunday)**
1. Wire up `cmd/jdextract/main.go` using standard `os.Args` or `flag` (no need for heavy CLI frameworks like Cobra for 4 commands).
2. Add basic error handling ("API key missing").

**Phase 3: Visual Mode (Next Weekend)**
1. Create `static/index.html` with Alpine.js and Tailwind (CDN).
2. Add `web.go` with standard `net/http` handlers.
3. Wire the web buttons to call the Go functions you already wrote in Phase 1 & 2.

