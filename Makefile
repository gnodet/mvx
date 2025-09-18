# mvx Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -ldflags "-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"
STATIC_FLAGS = CGO_ENABLED=0

# Default target
.PHONY: all
all: build

# Build the binary (static)
.PHONY: build
build:
	$(STATIC_FLAGS) go build $(LDFLAGS) -o mvx-binary .
	@echo "Installing global development binary..."
	@mkdir -p ~/.mvx/dev
	@cp mvx-binary ~/.mvx/dev/mvx
	@chmod +x ~/.mvx/dev/mvx
	@echo "✅ Global development binary installed at ~/.mvx/dev/mvx"
	@echo "   All projects using 'mvxVersion=dev' will now use this binary automatically"

# Build for multiple platforms (static binaries)
.PHONY: build-all
build-all:
	@mkdir -p dist
	$(STATIC_FLAGS) GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/mvx-linux-amd64 .
	$(STATIC_FLAGS) GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/mvx-linux-arm64 .
	$(STATIC_FLAGS) GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/mvx-darwin-amd64 .
	$(STATIC_FLAGS) GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/mvx-darwin-arm64 .
	$(STATIC_FLAGS) GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/mvx-windows-amd64.exe .
	@echo "Built binaries:"
	@ls -la dist/

# Run tests
.PHONY: test
test:
	go test ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -f mvx-binary mvx-dev
	rm -rf dist/
	rm -rf .mvx/local/ .mvx/tools/ .mvx/versions/

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Development build with race detection (dynamic linking for race detector)
.PHONY: dev
dev:
	go build -race $(LDFLAGS) -o mvx-dev .
	@echo "Installing global development binary..."
	@mkdir -p ~/.mvx/dev
	@cp mvx-dev ~/.mvx/dev/mvx
	@chmod +x ~/.mvx/dev/mvx
	@echo "✅ Global development binary installed at ~/.mvx/dev/mvx"
	@echo "   All projects using 'mvxVersion=dev' will now use this binary automatically"



# Deploy local binary to another project
.PHONY: deploy
deploy:
	@if [ -z "$(TARGET)" ]; then \
		echo "Usage: make deploy TARGET=/path/to/project [METHOD=copy|symlink|wrapper]"; \
		echo "Example: make deploy TARGET=~/projects/my-app METHOD=symlink"; \
		exit 1; \
	fi
	@./scripts/deploy-local.sh "$(TARGET)" "$(METHOD)"

# Install locally
.PHONY: install
install:
	go install $(LDFLAGS) .

# Generate checksums for release binaries
.PHONY: checksums
checksums:
	@echo "Generating checksums..."
	@cd dist && for file in mvx-*; do \
		if [ -f "$$file" ] && [ "$${file##*.}" != "sha256" ]; then \
			sha256sum "$$file" > "$$file.sha256"; \
			echo "Generated checksum for $$file"; \
		fi \
	done
	@cd dist && cat *.sha256 > checksums.txt
	@echo "Combined checksums in dist/checksums.txt"

# Build all platforms and generate checksums
.PHONY: release-build
release-build: build-all checksums

# Package mvx files
.PHONY: package-mvx
package-mvx:
	@echo "Packaging mvx files..."
	@mkdir -p dist/mvx
	@cp mvx dist/mvx/
	@cp mvx.cmd dist/mvx/
	@cp -r .mvx dist/mvx/
	@cp WRAPPER.md dist/mvx/
	@cp install-mvx.sh dist/mvx/
	@echo "mvx files packaged in dist/mvx/"

# Test mvx functionality
.PHONY: test-mvx
test-mvx: build
	@echo "Testing mvx..."
	@cp mvx-binary ./mvx.exe 2>/dev/null || cp mvx-binary ./mvx-binary-local
	@./mvx version || echo "mvx test completed"
	@rm -f ./mvx.exe ./mvx-binary-local

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary (static)"
	@echo "  deploy         - Deploy local binary to another project (requires TARGET=path)"
	@echo "  build-all      - Build for multiple platforms (static)"
	@echo "  release-build  - Build all platforms and generate checksums"
	@echo "  checksums      - Generate checksums for dist/ binaries"
	@echo "  test           - Run tests"
	@echo "  test-mvx       - Test mvx functionality"
	@echo "  package-mvx    - Package mvx files for distribution"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  dev            - Development build with race detection (creates mvx-dev)"
	@echo "  install        - Install locally"
	@echo "  help           - Show this help"
