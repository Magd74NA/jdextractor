# Work Order: jdextract

## Phase 1: Core Infrastructure

### Setup
- [x] Set up project structure: `cmd/main.go`, `cmd/web/`, `jdextract/`
- [x] Add `golang.org/x/net/html` dependency
- [x] `app.go`: `GetPortablePaths()` — resolve exe dir, follow symlinks, macOS `.app` bundle support
- [-] `app.go`: implement `NewApp()` and `Setup()` (creates `templates/`, `data/jobs/`, example templates)
- [ ] `config.go`: parse `<exe_dir>/config` (KEY=VALUE, `#` comments, env var > file > default)
- [ ] `config.go`: create config file with `0600` permissions (contains API key); job files use `0644`

### Fetching
- [ ] `fetch.go`: hybrid strategy — direct GET, then `r.jina.ai/{url}`, then error with `--local` hint
- [ ] `fetch.go`: size caps (500KB raw HTML, 100KB Jina)
- [ ] `fetch.go`: distinct errors — `ErrJinaRateLimited` (429, suggest retry) vs `ErrJinaExtraction` (suggest `--local`)

### Parsing
- [ ] `parse.go`: extract company/role from `<title>` and `<h1>` via `golang.org/x/net/html`
- [ ] `parse.go`: regex patterns for common job board formats
- [ ] `parse.go`: `slug()` — returns `"unknown"` when sanitized result is empty

---

## Phase 2: LLM Integration

- [ ] `llm.go`: HTTP client for `api.deepseek.com` with auth and error handling
- [ ] `llm.go`: exponential backoff on HTTP 429; non-429 errors return immediately
- [ ] `llm.go`: batched prompt using `response_format: {"type": "json_object"}`; decode into raw map first, check key existence before struct assignment
- [ ] `llm.go`: fallback prompt for company/role extraction (when HTML parsing fails)
- [ ] `llm.go`: implement `GenerateAll()` — pass `nil` for `baseCover` to skip cover letter (see design.md section 5)

---

## Phase 3: Generation Pipeline

- [ ] `generate.go`: define `JobInput` (URL / LocalFile / RawText — exactly one set), `JobResult`, `JobMetadata`
- [ ] `generate.go`: implement `Process()` following workflow in design.md section 5
- [ ] `generate.go`: URL path uses HTML parse then LLM fallback; LocalFile/RawText skip directly to LLM extraction
- [ ] `generate.go`: accept `context.Context` throughout (web callers set 300s timeout)
- [ ] `generate.go`: atomic write for `job.json` (write `.tmp`, then `os.Rename`)
- [ ] `generate.go`: on partial failure, leave folder on disk; error if folder already exists on re-run

---

## Phase 4: CLI Interface

- [ ] `cmd/main.go`: `flag.NewFlagSet` per subcommand
- [ ] `cmd/main.go`: root context via `signal.NotifyContext` (os.Interrupt, SIGTERM)
- [ ] `jdextract setup` — initialize portable directory structure and example templates
- [ ] `jdextract generate <url>` — fetch, generate, save; display paths and token count
- [ ] `jdextract generate --local <file>` — process saved text file
- [ ] `jdextract generate --no-cover` — skip cover letter (default: generate if `templates/cover.txt` exists)
- [ ] `jdextract list` — tabular output via `text/tabwriter`; UUID truncated to 8 chars; skip corrupt `job.json` with warning
- [ ] `jdextract status <prefix> <status>` — UUID prefix match, validate against `draft|applied|interviewing|offer|rejected`
- [ ] `jdextract serve --port <port>` — pass root context into `Serve()` for graceful shutdown (default 8080)
- [ ] Error messages: missing API key, fetch failure with `--local` hint

---

## Phase 5: Web Interface

### Backend (`web.go`)
- [ ] `Serve(ctx, port, ui)` — accepts context for graceful shutdown via `http.Server.Shutdown(ctx)`
- [ ] `POST /api/process` — wraps `Process()` with `context.WithTimeout` (300s)
- [ ] `GET /api/jobs` — list with filtering/sorting; tolerates corrupt `job.json` (log + skip)
- [ ] `GET /api/jobs/{id}` — returns `JobDetail` (job.json merged with notes.md)
- [ ] `PATCH /api/jobs/{id}` — writable: `status`, `date_applied`, `notes` only; reject read-only fields
- [ ] CSRF: reject when `Origin` present and doesn't match `localhost:{port}`; require `Content-Type: application/json` on POST/PATCH

### Frontend (`cmd/web/index.html`)
- [ ] Single HTML file: Alpine.js + Tailwind + DaisyUI (all CDN)
- [ ] Job URL input + textarea for raw text paste + "Generate" button
- [ ] Generated text display with file paths and token usage
- [ ] Job list with status badges, filtering, sorting
- [ ] Loading spinner with timeout-specific error message

---

## Future / Out of Scope for MVP1
- [ ] Markdown support for templates
- [ ] PDF generation via Pandoc
- [ ] Keyword extraction dashboard
- [ ] Multiple resume templates
- [ ] Batch URL processing
- [ ] Job board API integration (LinkedIn, Indeed)
