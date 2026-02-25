---

# Design Document: jdextract

## 1. Project Summary
**jdextract** is a local-first CLI and embedded web tool designed to eliminate the cognitive fatigue of tailoring job applications. It does not automate *applying*; it automates *context-switching*. 

By providing a target Job URL and a base resume text file, the tool uses an LLM to align the resume's experience bullets to the job's requirements, outputting a customized plain text file for human review. The generated files serve as a natural "CRM" for the user's job hunt.

## 2. Core Constraints (To Prevent Scope Creep)
*   **Zero Databases:** The filesystem is the database. Every job run creates a structured folder.
*   **Zero JS Build Steps:** The web UI is a single embedded HTML file using CDN-hosted Alpine.js and Tailwind.
*   **One LLM Dependency:** DeepSeek for all text analysis and generation. `r.jina.ai` is used as a fetch fallback for JS-heavy pages; it requires no account or API key.
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
│       ├── main.go           # CLI argument parsing; calls jdextract.App methods
│       └── web/              # ← web/ lives here so //go:embed works
│           └── index.html    # Embedded UI (HTML/Alpine/Tailwind)
└── jdextract/                # Core package (importable library)
    ├── app.go                # Central App struct (holds config, coordinates flow)
    ├── config.go             # Resolves ~/.jdextract paths, reads env vars + config.yaml
    ├── fetch.go              # Hybrid HTTP fetch with fallback strategies
    ├── parse.go              # HTML parsing with golang.org/x/net/html (company/role + slug)
    ├── llm.go                # DeepSeek HTTP client, system prompts, retry logic
    ├── generate.go           # Orchestration: fetch -> parse -> LLM -> save to disk
    └── web.go                # net/http server; accepts fs.FS from caller
```

`//go:embed` can only embed from the **same directory or below** the `.go` file containing the directive. Placing `web/` under `cmd/jdextract/` and embedding in `main.go` keeps the core `jdextract` package free of UI concerns:

```go
// cmd/jdextract/main.go
//go:embed web
var webFiles embed.FS

func main() {
    app := jdextract.NewApp()
    // ...
    app.Serve("8080", webFiles)
}
```

## 4. The Data Model (Filesystem as CRM)
Every time a user runs the tool against a job, it generates a "Run Folder" in their output directory. This creates a natural, searchable history of all job applications.

**Base Directory:** `~/.jdextract/`
```text
~/.jdextract/
├── config.yaml               # API key (optional, fallback to DEEPSEEK_API_KEY env var)
├── templates/
│   ├── resume.txt            # The user's master resume
│   └── cover.txt             # The user's base cover letter (optional)
└── jobs/
    └── 2026-02-24_acme-corp_copywriter_a7x9/    <-- "Run Folder" (with URL hash suffix)
        ├── job.json                          # Structured metadata (see below)
        ├── job_raw.txt                       # The scraped webpage content
        ├── resume_custom.txt                 # AI-tailored resume (user edits this)
        └── cover_letter.txt                  # AI-drafted cover letter (optional)
```

### Job Metadata (`job.json`)
Stores structured + freeform data together for both human readability and web UI functionality:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "company": "Acme Corp",
  "role": "Copywriter",
  "status": "draft",
  "date_created": "2026-02-24T14:30:00Z",
  "date_applied": null,
  "url": "https://acme.com/jobs/123",
  "tokens_used": 2847,
  "notes_md": "## Follow-up\n- Emailed hiring manager...\n## Research\n- Company focuses on..."
}
```
- **`id`**: UUID v4 generated on creation — clean API identifier for `/api/jobs/{id}`
- **`date_created`**: ISO 8601 timestamp — enables reliable sorting in web UI
- **`status`**: One of `draft`, `applied`, `interviewing`, `offer`, `rejected`
- **`notes_md`**: Freeform markdown for user notes

This enables status badges, filtering, sorting in the web UI while keeping human-editable notes.

## 5. System Components (The `jdextract` package)

### `App` (app.go)
The central orchestrator. It holds the configuration and provides the high-level methods used by both the CLI and Web interfaces.
*   `func NewApp() *App`
*   `func (a *App) Setup() error` (Creates directories and example templates)

### `Config` (config.go)
*   Finds the user's home directory.
*   Loads `DEEPSEEK_API_KEY` with standard precedence: **env var > config.yaml > error**
*   This follows 12-factor app conventions — deployed environments can override file config.

### `Fetch` (fetch.go) - Hybrid Strategy
Handles the "JS-Page Problem" without headless browsers (preserves "Zero System Dependencies"):
1. **First:** Try direct HTTP GET (works for Greenhouse, Lever static fallbacks)
2. **Second:** Try `https://r.jina.ai/http://URL` (free service extracts text from React SPAs)
3. **Third:** Return a clear error telling the user to save the job text to a file and re-run with `--local ./file.txt`

**Safety caps:** 500KB for raw HTML; 100KB for Jina responses (which return extracted markdown — far denser than raw HTML).

### `Parse` (parse.go)
Extract company/role from HTML without LLM (saves tokens and latency):
*   Use `golang.org/x/net/html` (zero external dependencies) to parse `<title>` and `<h1>` tags
*   Regex patterns for common job board formats
*   Only fallback to LLM if HTML parsing fails completely
*   **`slug()` helper:** Sanitize extracted strings before use in folder names. Company names like `"Acme & Co."` or roles like `"Sr. Engineer / Backend"` must be normalized:
    ```go
    // "Acme & Co." → "acme-co"
    func slug(s string) string {
        s = strings.ToLower(s)
        s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
        return strings.Trim(s, "-")
    }
    ```

### `LLM Client` (llm.go)
A `net/http` wrapper for `api.deepseek.com` with production-ready features:
*   **Optional Cover Letter:** Pass `nil` for `baseCover` to skip cover letter generation
*   **Batched Calls:** When cover letter is requested, combine into one prompt and use `response_format: {"type": "json_object"}` (DeepSeek JSON mode) to get a structured response — eliminates delimiter fragility and cuts API costs by ~40%:
    ```json
    { "resume": "...", "cover_letter": "..." }
    ```
*   **Exponential Backoff:** Retry with backoff for HTTP 429 (rate limit) responses

```go
func (l *LLM) GenerateAll(ctx context.Context, jobText, baseResume string, baseCover *string) (
    resume string,
    cover *string,  // nil if baseCover was nil
    tokensUsed int,
    err error,
)
```

### `Generator` (generate.go)
The heavy lifter with unified input and rich return values:

```go
// Unified input type for URL or local file — fields are mutually exclusive.
// Process() validates and returns an error if both or neither are set.
type JobInput struct {
    URL       string // Mutually exclusive with LocalFile
    LocalFile string // Mutually exclusive with URL
}

type JobResult struct {
    ID         string // UUID from job.json
    Company    string
    Role       string
    TokensUsed int
    FolderPath string
    ResumePath string
    CoverPath  string // Empty if no cover letter generated
}

func (a *App) Process(ctx context.Context, input JobInput) (*JobResult, error)
```

**Process() workflow:**
1. Reads job text from URL (hybrid fetch) or local file
2. Extracts Company Name & Role from HTML via golang.org/x/net/html (fallback to LLM)
3. Generates UUID for the job
4. Creates `~/.jdextract/jobs/YYYY-MM-DD_Company_Role_URLHash/` (hash prevents collisions)
5. Saves `job_raw.txt` (scraped/pasted content)
6. Calls LLM to generate resume (and optionally cover letter)
7. Writes `resume_custom.txt` and `cover_letter.txt` (if generated)
8. Creates `job.json` with metadata including `id` and `date_created`

### `Web Server` (web.go)
*   Accepts `fs.FS` from the caller — the `//go:embed web` directive lives in `cmd/jdextract/main.go`, not here
*   `func (a *App) Serve(port string, ui fs.FS)`
*   Exposes endpoints:
    *   `POST /api/process` - accepts `{ "url": "..." }` or `{ "local_text": "..." }`, returns `JobResult`
    *   `GET /api/jobs` - lists all jobs with filtering/sorting by status, date, company
    *   `GET /api/jobs/{id}` - get job metadata by UUID
    *   `PATCH /api/jobs/{id}` - update job metadata (status, notes)
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
$ jdextract resume --local ./my_job_paste.txt    # Provide a saved file (not stdin)
  > Processing local file...
$ jdextract resume --no-cover https://acme.com/job/123  # Skip cover letter
  > Skipping cover letter (no template or --no-cover)
$ jdextract list                                  # Print job history + status
  > ID       DATE        COMPANY     ROLE          STATUS
  > a7x9...  2026-02-24  Acme Corp   Copywriter    draft
$ jdextract status a7x9 applied                   # Update status in job.json
  > Updated: 2026-02-24_acme_copywriter_a7x9 → applied
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
1. Write `config.go` to establish the `~/.jdextract` folder structure + config.yaml support (env var precedence).
2. Write `fetch.go` with hybrid strategy (direct GET → r.jina.ai → error with --local hint).
3. Write `parse.go` with `golang.org/x/net/html` for company/role extraction.
4. Write `llm.go` with exponential backoff for 429s and optional cover letter support.
5. Combine them in `generate.go` with `JobInput` struct and `job.json` (including UUID + date_created).

**Phase 2: CLI & Polish (Sunday)**
1. Wire up `cmd/jdextract/main.go` using `flag.NewFlagSet` per subcommand (cleaner than global flag.Parse).
2. Add `--local <file>` flag — accepts a saved text file, not stdin.
3. Add `--no-cover` flag to skip cover letter generation.
4. Add `jdextract list` — tabular job history with status column.
5. Add `jdextract status <id|folder> <status>` — update status field in `job.json`.
6. Add basic error handling ("API key missing", "Fetch failed, save text to a file and re-run with --local ./file.txt").

**Phase 3: Visual Mode (Next Weekend)**
1. Create `static/index.html` with Alpine.js and Tailwind (CDN).
2. Add `web.go` with standard `net/http` handlers + Origin header validation.
3. Wire the web buttons to call the Go functions.
4. Add job list view with status badges and filtering.

This design gives you a robust, highly functional tool with practically zero boilerplate.