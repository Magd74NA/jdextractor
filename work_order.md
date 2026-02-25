# Work Order: jdextract

## Phase 1: Foundation (No Dependencies)

### Core Infrastructure
- [ ] Learn `go/embed` - embed and serve static files (for `static/index.html`)
- [ ] Learn `go/http` - basic HTTP client for fetching job posting URLs
- [ ] Learn `go/flag` - create CLI with commands: `setup`, `resume`, `cover`, `serve`
- [ ] Set up project structure following design (cmd/, jdextract/, static/)

### Configuration & Setup
- [ ] Implement `config.go` - resolve `~/.jdextract` paths, load `DEEPSEEK_API_KEY` from env
- [ ] Implement `app.go` - central App struct with `NewApp()` and `Setup()` methods
- [ ] Create directory structure: `~/.jdextract/`, `templates/`, `jobs/`
- [ ] Generate example templates (`resume.txt`, `cover.txt`) on first setup

### HTML Fetching & Parsing
- [ ] Implement `fetch.go` - simple HTTP GET to pull raw job HTML
- [ ] Parse HTML to strip unnecessary elements (HEAD, footer, scripts, styles)
- [ ] Extract clean job text for LLM processing

---

## Phase 2: LLM Integration (DeepSeek API)

### API Client
- [ ] Wire DeepSeek API in `llm.go` - HTTP client for `api.deepseek.com`
- [ ] Handle API authentication and error responses
- [ ] Implement retry logic for transient failures

### Prompt Engineering
- [ ] Design system prompt for resume customization (preserve structure, modify bullets)
- [ ] Design prompt for extracting company name and role from job HTML
- [ ] Design prompt for cover letter generation
- [ ] Create prompt templates for consistent output formatting

### LLM Functions
- [ ] `CustomizeResume(jobText, baseResume string) (string, error)` - tailor resume to job
- [ ] `ExtractJobMetadata(jobText string) (company, role string, error)` - for folder naming
- [ ] `GenerateCoverLetter(jobText, baseCover string) (string, error)` - draft cover letter

---

## Phase 3: Generation Pipeline

### Orchestration (`generate.go`)
- [ ] Implement `ProcessJob(url string) (folderPath string, err error)`
  - [ ] Fetch URL HTML
  - [ ] Extract company name & role via LLM
  - [ ] Create run folder: `~/.jdextract/jobs/YYYY-MM-DD_Company_Role/`
  - [ ] Save `job_raw.txt` (scraped webpage)
  - [ ] Call LLM to customize resume
  - [ ] Write `resume_custom.txt`
  - [ ] Generate and write `cover_letter.txt`

---

## Phase 4: CLI Interface

### Command Structure (`cmd/jdextract/main.go`)
- [ ] `jdextract setup` - initialize ~/.jdextract directory and templates
- [ ] `jdextract resume <url>` - fetch job, customize resume, save to disk
- [ ] `jdextract cover <url>` - fetch job, generate cover letter, save to disk
- [ ] `jdextract serve` - launch web UI

### User Experience
- [ ] Add progress indicators for long-running operations
- [ ] Implement meaningful error messages ("API key missing")
- [ ] Add `--help` documentation for each command

---

## Phase 5: Web Interface

### Backend (`web.go`)
- [ ] Embed static files with `//go:embed static/*`
- [ ] Implement `Serve(port string)` - start HTTP server
- [ ] Create `/api/process` endpoint - calls `ProcessJob()`
- [ ] Create `/api/cover` endpoint - calls cover letter generation
- [ ] Add CORS headers for local development

### Frontend (`static/index.html`)
- [ ] Create single HTML file with Alpine.js and Tailwind (CDN)
- [ ] Job URL input field with "Generate" button
- [ ] Display generated text in read-only text box
- [ ] Show file path after generation
- [ ] "Generate Cover Letter" button
- [ ] Loading states and error handling UI

---

## Future / Out of Scope for MVP1
- [ ] Markdown support for resume/cover letter templates
- [ ] PDF generation via Pandoc
- [ ] Keyword extraction and analysis dashboard
- [ ] Multiple resume templates
- [ ] Job application status tracking
- [ ] Batch processing multiple URLs
- [ ] Integration with job boards APIs (LinkedIn, Indeed)