# Building bujo

This guide covers building the CLI, TUI, and desktop application.

## Prerequisites

### CLI and TUI

- Go 1.24 or later
- Git

### Desktop Application (Optional)

- Node.js 18+ and npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verify installation
wails doctor
```

## Building the CLI

```bash
git clone https://github.com/typingincolor/bujo.git
cd bujo
go build -o bujo ./cmd/bujo
```

The binary includes both CLI and TUI modes:
- `./bujo` - CLI commands
- `./bujo tui` - Terminal UI

## Building the Desktop Application

The desktop app uses [Wails](https://wails.io/) with a React frontend.

### Development Mode

```bash
# Install frontend dependencies
cd frontend && npm install && cd ..

# Run in development mode (hot reload)
wails dev
```

### Production Build

```bash
# Build optimized application
wails build

# Output: build/bin/bujoapp (macOS) or build/bin/bujoapp.exe (Windows)
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
