# ==============================================================================
# River Raid - 21st Century Tactical Strike Makefile
# ==============================================================================

BINARY_NAME ?= riverraid
MODULE_NAME ?= riverraid
GO ?= go

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	OS_TYPE := macos
	CC := clang
else ifeq ($(UNAME_S),Linux)
	OS_TYPE := linux
	CC ?= gcc
else
	OS_TYPE := windows
	CC ?= gcc
endif

CGO_ENABLED ?= 1

# Build flags
LDFLAGS := -s -w
BUILD_ENV := CC=$(CC) CGO_ENABLED=$(CGO_ENABLED)

.PHONY: all help build run clean test vet fmt tidy deps check package-mac dist

all: build ## Default target: build the game executable

help: ## Show this help message
	@echo "=================================================================="
	@echo " River Raid - Build & Development Commands"
	@echo "=================================================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
	@echo "=================================================================="

deps: ## Download and verify module dependencies
	@echo "==> Downloading and verifying dependencies..."
	$(GO) mod download
	$(GO) mod verify

tidy: ## Tidy Go module dependencies
	@echo "==> Tidying dependencies in go.mod and go.sum..."
	$(GO) mod tidy

fmt: ## Format Go source code
	@echo "==> Formatting Go files..."
	$(GO) fmt ./...

vet: ## Run Go static analysis (vet)
	@echo "==> Running go vet..."
	$(BUILD_ENV) $(GO) vet ./...

check: fmt vet ## Run formatting and static code checks

test: ## Run unit and integration tests
	@echo "==> Running tests..."
	$(BUILD_ENV) $(GO) test -v ./...

build: ## Compile game binary
	@echo "==> Building $(BINARY_NAME) for $(OS_TYPE)..."
	$(BUILD_ENV) $(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "==> Build complete: $(BINARY_NAME)"

run: build ## Build and immediately launch the game
	@echo "==> Launching $(BINARY_NAME)..."
	./$(BINARY_NAME)

clean: ## Clean up compiled binaries and caches
	@echo "==> Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf dist/
	@echo "==> Clean complete."

package-mac: build ## Package macOS .app application bundle (macOS only)
ifeq ($(OS_TYPE),macos)
	@echo "==> Creating macOS App Bundle (RiverRaid.app)..."
	@mkdir -p dist/RiverRaid.app/Contents/MacOS
	@mkdir -p dist/RiverRaid.app/Contents/Resources
	@cp $(BINARY_NAME) dist/RiverRaid.app/Contents/MacOS/RiverRaid
	@if [ -d assets ]; then cp -r assets dist/RiverRaid.app/Contents/Resources/; fi
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > dist/RiverRaid.app/Contents/Info.plist
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> dist/RiverRaid.app/Contents/Info.plist
	@echo '<plist version="1.0"><dict>' >> dist/RiverRaid.app/Contents/Info.plist
	@echo '  <key>CFBundleExecutable</key><string>RiverRaid</string>' >> dist/RiverRaid.app/Contents/Info.plist
	@echo '  <key>CFBundleIdentifier</key><string>com.riverraid.game</string>' >> dist/RiverRaid.app/Contents/Info.plist
	@echo '  <key>CFBundleName</key><string>River Raid</string>' >> dist/RiverRaid.app/Contents/Info.plist
	@echo '  <key>CFBundlePackageType</key><string>APPL</string>' >> dist/RiverRaid.app/Contents/Info.plist
	@echo '  <key>CFBundleShortVersionString</key><string>1.0.0</string>' >> dist/RiverRaid.app/Contents/Info.plist
	@echo '</dict></plist>' >> dist/RiverRaid.app/Contents/Info.plist
	@echo "==> Packaged at dist/RiverRaid.app"
else
	@echo "package-mac target is only supported on macOS."
endif

dist: clean ## Build and package for all platforms (macOS, Linux, Windows)
	@echo "==> Preparing distribution packages..."
	@mkdir -p dist/windows dist/linux dist/macos
	
	@echo "==> Building for Windows (amd64)..."
	@# Note: Requires x86_64-w64-mingw32-gcc for cross-compilation
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc $(GO) build -ldflags="$(LDFLAGS)" -o dist/windows/$(BINARY_NAME).exe . || echo "Skipping Windows build (compiler not found)"
	@if [ -f dist/windows/$(BINARY_NAME).exe ]; then cp -r assets dist/windows/; fi
	
	@echo "==> Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GO) build -ldflags="$(LDFLAGS)" -o dist/linux/$(BINARY_NAME) . || echo "Skipping Linux build (compiler not found)"
	@if [ -f dist/linux/$(BINARY_NAME) ]; then cp -r assets dist/linux/; fi
	
	@echo "==> Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GO) build -ldflags="$(LDFLAGS)" -o dist/macos/$(BINARY_NAME) . || echo "Skipping macOS build (compiler not found)"
	@if [ -f dist/macos/$(BINARY_NAME) ]; then cp -r assets dist/macos/; fi
	
	@echo "==> Creating archives..."
	@if [ -d dist/windows ]; then cd dist/windows && tar -czf ../riverraid-windows-amd64.tar.gz *; fi
	@if [ -d dist/linux ]; then cd dist/linux && tar -czf ../riverraid-linux-amd64.tar.gz *; fi
	@if [ -d dist/macos ]; then cd dist/macos && tar -czf ../riverraid-macos-amd64.tar.gz *; fi
	
	@echo "==> Distribution builds complete. Files are in dist/"
