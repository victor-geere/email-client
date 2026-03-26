# SPEC.md — Email Thread Linearizer

> **This file is the editable "agentic source-code" for the project.**
> The `/build` command reads this spec, syncs it to `prompt-tree.json`, implements pending tasks, builds, and tests.
> Edit this file to drive development. Mark tasks `[x]` only when code is written, compiles, and tests pass.

---

## 1. Project Bootstrap

- [x] Initialize Go module (`go mod init github.com/victor/email-linearize`)
- [x] Create directory structure: `cmd/email-linearize/`, `internal/{auth,graph,thread,quote,linearize,render}/`, `testdata/`
- [x] Add `.gitignore` (Go binaries, token cache files, output dir)

## 2. Domain Types (`internal/domain/`)

- [x] `Message` struct: ID, ConversationID, Subject, From (EmailAddress), ToRecipients, CcRecipients, ReceivedDateTime, Body (content + contentType), BodyPreview
- [x] `EmailAddress` struct: Name, Address
- [x] `Thread` struct: ConversationID, Subject, Messages (ordered slice of Message)
- [x] `QuoteRegion` struct: StartOffset, EndOffset, AttributedToMessageOrdinal (nullable int), Original text snippet
- [x] `AnnotatedMessage` struct: embeds Message + Ordinal int + QuoteRegions slice + CleanBody string
- [x] `LinearizedThread` struct: ConversationID, Subject, Messages (slice of AnnotatedMessage)

## 3. Auth Module (`internal/auth/`)

- [x] `DeviceCodeFlow` function: initiates device code request to `https://login.microsoftonline.com/{tenant}/oauth2/v2.0/devicecode`
- [x] `PollForToken` function: polls token endpoint with device code until user approves or timeout
- [x] `TokenCache` interface: `Save(token)`, `Load() (token, error)`, `Clear()`
- [x] `KeychainCache` implementation: stores refresh token in OS keychain (macOS Keychain via `keychain` package)
- [x] `FileCache` fallback implementation: encrypted file with restrictive permissions (0600)
- [x] `TokenSource` function: returns valid access token — loads from cache, refreshes if expired, falls back to device code flow
- [x] Environment variables: `EMAIL_LINEAR_CLIENT_ID`, `EMAIL_LINEAR_TENANT_ID`
- [x] Scopes: `Mail.Read`, `offline_access`
- [x] Tests: token caching round-trip, refresh logic, expired token triggers re-auth prompt

## 4. Graph Client (`internal/graph/`)

- [x] `Client` struct with configurable base URL (for testing), HTTP client, token source
- [x] `ResolveFolderID(ctx, folderName) (string, error)` — calls `/me/mailFolders`, matches by displayName
- [x] `FetchMessages(ctx, folderID, options) ([]Message, error)` — paginated fetch via `@odata.nextLink`
- [x] `FetchThreadMessages(ctx, conversationID) ([]Message, error)` — filter by conversationId
- [x] Request: `$select` only needed fields, `$orderby=receivedDateTime asc`, `$top=50`
- [x] Response parsing: unmarshal into domain `Message` types within this module
- [x] Error handling: 401 → refresh + retry once; 429 → respect `Retry-After`; 500/503 → exponential backoff (max 3)
- [x] Never log response bodies
- [x] Tests: httptest.Server with fixture JSON; test pagination, 429 retry, 401 refresh

## 5. Thread Builder (`internal/thread/`)

- [x] `GroupByConversation(messages []Message) map[string][]Message` — group by ConversationID
- [x] `BuildThread(conversationID string, messages []Message) Thread` — sort by ReceivedDateTime, assign Subject from first message
- [x] `AssignOrdinals(thread Thread) []AnnotatedMessage` — number messages 1, 2, 3… within thread
- [x] Tests: sorting correctness, single-message thread, messages with identical timestamps

## 6. Quote Detector (`internal/quote/`)

- [x] `Detect(msg AnnotatedMessage, earlierMessages []AnnotatedMessage) []QuoteRegion`
- [x] Strategy: HTML blockquote detection — find `<blockquote>` elements, extract text
- [x] Strategy: Plain-text line-prefix detection — lines starting with `>` (nested: `> >`, `> > >`)
- [x] Strategy: Header pattern detection — "On {date}, {name} wrote:", "From: … Sent: … Subject: …"
- [x] Strategy: Content-hash matching — hash normalized text of earlier messages, compare against quote body
- [x] Attribution: match detected quote to the specific earlier message it came from (by ordinal)
- [x] Handle multi-level nesting (quote within a quote)
- [x] Strip HTML to extract text for comparison (use `golang.org/x/net/html` tokenizer)
- [x] Tests: blockquote HTML, plain-text `>` prefixes, Outlook-style headers, unattributable quotes, nested quotes

## 7. Linearizer (`internal/linearize/`)

- [x] `Linearize(thread Thread, annotatedMessages []AnnotatedMessage) LinearizedThread`
- [x] Replace each QuoteRegion with reference: `[→ see msg #N from {FirstName}, {DD Mon}]`
- [x] If quote is unattributed: `[→ quoted text omitted]`
- [x] Trim residual blank lines and trailing whitespace after replacement
- [x] Preserve the author's own text surrounding inline quotes
- [x] CleanBody field on each AnnotatedMessage holds the final result
- [x] Tests: single quote replacement, multiple quotes in one message, nested quotes, no-quote passthrough

## 8. Renderer (`internal/render/`)

- [x] `RenderMarkdown(thread LinearizedThread) string` — produce full Markdown document
- [x] Output format: heading with thread subject, then for each message: `### Message #N`, metadata block (From, To, Cc, Date, Subject), blank line, body
- [x] `RenderText(thread LinearizedThread) string` — plain-text variant, no Markdown formatting
- [x] `WriteFile(outputDir, thread LinearizedThread, format string) error` — write to disk
- [x] Filename: slugified subject, max 80 chars, collision-safe (append `-2`, `-3` if needed)
- [x] Tests: Markdown structure matches expected output, filename edge cases (special chars, long subjects, duplicates)

## 9. CLI Entry Point (`cmd/email-linearize/`)

- [x] `main.go`: flag parsing with `flag` package
- [x] Flags: `--folder` (default "Inbox"), `--thread`, `--output` (default "./output"), `--limit`, `--format` (default "markdown"), `--verbose`, `--auth-only`
- [x] Pipeline orchestration: auth → resolve folder → fetch messages → group into threads → for each thread: detect quotes → linearize → render
- [x] `--auth-only`: authenticate and print success, then exit
- [x] Error handling: print to stderr, exit code 1 on failure
- [x] Verbose mode: log metadata (folder name, thread count, message count per thread) — never email bodies
- [x] Tests: flag parsing, `--auth-only` exits cleanly

## 10. Test Fixtures (`testdata/`)

- [x] `simple_thread.json` — 2-message thread, one reply quoting the original
- [x] `deep_thread.json` — 5-message thread with nested quotes
- [x] `html_thread.json` — messages with HTML bodies and `<blockquote>` elements
- [x] `plaintext_thread.json` — messages with plain-text `>` quoted lines
- [x] `mixed_thread.json` — mix of HTML and plain-text messages
- [x] `edge_cases.json` — empty body, missing subject, single-message thread
- [x] `graph_folder_list.json` — sample `/me/mailFolders` response
- [x] `graph_paginated.json` + `graph_paginated_page2.json` — paginated message list

## 11. Build & Polish

- [x] `go build ./cmd/email-linearize` compiles without errors
- [x] `go test ./...` passes all tests
- [x] `go vet ./...` reports no issues
- [x] README.md with usage instructions, prerequisites (app registration), and example output
- [x] Verify no tokens, credentials, or email bodies appear in any log output
