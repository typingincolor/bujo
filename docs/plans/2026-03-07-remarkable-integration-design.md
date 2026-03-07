# reMarkable Cloud Integration — Test Harness Design

**Date:** 2026-03-07
**Status:** Approved

## Goal

Validate the reMarkable Cloud API integration end-to-end before building full UX. Prove that bujo can authenticate, list documents, download them, extract typed text (from "Convert to text"), and parse entries.

## Background

- reMarkable has no official cloud API
- An unofficial reverse-engineered API exists, used by rmapi (Go) and others
- reMarkable changed their sync protocol (~2022, "protocol 1.5") with new endpoints
- New endpoints (as of Feb 2025):
  - Auth: `webapp-prod.cloud.remarkable.engineering/token/json/2/device/new`
  - Files: `eu.tectonic.remarkable.com/doc/v2/files` (may be region-specific)
- Old endpoints may still work via discovery at `service-manager-production-dot-remarkable-production.appspot.com`
- Documents are ZIP archives containing binary `.rm` stroke files and metadata
- "Convert to text" on-device creates typed text pages within notebooks that sync to cloud
- User writes in bujo notation on the reMarkable, then uses "Convert to text"

## Commands

```bash
# One-time registration
bujo remarkable register <8-char-code>
# → Stores device token to ~/.config/bujo/remarkable.json

# List documents from cloud
bujo remarkable list
# → Prints document names, IDs, modified dates

# Download, extract text, parse, print to stdout
bujo remarkable import <doc-id>
# → Downloads ZIP, extracts typed text, runs TreeParser, prints entries
# → No DB writes, stdout only
```

## Architecture

```
internal/
  adapter/
    remarkable/
      client.go          # HTTP client: register, auth, list, download
      client_test.go     # Tests with HTTP test server mocks
      document.go        # ZIP extraction + typed text parsing
      document_test.go   # Tests with fixture ZIP files
  adapter/
    cli/
      remarkable.go      # Cobra commands (register, list, import)
```

## Auth Flow

1. User visits `my.remarkable.com/connect/desktop` to get 8-char code
2. `POST /token/json/2/device/new` with `{code, deviceDesc: "desktop-macos", deviceID: <generated-uuid>}`
3. Response: device token (JWT, text/plain)
4. Store device token + device ID in `~/.config/bujo/remarkable.json`
5. Before each API call: `POST /token/json/2/user/new` with device token as Bearer → short-lived user token

### Token Storage

```json
{
  "device_token": "eyJ...",
  "device_id": "d4605307-a145-48d2-b60a-3be2c46035ef"
}
```

## Document Listing

- Try new endpoint: `GET https://eu.tectonic.remarkable.com/doc/v2/files` with user token
- Fall back to old: discovery → `GET /document-storage/json/2/docs`
- Display: name (`VissibleName`), ID, modified date, type

## Document Download + Text Extraction

1. Fetch single doc with `?doc=<id>&withBlob=true` to get `BlobURLGet`
2. Download ZIP from blob URL
3. Extract ZIP, look for typed text content (pages created by "Convert to text")
4. Text format in bundle is TBD — this is what the test harness will discover

## What We Expect to Learn

- Whether new or old protocol endpoints work (or both)
- How "Convert to text" content is stored in document bundles
- Whether region-specific endpoints matter (`eu.` prefix)
- Any auth quirks or rate limiting

## Out of Scope

- Import history / duplicate tracking
- Frontend integration
- Entry persistence (no DB writes)
- Gemini fallback for OCR
- Robust error handling
- Folder browsing / tree navigation
