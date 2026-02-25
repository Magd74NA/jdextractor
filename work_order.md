# Work Order: jdextract

## Phase 1: Foundation (No Dependencies)

### Core Infrastructure
- [ ] Learn `go/embed` - embed `web/` directory from `cmd/jdextract/main.go`; pass `fs.FS` into `App.Serve()` (embed directive must be in same package as the embedded directory)
- [ ] Learn `go/http` - basic HTTP client for fetching job posting URLs
- [ ] Learn `go/flag` - create CLI with subcommands using `flag.NewFlagSet` per command
- [ ] Set up project structure following design (cmd/, jdextract/, cmd/jdextract/web/)
- [ ] Add `golang.org/x/net/html` dependency for HTML parsing (zero external deps, stays in Go ecosystem)

### Configuration & Setup
- [ ] Implement `config.go` - resolve `~/.jdextract` paths
- [ ] Support both config.yaml AND `DEEPSEEK_API_KEY` env var
- [ ] **Precedence: env var > config.yaml > error** (12-factor app convention)
- [ ] Implement `app.go` - central App struct with `NewApp()` and `Setup()` methods
- [ ] Create directory structure: `~/.jdextract/`, `templates/`, `jobs/`
- [ ] Generate example templates (`resume.txt`, `cover.txt`) on first setup

### HTML Fetching (Hybrid Strategy)
- [ ] Implement `fetch.go` with hybrid fetch strategy:
  - [ ] First: Direct HTTP GET (works for static sites, Greenhouse, Lever)
  - [ ] Second: Try `https://r.jina.ai/http://URL` (extracts text from React SPAs)
  - [ ] Third: Fail gracefully with `--local` flag suggestion
- [ ] Add size caps: 500KB for raw HTML, 100KB for Jina markdown responses (markdown is denser)
- [ ] Add `--local <file>` flag — reads a saved text file; update error message to say "save text to a file and re-run with --local ./file.txt"

### HTML Parsing (No LLM)
- [ ] Implement `parse.go` with `golang.org/x/net/html`
- [ ] Extract company name from `<title>` and `<h1>` tags
- [ ] Extract role from common job board patterns
- [ ] Add regex patterns for Greenhouse, Lever, Workday, etc.
- [ ] Only fallback to LLM if HTML parsing completely fails
- [ ] Implement `slug()` helper: normalize extracted strings for safe folder names (`"Acme & Co." → "acme-co"`)

---

## Phase 2: LLM Integration (DeepSeek API)

### API Client
- [ ] Wire DeepSeek API in `llm.go` - HTTP client for `api.deepseek.com`
- [ ] Handle API authentication and error responses
- [ ] Implement exponential backoff retry for HTTP 429 (rate limit) responses

### Prompt Engineering
- [ ] Design batched prompt combining resume + cover letter; use `response_format: {"type": "json_object"}` (DeepSeek JSON mode) to get `{"resume": "...", "cover_letter": "..."}` — avoids fragile delimiter splitting
- [ ] Design fallback prompt for LLM-based company/role extraction (only when HTML parsing fails)
- [ ] Handle optional cover letter (skip section when `baseCover` is nil; omit `cover_letter` key in response schema)
- [ ] Draft system prompt + user prompt templates in design doc before implementing `llm.go`

### LLM Functions
- [ ] `GenerateAll(ctx context.Context, jobText, baseResume string, baseCover *string) (resume string, cover *string, tokensUsed int, err error)`
  - Pass `nil` for `baseCover` to skip cover letter generation
  - Returns `nil` for `cover` when not generated

---

## Phase 3: Generation Pipeline

### Data Structures
- [ ] Define `JobInput` struct with URL and LocalFile fields (mutually exclusive)
- [ ] Define `JobResult` struct with ID, Company, Role, TokensUsed, FolderPath, ResumePath, CoverPath
- [ ] Define `JobMetadata` struct for `job.json`:
  - `id` (UUID v4)
  - `company`, `role`, `status`
  - `date_created` (ISO 8601), `date_applied`
  - `url`, `tokens_used`, `notes_md`

### Data Structures (validation)
- [ ] Validate `JobInput` mutual exclusivity at the start of `Process()`:
  - Error if both `URL` and `LocalFile` are set
  - Error if neither is set

### Orchestration (`generate.go`)
- [ ] Implement `Process(ctx context.Context, input JobInput) (*JobResult, error)`
  - [ ] Validate `JobInput` (mutually exclusive fields)
  - [ ] Read job text from URL (hybrid fetch) or local file
  - [ ] Extract company name & role via golang.org/x/net/html (fallback to LLM)
  - [ ] Apply `slug()` to company and role before building folder name
  - [ ] Generate UUID v4 for the job
  - [ ] Generate URL hash suffix for collision prevention
  - [ ] Create run folder: `~/.jdextract/jobs/YYYY-MM-DD_slug-company_slug-role_URLHash/`
  - [ ] Save `job_raw.txt` (scraped/pasted content)
  - [ ] Call LLM batched generate (resume + optional cover letter)
  - [ ] Write `resume_custom.txt` and `cover_letter.txt` (if generated)
  - [ ] Create `job.json` with metadata including `id` and `date_created`
- [ ] Thread `context.Context` through `Process()`, `Fetch()`, and `GenerateAll()` from the start (retrofitting later is painful)

---

## Phase 4: CLI Interface

### Command Structure (`cmd/jdextract/main.go`)
- [ ] Use `flag.NewFlagSet` per subcommand (cleaner than global `flag.Parse`)
- [ ] `jdextract setup` - initialize ~/.jdextract directory and templates
- [ ] `jdextract resume <url>` - fetch job, generate resume + cover letter, save to disk
- [ ] `jdextract resume --local <file>` - process a saved text file instead of URL
- [ ] `jdextract resume --no-cover <url>` - skip cover letter generation
- [ ] `jdextract list` - print tabular job history (ID, date, company, role, status)
- [ ] `jdextract status <id|folder-name> <status>` - update status field in job.json
- [ ] `jdextract serve` - launch web UI

### User Experience
- [ ] Display folder path, resume path, cover path, and tokens used after generation
- [ ] Show "Skipping cover letter" message when template missing or `--no-cover`
- [ ] Add progress indicators for long-running operations
- [ ] Implement meaningful error messages:
  - [ ] "API key missing (set DEEPSEEK_API_KEY or add to config.yaml)"
  - [ ] "Fetch failed — save the job text to a file and re-run with --local ./file.txt"
- [ ] Add `--help` documentation for each command

---

## Phase 5: Web Interface

### Backend (`web.go`)
- [ ] `//go:embed web` lives in `cmd/jdextract/main.go`; `web.go` accepts `fs.FS` as a parameter
- [ ] Implement `Serve(port string, ui fs.FS)` - start HTTP server
- [ ] `POST /api/process` - accepts `{ "url": "..." }` or `{ "local_text": "..." }`, returns `JobResult`
- [ ] `GET /api/jobs` - lists all jobs with filtering/sorting by status, date, company
- [ ] `GET /api/jobs/{id}` - get job metadata by UUID
- [ ] `PATCH /api/jobs/{id}` - update job metadata (status, notes)
- [ ] Add Origin header validation for CSRF protection (even on localhost)

### Frontend (`static/index.html`)
- [ ] Create single HTML file with Alpine.js and Tailwind (CDN)
- [ ] Job URL input field with "Generate" button
- [ ] Optional: textarea for pasting job text directly (for JS-heavy sites)
- [ ] Display generated text in read-only text box
- [ ] Show file paths and token usage after generation
- [ ] Job list view with status badges (draft, applied, interviewing, offer, rejected)
- [ ] Filtering and sorting by status, company, date
- [ ] Loading states and error handling UI

---

## Future / Out of Scope for MVP1
- [ ] Markdown support for resume/cover letter templates
- [ ] PDF generation via Pandoc
- [ ] Keyword extraction and analysis dashboard
- [ ] Multiple resume templates
- [ ] Batch processing multiple URLs
- [ ] Integration with job boards APIs (LinkedIn, Indeed)