#!/usr/bin/env bash
# Build the email-linearize binary.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

echo "Running tests..."
go test ./...

echo "Building email-linearize..."
go build -o email-linearize ./cmd/email-linearize

echo "Build complete: ./email-linearize"
