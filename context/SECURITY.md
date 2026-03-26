# Security

## Authentication

- **OAuth 2.0 device code flow** against Microsoft Entra ID (Azure AD).
- The device code flow is chosen because this is a CLI tool — no browser redirect is needed on headless machines; the user approves on any device.
- Required scopes: `Mail.Read` (read-only access to mailbox). No write scopes.
- Client ID is configured via environment variable (`EMAIL_LINEAR_CLIENT_ID`), never hardcoded.

## Token Handling

- Access tokens are cached locally for reuse within their validity window.
- Refresh tokens are stored encrypted on disk or in the OS keychain (macOS Keychain, Linux secret-service, Windows Credential Manager).
- Tokens are never logged, printed to stdout, or included in error messages.
- On 401 from Graph API: attempt one token refresh, then fail with a clear "re-authenticate" message.

## Credential Storage

- **Environment variables** for app registration values: client ID, tenant ID.
- **OS keychain** (preferred) or encrypted file for cached tokens.
- Never store credentials in config files committed to version control.
- `.gitignore` must exclude any local token cache files.

## Data Privacy

- Email content is fetched, processed in memory, and written to local Markdown files. It is never sent to any third-party service.
- Email body content is never logged, even in verbose/debug mode. Log metadata (subject line, sender, date) only when `--verbose` is set.
- No telemetry or analytics.

## Email Content Safety

- Email HTML may contain malicious content (scripts, tracking pixels, encoded payloads).
- The tool strips HTML to extract text content. It does not render HTML in a browser.
- Remote image URLs in email bodies are NOT fetched or followed — they are stripped or ignored.
- The output is Markdown plain text — no executable content survives to the output files.

## Network

- All API calls use HTTPS (Microsoft Graph API enforces this).
- TLS certificate validation must remain enabled. Never skip certificate checks.
- Respect Microsoft Graph rate limits (429 + `Retry-After`). Do not retry aggressively.

## Threat Model (Lightweight)

| Threat | Mitigation |
|--------|-----------|
| Token theft from disk | OS keychain or encrypted storage; restrictive file permissions |
| Token leak via logs | Never log tokens; redact in error messages |
| Malicious email HTML | Strip HTML to text; do not render or execute |
| Tracking pixels | Do not fetch remote resources |
| Man-in-the-middle | HTTPS only; validate TLS certificates |
| Excessive API calls | Rate limit enforcement; `--limit` flag caps threads processed |
