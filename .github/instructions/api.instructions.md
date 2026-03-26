---
description: "Use when writing or modifying the Microsoft Graph API client — HTTP calls, pagination, authentication headers, rate limiting, or response parsing."
applyTo: "**/graph/**,**/auth/**"
---
# Graph API Client Guidelines

- Accept a configurable base URL so tests can point the client at `httptest.Server`.
- Set the `Authorization: Bearer {token}` header on every request. Get the token from the auth module.
- Handle pagination: follow `@odata.nextLink` until absent. Never assume a single-page response.
- Respect rate limits: on 429, read `Retry-After` header and wait before retrying.
- Retry server errors (500, 503) with exponential backoff, max 3 attempts.
- On 401, attempt one token refresh then retry. If still 401, return an error prompting re-authentication.
- Parse Graph API JSON into domain types within this module — do not leak raw JSON or `map[string]any` to callers.
- Select only needed fields via `$select` to minimize response size and reduce quota usage.
- Never log response bodies (they contain email content). Log request metadata (URL, status, duration) only in verbose mode.
