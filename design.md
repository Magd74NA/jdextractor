# Design Document: jdextract

## 1. Project Summary
**jdextract** is a local-first CLI and embedded web tool designed to eliminate the cognitive fatigue of tailoring job applications. It does not automate *applying*; it automates *context-switching*.

By providing a target Job URL and a base resume text file, the tool uses an LLM to align the resume's experience bullets to the job's requirements, outputting a customized plain text file for human review. The generated files serve as a natural "CRM" for the user's job hunt.

## 2. Core Constraints
*   **Zero Databases:** The filesystem is the database. Every job run creates a structured folder.
*   **Zero JS Build Steps:** The web UI is a single embedded HTML file using CDN-hosted Alpine.js, Tailwind, and DaisyUI.
*   **One LLM Dependency:** DeepSeek for all text analysis and generation. `r.jina.ai` is used for all URL fetching; it requires no account or API key.
*   **One Go Module Dependency:** `github.com/toon-format/toon-go` (vendored). Used to serialize the parsed AST into TOON format before sending to the LLM — compact, token-efficient, structured. All other functionality uses Go stdlib.
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
    ├── config.go            # JSON config loading and file creation
    ├── fetch.go             # HTTP fetch via r.jina.ai; exponential backoff on 429
    ├── parse.go             # Line-level AST classifier; returns []JobDescriptionNode
    ├── llm.go               # DeepSeek HTTP client, retry logic
    ├── generate.go          # LLM orchestration: TOON encode → prompt → GenerateAll
    ├── storage.go           # Filesystem/job management: slug, structs, Process()
    └── web.go               # net/http server; accepts fs.FS from caller
```

The `//go:embed web` directive in `cmd/main.go` embeds the UI. `fs.Sub(webFiles, "web")` strips the prefix before passing to `App.Serve()`.

## 4. Data Model (Filesystem as CRM)
Every run creates a "Run Folder" under the data directory, forming a searchable application history.

**Portable layout** — all paths resolve relative to the executable via `getPortablePaths()`. On macOS inside a `.app` bundle, the root is the directory containing the `.app`. No hard-coded home directory.

```text
<exe_dir>/
├── jdextract                 # binary
├── config/                   # Configuration directory
│   ├── config.json           # JSON config (optional)
│   └── templates/
│       ├── resume.txt        # The user's master resume
│       └── cover.txt         # The user's base cover letter (optional)
└── data/
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
Central orchestrator holding configuration and portable paths. Used by both CLI and Web interfaces.

```go
type PortablePaths struct {
    Root      string
    Jobs      string
    Data      string
    Config    string
    Templates string
}

type App struct {
    Paths  PortablePaths
    Config Config
}

func getPortablePaths() (PortablePaths, error)
func NewApp() (*App, error)
func (a *App) Setup() error
func (a *App) createExampleTemplates() error
```

`getPortablePaths()` resolves the executable's location (following symlinks), then derives `config`, `templates/`, and `data/` paths relative to it. On macOS inside a `.app` bundle, it walks up to the directory containing the `.app`. The caller (`main.go`) creates the root context via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` and passes it into App methods. App does not handle signals itself.

`NewApp()` returns an error rather than calling `os.Exit()`, making it testable and allowing callers to handle errors gracefully. It calls `Setup()` which creates all required directories and example templates.

`createExampleTemplates()` creates `resume.txt` and `cover.txt` template files in `config/templates/` if they don't already exist (won't overwrite user customizations). Files are created with `0644` permissions.

### `Config` (config.go)
*   Reads `<exe_dir>/config/config.json` (JSON format). Path provided by `App.Paths.Config`.
*   `DEEPSEEK_MODEL` defaults to `deepseek-chat`. Env var override and precedence rules deferred to post-MVP1.
*   **Permissions:** Config file created with `0600` (contains API key). Job output files use `0644`.

### `Fetch` (fetch.go)
All URL fetching goes through `https://r.jina.ai/{fullURL}`, which handles both static pages and JS-heavy SPAs without a headless browser.

**Safety cap:** 100KB response limit.

**Errors:** HTTP 429 is handled internally with exponential backoff and retry. `FetchJobDescription` accepts a `context.Context` as its first parameter so the backoff loop can be interrupted by the caller's timeout (e.g. the 300s web deadline) or Ctrl+C. All other failures return the error directly; user can fall back to `--local`.

### `Parse` (parse.go)
Converts the markdown returned by `r.jina.ai` into a typed, filtered line-level AST.

Each non-empty line is classified as one of 15 `NodeType` constants ordered from most generic (`body`) to most specific (`jina_title`). Noise types (`jina_marker`, `setext_underline`, `nav_link`) are stripped; long low-information body lines (> 300 chars) are dropped to preserve LLM context window. Returns `[]JobDescriptionNode` — serialization to TOON is handled downstream in `generate.go`.

```go
// classifyLine returns the most specific NodeType for a single non-empty line,
// or "" to signal the line should be dropped.
func classifyLine(line string) string

// Parse returns the filtered AST: noise removed, long body lines dropped.
func Parse(s string) []JobDescriptionNode
```

### `LLM Client` (llm.go)
Pure HTTP interface — no prompt text or business logic. Contains the wire-format types and `InvokeDeepseekApi`.

*   **Wire types:** `deepseekRequest`, `deepseekResponse`, `deepseekMessage` (unexported). Request uses `response_format: {"type": "json_object"}` and `stream: false`.
*   **Exponential Backoff:** `InvokeDeepseekApi` retries on HTTP 429 recursively. Non-429 errors return immediately.

```go
func InvokeDeepseekApi(ctx context.Context, apiKey string, c *http.Client, backoff int, requestBody json.RawMessage) (string, error)
```

### `Generator` (generate.go)
Pure LLM orchestration — no filesystem access. Contains the prompt, TOON wrapper, and `GenerateAll`.

*   **TOON encoding:** `[]JobDescriptionNode` is wrapped in `jobDescriptionPayload{Nodes: nodes}` and serialized via `toon.MarshalString` inside `GenerateAll` — compact tabular format, token-efficient.
*   **Batched prompt:** Single LLM call returns company, role, resume, cover letter (optional), and a match score. Uses `response_format: {"type": "json_object"}`. Response decoded into a typed inline struct — no defensive map needed.
*   **JSON schema from LLM:**
```json
{
  "Result":  { "company": "string", "role": "string" },
  "Resume":  "full tailored resume text",
  "Cover":   "tailored cover letter — omitted if no base cover provided",
  "Score":   7
}
```
*   **Score:** Integer 1–10 subjective rating of how well the base resume matches the job requirements.
*   **Optional Cover Letter:** Pass `nil` for `baseCover` to skip. Default: generate if `templates/cover.txt` exists AND `--no-cover` is not set.
*   **Company/Role Extraction:** Extracted by the LLM from the TOON payload in the same call — no separate fallback prompt.

```go
func GenerateAll(
    ctx        context.Context,
    apiKey     string,
    model      string,
    c          *http.Client,
    nodes      []JobDescriptionNode,
    baseResume string,
    baseCover  *string,
) (company, role, resume string, cover *string, score, tokensUsed int, err error)
```

### `Storage` (storage.go)
All filesystem and job-management concerns. No LLM logic — calls `GenerateAll` as a black box.

```go
// slug normalizes s for use in folder names ("Acme & Co." -> "acme-co").
// Returns "unknown" if the sanitized result is empty.
func slug(s string) string

type JobInput struct {
    URL       string // mutually exclusive
    LocalFile string // mutually exclusive
    RawText   string // mutually exclusive (web UI paste path)
}

type JobMetadata struct {
    ID          string     `json:"id"`
    Company     string     `json:"company"`
    Role        string     `json:"role"`
    Status      string     `json:"status"`
    DateCreated time.Time  `json:"date_created"`
    DateApplied *time.Time `json:"date_applied"`
    URL         string     `json:"url"`
    TokensUsed  int        `json:"tokens_used"`
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
   - URL: fetch via `fetch.go`
   - LocalFile: read file from disk
   - RawText: use directly
3. Parse: `Parse()` → `[]JobDescriptionNode` (typed, filtered AST)
4. Load templates (`resume.txt` required; `cover.txt` optional)
5. Call `GenerateAll()` — returns company, role, resume, cover, score, tokensUsed
6. Apply `slug()` to company and role
7. Generate UUID v4 (`crypto/rand`)
8. Generate folder hash suffix: first 4 hex chars of SHA-256 of the source (URL string, absolute file path, or raw text)
9. Create run folder `<exe_dir>/data/jobs/YYYY-MM-DD_company_role_hash/`
10. Write `job_raw.txt`
11. Write output files (`resume_custom.txt`, `cover_letter.txt`)
12. Write `job.json` atomically (write to `.tmp`, then `os.Rename`)

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
  > Saved to: ./data/jobs/2026-02-24_acme_copywriter_a7x9/
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