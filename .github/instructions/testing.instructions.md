---
description: "Use when writing or modifying tests, creating test fixtures, or setting up test infrastructure. Covers test naming, structure, mocking Graph API, and patterns for the email thread linearizer."
applyTo: "**/*_test.go,**/testdata/**"
---
# Testing Guidelines

- Name tests to describe behavior: `TestLinearize_ReplacesNestedQuoteWithReference`, not `TestLinearize`.
- Structure tests as Arrange → Act → Assert.
- Use builder/factory functions for test messages — avoid copy-pasting large structs across tests.
- Store shared JSON fixtures (sample Graph API responses) in `testdata/`.
- Mock Microsoft Graph API using `httptest.Server`. The Graph client must accept a configurable base URL.
- Never use real credentials or call the real Graph API in tests.
- Test quote detection with both HTML (`<blockquote>`) and plain-text (`>` prefix) inputs.
- Keep tests independent — no shared mutable state between test cases.
- Run `go test ./...` before considering work complete.
