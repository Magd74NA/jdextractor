# JDExtractor

A local-first tool that fetches job postings, parses them, and uses AI (DeepSeek or Kimi) to tailor your resume and cover letter to each role. Includes contact management and AI-powered networking follow-ups. Ships as a single self-contained binary with a modern Svelte web UI.

## How it works

1. Paste a job posting URL (or raw text) into the web UI
2. JDExtractor fetches and parses the job description via [jina.ai](https://jina.ai)
3. Your selected AI backend rewrites your resume and cover letter templates to match the role
4. Track applications, manage networking contacts, and generate AI follow-up messages — all from one place

## Prerequisites

- A [DeepSeek API key](https://platform.deepseek.com/) (pay-as-you-go, very cheap), or
- A [Kimi API key](https://platform.moonshot.cn/) (Moonshot AI, experimental support)

## Installation

Download the latest binary for your platform from the [Releases](https://github.com/Magd74NA/jdextractor/releases) page.

```bash
# Linux / macOS
tar -xzf jdextractor_linux_amd64.tar.gz
./jdextractor
```

```powershell
# Windows — extract the zip, then run:
.\jdextractor.exe
```

## Quick start

```bash
# First run: create the config and template directories
./jdextractor setup

# Add your resume and cover letter templates to:
#   config/templates/resume.txt
#   config/templates/cover.txt (optional)

# Start the web UI (default: http://localhost:8080)
./jdextractor serve          # or: serve --open to auto-launch browser
```

Then open `http://localhost:8080` and configure your AI backend in Settings.

## Web UI

The Svelte UI has five main sections:

- **Dashboard** — Application stats, activity charts, match score distribution, and an overdue follow-up queue
- **Jobs** — Filterable table of all applications with inline editing, file viewers, and token usage tracking
- **Process** — Submit job URLs (single or batch) or paste raw text; streaming output shows real-time progress
- **Contacts** — Manage networking contacts, log conversations, generate AI follow-up messages, and link contacts to jobs
- **Settings** — Backend selection (DeepSeek / Kimi), API key, model, templates, and networking prompt configuration

A unified search bar in the header searches across both jobs and contacts.

## CLI

```bash
# Process a URL
./jdextractor generate https://jobs.example.com/some-role

# Batch-process multiple URLs concurrently
./jdextractor generate --batch urls.txt

# Process a local file
./jdextractor generate --local path/to/job.txt

# List tracked applications
./jdextractor list

# Update application status (draft, applied, interviewing, offer, rejected)
./jdextractor status <dir-prefix> applied

# Contacts
./jdextractor contacts add --name "Jane Doe" --company Acme --role "Eng Manager"
./jdextractor contacts list
./jdextractor contacts log <id> --channel email --summary "Discussed the role"
./jdextractor contacts followup <id>          # AI-generated follow-up message
./jdextractor contacts overdue                # List contacts past their follow-up date
./jdextractor contacts link <contact-id> <job-dir>
./jdextractor contacts status <id> connected
./jdextractor contacts delete <id>
```

## Build from source

Requires Go 1.25+ and Node.js (for the UI).

```bash
git clone https://github.com/Magd74NA/jdextractor.git
cd jdextractor
make build        # builds UI assets, then the Go binary
./out/jdextractor
```

Other targets: `make clean`, `make run`, `make snapshot`, `make fmt`, `make vet`.

## Features

- **Modern Svelte UI** — Dashboard with activity charts, job statistics, and inline editing
- **Dual backend support** — DeepSeek (`deepseek-chat`, `deepseek-reasoner`) or Kimi K2.5 (experimental)
- **Contact management** — Track networking contacts with relationship status, tags, and conversation threads
- **AI follow-ups** — Generate context-aware follow-up messages with suggested timing and channel
- **PII sanitization** — Emails and phone numbers are redacted before sending data to the LLM
- **Unified search** — Search across jobs and contacts from a single search bar
- **Streaming processing** — Real-time progress updates as documents are generated
- **Batch processing** — Process multiple job URLs concurrently
- **Local-first** — All data stored relative to the binary; no cloud uploads or hidden config directories
- **Customizable prompts** — Edit system prompts and task lists for both resume generation and networking

## Notes

- URL fetching depends on [jina.ai](https://jina.ai) to convert web pages to clean text
- **DeepSeek**: `deepseek-chat` recommended for most cases; `deepseek-reasoner` for complex roles
- **Kimi**: K2.5 model is experimental and still being tested

## License

MIT
