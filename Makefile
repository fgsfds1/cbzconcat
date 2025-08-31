# Makefile for cbztools
# Build with version information from git tags

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%m:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags for version injection
VERSION_FLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

# Default target
.PHONY: all
all: build

# Install dependencies
.PHONY: deps
deps:
	go mod tidy

# Build the binary with version information (local development)
.PHONY: build
build:
	@echo "Building cbztools v$(VERSION) (commit: $(GIT_COMMIT))"
	make deps
	mkdir -p build
	go build -ldflags "$(VERSION_FLAGS)" -o build/cbztools

# Run the binary
.PHONY: run
run:
	make build
	./build/cbztools

# Build all platforms like release workflow
.PHONY: release
release:
	@echo "Building all platforms like release workflow"
	make deps
	mkdir -p build
	# Linux amd64
	go build -ldflags "$(VERSION_FLAGS) -s -w" -v -o build/cbztools-$(VERSION).linux.amd64
	# Linux x86
	go build -ldflags "$(VERSION_FLAGS) -s -w" -v -o build/cbztools-$(VERSION).linux.x86
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(VERSION_FLAGS) -s -w" -v -o build/cbztools-$(VERSION).win_amd64.exe
	# Windows x86
	GOOS=windows GOARCH=386 go build -ldflags "$(VERSION_FLAGS) -s -w" -v -o build/cbztools-$(VERSION).win_x86.exe

# Clean build artifacts
.PHONY: clean
clean:
	rm -f cbztools
	rm -rf build/

# Run tests
.PHONY: test
test:
	make deps
	go test -v

# Show current version info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build cbztools with version info (local)"
	@echo "  release       - Build all platforms like release workflow"
	@echo "  install       - Install to system"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  version       - Show version information"
	@echo "  help          - Show this help" 