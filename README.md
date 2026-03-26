# Email Thread Linearizer

CLI tool that reads multithreaded emails from a Microsoft 365 account and converts them into a linear dialogue in Markdown format — without duplicating emails nested in threads.

## Prerequisites

- Go 1.21+
- A Microsoft Entra ID (Azure AD) app registration with `Mail.Read` and `offline_access` permissions
- Environment variables:
  - `EMAIL_LINEAR_CLIENT_ID` — Application (client) ID
  - `EMAIL_LINEAR_TENANT_ID` — Tenant ID (defaults to `common` if unset)

## Build

```sh
./scripts/build.sh
```

Or manually:

```sh
go build -o email-linearize ./cmd/email-linearize
```

## Usage

### Authenticate

```sh
./scripts/authenticate.sh
```

This runs the OAuth 2.0 device code flow. Follow the on-screen instructions to sign in via browser. The token is cached in your OS keychain.

### Fetch and linearize mail

```sh
./scripts/fetch-mail.sh
```

Or run the binary directly with options:

```sh
email-linearize [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--folder` | `Inbox` | Mail folder to read |
| `--thread` | | Specific conversation ID to process |
| `--output` | `./output` | Output directory |
| `--limit` | `0` | Maximum messages to fetch (0 = unlimited) |
| `--format` | `markdown` | Output format: `markdown` or `text` |
| `--verbose` | `false` | Enable verbose logging |
| `--auth-only` | `false` | Authenticate and exit |

### Examples

```sh
# Linearize all threads in Inbox
email-linearize

# Linearize a specific thread as plain text
email-linearize --thread "AAQkADc..." --format text

# Fetch from a different folder, verbose output
email-linearize --folder "Sent Items" --verbose --output ./threads

# Authenticate only (verify credentials work)
email-linearize --auth-only
```

## Output

Each thread produces a Markdown (or text) file in the output directory with:

- Thread subject as heading
- Messages in chronological order with sender, recipients, and timestamp
- Quoted content replaced with human-readable references like `[→ see msg #2 from Alice, 15 Jan]`

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
  render/              Markdown/text output and file writing
context/               Project context documents
testdata/              Test fixtures (JSON)
scripts/               Build, auth, and fetch scripts
docs/                  Additional documentation
```

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
