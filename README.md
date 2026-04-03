# Email Thread Linearizer

CLI tool that reads multithreaded emails from a Microsoft 365 account and converts them into a linear dialogue — without duplicating emails nested in threads. Quoted content is replaced with human-readable references to earlier messages.

## Quick Start

```sh
git clone <repo-url> && cd email-client
cp .env.example .env   # fill in your credentials
./scripts/build.sh
./scripts/authenticate.sh
./email-linearize --format html
```

## Prerequisites

- **Go 1.21+**
- A **Microsoft Entra ID** (Azure AD) app registration with `Mail.Read` and `offline_access` permissions
- Environment variables (set in `.env` or export in shell):

| Variable | Required | Description |
|---|---|---|
| `EMAIL_LINEAR_CLIENT_ID` | Yes | Application (client) ID |
| `EMAIL_LINEAR_TENANT_ID` | No | Tenant ID (defaults to `common`) |
| `EMAIL_LINEAR_CLIENT_SECRET` | No | Client secret (if configured) |

## Clone & Build

```sh
git clone <repo-url>
cd email-client
go build -o email-linearize ./cmd/email-linearize
```

Or use the build script (runs tests first):

```sh
./scripts/build.sh
```

## Authentication

The tool uses OAuth 2.0 **device code flow**. Run:

```sh
./scripts/authenticate.sh
```

1. A code is displayed in the terminal
2. Open https://microsoft.com/devicelogin in your browser
3. Enter the code and sign in (MFA is supported)
4. The token is cached in your OS keychain

To verify credentials without fetching mail:

```sh
./email-linearize --auth-only
```

## Usage

```sh
./email-linearize [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--folder` | `Inbox` | Mail folder to read |
| `--thread` | | Specific conversation ID to process |
| `--output` | `./output` | Output directory |
| `--limit` | `0` | Maximum messages to fetch (0 = unlimited) |
| `--format` | `markdown` | Output format: `markdown`, `text`, or `html` |
| `--verbose` | `false` | Enable verbose logging |
| `--auth-only` | `false` | Authenticate and exit |
| `--latest` | `false` | Only fetch messages newer than the latest file in the output directory |

### Output Formats

| Format | Description |
|---|---|
| `markdown` | `.md` files with hyperlinked references and attachments |
| `text` | Plain `.txt` files |
| `html` | Converts markdown to styled HTML with an index page, navigation banner, and accordion views. Report emails are routed to subdirectories. |

### Examples

```sh
# Fetch all threads as styled HTML (recommended)
./email-linearize --format html

# Incremental fetch — only new messages since last run
./email-linearize --format html --latest

# Plain markdown output
./email-linearize --format markdown

# Specific thread as text
./email-linearize --thread "AAQkADc..." --format text

# Verbose logging
./email-linearize --format html --verbose
```

## HTML Output

When `--format html` is used, the output directory contains:

```
output/
├── index.html                     ← Main index (accordion by day)
├── style.css                      ← Shared stylesheet
├── *.html                         ← Individual thread pages
├── attachments/                   ← Saved email attachments
├── daily-activity-reports/
│   ├── index.html                 ← Sub-index (accordion by month)
│   └── *.html
└── sprint-completion-reports/
    ├── index.html                 ← Sub-index (accordion by month)
    └── *.html
```

Every HTML page includes a **navigation banner** with links to Home, Sprint Reports, and Activity Reports.

Report emails (daily activity reports, sprint completion reports) are automatically routed to their respective subdirectories.

## Topic Summary

Generate a summary and index page for emails matching a specific topic:

```sh
./email-linearize summary --topic <topic> [--output <dir>]
```

| Flag | Default | Description |
|---|---|---|
| `--topic` | *(required)* | Topic to search for (case-insensitive match in titles and filenames) |
| `--output` | `./output` | Output directory containing existing HTML files |

This produces two files:

- **`<topic>-index.html`** — Flat list of all matching threads with dates
- **`<topic>.html`** — Summary page with threads grouped by month, including dates, snippets, and hyperlinks to the original thread pages

Example:

```sh
# Summarize all emails about "axiom"
./email-linearize summary --topic "axiom"

# Summarize emails about "roadmap"
./email-linearize summary --topic "roadmap"
```

> **Note:** The output directory must already contain HTML files. Run `./email-linearize --format html` first.

### Claude Skill (VS Code / Claude Code)

A skill is included at `.claude/skills/summary/` and a prompt at `.github/prompts/summary.prompt.md`. In VS Code Copilot Chat or Claude Code, type:

```
/summary <topic>
```

This invokes the `email-linearize summary` subcommand and reports the results.

## Scripts

| Script | Description |
|---|---|
| `scripts/build.sh` | Run tests and build the binary |
| `scripts/authenticate.sh` | Run OAuth device code flow |
| `scripts/fetch-mail.sh` | Build (if needed) and run with any flags passed through |

## Project Structure

```
cmd/email-linearize/   CLI entry point
internal/
  auth/                OAuth 2.0 device code flow + token caching
  domain/              Shared types (Message, Thread, QuoteRegion, etc.)
  graph/               Microsoft Graph API client
  thread/              Thread grouping and ordinal assignment
  quote/               Multi-strategy quote detection
  linearize/           Quote replacement with references
  render/              Output rendering, HTML conversion, file writing
context/               Project context documents
testdata/              Test fixtures (JSON)
scripts/               Build, auth, and fetch scripts
.claude/skills/        Claude Code skills (topic summary)
.github/prompts/       VS Code Copilot prompt files
.github/instructions/  File-scoped coding instructions
```

## How It Works

1. **Fetch** — Messages are retrieved from Microsoft Graph API (Inbox + Sent Items) with optional date filtering
2. **Group** — Messages are grouped by conversation ID into threads
3. **Detect** — Quoted content is identified using header patterns and content hashing
4. **Linearize** — Quoted regions are replaced with references like `[→ quoted text omitted. see message #2 from Alice, 15 Jan]`
5. **Render** — Threads are written as markdown, text, or HTML with attachments saved separately
6. **Index** — HTML format generates index pages with accordion navigation and a shared stylesheet

## Testing

```sh
go test ./...
```

All tests mock the Microsoft Graph API at the HTTP boundary — no real API calls are made.

## Security

- OAuth tokens stored in OS keychain (macOS Keychain, Linux secret-service), never in plaintext
- Email HTML stripped to text — remote resources are never fetched
- Email body content is never logged, even in verbose mode
- HTTPS enforced for all API calls; TLS validation is never disabled
