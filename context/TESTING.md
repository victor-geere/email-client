# Testing

## Strategy

- **Unit tests** for business logic: quote detection, linearization, reference generation, thread ordering.
- **Integration tests** for the Graph API client: use an HTTP test server returning fixture data.
- **No real API calls** in tests. Ever. Mock at the HTTP boundary.
- **No real credentials** in test code or fixtures.

## Test Naming

Describe the behavior, not the function:

```go
// Good
func TestLinearize_ReplacesNestedQuoteWithReference(t *testing.T) { ... }
func TestQuoteDetector_IdentifiesBlockquoteElements(t *testing.T) { ... }
func TestThreadBuilder_SortsMessagesByDate(t *testing.T) { ... }

// Bad
func TestLinearize(t *testing.T) { ... }
func TestProcess(t *testing.T) { ... }
```

## Structure

Each test follows Arrange → Act → Assert:

```go
func TestLinearize_ReplacesNestedQuoteWithReference(t *testing.T) {
    // Arrange: build a thread with a quoted reply
    thread := buildTestThread(...)

    // Act: linearize the thread
    result := linearize(thread)

    // Assert: quoted block replaced with reference
    assert(result.Messages[1].Body, contains("[→ see msg #1"))
}
```

## Test Data

- Use `testdata/` at the repo root for shared JSON fixtures (sample Graph API responses).
- Use builder/factory functions for constructing test messages — avoid copy-pasting large structs.
- Include fixtures for:
  - Simple 2-message thread
  - Deep thread (5+ replies with nested quotes)
  - Thread with HTML bodies and `<blockquote>` elements
  - Thread with plain-text quote markers (`>`)
  - Thread with mixed HTML and plain-text messages
  - Edge cases: empty body, missing subject, single-message "thread"

## Mocking the Graph API

- Use `net/http/httptest.Server` (or language equivalent) to serve fixture JSON.
- The Graph client should accept a base URL parameter so tests can point it at a local server.
- Return realistic status codes: 200, 401 (expired token), 429 (rate limited with `Retry-After`), 500.

## What to Test

| Module | Test Focus |
|--------|-----------|
| `auth` | Token caching, refresh logic, error on invalid credentials |
| `graph` | Pagination handling, retry on 429, mapping JSON to domain types |
| `thread` | Chronological sorting, conversation grouping, ordinal assignment |
| `quote` | Blockquote detection, line-prefix detection, header-pattern detection, content-hash matching |
| `linearize` | Reference formatting, trimming residual whitespace, multi-level nested quotes |
| `render` | Markdown output structure, filename generation, special character handling |
| `cmd` | Flag parsing, error exit codes (no end-to-end via real API) |

## Running Tests

```sh
go test ./...              # all tests
go test ./internal/quote/  # single package
go test -v -run QuoteDetector ./internal/quote/  # specific tests
```
