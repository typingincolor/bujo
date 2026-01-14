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

## CI/CD and Distribution Changes

### Current State (Pure Go)

```yaml
# .goreleaser.yaml - simple cross-compilation
builds:
  - env:
      - CGO_ENABLED=0  # No C dependencies
    goos: [darwin, linux, windows]
    goarch: [amd64, arm64]
```

- Single `go build` works everywhere
- Cross-compilation from any platform
- `brew install typingincolor/tap/bujo` downloads pre-built binary

### With CGO: Build Strategy Options

#### Option A: Platform-Specific Builds (Recommended)

Build natively on each platform instead of cross-compiling:

```yaml
# .github/workflows/release.yml
jobs:
  build-macos:
    runs-on: macos-latest
    strategy:
      matrix:
        arch: [amd64, arm64]
    steps:
      - run: |
          brew install llama.cpp
          CGO_ENABLED=1 GOARCH=${{ matrix.arch }} go build -o bujo-darwin-${{ matrix.arch }}

  build-linux:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        arch: [amd64, arm64]
    steps:
      - run: |
          sudo apt-get install -y build-essential
          # Build llama.cpp from source or use pre-built
          CGO_ENABLED=1 go build -o bujo-linux-${{ matrix.arch }}
```

**Pros:** Native builds, optimal performance, Metal acceleration on macOS
**Cons:** Slower CI (4 separate jobs), more complex workflow

#### Option B: Two Binary Variants

Ship both `bujo` (pure Go, no AI) and `bujo-ai` (with CGO):

```yaml
builds:
  - id: bujo-core
    env: [CGO_ENABLED=0]
    # ... cross-compile all platforms

  - id: bujo-ai
    env: [CGO_ENABLED=1]
    # ... native builds only
```

**Pros:** Users choose complexity level
**Cons:** Two binaries to maintain, confusing UX

#### Option C: Runtime Plugin

Keep bujo pure Go, download AI engine as separate binary on first use:

```bash
$ bujo ask "question"
# AI engine not found. Downloading bujo-ai-engine for darwin-arm64...
# Downloaded to ~/.bujo/bin/bujo-ai-engine
```

**Pros:** Simple main binary, optional AI
**Cons:** Complex architecture, two binaries anyway

### Recommended: Option A with Homebrew Changes

#### Updated GoReleaser Config

```yaml
# .goreleaser.yaml
version: 2

builds:
  - id: bujo-darwin-amd64
    main: ./cmd/bujo
    binary: bujo
    goos: [darwin]
    goarch: [amd64]
    env: [CGO_ENABLED=1]
    flags: [-tags=cgo]

  - id: bujo-darwin-arm64
    main: ./cmd/bujo
    binary: bujo
    goos: [darwin]
    goarch: [arm64]
    env: [CGO_ENABLED=1]
    flags: [-tags=cgo]

  - id: bujo-linux-amd64
    main: ./cmd/bujo
    binary: bujo
    goos: [linux]
    goarch: [amd64]
    env: [CGO_ENABLED=1]

  # Windows: CGO_ENABLED=0 fallback (no local AI, Gemini only)
  - id: bujo-windows
    main: ./cmd/bujo
    binary: bujo
    goos: [windows]
    goarch: [amd64, arm64]
    env: [CGO_ENABLED=0]
    flags: [-tags=noai]
```

#### Updated CI Workflow

```yaml
# .github/workflows/ci.yml
jobs:
  test:
    strategy:
      matrix:
        include:
          - os: macos-latest
            cgo: "1"
          - os: ubuntu-latest
            cgo: "1"
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
      - name: Install llama.cpp (macOS)
        if: runner.os == 'macOS'
        run: brew install llama.cpp
      - name: Install llama.cpp (Linux)
        if: runner.os == 'Linux'
        run: |
          git clone https://github.com/ggerganov/llama.cpp
          cd llama.cpp && make && sudo make install
      - run: CGO_ENABLED=${{ matrix.cgo }} go test -v ./...
```

#### Updated Homebrew Formula

```ruby
# Formula/bujo.rb
class Bujo < Formula
  desc "Command-line Bullet Journal with local AI"
  homepage "https://github.com/typingincolor/bujo"
  license "MIT"

  depends_on "llama.cpp"  # NEW: runtime dependency

  # Pre-built bottles for common platforms
  bottle do
    root_url "https://github.com/typingincolor/bujo/releases/download/v#{version}"
    sha256 cellar: :any, arm64_sonoma: "..."
    sha256 cellar: :any, ventura: "..."
  end

  def install
    bin.install "bujo"
  end

  def caveats
    <<~EOS
      To use local AI features, download a model:
        bujo model pull tinyllama

      Models are stored in ~/.bujo/models/
    EOS
  end
end
```

### Homebrew: Will It Still Work?

**Yes**, but with changes:

| Aspect | Before | After |
|--------|--------|-------|
| Install command | `brew install typingincolor/tap/bujo` | Same |
| Dependencies | None | `llama.cpp` |
| Binary type | Static | Dynamic (links llama.cpp) |
| First run | Works immediately | Needs `bujo model pull` |
| Binary size | ~15 MB | ~20-25 MB |

### User Experience

```bash
# Install (same as before)
$ brew install typingincolor/tap/bujo

# First AI command prompts for model
$ bujo ask "What did I do today?"
No AI model found. Available models:
  tinyllama     (637 MB)  - Fast, good for testing
  llama3.2:1b   (1.3 GB)  - Recommended

Download tinyllama? [Y/n] y
Downloading tinyllama... 637 MB / 637 MB [================] 100%
Model ready!

Based on your entries today...
```

### Windows Considerations

Windows builds will use `CGO_ENABLED=0` with a `noai` build tag:

```go
// +build noai

package ai

func NewLocalClient() (GenAIClient, error) {
    return nil, errors.New("local AI not available on this platform, use BUJO_AI_PROVIDER=gemini")
}
```

Windows users can still use Gemini for AI features.

## Implementation Order

1. **Domain types** (`internal/domain/model.go`) - ModelSpec, ModelVersion, ModelInfo
2. **Model service** (`internal/service/model.go`) - list, pull, rm, check, update operations
3. **Downloader** (`internal/adapter/ai/local/download.go`) - Hugging Face fetcher with progress
4. **Version checker** (`internal/adapter/ai/local/version.go`) - HF API client for commit comparison
5. **Local client** (`internal/adapter/ai/local/client.go`) - llama.cpp wrapper with streaming support
6. **CLI: model commands** - `bujo model list|pull|rm|check|update|status|verify`
7. **CLI: ask command** - `bujo ask "question"` with streaming output
8. **Wire up summary** - Update summary command to use local by default
9. **CI/CD updates** - Platform-specific builds, Homebrew formula
10. **Configuration** - Environment variables, defaults

## Suggested Models (Curated List)

| Model | Size | Notes |
|-------|------|-------|
| tinyllama | 637 MB | Fastest, good for testing |
| llama3.2:1b | 1.3 GB | Good balance |
| llama3.2:3b | 2.0 GB | Better quality |
| phi-3-mini | 2.3 GB | Microsoft, good reasoning |
| mistral:7b | 4.1 GB | High quality, needs more RAM |

## Additional Considerations

### Context Window Management

LLMs have token limits. Strategy for handling large entry sets:

```go
const (
    MaxContextTokens = 4096  // Conservative limit for small models
    TokensPerEntry   = ~50   // Average estimate
    MaxEntries       = 80    // Safe default
)
```

**When entries exceed limit:**
1. Prioritize recent entries
2. Summarize older entries into bullet points
3. Show warning: "Analyzing 80 of 250 entries (most recent)"

```bash
$ bujo ask "What did I work on this year?" --from "jan 1"
⚠️ 1,247 entries found. Analyzing most recent 80 entries.
   Use --limit to adjust, or narrow date range with --from/--to.
```

### Memory Requirements

Different models need different RAM. Show requirements and warn:

| Model | RAM Required | GPU VRAM |
|-------|--------------|----------|
| tinyllama | 1 GB | 512 MB |
| llama3.2:1b | 2 GB | 1 GB |
| llama3.2:3b | 4 GB | 2 GB |
| mistral:7b | 8 GB | 4 GB |

```bash
$ bujo model pull mistral:7b
⚠️ mistral:7b requires 8 GB RAM. Your system has 8 GB.
   This may cause slowdowns. Continue? [y/N]
```

### Model Integrity (Checksums)

Verify downloads haven't been tampered with:

```json
// manifest.json
{
  "models": {
    "llama3.2:3b": {
      "sha256": "a1b2c3d4...",
      "verified": true
    }
  }
}
```

```bash
$ bujo model pull llama3.2:3b
Downloading llama3.2:3b... 2.0 GB [================] 100%
Verifying checksum... ✓ OK

$ bujo model verify
llama3.2:3b    ✓ OK
tinyllama      ✓ OK
```

### Disk Space Management

Warn before downloads and provide cleanup:

```bash
$ bujo model pull mistral:7b
⚠️ This download requires 4.1 GB. You have 3.2 GB free.
   Free up space or remove unused models with 'bujo model rm'.

$ bujo model status
Models directory: ~/.bujo/models/
Total size: 2.6 GB
Available: 3.2 GB

Downloaded models:
  llama3.2:3b    2.0 GB    (last used: 2 hours ago)
  tinyllama      637 MB    (last used: 3 days ago)
```

### Error Handling & Fallback

What happens when local inference fails:

```go
type AIError struct {
    Code    AIErrorCode
    Message string
    Model   string
}

const (
    ErrNoModel       AIErrorCode = "NO_MODEL"
    ErrModelCorrupt  AIErrorCode = "MODEL_CORRUPT"
    ErrOutOfMemory   AIErrorCode = "OUT_OF_MEMORY"
    ErrInferenceFail AIErrorCode = "INFERENCE_FAIL"
)
```

**Fallback behavior:**
```bash
$ bujo ask "question"
Error: Model inference failed (out of memory)

Options:
  1. Try a smaller model: bujo ask "question" --model tinyllama
  2. Use cloud AI: bujo ask "question" --provider gemini
  3. Reduce context: bujo ask "question" --limit 20
```

### Testing Strategy

How to test AI features without real models:

```go
// internal/adapter/ai/local/mock.go
type MockLLMClient struct {
    responses map[string]string
}

func (m *MockLLMClient) Generate(ctx context.Context, prompt string) (string, error) {
    // Return canned responses based on prompt keywords
    if strings.Contains(prompt, "habits") {
        return "You logged exercise 5 times this week.", nil
    }
    return "Mock AI response", nil
}
```

**Test categories:**
1. **Unit tests** - Mock the `GenAIClient` interface
2. **Integration tests** - Use tinyllama with `--short` flag in CI (slow, optional)
3. **Prompt tests** - Verify prompt construction without inference

### Migration from Gemini

For users with `GEMINI_API_KEY` set:

```bash
$ bujo summary weekly
Note: Using Gemini API. To switch to local AI:
  1. bujo model pull llama3.2:3b
  2. unset GEMINI_API_KEY (or set BUJO_AI_PROVIDER=local)
```

**Config precedence:**
1. `--provider` flag (highest)
2. `BUJO_AI_PROVIDER` env var
3. If `GEMINI_API_KEY` set and no local model → use Gemini
4. If local model exists → use local (default)

### Response Streaming (v1)

Show AI output as it generates (essential for slow local models):

```bash
$ bujo ask "What patterns do you see?"
Based on your ▌                    # appears immediately
Based on your entries, I notice ▌   # builds up in real-time
Based on your entries, I notice you exercise more on weekends...
```

**Why required for v1:**
- Local models can take 10-60 seconds for longer responses
- Without streaming, users may think it's frozen
- Allows early cancellation (Ctrl+C) if response is wrong direction

**Implementation:**

```go
// GenAIClient interface extended for streaming
type GenAIClient interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateStream(ctx context.Context, prompt string, callback func(token string)) error
}

// CLI usage
func (c *AskCmd) Run() error {
    return c.client.GenerateStream(ctx, prompt, func(token string) {
        fmt.Print(token)  // Print each token as it arrives
    })
}
```

**llama.cpp support:** The `go-llama.cpp` bindings support token callbacks via `llama.SetCallback()`.

### Prompt Templates

The `ask` command needs well-crafted prompts:

```go
var askPromptTemplate = `You are a helpful assistant analyzing a personal bullet journal.

Here are the journal entries:
{{range .Entries}}
{{.Type.Symbol}} {{.Content}} ({{.CreatedAt.Format "Jan 2"}})
{{end}}

User question: {{.Question}}

Provide a helpful, concise answer based only on the entries above.
If the entries don't contain relevant information, say so.`
```

## Open Questions

1. Should we keep Gemini as a fallback option or remove entirely?
2. Default model for first-time users - auto-download tinyllama or prompt?
3. GPU acceleration (Metal on macOS) - worth the complexity?
4. Should `bujo model` be a separate binary to keep core bujo pure Go?

## Related Files

- `internal/adapter/ai/client.go` - Existing Gemini client
- `internal/adapter/ai/generator.go` - SummaryGenerator interface
- `internal/adapter/ai/gemini.go` - Prompt building (reusable)
- `internal/service/summary.go` - Summary service (uses generator)
- `cmd/bujo/cmd/summary.go` - Summary CLI command
