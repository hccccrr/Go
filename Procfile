# Procfile for Heroku deployment

# Build the Go binary and run it
web: go build -o shizumusic main.go && ./shizumusic

# Alternative: If you commit the binary
# worker: ./shizumusic

# With release phase (builds before running)
release: go build -ldflags="-s -w" -o shizumusic main.go
worker: ./shizumusic
