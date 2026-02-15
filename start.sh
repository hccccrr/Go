#!/bin/bash
# ShizuMusic Go Bot - Startup Script
# Usage: bash start.sh

set -e  # Exit on error

echo "🎵 ShizuMusic Bot - Go Edition"
echo "================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed!"
    echo "Please install Go 1.21+ from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
print_success "Go detected: $GO_VERSION"

# Check if .env exists
if [ ! -f ".env" ]; then
    print_error ".env file not found!"
    echo "Please create .env file:"
    echo "  cp .env.example .env"
    echo "  nano .env  # Add your credentials"
    exit 1
fi

print_success ".env file found"

# Create required directories
print_info "Creating required directories..."
mkdir -p downloads cache logs
print_success "Directories created"

# Download dependencies
print_info "Downloading Go modules..."
go mod download
if [ $? -eq 0 ]; then
    print_success "Dependencies downloaded"
else
    print_error "Failed to download dependencies"
    exit 1
fi

# Build the binary
print_info "Building ShizuMusic..."
go build -ldflags="-s -w" -o shizumusic main.go

if [ $? -eq 0 ]; then
    print_success "Build successful!"
    
    # Get binary size
    BINARY_SIZE=$(du -h shizumusic | cut -f1)
    echo "   Binary size: $BINARY_SIZE"
else
    print_error "Build failed!"
    exit 1
fi

echo ""
echo "================================"
echo "🚀 Starting ShizuMusic Bot..."
echo "================================"
echo ""

# Run the bot
./shizumusic

# Exit code handling
EXIT_CODE=$?
if [ $EXIT_CODE -eq 0 ]; then
    print_success "Bot stopped gracefully"
else
    print_error "Bot stopped with error code: $EXIT_CODE"
    exit $EXIT_CODE
fi
