# Microsoft Graph API Usage

## Overview

This tool uses the Microsoft Graph REST API v1.0 to read emails from a user's Microsoft 365 mailbox. It is read-only — no messages are sent, modified, or deleted.

## Authentication Endpoint

```
POST https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/devicecode
POST https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/token
```

- **Flow**: OAuth 2.0 device authorization grant
- **Scopes**: `Mail.Read`, `offline_access`
- `offline_access` is needed to receive a refresh token for subsequent runs

## Key Endpoints

### List messages in a folder

```
GET https://graph.microsoft.com/v1.0/me/mailFolders/{folder_id}/messages
```

Query parameters:
- `$top=50` — page size (max 1000, default 10)
- `$orderby=receivedDateTime asc`
- `$select=id,conversationId,subject,from,toRecipients,ccRecipients,receivedDateTime,body,bodyPreview`
- `$filter=conversationId eq '{id}'` — to fetch a specific thread

Pagination: follow `@odata.nextLink` until absent.

### List mail folders

```
GET https://graph.microsoft.com/v1.0/me/mailFolders
```

Used to resolve folder name (e.g., "Inbox") to folder ID.

### Get a specific message

```
GET https://graph.microsoft.com/v1.0/me/messages/{message_id}
```

Rarely needed — bulk list is preferred.

## Response Shape (Message)

```json
{
  "id": "AAMkAG...",
  "conversationId": "AAQkAD...",
  "subject": "Re: Q4 Budget Review",
  "receivedDateTime": "2025-01-14T14:05:00Z",
  "from": {
    "emailAddress": {
      "name": "Bob Chen",
      "address": "bob@example.com"
    }
  },
  "toRecipients": [
    {
      "emailAddress": {
        "name": "Alice Wang",
        "address": "alice@example.com"
      }
    }
  ],
  "ccRecipients": [],
  "body": {
    "contentType": "html",
    "content": "<html><body>...</body></html>"
  },
  "bodyPreview": "The numbers look good..."
}
```

## Important Fields

| Field | Purpose |
|-------|---------|
| `conversationId` | Groups messages into threads. All messages in a conversation share this value. |
| `body.contentType` | `"html"` or `"text"` — determines which quote detection strategy to use. |
| `body.content` | Full message body. HTML bodies contain `<blockquote>` for quoted replies. |
| `receivedDateTime` | ISO 8601 timestamp. Used for chronological sorting. |
| `from.emailAddress` | Sender info for display and reference generation. |

## Rate Limiting

- Graph API returns `429 Too Many Requests` with a `Retry-After` header (seconds).
- The client must respect `Retry-After` — do not retry before the specified delay.
- Typical limits: ~10,000 requests per 10 minutes per app per tenant (varies by endpoint).

## Error Handling

| Status | Meaning | Action |
|--------|---------|--------|
| 200 | Success | Process response |
| 401 | Token expired or invalid | Refresh token, retry once |
| 403 | Insufficient permissions | Ask user to re-consent with correct scopes |
| 404 | Folder or message not found | Report and skip |
| 429 | Rate limited | Wait `Retry-After` seconds, then retry |
| 500/503 | Server error | Retry with exponential backoff (max 3 attempts) |

## App Registration

The tool requires an Azure AD app registration with:
- **Redirect URI**: `https://login.microsoftonline.com/common/oauth2/nativeclient` (for device code flow)
- **API permissions**: `Mail.Read` (delegated)
- **Supported account types**: depends on target audience (single tenant or multi-tenant)

Configuration values provided by the user via environment variables:
- `EMAIL_LINEAR_CLIENT_ID` — Application (client) ID
- `EMAIL_LINEAR_TENANT_ID` — Directory (tenant) ID (or `common` for multi-tenant)
