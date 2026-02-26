# Design Document: jdextract

## 1. Project Summary
**jdextract** is a local-first CLI and embedded web tool designed to eliminate the cognitive fatigue of tailoring job applications. It does not automate *applying*; it automates *context-switching*.

By providing a target Job URL and a base resume text file, the tool uses an LLM to align the resume's experience bullets to the job's requirements, outputting a customized plain text file for human review. The generated files serve as a natural "CRM" for the user's job hunt.

## 2. Core Constraints
*   **Zero Databases:** The filesystem is the database. Every job run creates a structured folder.
*   **Zero JS Build Steps:** The web UI is a single embedded HTML file using CDN-hosted Alpine.js, Tailwind, and DaisyUI.
*   **One LLM Dependency:** DeepSeek for all text analysis and generation. `r.jina.ai` is used for all URL fetching; it requires no account or API key.
*   **Zero Go Module Dependencies:** All functionality uses Go stdlib only. The AST is serialized to minified JSON via `encoding/json` before sending to the LLM.
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
    ├── app.go               # Central App struct, PortablePaths, NewApp()
    ├── setup.go             # Setup() and createExampleTemplates()
    ├── config.go            # JSON config loading and file creation
    ├── fetch.go             # HTTP fetch via r.jina.ai; exponential backoff on 429
    ├── parse.go             # Line-level AST classifier; returns []JobDescriptionNode
    ├── llm.go               # DeepSeek HTTP client, retry logic
    ├── generate.go          # LLM orchestration: JSON encode → prompt → GenerateAll
    ├── storage.go           # FS primitives: slugify, createApplicationDirectory, fetchResume, fetchCover
    ├── process.go           # Orchestration: (a *App) Process(), applicationMeta struct
    └── web.go               # (Phase 5, not yet created) net/http server; accepts fs.FS from caller
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
        └── 2026-02-24-a7x9k3m2-intermediate-copywriter/    <-- "Run Folder"
            ├── meta.json        # Structured metadata (applicationMeta)
            ├── resume.txt       # AI-tailored resume
            └── cover.txt        # AI-drafted cover letter (if base cover provided)
```

### Job Metadata (`meta.json`)
```json
{
  "company": "Acme Corp",
  "role": "Intermediate Copywriter",
  "score": 7,
  "tokens": 2847,
  "date": "2026-02-24"
}
```
- **`company`**, **`role`**: Extracted by the LLM from the job description.
- **`score`**: Integer 1–10 subjective fit rating from the LLM. Defaults to 0 on parse failure.
- **`tokens`**: Total tokens used for the LLM call.
- **`date`**: `YYYY-MM-DD` from `currentDate()`.

Richer metadata (status tracking, URL storage, UUID-based IDs, notes sidecar) is deferred to Phase 4/5.

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
    Client http.Client
}

func getPortablePaths() (PortablePaths, error)
func NewApp() (*App, error)
func (a *App) Setup() error              // in setup.go
func (a *App) createExampleTemplates() error  // in setup.go
```

`getPortablePaths()` resolves the executable's location (following symlinks), then derives `config`, `templates/`, and `data/` paths relative to it. On macOS inside a `.app` bundle, it walks up to the directory containing the `.app`. The caller (`main.go`) creates the root context via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` and passes it into App methods. App does not handle signals itself.

`NewApp()` returns an error rather than calling `os.Exit()`, making it testable and allowing callers to handle errors gracefully. It calls `Setup()` which creates all required directories and example templates.

`createExampleTemplates()` creates `resume.txt` and `cover.txt` template files in `config/templates/` if they don't already exist (won't overwrite user customizations). Files are created with `0644` permissions.

### `Config` (config.go)
*   Reads `<exe_dir>/config/config.json` (JSON format). Path provided by `App.Paths.Config`.
*   Config struct fields: `DeepSeekApiKey` (string), `DeepSeekModel` (string, defaults to `"deepseek-chat"`), `Port` (int). Env var override deferred to post-MVP1.
*   **Permissions:** Config file created with `0600` (contains API key). Job output files use `0644`.

### `Fetch` (fetch.go)
All URL fetching goes through `https://r.jina.ai/{fullURL}`, which handles both static pages and JS-heavy SPAs without a headless browser.

**Safety cap:** 100KB response limit.

**Errors:** HTTP 429 is handled internally with exponential backoff and retry. `FetchJobDescription` accepts a `context.Context` as its first parameter so the backoff loop can be interrupted by the caller's timeout (e.g. the 300s web deadline) or Ctrl+C. All other failures return the error directly; user can fall back to `--local`.

### `Parse` (parse.go)
Converts the markdown returned by `r.jina.ai` into a typed, filtered line-level AST.

Each non-empty line is classified as one of 15 `NodeType` constants ordered from most generic (`body`) to most specific (`jina_title`). Noise types (`jina_marker`, `setext_underline`, `nav_link`) are stripped; long low-information body lines (> 300 chars) are dropped to preserve LLM context window. Returns `[]JobDescriptionNode` — serialization to JSON is handled downstream in `generate.go`.

```go
// classifyLine returns the most specific NodeType for a single non-empty line,
// or "" to signal the line should be dropped.
func classifyLine(line string) string

// Parse returns the filtered AST: noise removed, long body lines dropped.
func Parse(s string) []JobDescriptionNode
```

### `LLM Client` (llm.go)
Pure HTTP interface — no prompt text or business logic. Contains the wire-format types and `InvokeDeepseekApi`.

*   **Wire types:** `deepseekRequest`, `deepseekResponse`, `deepseekMessage` (unexported). Request uses `stream: false`. No `response_format` field — plain text mode (see Generator section).
*   **Exponential Backoff:** `InvokeDeepseekApi` retries on HTTP 429 recursively. Non-429 errors return immediately.

```go
func InvokeDeepseekApi(ctx context.Context, apiKey string, c *http.Client, backoff int, requestBody json.RawMessage) (string, error)
```

### `Generator` (generate.go)
Pure LLM orchestration — no filesystem access. Contains the system prompt, tag-extraction helpers, and `GenerateAll`.

*   **Job description encoding:** `[]JobDescriptionNode` is serialized to minified JSON via `json.Marshal` and sent as the job description payload. The JSON keys (`"content"`, `"type"`) are set by the struct tags on `JobDescriptionNode`.

*   **Plain text output mode (not JSON mode):** The DeepSeek API supports a `response_format: {"type": "json_object"}` flag that constrains the model to emit valid JSON. This was dropped. Forcing JSON mode is known to degrade output quality — the model has to simultaneously reason about content *and* maintain JSON syntax, which competes for the same generation capacity. Plain text mode lets the model reason freely; we impose structure on the output ourselves via XML-like delimiter tags.

*   **Output format — XML delimiter tags:** The system prompt instructs the model to wrap each output field in a unique tag. Five compiled regexps with `(?s)` (dot-matches-newline) extract each section. Tag-based parsing is robust for this use case because plain text resumes and cover letters will never naturally contain strings like `<resume>`.

```text
<company>Acme Corp</company>
<role>Senior Copywriter</role>
<score>7</score>
<resume>
full tailored resume text
</resume>
<cover>
tailored cover letter (only present if base cover was provided)
</cover>
```

*   **Validation:** After extraction, `GenerateAll` errors immediately if `company`, `role`, or `resume` are empty, with a diagnostic message showing what was (and wasn't) parsed. This surfaces prompt compliance failures rather than silently writing empty files.

*   **Score:** Integer 1–10 subjective rating of how well the base resume matches the job requirements. Defaults to 0 on parse failure — non-fatal.
*   **Optional Cover Letter:** Pass `nil` for `baseCover` to skip. `<cover>` tag is omitted from the prompt instruction when no base cover is provided; the regexp only runs when `baseCover != nil`.
*   **Company/Role Extraction:** Extracted by the LLM from the JSON payload in the same call — no separate fallback prompt.

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
Pure filesystem primitives. No orchestration logic, no LLM calls.

```go
func currentDate() string  // returns YYYY-MM-DD via time.Now().Format

// slugify derives a folder name from parsed AST nodes: prefers jina_title
// (with "Title:" prefix stripped), falls back to job_title (with leading
// #/* stripped). Format: YYYY-MM-DD-{rand8}-{title-slug}.
// Returns bare date prefix if no title node is found.
func slugify(nodes []JobDescriptionNode) string

// createApplicationDirectory creates the run folder under App.Paths.Jobs.
// On os.ErrExist, appends "col" suffix and retries once.
func createApplicationDirectory(slug string, a *App) error

func fetchResume(a *App) (string, error)  // reads config/templates/resume.txt
func fetchCover(a *App) (string, error)   // reads config/templates/cover.txt
```

### `Process` (process.go)
Orchestrates the full pipeline: parse, load templates, call LLM, create directory, write files. Source-routing (URL vs local file vs stdin) is the CLI's concern — `Process` receives raw text directly.

```go
type applicationMeta struct {
    Company string `json:"company"`
    Role    string `json:"role"`
    Score   int    `json:"score"`
    Tokens  int    `json:"tokens"`
    Date    string `json:"date"`
}

func (a *App) Process(ctx context.Context, rawText string) (string, error)
```

**Process() workflow:**
1. Parse raw text: `Parse(rawText)` → `[]JobDescriptionNode`
2. Load templates: `fetchResume()` required (fail fast); `fetchCover()` optional (nil if absent)
3. Call `GenerateAll()` — the only expensive/fallible operation; no filesystem has been touched yet
4. Build folder name: `slugify(nodes)` using AST title nodes (format: `YYYY-MM-DD-{rand8}-{title-slug}`)
5. `createApplicationDirectory()` — on `os.ErrExist`, appends `"col"` suffix and retries
6. Write files sequentially: `resume.txt`, `cover.txt` (if cover returned), `meta.json`
7. Return the output directory path

**Failure behavior:** If any write after folder creation fails, the partial folder remains on disk for inspection. The random component in the slug means re-running the same source produces a new, unique folder.

### `Web Server` (web.go)

```go
func (a *App) Serve(ctx context.Context, port string, ui fs.FS) error
```

Accepts bare port number (e.g. `"8080"`), prepends `:` internally. Uses `http.Server.Shutdown(ctx)` for graceful shutdown — on context cancellation, in-flight requests (including long LLM calls) finish before the server exits.

**Endpoints:**
*   `POST /api/process` — accepts `{"url": "..."}` or `{"local_text": "..."}`, wraps `Process()` with `context.WithTimeout` (300s). Returns output directory path and metadata.
*   `GET /api/jobs` — lists jobs by reading `meta.json` from each run folder; query params for filtering/sorting. Tolerates corrupt `meta.json` entries (logs warning, skips). Note: the current `applicationMeta` struct will need extending with `status`, `url`, and date fields for full Phase 5 functionality.
*   `GET /api/jobs/{id}` — returns job metadata from `meta.json`. ID scheme and extended metadata type TBD for Phase 5.
*   `PATCH /api/jobs/{id}` — **writable fields: `status`, `date_applied` only.** Read-only fields rejected if present. Notes sidecar TBD for Phase 5.

**CSRF:** Reject requests where `Origin` header is present and does not match `http://localhost:{port}` or `http://127.0.0.1:{port}`. Requests without `Origin` (e.g. curl) pass through. Additionally, POST/PATCH endpoints require `Content-Type: application/json` to block simple form submissions.

## 6. User Interface

**CLI:**
```bash
$ jdextract setup
$ jdextract generate https://acme.com/job/123
  > Saved to: ./data/jobs/2026-02-24-a7x9k3m2-intermediate-copywriter/
  > Resume: resume.txt  |  Cover: cover.txt
  > Tokens used: 2847
$ jdextract generate --local ./my_job_paste.txt
$ jdextract list                                  # folder-prefix identification
  > FOLDER                                          COMPANY     ROLE                   STATUS
  > 2026-02-24-a7x9k3m2-intermediate-copywriter    Acme Corp   Intermediate Copywriter draft
$ jdextract status 2026-02-24-a7x9 applied        # folder prefix match
$ jdextract serve --port 9090                     # Default: 8080
```

`jdextract status` validates against `draft|applied|interviewing|offer|rejected` before writing.

**Web:** User pastes a Job URL (or raw text), clicks "Generate". UI shows the generated text, file paths, and token usage. Job list view with status badges, filtering, and sorting. Loading spinner with timeout-specific error if the server deadline is exceeded.