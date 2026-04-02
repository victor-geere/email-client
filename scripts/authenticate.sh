#!/usr/bin/env bash
# Authenticate with Microsoft 365 using the device code flow.
# A code is displayed in the terminal — go to https://microsoft.com/devicelogin
# and enter it. MFA prompts (if enabled) are handled during that browser session.
# Requires EMAIL_LINEAR_CLIENT_ID to be set.
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
  echo "Set it to your Microsoft Entra ID application (client) ID." >&2
  exit 1
fi

BINARY="./email-linearize"
if [[ ! -x "$BINARY" ]]; then
  echo "Binary not found. Building first..."
  ./scripts/build.sh
fi

# The binary will display a user code and verification URL.
# Sign in at the URL, enter the code, and complete any MFA prompts.
"$BINARY" --auth-only
