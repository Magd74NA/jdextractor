# JDExtractor

A local-first tool that fetches a job posting URL, parses it, and uses AI (DeepSeek or Kimi) to tailor your resume and cover letter to that specific role. Ships as a single self-contained binary with a modern Svelte-based web UI.

## How it works

1. Paste a job posting URL (or raw text) into the web UI
2. JDExtractor fetches and parses the job description via [jina.ai](https://jina.ai)
3. Your selected AI backend rewrites your resume and cover letter templates to match the role
4. The tailored documents are saved locally and tracked with activity charts and statistics
5. Edit job metadata directly in the UI and monitor application progress

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
./jdextractor serve
```

Then open `http://localhost:8080` and configure your AI backend:
- **Configuration tab**: Choose DeepSeek or Kimi (experimental), enter your API key, and optionally customize the system prompt and task list
- **Jobs tab**: View all processed applications with activity charts and statistics
- **Process tab**: Submit new job URLs or raw job descriptions for processing

Stream processing shows real-time progress as your resume and cover letter are being tailored.

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

## Features

- **Modern Svelte UI**: Interactive dashboard with activity charts, job statistics, and inline editing
- **Dual backend support**: Choose between DeepSeek (recommended) or Kimi (experimental)
- **Streaming processing**: Real-time progress updates as your documents are being tailored
- **Local-first**: All data stored relative to binary location — no cloud uploads or hidden config directories
- **Customizable prompts**: Edit system prompts and task lists directly in the UI
- **Batch processing**: Process multiple job URLs in a single operation
- **Job tracking**: Monitor application status, match scores, and token usage

## Notes

- URL fetching depends on [jina.ai](https://jina.ai) to convert web pages to clean text
- **DeepSeek**: Use `deepseek-chat` for most cases; `deepseek-reasoner` available for complex roles
- **Kimi**: K2.5 model is experimental and still being tested

## License

MIT
