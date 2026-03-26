# Email Thread Linearizer

CLI tool that reads multithreaded emails from a Microsoft 365 account and converts them into a linear dialogue in Markdown format — without duplicating emails nested in threads. It adds unique, human-readable references to previous emails that would otherwise have been quoted inline.

Read `context/` for detailed project context:

- [context/PRODUCT.md](context/PRODUCT.md) — Product vision, features, and CLI interface design
- [context/ARCHITECTURE.md](context/ARCHITECTURE.md) — System components, data flow, and module layout
- [context/CONVENTIONS.md](context/CONVENTIONS.md) — Code style, naming, file organization, and commit conventions
- [context/TESTING.md](context/TESTING.md) — Test strategy, patterns, and mocking approach
- [context/SECURITY.md](context/SECURITY.md) — OAuth handling, credential storage, and data privacy
- [context/API.md](context/API.md) — Microsoft Graph API usage for mail threads

## Quick Reference

- **Language**: Go (preferred) — final decision pending
- **Auth**: OAuth 2.0 device code flow against Microsoft Entra ID (Azure AD)
- **API**: Microsoft Graph REST API v1.0 (`/me/messages`, `/me/mailFolders`)
- **Output**: Markdown files with linear dialogue, one per thread
- **Secrets**: Environment variables or OS keychain — never in code or config files
- **Tests**: Required for new functionality. Mock the Graph API at the HTTP boundary.
