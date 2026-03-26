---
description: "Review code for security issues specific to the email thread linearizer CLI"
agent: "agent"
argument-hint: "The file or module to review..."
---
Review the specified code for security vulnerabilities, focusing on threats relevant to this CLI tool:

1. **Token leakage**: Are OAuth tokens, refresh tokens, or credentials ever logged, printed, or included in error messages?
2. **Credential storage**: Are tokens stored securely (OS keychain / encrypted file)? Never plaintext config.
3. **Email content in logs**: Is email body content logged anywhere — even in debug/verbose mode?
4. **Remote resource fetching**: Does the code fetch remote images, stylesheets, or other resources from email HTML? (It must not.)
5. **TLS validation**: Are HTTPS calls made with proper TLS certificate validation? No skipping.
6. **Input validation**: Are Graph API responses validated before use? Are CLI flags sanitized?
7. **Retry behavior**: Does token refresh have a bound? No infinite retry loops.

Report findings as a list with severity (critical/high/medium/low), location, and a fix recommendation.
