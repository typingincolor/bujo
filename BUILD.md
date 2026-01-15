# Building bujo

## Prerequisites

- Go 1.24 or later
- Git

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
ollama pull llama3.2:1b
```

## Building

```bash
git clone https://github.com/typingincolor/bujo.git
cd bujo
go build -o bujo ./cmd/bujo
```

## Running Tests

```bash
go test ./...
```

## Configuration

- `BUJO_MODEL` - Ollama model name (default: `llama3.2:1b`)
- `GEMINI_API_KEY` - Use Gemini instead of local AI
- `BUJO_AI_PROVIDER` - Force `local` or `gemini`
