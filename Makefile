.PHONY: all cli ocr desktop clean test check-tools lint

all: cli ocr

cli:
	go build -o bujo ./cmd/bujo

ocr:
ifeq ($(shell uname),Darwin)
	swiftc -O -o tools/remarkable-ocr/remarkable-ocr tools/remarkable-ocr/main.swift \
		-framework Vision -framework AppKit
else
	@echo "Skipping OCR build (macOS only)"
endif

desktop: ocr
	wails build
ifeq ($(shell uname),Darwin)
	cp tools/remarkable-ocr/remarkable-ocr build/bin/bujoapp.app/Contents/MacOS/remarkable-ocr
endif

dev: ocr
	wails dev

test:
	go test ./...
	cd frontend && npm run test

lint:
	go vet ./cmd/... ./internal/...
	golangci-lint run ./cmd/... ./internal/...

check-tools:
	@echo "Checking tool versions..."
	@failed=0; \
	GO_MOD_VER=$$(grep '^go ' go.mod | awk '{print $$2}'); \
	GO_LOCAL_VER=$$(go version | awk '{print $$3}' | sed 's/go//'); \
	GO_LOCAL_MAJOR=$$(echo $$GO_LOCAL_VER | cut -d. -f1-2); \
	GO_MOD_MAJOR=$$(echo $$GO_MOD_VER | cut -d. -f1-2); \
	printf "  go:             %-12s (go.mod: %s) " "$$GO_LOCAL_VER" "$$GO_MOD_VER"; \
	if [ "$$(printf '%s\n%s' "$$GO_MOD_MAJOR" "$$GO_LOCAL_MAJOR" | sort -V | head -1)" = "$$GO_MOD_MAJOR" ]; then \
		echo "✓"; \
	else \
		echo "✗ (local Go older than go.mod)"; failed=1; \
	fi; \
	if command -v golangci-lint >/dev/null 2>&1; then \
		LINT_VER=$$(golangci-lint version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
		LINT_MAJOR=$$(echo $$LINT_VER | cut -d. -f1); \
		printf "  golangci-lint:  %-12s (CI: v2 latest)    " "$$LINT_VER"; \
		if [ "$$LINT_MAJOR" -ge 2 ] 2>/dev/null; then \
			echo "✓"; \
		else \
			echo "✗ (need v2+: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest)"; failed=1; \
		fi; \
	else \
		echo "  golangci-lint:  NOT INSTALLED ✗"; \
		echo "    Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		failed=1; \
	fi; \
	if [ -d "frontend" ]; then \
		if command -v node >/dev/null 2>&1; then \
			NODE_VER=$$(node --version | sed 's/v//'); \
			NODE_MAJOR=$$(echo $$NODE_VER | cut -d. -f1); \
			printf "  node:           %-12s (CI: 22+)           " "$$NODE_VER"; \
			if [ "$$NODE_MAJOR" -ge 22 ] 2>/dev/null; then \
				echo "✓"; \
			else \
				echo "✗ (need Node 22+)"; failed=1; \
			fi; \
		else \
			echo "  node:           NOT INSTALLED ✗"; failed=1; \
		fi; \
	fi; \
	if command -v actionlint >/dev/null 2>&1; then \
		ACTIONLINT_VER=$$(actionlint --version 2>&1 | head -1); \
		printf "  actionlint:     %-12s " "$$ACTIONLINT_VER"; \
		echo "✓"; \
	else \
		echo "  actionlint:     not installed (optional, for workflow linting)"; \
	fi; \
	echo ""; \
	if [ "$$failed" -eq 1 ]; then \
		echo "❌ Some tools need updating"; exit 1; \
	else \
		echo "✅ All tools up to date"; \
	fi

clean:
	rm -f bujo
	rm -f tools/remarkable-ocr/remarkable-ocr
	rm -rf build/bin
