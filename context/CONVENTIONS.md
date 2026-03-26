# Conventions

## Code Style

- Clear, descriptive names. No abbreviations except well-known ones (`id`, `url`, `html`, `msg`).
- Small, focused functions. One function = one thing.
- Early returns to reduce nesting.
- Files under 300 lines. Extract into a new file when growing beyond that.
- Inline comments explain *why*, not *what*.

## Go-Specific (if Go is chosen)

- Follow standard `gofmt` / `goimports` formatting.
- Use `errors.New` / `fmt.Errorf` with `%w` for error wrapping.
- Prefer returning errors over panicking. Reserve `panic` for truly unrecoverable bugs.
- Exported names get doc comments. Unexported helpers don't need them unless non-obvious.
- Use `context.Context` for cancellation and timeouts in HTTP and long-running operations.
- Interfaces should be small (1–3 methods). Define them where they're consumed, not where they're implemented.
- Struct field tags: `json:"fieldName"` for API responses, nothing unnecessary.

## Naming

| Concept | Convention | Example |
|---------|-----------|---------|
| CLI binary | lowercase, hyphenated | `email-linearize` |
| Go packages | short, singular, lowercase | `auth`, `graph`, `thread`, `quote` |
| Types | PascalCase | `Message`, `Thread`, `QuoteRegion` |
| Functions | PascalCase (exported), camelCase (unexported) | `FetchMessages`, `parseQuoteBlock` |
| Constants | PascalCase or ALL_CAPS for env var names | `DefaultPageSize`, `ENV_CLIENT_ID` |
| Test files | `*_test.go` alongside source | `thread_test.go` |
| Test functions | `Test` prefix + descriptive | `TestLinearize_StripsNestedQuotes` |

## File Organization

- One primary type per file when the type is substantial.
- Group related small types/functions in a single file.
- Keep `cmd/` thin — flag parsing and orchestration only. No business logic.
- Internal packages under `internal/` to prevent external import.

## Commit Messages

- Imperative mood, concise subject line (≤72 chars).
- Examples: `Add quote detection for blockquote elements`, `Fix token refresh on 401 retry`.
- No prefix conventions required (no `feat:`, `fix:` — keep it simple).

## Error Handling

- Wrap errors with context: `fmt.Errorf("fetch messages for folder %q: %w", folder, err)`.
- Surface actionable messages to the user via stderr.
- Log debug details only when `--verbose` is set.
- Never log tokens, credentials, or email body content.

## Dependencies

- Minimize external dependencies. Prefer the standard library where practical.
- Vendor or pin dependency versions for reproducible builds.
- Evaluate any new dependency for maintenance health and security posture.
