# Architecture

## High-Level Data Flow

```
Microsoft 365 Mailbox
        │
        ▼
┌──────────────────┐
│   Auth Module     │  OAuth 2.0 device code flow
│   (token mgmt)   │  Token caching & refresh
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│   Graph Client    │  HTTP calls to Microsoft Graph API v1.0
│   (fetch threads) │  Pagination, rate limiting, retries
└──────┬───────────┘
       │  Raw email JSON (messages grouped by conversationId)
       ▼
┌──────────────────┐
│   Thread Builder  │  Groups messages into conversations
│                   │  Sorts chronologically within each thread
└──────┬───────────┘
       │  Ordered message list per thread
       ▼
┌──────────────────┐
│   Quote Detector  │  Identifies quoted/nested content in each message
│                   │  Matches quotes to earlier messages in the thread
└──────┬───────────┘
       │  Messages with quote regions annotated
       ▼
┌──────────────────┐
│   Linearizer      │  Replaces quoted blocks with human-readable refs
│                   │  Assigns ordinal IDs per thread (msg #1, #2, …)
└──────┬───────────┘
       │  Clean linear message sequence
       ▼
┌──────────────────┐
│   Renderer        │  Formats output as Markdown (or plain text)
│                   │  Writes one file per thread
└──────────────────┘
```

## Module Responsibilities

### `auth`
- Implements OAuth 2.0 device code flow against Microsoft Entra ID
- Caches tokens to disk (encrypted or OS keychain)
- Handles token refresh transparently
- Exposes a function that returns a valid access token for Graph API calls

### `graph`
- Wraps Microsoft Graph REST API v1.0
- Fetches messages by folder, handles `$top` / `$skip` / `@odata.nextLink` pagination
- Groups messages by `conversationId`
- Handles HTTP retries, 429 rate limiting (respects `Retry-After`)
- Returns structured message types — not raw JSON outside this module

### `thread`
- Takes a set of messages sharing a `conversationId`
- Sorts them chronologically by `receivedDateTime`
- Assigns each message an ordinal within the thread (1, 2, 3, …)
- Attaches metadata: sender, recipients, date, subject

### `quote`
- Parses email body content (HTML or plain text) to find quoted regions
- Detection strategies (combined, not exclusive):
  - Line-prefix patterns (`>`, `> >`, etc.)
  - `<blockquote>` elements and known email client quote wrappers
  - Header patterns ("On Jan 14, Alice wrote:", "From: … Sent: … Subject: …")
  - Content hashing — match quoted text against earlier messages in the thread
- Returns annotated message bodies with quote regions marked and attributed

### `linearize`
- Replaces each detected quote region with a reference: `[→ see msg #N from Author, DD Mon]`
- Trims trailing whitespace, signatures, and empty lines left by removal
- Ensures the resulting message body is coherent as standalone text

### `render`
- Formats a linearized thread into Markdown (primary) or plain text
- Writes output files to the specified directory
- Handles filename generation from thread subject (slugified, collision-safe)

### `cmd` (CLI entry point)
- Flag parsing (folder, thread ID, output dir, limit, format, verbosity)
- Orchestrates the pipeline: auth → fetch → build threads → detect quotes → linearize → render
- Exits with appropriate codes and error messages

## Planned Directory Layout

```
cmd/                  CLI entry point and flag parsing
internal/
  auth/               OAuth device code flow, token cache
  graph/              Microsoft Graph API client
  thread/             Thread building and ordering
  quote/              Quoted content detection and attribution
  linearize/          Quote replacement with references
  render/             Markdown and plain text formatters
context/              Project context docs (this folder)
testdata/             Sample email JSON fixtures for tests
```

## Key Design Decisions

- **Language**: Go is the leading candidate. Good for CLI tools, single binary distribution, strong standard library for HTTP and JSON. Final decision pending.
- **No database**: Threads are processed in-memory, streamed to disk. No persistence layer needed.
- **HTML and plain text**: Email bodies arrive in both formats. The quote detector should handle both. The renderer works from the linearized output, which is plain text with references.
- **Stateless processing**: Each run is independent. No incremental sync or state between runs (may change if needed).
