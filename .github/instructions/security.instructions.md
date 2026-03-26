---
description: "Use when writing code that handles OAuth tokens, credential storage, email content processing, HTTP calls to Microsoft Graph, or any trust boundary crossing. Covers token safety, data privacy, and secure API communication."
applyTo: "**"
---
# Security Guidelines

- Never log OAuth tokens, refresh tokens, or any credentials — not even in verbose/debug mode.
- Store tokens in OS keychain (macOS Keychain, Linux secret-service) or encrypted file with restrictive permissions. Never in plaintext config.
- Strip email HTML to extract text. Do not render HTML or execute any embedded content.
- Do not fetch remote resources (images, stylesheets) from email bodies — they may be tracking pixels.
- Never log email body content. Log only metadata (subject, sender, date) and only when `--verbose` is set.
- Use HTTPS for all API calls. Never disable TLS certificate validation.
- Validate and sanitize all data from Microsoft Graph API responses before processing.
- Environment variables for app registration values (`EMAIL_LINEAR_CLIENT_ID`, `EMAIL_LINEAR_TENANT_ID`). Never hardcode.
- On token refresh failure, prompt the user to re-authenticate. Do not retry indefinitely.
