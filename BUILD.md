# Building bujo

This guide covers building the CLI, TUI, and desktop application.

## Prerequisites

### CLI and TUI

- Go 1.24 or later
- Git

### Desktop Application (Optional)

- Node.js 18+ and npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- macOS (required for reMarkable OCR via Apple Vision framework)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verify installation
wails doctor
```

## Using the Makefile

The Makefile provides convenient targets for common builds:

```bash
make all       # Build CLI + OCR tool
make cli       # Build CLI only
make ocr       # Build OCR tool (macOS only)
make desktop   # Build desktop app with OCR bundled
make dev       # Run desktop app in dev mode
make test      # Run all tests (Go + frontend)
make clean     # Remove build artifacts
```

## Building the CLI

```bash
git clone https://github.com/typingincolor/bujo.git
cd bujo
make cli
# or: go build -o bujo ./cmd/bujo
```

The binary includes both CLI and TUI modes:
- `./bujo` - CLI commands
- `./bujo tui` - Terminal UI

## Building the OCR Tool (macOS only)

The reMarkable import feature requires a Swift OCR tool that uses Apple's Vision framework:

```bash
make ocr
```

This compiles `tools/remarkable-ocr/main.swift` into a binary at `tools/remarkable-ocr/remarkable-ocr`. The CLI and desktop app automatically detect this binary at runtime.

## Building the Desktop Application

The desktop app uses [Wails](https://wails.io/) with a React frontend.

### Development Mode

```bash
# Install frontend dependencies
cd frontend && npm install && cd ..

# Build OCR tool + run in dev mode (hot reload)
make dev
```

### Production Build

```bash
# Build desktop app with OCR tool bundled into .app
make desktop

# Output: build/bin/bujoapp.app (macOS) with remarkable-ocr in Contents/MacOS/
```

### Frontend Only

For frontend development without the Go backend:

```bash
cd frontend
npm install
npm run dev      # Start dev server
npm run build    # Production build
npm run test     # Run tests
npm run lint     # Lint code
```

## Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/domain/...

# Frontend tests
cd frontend && npm run test
```

## For Local AI (Optional)

Install [Ollama](https://ollama.ai):

```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh
```

Pull a model:
```bash
ollama pull llama3.2:3b
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `BUJO_AI_ENABLED` | Enable AI features | `false` |
| `BUJO_AI_PROVIDER` | AI provider: `local` or `gemini` | `local` |
| `BUJO_MODEL` | Local AI model name | `llama3.2:3b` |
| `GEMINI_API_KEY` | API key for Gemini provider | (none) |
| `DB_PATH` | Database file location | `~/.bujo/bujo.db` |

See [AI Setup](docs/AI_SETUP.md) for detailed AI configuration.
