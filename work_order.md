# Work Order: jdextract

## Phase 1: Foundation (No Dependencies)

### Core Infrastructure
- [ ] Learn `go/embed` - embed and serve static files (for `static/index.html`)
- [ ] Learn `go/http` - basic HTTP client for fetching job posting URLs
- [ ] Learn `go/flag` - create CLI with commands: `setup`, `resume`, `serve`
- [ ] Set up project structure following design (cmd/, jdextract/, static/)
- [ ] Add `goquery` dependency for HTML parsing

### Configuration & Setup
- [ ] Implement `config.go` - resolve `~/.jdextract` paths
- [ ] Support both config.yaml AND `DEEPSEEK_API_KEY` env var (config.yaml takes precedence)
- [ ] Implement `app.go` - central App struct with `NewApp()` and `Setup()` methods
- [ ] Create directory structure: `~/.jdextract/`, `templates/`, `jobs/`
- [ ] Generate example templates (`resume.txt`, `cover.txt`) on first setup

### HTML Fetching (Hybrid Strategy)
- [ ] Implement `fetch.go` with hybrid fetch strategy:
  - [ ] First: Direct HTTP GET (works for static sites, Greenhouse, Lever)
  - [ ] Second: Try `https://r.jina.ai/http://URL` (extracts text from React SPAs)
  - [ ] Third: Fail gracefully with `--local` flag suggestion
- [ ] Add 500KB cap on HTML fetch to prevent downloading massive SPAs
- [ ] Add `--local` flag for manual text file input

### HTML Parsing (No LLM)
- [ ] Implement `parse.go` with goquery
- [ ] Extract company name from `<title>` and `<h1>` tags
- [ ] Extract role from common job board patterns
- [ ] Add regex patterns for Greenhouse, Lever, Workday, etc.
- [ ] Only fallback to LLM if HTML parsing completely fails

---

## Phase 2: LLM Integration (DeepSeek API)

### API Client
- [ ] Wire DeepSeek API in `llm.go` - HTTP client for `api.deepseek.com`
- [ ] Handle API authentication and error responses
- [ ] Implement exponential backoff retry for HTTP 429 (rate limit) responses

### Prompt Engineering
- [ ] Design batched prompt combining resume + cover letter with `--- COVER LETTER ---` delimiter
- [ ] Design fallback prompt for LLM-based company/role extraction (only when HTML parsing fails)
- [ ] Create prompt templates for consistent output formatting

### LLM Functions
- [ ] `GenerateAll(jobText, baseResume, baseCover string) (resume, cover string, tokensUsed int, error)` - batched call (~40% cost savings)

---

## Phase 3: Generation Pipeline

### Data Structures
- [ ] Define `JobResult` struct with Company, Role, TokensUsed, FolderPath, ResumePath, CoverPath
- [ ] Define `JobMetadata` struct for `job.json` (company, role, status, date_applied, url, tokens_used, notes_md)

### Orchestration (`generate.go`)
- [ ] Implement `ProcessJob(url string) (*JobResult, error)`
  - [ ] Fetch URL HTML (with hybrid strategy)
  - [ ] Extract company name & role via goquery (fallback to LLM)
  - [ ] Generate URL hash suffix for collision prevention
  - [ ] Create run folder: `~/.jdextract/jobs/YYYY-MM-DD_Company_Role_URLHash/`
  - [ ] Save `job_raw.txt` (scraped content)
  - [ ] Call LLM batched generate (resume + cover letter)
  - [ ] Write `resume_custom.txt` and `cover_letter.txt`
  - [ ] Create `job.json` with metadata

---

## Phase 4: CLI Interface

### Command Structure (`cmd/jdextract/main.go`)
- [ ] `jdextract setup` - initialize ~/.jdextract directory and templates
- [ ] `jdextract resume <url>` - fetch job, generate resume + cover letter, save to disk
- [ ] `jdextract resume --local <file>` - process local text file instead of URL
- [ ] `jdextract serve` - launch web UI

### User Experience
- [ ] Display folder path, resume path, cover path, and tokens used after generation
- [ ] Add progress indicators for long-running operations
- [ ] Implement meaningful error messages:
  - [ ] "API key missing (set DEEPSEEK_API_KEY or add to config.yaml)"
  - [ ] "Fetch failed, try --local flag with pasted text"
- [ ] Add `--help` documentation for each command

---

## Phase 5: Web Interface

### Backend (`web.go`)
- [ ] Embed static files with `//go:embed static/*`
- [ ] Implement `Serve(port string)` - start HTTP server
- [ ] Create `/api/process` endpoint - calls `ProcessJob()`, returns JSON `JobResult`
- [ ] Create `/api/jobs` endpoint - lists all jobs with filtering/sorting
- [ ] Create `/api/jobs/{id}` endpoint - get/update job metadata (status, notes)
- [ ] Add Origin header validation for CSRF protection (even on localhost)

### Frontend (`static/index.html`)
- [ ] Create single HTML file with Alpine.js and Tailwind (CDN)
- [ ] Job URL input field with "Generate" button
- [ ] Display generated text in read-only text box
- [ ] Show file paths and token usage after generation
- [ ] Job list view with status badges (draft, applied, interviewing)
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