#!/bin/bash
set -e

echo "Building for all platforms..."
mkdir -p dist

# Build flags
LDFLAGS="-s -w -X main.Version=$(git describe --tags --always --dirty) -X main.Commit=$(git rev-parse HEAD) -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Build for all platforms
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/mvx-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/mvx-linux-arm64 .
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/mvx-darwin-amd64 .
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/mvx-darwin-arm64 .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/mvx-windows-amd64.exe .

echo "Built binaries:"
ls -la dist/
