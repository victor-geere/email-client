# Product

## Problem

Email threads in Microsoft 365 accumulate deeply nested quoted replies. Reading them requires scrolling through repeated blocks of previously seen content. There is no clean way to read a thread as a straightforward linear conversation.

## Solution

A CLI tool that:

1. Authenticates against a Microsoft 365 account via OAuth
2. Fetches email threads (conversations) from the user's mailbox
3. Deduplicates quoted/nested email content within each thread
4. Replaces inline-quoted blocks with short, human-readable references (e.g., `[→ see msg #3 from Alice, 14 Jan]`)
5. Outputs a Markdown file per thread with the emails rendered as a clean, linear dialogue

## Key Behaviors

- **No duplication**: If an email quotes a previous message, the quoted portion is replaced by a reference — not repeated.
- **Human-readable refs**: Each reference is meaningful at a glance — includes author, date, and a short ordinal within the thread.
- **Chronological order**: Emails appear in the order they were sent, oldest first.
- **Metadata preserved**: Each message in the output includes sender, recipients (To/Cc), date, and subject.
- **Thread grouping**: Emails are grouped by the `conversationId` field from Microsoft Graph.

## CLI Interface (Planned)

```
email-linearize [flags]

Flags:
  --folder <name>      Mailbox folder to read (default: "Inbox")
  --thread <id>        Specific conversation ID to process
  --output <dir>       Output directory for Markdown files (default: "./output")
  --limit <n>          Max number of threads to process
  --format <fmt>       Output format: "markdown" (default), "text"
  --verbose            Enable verbose/debug logging
  --auth-only          Authenticate and exit (useful for first-time setup)
```

## Output Example

```markdown
# Thread: Q4 Budget Review

---

### Message #1
**From:** Alice Wang <alice@example.com>
**To:** Bob Chen <bob@example.com>, Carol Diaz <carol@example.com>
**Date:** 2025-01-14 09:32 UTC
**Subject:** Q4 Budget Review

Hi team, here's the Q4 budget draft for review. Please share feedback by Friday.

---

### Message #2
**From:** Bob Chen <bob@example.com>
**To:** Alice Wang <alice@example.com>
**Cc:** Carol Diaz <carol@example.com>
**Date:** 2025-01-14 14:05 UTC
**Subject:** Re: Q4 Budget Review

[→ see msg #1 from Alice, 14 Jan]

The numbers look good. One question — did we account for the new contractor line item?

---
```

## Non-Goals

- This is not a full email client. No sending, composing, or mailbox management.
- No GUI — CLI only.
- No support for IMAP/SMTP. Microsoft Graph API only.
- No real-time sync or notifications.
