# Work Order: jdextract

## Phase 1: Core Infrastructure

### Setup
- [x] Set up project structure: `cmd/main.go`, `cmd/web/`, `jdextract/`
- [x] ~~Add `golang.org/x/net/html` dependency~~ — dropped; `r.jina.ai` returns markdown, parsed with stdlib `regexp`
- [x] ~~Add `github.com/toon-format/toon-go` dependency~~ — dropped; AST serialized to minified JSON via `encoding/json`; zero external dependencies
- [x] `app.go`: `GetPortablePaths()` — resolve exe dir, follow symlinks, macOS `.app` bundle support
- [x] `app.go`: `NewApp()` — constructor only; types `App` and `PortablePaths`
- [x] `setup.go`: `Setup()` and `createExampleTemplates()` — creates `templates/`, `data/jobs/`, example templates; split from `app.go`
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
- [x] `llm.go`: wire format types (`deepseekRequest`, `deepseekResponse`, `deepseekMessage`); no `response_format` field — plain text mode; no business logic in llm.go
- [x] `generate.go`: `GenerateAll()` — serializes `[]JobDescriptionNode` to minified JSON, builds batched prompt, calls `InvokeDeepseekApi`; parses response via XML delimiter tags (`<company>`, `<role>`, `<score>`, `<resume>`, `<cover>`); returns `company, role, resume, cover, score, tokensUsed`; errors if required fields are empty

---

## Phase 3: Generation Pipeline

- [x] `storage.go`: pure FS primitives — `slugify()`, `createApplicationDirectory()`, `fetchResume()`, `fetchCover()`; no orchestration logic
- [x] `generate.go`: pure LLM orchestration only — no filesystem access; seam with storage.go is `[]JobDescriptionNode` + plain strings in, plain strings out
- [x] `process.go`: `(a *App) Process(ctx, rawText string) (string, error)` — returns output dir; parse → load templates → `GenerateAll` → create dir → write files; source routing is CLI concern
- [x] `storage.go`: `currentDate()` — returns `YYYY-MM-DD` via `time.Now().Format`
- [x] `storage.go`: `slugify()` uses `currentDate()` as prefix; current format: `YYYY-MM-DD-{rand8}-{title-slug}`
- [x] `process.go`: `applicationMeta` has `date` field; `currentDate()` called and stored in metadata
- [ ] `storage.go`: complete folder naming — change `slugify(nodes)` to `slugify(company, role, rawText string)`; replace `{rand8}` midfix with first 4 hex chars of SHA-256 of `rawText`; replace `{title-slug}` with `{company-slug}_{role-slug}`; final format: `YYYY-MM-DD_{company-slug}_{role-slug}_{hash}`; error on `ErrExist`
- [ ] `process.go`: pass `company`, `role`, `rawText` into `slugify`; write files: `job_raw.txt`, `resume_custom.txt`, `cover_letter.txt` (if cover), `job.json` atomically (`.tmp` → `os.Rename`)

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
