---

# Design Document: jdextract

## 1. Project Summary
**jdextract** is a local-first CLI and embedded web tool designed to eliminate the cognitive fatigue of tailoring job applications. It does not automate *applying*; it automates *context-switching*. 

By providing a target Job URL and a base resume text file, the tool uses an LLM to align the resume's experience bullets to the job's requirements, outputting a customized plain text file for human review. The generated files serve as a natural "CRM" for the user's job hunt.

## 2. Core Constraints (To Prevent Scope Creep)
*   **Zero Databases:** The filesystem is the database. Every job run creates a structured folder.
*   **Zero JS Build Steps:** The web UI is a single embedded HTML file using CDN-hosted Alpine.js and Tailwind.
*   **One API Dependency:** DeepSeek for all text analysis and generation.
*   **Zero System Dependencies:** No external tools required (no Pandoc, no headless browsers for MVP1).
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
│   ├── config.go             # Resolves ~/.jdextract paths, reads env vars + config.yaml
│   ├── fetch.go              # Hybrid HTTP fetch with fallback strategies
│   ├── parse.go              # HTML parsing with goquery (company/role extraction)
│   ├── llm.go                # DeepSeek HTTP client, system prompts, retry logic
│   ├── generate.go           # Orchestration: fetch -> parse -> LLM -> save to disk
│   └── web.go                # net/http server wrapping App methods
└── static/
    └── index.html            # Embedded UI (HTML/Alpine/Tailwind)
```

## 4. The Data Model (Filesystem as CRM)
Every time a user runs the tool against a job, it generates a "Run Folder" in their output directory. This creates a natural, searchable history of all job applications.

**Base Directory:** `~/.jdextract/`
```text
~/.jdextract/
├── config.yaml               # API key (optional, fallback to DEEPSEEK_API_KEY env var)
├── templates/
│   ├── resume.txt            # The user's master resume
│   └── cover.txt             # The user's base cover letter
└── jobs/
    └── 2026-02-24_acme-corp_copywriter_a7x9/    <-- "Run Folder" (with URL hash suffix)
        ├── job.json                          # Structured metadata (see below)
        ├── job_raw.txt                       # The scraped webpage content
        ├── resume_custom.txt                 # AI-tailored resume (user edits this)
        └── cover_letter.txt                  # AI-drafted cover letter
```

### Job Metadata (`job.json`)
Stores structured + freeform data together for both human readability and web UI functionality:
```json
{
  "company": "Acme Corp",
  "role": "Copywriter",
  "status": "draft",
  "date_applied": null,
  "url": "https://acme.com/jobs/123",
  "tokens_used": 2847,
  "notes_md": "## Follow-up\n- Emailed hiring manager...\n## Research\n- Company focuses on..."
}
```
This enables status badges, filtering, sorting in the web UI while keeping human-editable notes.

## 5. System Components (The `jdextract` package)

### `App` (app.go)
The central orchestrator. It holds the configuration and provides the high-level methods used by both the CLI and Web interfaces.
*   `func NewApp() *App`
*   `func (a *App) Setup() error` (Creates directories and example templates)

### `Config` (config.go)
*   Finds the user's home directory.
*   Loads `DEEPSEEK_API_KEY` from config.yaml first, falls back to environment variable.
*   Config file preferred for non-technical users who struggle with env vars.

### `Fetch` (fetch.go) - Hybrid Strategy
Handles the "JS-Page Problem" without headless browsers (preserves "Zero System Dependencies"):
1. **First:** Try direct HTTP GET (works for Greenhouse, Lever static fallbacks)
2. **Second:** Try `https://r.jina.ai/http://URL` (free service extracts text from React SPAs)
3. **Third:** Fail gracefully—prompt user to paste text manually with `--local` flag

**Safety:** Cap HTML fetch at 500KB to prevent downloading massive SPAs.

### `Parse` (parse.go)
Extract company/role from HTML without LLM (saves tokens and latency):
*   Use `goquery` to parse `<title>` and `<h1>` tags
*   Regex patterns for common job board formats
*   Only fallback to LLM if HTML parsing fails completely

### `LLM Client` (llm.go)
A `net/http` wrapper for `api.deepseek.com` with production-ready features:
*   **Batched Calls:** Combine resume customization + cover letter generation into one prompt with `--- COVER LETTER ---` delimiter (cuts API costs by ~40%)
*   **Exponential Backoff:** Retry with backoff for HTTP 429 (rate limit) responses
*   `func (l *LLM) GenerateAll(jobText, baseResume, baseCover string) (resume, cover string, tokensUsed int, error)`
    *   *Prompt:* "You are a resume editor. Modify the bullets under 'Experience' to highlight overlaps with the job description. Output only plain text. Then write a cover letter after the delimiter --- COVER LETTER ---. Do not change the resume structure."

### `Generator` (generate.go)
The heavy lifter with rich return values for web UI:
```go
type JobResult struct {
    Company     string
    Role        string
    TokensUsed  int
    FolderPath  string
    ResumePath  string
    CoverPath   string
}

func (a *App) ProcessJob(url string) (*JobResult, error)
```
1. Fetches URL HTML (with hybrid strategy)
2. Extracts Company Name & Role from HTML via goquery (fallback to LLM)
3. Creates `~/.jdextract/jobs/YYYY-MM-DD_Company_Role_URLHash/` (hash prevents collisions on re-applications)
4. Saves `job_raw.txt` (scraped content)
5. Calls LLM to generate both resume and cover letter (batched)
6. Writes `resume_custom.txt` and `cover_letter.txt`
7. Creates `job.json` with metadata

### `Web Server` (web.go)
*   `//go:embed static/*`
*   `func (a *App) Serve(port string)`
*   Exposes endpoints:
    *   `/api/process` - calls `ProcessJob()`, returns JSON with `JobResult`
    *   `/api/jobs` - lists all jobs with filtering/sorting
    *   `/api/jobs/{id}` - get/update job metadata
*   **CSRF Protection:** Validate `Origin` headers even on localhost

## 6. User Interface (CLI & Web)

**CLI Flow (for power users / scripting):**
```bash
$ jdextract setup
$ jdextract resume https://acme.com/job/123
  > Saved to: ~/.jdextract/jobs/2026-02-24_acme_copywriter_a7x9/
  > Resume: resume_custom.txt
  > Cover: cover_letter.txt
  > Tokens used: 2847
$ jdextract resume --local ./my_job_paste.txt    # Manual paste mode
  > Processing local file...
```

**Web Flow (for the copywriter friend):**
```bash
$ jdextract serve
  > Starting UI at http://localhost:8080
```
*   User opens browser.
*   Pastes Job URL into an input field. Clicks "Generate".
*   UI hits `/api/process`.
*   UI displays the generated text in a read-only text box with file paths and token usage.
*   Job list shows status badges (draft, applied, interviewing) with sorting/filtering.

## 7. Execution Plan (Weekend Roadmap)

**Phase 1: The Engine (Saturday)**
1. Write `config.go` to establish the `~/.jdextract` folder structure + config.yaml support.
2. Write `fetch.go` with hybrid strategy (direct GET → r.jina.ai → error with --local hint).
3. Write `parse.go` with goquery for company/role extraction.
4. Write `llm.go` with exponential backoff for 429s.
5. Combine them in `generate.go` to fetch, parse, generate, and save with `job.json`.

**Phase 2: CLI & Polish (Sunday)**
1. Wire up `cmd/jdextract/main.go` using standard `os.Args` or `flag`.
2. Add `--local` flag for manual text input.
3. Add basic error handling ("API key missing", "Fetch failed, try --local").

**Phase 3: Visual Mode (Next Weekend)**
1. Create `static/index.html` with Alpine.js and Tailwind (CDN).
2. Add `web.go` with standard `net/http` handlers + Origin header validation.
3. Wire the web buttons to call the Go functions.
4. Add job list view with status badges and filtering.

This design gives you a robust, highly functional tool with practically zero boilerplate.