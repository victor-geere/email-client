#!/usr/bin/env bash
# Fetch and linearize email threads from Microsoft 365.
# Requires EMAIL_LINEAR_CLIENT_ID to be set and a valid cached token.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

if [[ -f .env ]]; then
  set -a
  source .env
  set +a
fi

if [[ -z "${EMAIL_LINEAR_CLIENT_ID:-}" ]]; then
  echo "Error: EMAIL_LINEAR_CLIENT_ID is not set." >&2
  exit 1
fi

BINARY="./email-linearize"
if [[ ! -x "$BINARY" ]]; then
  echo "Binary not found. Building first..."
  ./scripts/build.sh
fi

# Pass through any arguments from the command line
"$BINARY" "$@"
