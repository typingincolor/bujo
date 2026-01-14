# feat: Add local AI support for notes Q&A and summaries

## Summary

Add the ability to query notes using AI that runs entirely locally, without sending data to cloud services. This replaces the existing Gemini integration with a local llama.cpp-based solution.

## Motivation

- **Privacy**: Journal entries are personal - users shouldn't need to send them to cloud APIs
- **Offline usage**: Local AI works without internet connectivity
- **Cost**: No API fees for AI features

## Proposed Features

### 1. Model Management Commands

```bash
# List available models (shows download status)
bujo model list

# Output:
# Available models:
#   llama3.2:3b      (2.0 GB)  [downloaded]
#   llama3.2:1b      (1.3 GB)
#   phi-3-mini       (2.3 GB)
#   mistral:7b       (4.1 GB)
#   tinyllama        (637 MB)

# Download a model
bujo model pull llama3.2:3b

# Check for updates
bujo model check

# Output:
# Updates available:
#   llama3.2:3b      v1.0.0 → v1.1.0  (2.1 GB)
#   tinyllama        (up to date)

# Update a specific model
bujo model update llama3.2:3b

# Update all downloaded models
bujo model update --all

# Remove a model
bujo model rm mistral:7b
```

### 2. Ask Command (New)

Query your notes with natural language:

```bash
bujo ask "What patterns do you see in my habits?"
bujo ask "Summarize my notes about the API project"
bujo ask "What did I accomplish last week?" --from "last monday"
bujo ask "What questions are still open?" --model tinyllama
```

### 3. Local Summaries (Replaces Gemini)

Existing summary command uses local AI:

```bash
bujo summary weekly    # Uses local model
bujo summary daily
```

## Technical Design

### Architecture

```
internal/adapter/ai/
├── interface.go       # GenAIClient interface (existing)
├── gemini.go          # Gemini adapter (keep as optional fallback)
├── local/
│   ├── client.go      # LocalLLMClient implementing GenAIClient
│   ├── llama.go       # llama.cpp CGO bindings wrapper
│   └── download.go    # Model downloader (Hugging Face GGUF files)

internal/domain/
├── model.go           # ModelSpec, ModelInfo types

internal/service/
├── model.go           # ModelService (list, pull, rm, default)
├── ask.go             # AskService (Q&A with context retrieval)
```

### Domain Types

```go
type ModelSpec struct {
    Name    string // e.g., "llama3.2"
    Variant string // e.g., "3b" (optional)
}

type ModelVersion struct {
    Major int
    Minor int
    Patch int
}

type ModelInfo struct {
    Spec           ModelSpec
    Version        ModelVersion // latest available version
    Size           int64        // bytes
    Description    string
    HFRepo         string       // Hugging Face repo path
    HFFile         string       // GGUF filename
    LocalPath      string       // empty if not downloaded
    LocalVersion   *ModelVersion // nil if not downloaded
}

func (m ModelInfo) HasUpdate() bool {
    return m.LocalVersion != nil && m.Version.NewerThan(*m.LocalVersion)
}

func AvailableModels() []ModelInfo // curated list, fetches latest versions
func ParseModelName(s string) (ModelSpec, error)
```

### Unified AI Interface

The existing `GenAIClient` interface already supports this:

```go
type GenAIClient interface {
    Generate(ctx context.Context, prompt string) (string, error)
}
```

Both `GeminiClient` and the new `LocalLLMClient` implement this interface. The existing `GeminiGenerator` (prompt builder) works with either backend.

### Configuration

```bash
# Provider selection (default: local)
export BUJO_AI_PROVIDER=local   # or "gemini"

# Local model settings
export BUJO_MODEL=llama3.2:3b
export BUJO_MODEL_DIR=~/.bujo/models

# Gemini (optional fallback)
export GEMINI_API_KEY=...

# Update notifications
export BUJO_UPDATE_CHECK=true           # enable/disable (default: true)
export BUJO_UPDATE_CHECK_INTERVAL=24h   # how often to check (default: 24h)
```

### Model Storage

Models downloaded to `~/.bujo/models/` (configurable via `BUJO_MODEL_DIR`):

```
~/.bujo/models/
├── llama3.2-3b-q4.gguf
├── tinyllama-q4.gguf
└── manifest.json  # tracks downloaded models and versions
```

**manifest.json structure:**

```json
{
  "models": {
    "llama3.2:3b": {
      "version": "1.0.0",
      "file": "llama3.2-3b-q4.gguf",
      "size": 2147483648,
      "downloaded_at": "2025-01-14T10:30:00Z",
      "hf_repo": "TheBloke/Llama-3.2-3B-GGUF",
      "hf_commit": "abc123"
    }
  },
  "version_cache": {
    "llama3.2:3b": {
      "latest_commit": "def456",
      "checked_at": "2025-01-14T12:00:00Z"
    }
  },
  "notified_updates": ["llama3.2:3b"]
}
```

### Version Resolution

Models are versioned by tracking Hugging Face commit SHAs:

1. **On `model pull`**: Store the HF commit SHA in manifest
2. **On `model check`**: Query HF API for latest commit, compare with stored
3. **On `model update`**: Download new file, update manifest, remove old file

### Update Notifications

Users are automatically notified of available updates when using AI features:

```bash
$ bujo ask "What did I work on today?"

💡 Update available: llama3.2:3b v1.0.0 → v1.1.0
   Run 'bujo model update llama3.2:3b' to update

Based on your entries today, you focused on...
```

**Notification behavior:**
- Checks for updates in background (cached for 24 hours)
- Shows notification once per session per model
- Can be disabled with `--quiet` flag or `BUJO_UPDATE_CHECK=false`
- Never blocks on network - uses cached version info if offline

### CGO Dependency

Uses `github.com/go-skynet/go-llama.cpp` or similar for inference:

- Requires C compiler (Xcode CLI tools on macOS)
- Models in GGUF format from Hugging Face
- First-run downloads model if not present

## Implementation Order

1. **Domain types** (`internal/domain/model.go`) - ModelSpec, ModelVersion, ModelInfo
2. **Model service** (`internal/service/model.go`) - list, pull, rm, check, update operations
3. **Downloader** (`internal/adapter/ai/local/download.go`) - Hugging Face fetcher with progress
4. **Version checker** (`internal/adapter/ai/local/version.go`) - HF API client for commit comparison
5. **Local client** (`internal/adapter/ai/local/client.go`) - llama.cpp wrapper implementing GenAIClient
6. **CLI: model commands** - `bujo model list|pull|rm|check|update`
7. **CLI: ask command** - `bujo ask "question"`
8. **Wire up summary** - Update summary command to use local by default
9. **Configuration** - Environment variables, defaults

## Suggested Models (Curated List)

| Model | Size | Notes |
|-------|------|-------|
| tinyllama | 637 MB | Fastest, good for testing |
| llama3.2:1b | 1.3 GB | Good balance |
| llama3.2:3b | 2.0 GB | Better quality |
| phi-3-mini | 2.3 GB | Microsoft, good reasoning |
| mistral:7b | 4.1 GB | High quality, needs more RAM |

## Open Questions

1. Should we keep Gemini as a fallback option or remove entirely?
2. Default model for first-time users - auto-download tinyllama or prompt?
3. GPU acceleration (Metal on macOS) - worth the complexity?

## Related Files

- `internal/adapter/ai/client.go` - Existing Gemini client
- `internal/adapter/ai/generator.go` - SummaryGenerator interface
- `internal/adapter/ai/gemini.go` - Prompt building (reusable)
- `internal/service/summary.go` - Summary service (uses generator)
- `cmd/bujo/cmd/summary.go` - Summary CLI command
