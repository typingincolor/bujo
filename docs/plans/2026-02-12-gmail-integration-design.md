# Gmail-to-Bujo Integration Design

> **Issue:** #492
> **Date:** 2026-02-12
> **Status:** Ready for implementation

## Goal

Quick capture from Gmail: turn emails into bujo tasks without leaving the browser. The bujo desktop app (Wails) embeds an HTTP API that a bookmarklet calls directly.

## Architecture

```
Gmail (browser)                    Bujo (desktop app)
┌──────────────┐                  ┌─────────────────────────┐
│  Bookmarklet │──POST JSON──────▶│  HTTP Adapter (:8743)   │
│  (Gmail DOM  │◀─JSON response───│  internal/adapter/http/  │
│   scraping)  │                  │         │                │
└──────────────┘                  │         ▼                │
                                  │  Service Layer           │
                                  │  (BujoService.LogEntries)│
                                  │         │                │
                                  │         ▼                │
                                  │  SQLite (bujo.db)        │
                                  └─────────────────────────┘
```

The HTTP server runs inside the existing Wails app process. No new binary, no `bujo serve` command, no separate daemon. The server starts when bujo starts and stops when bujo quits.

## API Design

### `POST /api/entries` — Create entries with parent-child relationships

**Request:**
```json
{
  "entries": [
    {
      "type": "task",
      "content": "Follow up: Q1 Planning Follow-up @john #email",
      "children": [
        { "type": "note", "content": "Context: Thanks for the meeting..." },
        { "type": "note", "content": "Email: https://mail.google.com/..." }
      ]
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "entries": [
    {
      "id": 1234,
      "children": [{ "id": 1235 }, { "id": 1236 }]
    }
  ]
}
```

**Error Response (400 Bad Request):**
```json
{
  "success": false,
  "error": "Missing required field: content"
}
```

The handler translates JSON into the indented text format that `TreeParser` expects, then calls `BujoService.LogEntries()`. No new service methods needed.

### `GET /api/health` — Connectivity check

```json
{ "status": "ok" }
```

The bookmarklet pings this before attempting the POST.

### `GET /install` — Bookmarklet install page

Serves an embedded HTML page where the user drags a link to their bookmarks bar. One-time setup taking ~10 seconds.

### Design Choices

- **Localhost only:** Server binds to `127.0.0.1`, not `0.0.0.0`. Not network-accessible.
- **No auth:** Unnecessary for localhost-only connections.
- **CORS:** Allows `https://mail.google.com` origin so the bookmarklet's `fetch()` works.
- **Port 8743:** Uncommon port, avoids conflicts with common dev servers.
- **Entries created for today's date.** Rescheduling happens in bujo.

## Entry Structure

Each captured email creates a parent task with two child notes:

```
. Follow up: Q1 Planning Follow-up @john #email
  - Context: Thanks for the meeting yesterday...
  - Email: https://mail.google.com/mail/u/0/#inbox/...
```

Tags (`#email`) and mentions (`@john`) are extracted by the existing parser and stored in their respective tables.

## Bookmarklet

A JavaScript function that:
1. Extracts email data from Gmail's DOM (`h2.hP`, `span.gD[email]`, `div.a3s.aiL`)
2. POSTs JSON to `http://127.0.0.1:8743/api/entries`
3. Shows a toast notification in Gmail with success/error

No shell commands, no clipboard, no terminal. Direct HTTP call to the running bujo app.

### Install Flow

1. Open bujo (HTTP server starts automatically)
2. Visit `http://127.0.0.1:8743/install` in Chrome
3. Drag the "Gmail → Bujo" button to bookmarks bar
4. Done

### Gmail Selectors

- Subject: `document.querySelector('h2.hP')`
- Sender: `document.querySelector('span.gD[email]')`
- Body: `document.querySelector('div.a3s.aiL')`
- URL: `window.location.href`

These are fragile and may break if Gmail updates its DOM.

## New Code

### `internal/adapter/http/` (new package)

| File | Purpose |
|------|---------|
| `server.go` | HTTP server setup, start/stop lifecycle |
| `routes.go` | Route registration |
| `handlers.go` | Request handling, JSON parsing, response formatting |
| `install.html` | Embedded install page (via `go:embed`) |

### `integrations/gmail/` (new directory)

| File | Purpose |
|------|---------|
| `bookmarklet.js` | Readable source code |
| `minify-bookmarklet.js` | Build script |
| `install.html` | Source for the embedded install page |
| `README.md` | Usage guide |

### Modified Files

| File | Change |
|------|--------|
| `internal/adapter/wails/app.go` | Start HTTP server in `Startup()`, stop in `Shutdown()` |

### Unchanged

Domain layer, service layer, repository layer, frontend — all untouched. The existing `BujoService.LogEntries()` already handles entry creation with parent-child relationships, tag extraction, and mention extraction.

## Wails Integration

**Startup:** `App.Startup()` creates and starts the HTTP server on `127.0.0.1:8743` in a goroutine. Both the Wails runtime and the HTTP server share the same `*app.Services` instance.

**Shutdown:** `App.Shutdown()` calls `httpServer.Shutdown(ctx)` for graceful cleanup. In-flight requests complete before the server stops.

## Testing Strategy

Unit test HTTP handlers with `httptest.NewServer` backed by an in-memory SQLite database. Test through the real service layer (no mocks).

**Test cases (~6-8):**
- Valid entry creation returns 201 + entry IDs
- Parent-child linking creates correct hierarchy
- Validation: missing content returns 400
- Validation: invalid entry type returns 400
- CORS preflight from `mail.google.com` returns correct headers
- Health endpoint returns 200
- Install page serves HTML

The bookmarklet is tested manually against Gmail (browser JS, not Go code).

## Future: Chrome Extension

If the bookmarklet feels limiting, a Chrome extension provides:
- Native "Send to Bujo" button in Gmail toolbar
- Settings UI for customization
- Offline queue with retry logic
- Auto-update capability

The API designed here serves both the bookmarklet and a future extension without changes.
