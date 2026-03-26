---
description: "Generate Go tests for a module in the email thread linearizer"
agent: "agent"
argument-hint: "The file or package to generate tests for..."
---
Generate Go tests for the specified code. Follow these rules:

- Name tests to describe behavior: `TestQuoteDetector_IdentifiesBlockquoteElements`, not `TestDetect`
- Use Arrange → Act → Assert structure
- Cover happy paths, edge cases, and error scenarios
- For Graph client code: use `httptest.Server` to serve fixture JSON responses
- For quote detection: test both HTML (`<blockquote>`) and plain-text (`>` prefix) inputs
- For linearization: verify reference format `[→ see msg #N from Author, DD Mon]`
- Use builder functions for constructing test messages — avoid large struct literals
- Store reusable fixture JSON in `testdata/`
- Never use real credentials or call the real Microsoft Graph API
- Match existing test patterns in the codebase
