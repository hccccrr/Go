#!/bin/bash
# Build script for multiple platforms

echo "🔨 Building ShizuMusic for multiple platforms..."

# Build flags
LDFLAGS="-s -w"

# Linux AMD64
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o bin/shizumusic-linux-amd64 main.go

# Linux ARM64 (for Raspberry Pi, etc.)
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o bin/shizumusic-linux-arm64 main.go

# Windows AMD64
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o bin/shizumusic-windows-amd64.exe main.go

# macOS AMD64
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o bin/shizumusic-darwin-amd64 main.go

# macOS ARM64 (M1/M2)
echo "Building for macOS (arm64 - M1/M2)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o bin/shizumusic-darwin-arm64 main.go

echo ""
echo "✅ Build complete! Binaries in bin/ directory:"
ls -lh bin/
