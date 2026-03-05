# JDExtractor

A local-first tool that fetches a job posting URL, parses it, and uses DeepSeek AI to tailor your resume and cover letter to that specific role. Ships as a single self-contained binary with a built-in web UI.

## How it works

1. Paste a job posting URL (or raw text) into the web UI
2. JDExtractor fetches and parses the job description via [jina.ai](https://jina.ai)
3. DeepSeek rewrites your resume and cover letter templates to match the role
4. The tailored documents are saved locally and tracked in the Applications table

## Prerequisites

- A [DeepSeek API key](https://platform.deepseek.com/) (pay-as-you-go, very cheap)

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
./jdextractor serve
```

Then open `http://localhost:8080`, enter your DeepSeek API key in Configuration, and paste a job URL.

You can also use the CLI directly:

```bash
# Process a URL
./jdextractor generate https://jobs.example.com/some-role

# Process a local file
./jdextractor generate --local path/to/job.txt

# List tracked applications
./jdextractor list

# Update application status
./jdextractor status <dir-prefix> applied
```

## Build from source

Requires Go 1.25+.

```bash
git clone https://github.com/Magd74NA/jdextractor.git
cd jdextractor
make build
./out/jdextractor
```

## Notes

- All data (config, templates, applications) is stored relative to the binary location — no hidden config directories
- URL fetching depends on [jina.ai](https://jina.ai) to convert web pages to clean text
- The `deepseek-chat` model is recommended for most use cases; `deepseek-reasoner` is available for harder roles

## License

MIT
