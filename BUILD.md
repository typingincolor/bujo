# Building bujo

This guide explains how to build bujo from source.

## Prerequisites

### Required
- Go 1.24 or later
- Git

### For Local AI Support (Recommended)
- C compiler (Xcode Command Line Tools on macOS, GCC on Linux)
- llama.cpp library

## Installing Dependencies

### macOS

```bash
# Install Xcode Command Line Tools
xcode-select --install

# Install llama.cpp
brew install llama.cpp
```

### Linux (Ubuntu/Debian)

```bash
# Install build tools
sudo apt-get update
sudo apt-get install -y build-essential git

# Build and install llama.cpp
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make
sudo make install
```

### Windows

Windows builds do not support local AI. Use Gemini API instead by setting `GEMINI_API_KEY`.

## Building

### With Local AI Support (CGO Enabled)

```bash
# Clone the repository
git clone https://github.com/typingincolor/bujo.git
cd bujo

# Build with CGO
CGO_ENABLED=1 go build -tags=cgo -o bujo ./cmd/bujo

# Install to GOPATH/bin
CGO_ENABLED=1 go install -tags=cgo ./cmd/bujo
```

### Without Local AI (CGO Disabled)

```bash
# Build without CGO (Gemini only)
CGO_ENABLED=0 go build -o bujo ./cmd/bujo

# Install to GOPATH/bin
CGO_ENABLED=0 go install ./cmd/bujo
```

## Running Tests

Tests run with CGO disabled by default for compatibility:

```bash
# Run all tests
CGO_ENABLED=0 go test ./...

# Run with coverage
CGO_ENABLED=0 go test -coverprofile=coverage.out ./...

# Run with race detector
CGO_ENABLED=0 go test -race ./...
```

## Development Workflow

### Pre-push Checks

The repository includes a pre-push hook that runs:
1. `gofmt` formatting check
2. `go test` with CGO_ENABLED=0
3. `go vet` static analysis
4. `golangci-lint` linting

All checks run with CGO disabled for consistency.

### Running Locally

```bash
# Run without installing
CGO_ENABLED=1 go run ./cmd/bujo --help

# Use a test database
CGO_ENABLED=1 go run ./cmd/bujo --db-path ./test.db add ". Test entry"
```

## Build Tags

bujo uses build tags to handle CGO dependencies:

- `+build cgo`: Full local AI support with llama.cpp
- `+build !cgo`: Stub implementation that returns helpful error

The build system automatically selects the correct implementation based on `CGO_ENABLED`.

## CI/CD

GitHub Actions builds platform-specific binaries:

- **macOS (amd64/arm64)**: CGO enabled with llama.cpp
- **Linux (amd64/arm64)**: CGO enabled with llama.cpp
- **Windows (amd64/arm64)**: CGO disabled (no local AI)

See `.github/workflows/ci.yml` and `.goreleaser.yaml` for details.

## Troubleshooting

### "common.h file not found"

llama.cpp is not installed or not in the include path:
- macOS: `brew install llama.cpp`
- Linux: Build from source and `sudo make install`

### "undefined: llama"

Building with CGO disabled but no build tags:
- Use `CGO_ENABLED=0 go build` for Gemini-only build
- Use `CGO_ENABLED=1 go build -tags=cgo` for full build

### Tests fail with CGO errors

Tests should always run with `CGO_ENABLED=0`:
```bash
CGO_ENABLED=0 go test ./...
```

## Platform-Specific Notes

### macOS
- Homebrew installation includes llama.cpp dependency
- Supports Metal acceleration for faster inference
- Both Intel and Apple Silicon are supported

### Linux
- llama.cpp must be built from source
- Distribution packages may not include Go bindings
- ARM64 builds supported but less tested

### Windows
- No local AI support (llama.cpp requires MSVC/MinGW)
- Use Gemini API with `GEMINI_API_KEY`
- Pure Go build works without C compiler
