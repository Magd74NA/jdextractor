---

# Design Document: jdextract

## 1. Project Summary
**jdextract** is a local-first CLI and embedded web tool designed to eliminate the cognitive fatigue of tailoring job applications. It does not automate *applying*; it automates *context-switching*.

By providing a target Job URL and a base resume text file, the tool uses an LLM to align the resume's experience bullets to the job's requirements, outputting a customized plain text file for human review. The generated files serve as a natural "CRM" for the user's job hunt.

## 2. Core Constraints
*   **Zero Databases:** The filesystem is the database. Every job run creates a structured folder.
*   **Zero JS Build Steps:** The web UI is a single embedded HTML file using CDN-hosted Alpine.js, Tailwind, and DaisyUI.
*   **One LLM Dependency:** DeepSeek for all text analysis and generation. `r.jina.ai` is used as a fetch fallback for JS-heavy pages; it requires no account or API key.
*   **One Go Module Dependency:** `golang.org/x/net/html` for HTML parsing. All other functionality uses Go stdlib.
*   **Zero System Dependencies:** No external tools required (no Pandoc, no headless browsers for MVP1).
*   **Plain Text Output:** All generated files are `.txt` for simplicity.
*   **Human-in-the-Loop:** The AI generates text. The human reviews and edits the text.

## 3. Directory Structure

```text
jdextract/
├── go.mod
├── cmd/
│   ├── main.go              # CLI entry point; //go:embed web; calls jdextract.App
│   └── web/
│       └── index.html       # Embedded UI (HTML/Alpine/Tailwind/DaisyUI)
└── jdextract/               # Core package (importable library)
    ├── app.go               # Central App struct
    ├── config.go            # ~/.jdextract paths, env vars + config file
    ├── fetch.go             # Hybrid HTTP fetch with fallback strategies
    ├── parse.go             # HTML parsing (company/role + slug)
    ├── llm.go               # DeepSeek HTTP client, retry logic
    ├── generate.go          # Orchestration: fetch -> parse -> LLM -> save
    └── web.go               # net/http server; accepts fs.FS from caller
```

The `//go:embed web` directive in `cmd/main.go` embeds the UI. `fs.Sub(webFiles, "web")` strips the prefix before passing to `App.Serve()`.

## 4. Data Model (Filesystem as CRM)
Every run creates a "Run Folder" under the output directory, forming a searchable application history.

**Base Directory:** `~/.jdextract/`
```text
~/.jdextract/
├── config                    # KEY=VALUE config (optional, fallback to env vars)
├── templates/
│   ├── resume.txt            # The user's master resume
│   └── cover.txt             # The user's base cover letter (optional)
└── jobs/
    └── 2026-02-24_acme-corp_copywriter_a7x9/    <-- "Run Folder"
        ├── job.json                          # Structured metadata
        ├── job_raw.txt                       # Scraped webpage content
        ├── resume_custom.txt                 # AI-tailored resume
        ├── cover_letter.txt                  # AI-drafted cover letter (optional)
        └── notes.md                          # Freeform user notes
```

### Job Metadata (`job.json`)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "company": "Acme Corp",
  "role": "Copywriter",
  "status": "draft",
  "date_created": "2026-02-24T14:30:00Z",
  "date_applied": null,
  "url": "https://acme.com/jobs/123",
  "tokens_used": 2847
}
```
- **`id`**: UUID v4 via `crypto/rand`. CLI lookup uses **UUID prefix matching** (mirrors git short SHA). Error on zero or multiple matches.
- **`status`**: One of `draft`, `applied`, `interviewing`, `offer`, `rejected`

### User Notes (`notes.md`)
Sidecar file alongside `job.json` — keeps metadata machine-readable and notes human-editable. Created empty on first run; absence is fine.

### API Response (`JobDetail`)
`GET /api/jobs/{id}` returns a merged view of `job.json` + `notes.md`:

```go
type JobDetail struct {
    JobMetadata        // all job.json fields
    Notes       string // contents of notes.md; empty if absent
}
```

## 5. System Components (The `jdextract` package)

### `App` (app.go)
Central orchestrator holding configuration. Used by both CLI and Web interfaces.

```go
func NewApp() *App
func (a *App) Setup() error
```

The caller (`main.go`) creates the root context via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` and passes it into App methods. App does not handle signals itself.

### `Config` (config.go)
*   Resolves `~/.jdextract` paths and reads `~/.jdextract/config` (KEY=VALUE format, `#` comments).
*   Precedence: **env var > config file > default/error**. `DEEPSEEK_MODEL` defaults to `deepseek-chat`.
*   **Permissions:** Config file created with `0600` (contains API key). Job output files use `0644`.

### `Fetch` (fetch.go) — Hybrid Strategy
1. **First:** Direct HTTP GET (works for Greenhouse, Lever static pages)
2. **Second:** `https://r.jina.ai/{fullURL}` (extracts text from React SPAs, free, no auth)
3. **Third:** Return error with `--local` hint

**Safety caps:** 500KB raw HTML; 100KB Jina responses.

**Distinct errors:** Jina HTTP 429 returns `ErrJinaRateLimited` (suggests wait + retry). Other Jina failures return `ErrJinaExtraction` (suggests `--local` fallback). Messages differ so the user knows which action to take.

### `Parse` (parse.go)
Extracts company/role from HTML without LLM (saves tokens and latency):
*   Parses `<title>` and `<h1>` tags via `golang.org/x/net/html`
*   Regex patterns for common job board formats (Greenhouse, Lever, Workday)
*   Falls back to LLM only if HTML parsing fails completely

```go
// slug normalizes s for use in folder names ("Acme & Co." -> "acme-co").
// Returns "unknown" if the sanitized result is empty.
func slug(s string) string
```

### `LLM Client` (llm.go)
*   **Batched Calls:** Combines resume + cover letter into one prompt using `response_format: {"type": "json_object"}` — returns `{"resume": "...", "cover_letter": "..."}`. Eliminates delimiter fragility, cuts API costs ~40%. Parse defensively: decode into raw map first, check key existence before assigning to struct fields.
*   **Optional Cover Letter:** Pass `nil` for `baseCover` to skip; omits `cover_letter` from response. **Default behavior:** generate cover letter if `templates/cover.txt` exists AND `--no-cover` is not set. No template = no cover letter, silently.
*   **Exponential Backoff:** Retries on HTTP 429. Non-429 errors return immediately.

```go
func (l *LLM) GenerateAll(ctx context.Context, jobText, baseResume string, baseCover *string) (
    resume string,
    cover *string,
    tokensUsed int,
    err error,
)
```

### `Generator` (generate.go)

```go
type JobInput struct {
    URL       string // mutually exclusive
    LocalFile string // mutually exclusive
    RawText   string // mutually exclusive (web UI paste path)
}

type JobResult struct {
    ID         string
    Company    string
    Role       string
    TokensUsed int
    FolderPath string
    ResumePath string
    CoverPath  string // empty if no cover letter generated
}

func (a *App) Process(ctx context.Context, input JobInput) (*JobResult, error)
```

**Process() workflow:**
1. Validate `JobInput` — exactly one of URL / LocalFile / RawText must be set
2. Acquire job text:
   - URL: hybrid fetch (`fetch.go`)
   - LocalFile: read file from disk
   - RawText: use directly
3. Extract company & role:
   - **URL path:** HTML parse via `golang.org/x/net/html`, fallback to LLM
   - **LocalFile / RawText path:** skip HTML parse, go directly to LLM extraction
4. Apply `slug()` to company and role
5. Generate UUID v4 (`crypto/rand`)
6. Generate folder hash suffix: first 4 hex chars of SHA-256 of the source (URL string, absolute file path, or raw text)
7. Create run folder `~/.jdextract/jobs/YYYY-MM-DD_company_role_hash/`
8. Save `job_raw.txt`
9. Call LLM (resume + optional cover letter)
10. Write output files (`resume_custom.txt`, `cover_letter.txt`)
11. Write `job.json` (atomic: write to `.tmp`, then `os.Rename`)

**Failure behavior:** On partial failure after folder creation, the folder is left on disk for inspection. Re-running against the same source errors if the folder already exists (user must delete manually). When invoked via the web API, `Process()` receives a context with `context.WithTimeout` (default 300s).

### `Web Server` (web.go)

```go
func (a *App) Serve(ctx context.Context, port string, ui fs.FS) error
```

Accepts bare port number (e.g. `"8080"`), prepends `:` internally. Uses `http.Server.Shutdown(ctx)` for graceful shutdown — on context cancellation, in-flight requests (including long LLM calls) finish before the server exits.

**Endpoints:**
*   `POST /api/process` — accepts `{"url": "..."}` or `{"local_text": "..."}`, returns `JobResult`. Wraps `Process()` with `context.WithTimeout` (300s).
*   `GET /api/jobs` — lists jobs; query params: `?status=applied&sort=date_desc&company=acme`. Returns `id` truncated to 8 chars. Tolerates corrupt `job.json` entries (logs warning, skips).
*   `GET /api/jobs/{id}` — returns `JobDetail` (job.json merged with notes.md content).
*   `PATCH /api/jobs/{id}` — **writable fields: `status`, `date_applied`, `notes` only.** All other fields (`id`, `company`, `role`, `date_created`, `url`, `tokens_used`) are read-only and rejected if present. `notes` writes to `notes.md` sidecar.

**CSRF:** Reject requests where `Origin` header is present and does not match `http://localhost:{port}` or `http://127.0.0.1:{port}`. Requests without `Origin` (e.g. curl) pass through. Additionally, POST/PATCH endpoints require `Content-Type: application/json` to block simple form submissions.

## 6. User Interface

**CLI:**
```bash
$ jdextract setup
$ jdextract generate https://acme.com/job/123
  > Saved to: ~/.jdextract/jobs/2026-02-24_acme_copywriter_a7x9/
  > Resume: resume_custom.txt  |  Cover: cover_letter.txt
  > Tokens used: 2847
$ jdextract generate --local ./my_job_paste.txt
$ jdextract generate --no-cover https://acme.com/job/123
$ jdextract list                                  # UUID truncated to 8 chars
  > ID        DATE        COMPANY     ROLE          STATUS
  > 550e8400  2026-02-24  Acme Corp   Copywriter    draft
$ jdextract status 550e applied                   # UUID prefix match
$ jdextract serve --port 9090                     # Default: 8080
```

`jdextract status` validates against `draft|applied|interviewing|offer|rejected` before writing.

**Web:** User pastes a Job URL (or raw text), clicks "Generate". UI shows the generated text, file paths, and token usage. Job list view with status badges, filtering, and sorting. Loading spinner with timeout-specific error if the server deadline is exceeded.
