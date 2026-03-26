# Email Thread Linearizer — Project Guidelines

## Overview

CLI tool that reads multithreaded emails from a Microsoft 365 account and converts them into a linear dialogue in Markdown format — without duplicating emails nested in threads. See `context/` for detailed project context.

## Language

Go (preferred) — final decision pending. Conventions below assume Go.

## Code Style

- Clear, descriptive names. No abbreviations except well-known ones (`id`, `url`, `html`, `msg`).
- Small, focused functions. One function = one thing.
- Early returns to reduce nesting.
- Files under 300 lines. Extract when growing beyond that.

## Architecture

- Modules: `auth`, `graph`, `thread`, `quote`, `linearize`, `render`, `cmd`.
- Business logic in `internal/`. CLI entry point in `cmd/`.
- Never store secrets in code. Use environment variables or OS keychain.
- Email HTML is stripped to text — not rendered in a browser.

## Security

- Never log tokens, credentials, or email body content.
- Store OAuth tokens in OS keychain or encrypted file, never in plaintext config.
- Strip HTML to text. Do not fetch remote resources (images, tracking pixels).
- HTTPS for all external API calls. Validate TLS certificates.
- Validate data at system boundaries (API responses, CLI flags).

## Testing

- Tests required for new functionality. Unit tests for logic, integration tests for Graph client.
- Mock Microsoft Graph API at the HTTP boundary (`httptest.Server`). No real API calls.
- Test names describe behavior: `TestLinearize_ReplacesNestedQuoteWithReference`.
- Run tests before considering work complete.

## Conventions

- Environment variables for configuration: `EMAIL_LINEAR_CLIENT_ID`, `EMAIL_LINEAR_TENANT_ID`.
- Commit messages: imperative mood, concise (`"Add quote detection for blockquotes"`).
- Inline comments explain *why*, not *what*.
