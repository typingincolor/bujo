.PHONY: all cli ocr desktop clean test

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

clean:
	rm -f bujo
	rm -f tools/remarkable-ocr/remarkable-ocr
	rm -rf build/bin
