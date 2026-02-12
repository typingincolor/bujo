# Gmail → Bujo Integration

Capture emails as bujo tasks directly from Gmail using a bookmarklet.

## Setup

1. Open bujo (the HTTP API starts automatically on port 8743)
2. Visit http://127.0.0.1:8743/install in your browser
3. Drag the "Gmail → Bujo" button to your bookmarks bar

## Usage

1. Open an email in Gmail
2. Click the bookmarklet in your bookmarks bar
3. The email becomes a task in bujo with context and link

## What Gets Captured

Each email creates a task with two child notes:

```
. Follow up: <subject> @<sender> #email
  - Context: <first 200 chars of body>
  - Email: <gmail url>
```

Tags (`#email`) and mentions (`@sender`) are extracted automatically.

## API

The bookmarklet calls `POST http://127.0.0.1:8743/api/entries`. The server only accepts connections from localhost.

## Files

- `bookmarklet.js` — Readable source code for the bookmarklet
- `../../internal/adapter/http/install.html` — Embedded install page with minified bookmarklet
