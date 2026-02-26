# Work Order: jdextract

## Phase 1: Core Infrastructure

### Setup
- [x] Set up project structure: `cmd/main.go`, `cmd/web/`, `jdextract/`
- [x] ~~Add `golang.org/x/net/html` dependency~~ — dropped; `r.jina.ai` returns markdown, parsed with stdlib `regexp`
- [x] Add `github.com/toon-format/toon-go` dependency (vendored via `go mod vendor`); used to serialize the parsed AST to TOON format for LLM consumption
- [x] `app.go`: `GetPortablePaths()` — resolve exe dir, follow symlinks, macOS `.app` bundle support
- [x] `app.go`: implement `NewApp()` and `Setup()` (creates `templates/`, `data/jobs/`, example templates)
- [x] `config.go`: parse `<exe_dir>/config/config.json` (JSON); env var override deferred to post-MVP1
- [x] `config.go`: create config file with `0600` permissions (contains API key); job files use `0644`

### Fetching
- [x] `fetch.go`: fetch via `r.jina.ai/{url}`; 100KB response cap
- [x] `fetch.go`: exponential backoff retry loop on HTTP 429, handled internally within `FetchJobDescription`; accepts `context.Context` to allow cancellation; all other failures return error directly

### Parsing
- [x] `parse.go`: line-level AST via `buildProtoAST()` — classifies each line into 15 `NodeType` constants (generic → specific); drops noise and long body lines
- [x] `parse.go`: `filterNodes()` removes always-drop types; `Parse()` returns the filtered `[]JobDescriptionNode`

---

## Phase 2: LLM Integration

- [x] `llm.go`: HTTP client for `api.deepseek.com` with auth and error handling | Just reuse App.Client
- [x] `llm.go`: exponential backoff on HTTP 429; non-429 errors return immediately | `InvokeDeepseekApi(ctx, apiKey, client, backoff, requestBody)` — same recursive pattern as fetch.go
- [x] `llm.go`: wire format types (`deepseekRequest`, `deepseekResponse`, `deepseekMessage`) — `response_format: {"type": "json_object"}`; no business logic in llm.go
- [x] `generate.go`: `GenerateAll()` — builds prompt, calls `InvokeDeepseekApi`, decodes outer API response then inner LLM JSON defensively via `map[string]interface{}`; `nil` baseCover skips cover letter

---

## Phase 3: Generation Pipeline

- [ ] `generate.go`: `slug()` — normalizes company/role strings for folder names ("Acme & Co." → "acme-co"); returns `"unknown"` when sanitized result is empty
- [ ] `generate.go`: define `JobInput` (URL / LocalFile / RawText — exactly one set), `JobResult`, `JobMetadata`
- [ ] `generate.go`: implement `Process()` following workflow in design.md section 5
- [ ] `generate.go`: all paths run through `Parse()` → TOON serialization (`toon-go`) → LLM prompt; TOON encoding is the last-mile step before the LLM call, not a parse.go concern
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
