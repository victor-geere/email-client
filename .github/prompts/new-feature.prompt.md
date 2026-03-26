---
description: "Scaffold a new internal module for the email thread linearizer — package, types, core logic, and tests"
agent: "agent"
argument-hint: "Describe the module to scaffold..."
---
Scaffold a new module for the email thread linearizer CLI. Create:

1. **Package** under `internal/<name>/` with a clear, focused responsibility
2. **Types** — domain structs relevant to the module
3. **Core logic** — exported functions implementing the module's behavior
4. **Tests** — `*_test.go` with behavior-descriptive names, using Arrange → Act → Assert

Follow `context/ARCHITECTURE.md` for module layout. Follow `context/CONVENTIONS.md` for naming and style.
Never log tokens or email body content. Use `context.Context` for cancellation where appropriate.
