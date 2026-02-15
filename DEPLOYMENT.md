# 🚀 ShizuMusic Deployment Guide

Complete guide for deploying ShizuMusic Go Bot on various platforms.

---

## 📋 **Prerequisites:**

### **Required:**
- Go 1.21+
- FFmpeg
- yt-dlp
- MongoDB

### **Optional:**
- systemd (for service deployment)
- Docker (for containerized deployment)

---

## 🛠️ **Quick Setup:**

### **1. Install Dependencies:**
```bash
# Run automated installer
bash install.sh

# Or manually:
# Ubuntu/Debian
sudo apt install ffmpeg
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp
```

### **2. Configure:**
```bash
# Copy environment template
cp .env.example .env

# Edit with your credentials
nano .env
```

### **3. Run:**
```bash
# Simple start
bash start.sh

# Or manual
go build -o shizumusic main.go
./shizumusic
```

---

## 📦 **Deployment Methods:**

### **Method 1: Direct Run (Development)**

Best for: Testing, development

```bash
# Quick start
bash start.sh

# With auto-restart on crash
while true; do
    ./shizumusic
    echo "Bot crashed! Restarting in 5 seconds..."
    sleep 5
done
```

---

### **Method 2: Systemd Service (Production)**

Best for: VPS, dedicated servers

```bash
# Automated deployment
bash deploy.sh

# Manual deployment:
# 1. Build binary
go build -ldflags="-s -w" -o shizumusic main.go

# 2. Create service file
sudo nano /etc/systemd/system/shizumusic.service
```

Service file content:
```ini
[Unit]
Description=ShizuMusic Telegram Bot
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/ShizuMusic
ExecStart=/path/to/ShizuMusic/shizumusic
Restart=always
RestartSec=10

# Optional: Environment variables
Environment="API_ID=12345"
Environment="API_HASH=xxx"
# Or use EnvironmentFile=/path/to/.env

[Install]
WantedBy=multi-user.target
```

```bash
# 3. Enable and start
sudo systemctl daemon-reload
sudo systemctl enable shizumusic
sudo systemctl start shizumusic

# 4. Check status
sudo systemctl status shizumusic

# 5. View logs
sudo journalctl -u shizumusic -f
```

---

### **Method 3: Docker (Containerized)**

Best for: Cloud platforms, Kubernetes

**Dockerfile:**
```dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN go build -ldflags="-s -w" -o shizumusic main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ffmpeg ca-certificates

# Install yt-dlp
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
    -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

# Create app user
RUN adduser -D -u 1000 shizu
USER shizu

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/shizumusic .

# Create directories
RUN mkdir -p downloads cache logs

# Run bot
CMD ["./shizumusic"]
```

**Build and run:**
```bash
# Build image
docker build -t shizumusic .

# Run container
docker run -d \
  --name shizumusic \
  --env-file .env \
  --restart unless-stopped \
  shizumusic

# View logs
docker logs -f shizumusic

# Stop
docker stop shizumusic

# Remove
docker rm shizumusic
```

---

### **Method 4: Heroku**

Best for: Free hosting, beginners

**Using Procfile:**
```
release: go build -ldflags="-s -w" -o shizumusic main.go
worker: ./shizumusic
```

**Deploy:**
```bash
# Install Heroku CLI
# https://devcenter.heroku.com/articles/heroku-cli

# Login
heroku login

# Create app
heroku create your-app-name

# Add buildpack
heroku buildpacks:set heroku/go

# Set config vars
heroku config:set API_ID=12345
heroku config:set API_HASH=xxx
heroku config:set BOT_TOKEN=xxx
# ... set all required variables

# Deploy
git add .
git commit -m "Deploy ShizuMusic"
git push heroku main

# Scale worker
heroku ps:scale worker=1

# View logs
heroku logs --tail
```

---

### **Method 5: PM2 (Process Manager)**

Best for: Multiple bots, easy management

```bash
# Install PM2
npm install -g pm2

# Create ecosystem.config.js
cat > ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'shizumusic',
    script: './shizumusic',
    cwd: '/path/to/ShizuMusic',
    interpreter: 'none',
    instances: 1,
    autorestart: true,
    watch: false,
    max_memory_restart: '200M',
    env: {
      NODE_ENV: 'production'
    }
  }]
}
EOF

# Start with PM2
pm2 start ecosystem.config.js

# Save PM2 config
pm2 save

# Setup PM2 startup
pm2 startup

# Useful commands:
pm2 status              # Check status
pm2 logs shizumusic     # View logs
pm2 restart shizumusic  # Restart
pm2 stop shizumusic     # Stop
pm2 delete shizumusic   # Remove
```

---

## 🏗️ **Build for Multiple Platforms:**

```bash
# Use build script
bash build.sh

# Or manually:

# Linux (AMD64)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shizumusic-linux main.go

# Linux (ARM64) - Raspberry Pi
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o shizumusic-arm main.go

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o shizumusic.exe main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o shizumusic-mac main.go

# macOS (M1/M2)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o shizumusic-mac-arm main.go
```

---

## 🔧 **Environment Variables:**

Required in `.env` or system environment:

```env
# Required
API_ID=12345
API_HASH=your_hash
BOT_TOKEN=123456:ABC-DEF
DATABASE_URL=mongodb://localhost:27017
STRING_SESSION=gogram_session
LOGGER_ID=-1001234567890
OWNER_ID=123456789

# Optional
BOT_NAME=ShizuMusic
PLAY_LIMIT=0
MAX_FAVORITES=30
```

---

## 📊 **Monitoring:**

### **System Resources:**
```bash
# CPU and Memory usage
top -p $(pgrep shizumusic)

# Detailed stats
htop

# Binary size
ls -lh shizumusic
```

### **Logs:**
```bash
# Systemd
sudo journalctl -u shizumusic -f

# Docker
docker logs -f shizumusic

# PM2
pm2 logs shizumusic

# File logs
tail -f ShizuMusic.log
```

---

## 🆘 **Troubleshooting:**

### **Bot not starting:**
```bash
# Check Go version
go version  # Should be 1.21+

# Check dependencies
which ffmpeg
which yt-dlp

# Rebuild
go clean
go build -o shizumusic main.go

# Check logs
cat ShizuMusic.log
```

### **Can't connect to database:**
```bash
# Test MongoDB connection
mongosh $DATABASE_URL

# Check if MongoDB is running
sudo systemctl status mongodb
```

### **Permission errors:**
```bash
# Fix binary permissions
chmod +x shizumusic

# Fix directory permissions
chmod 755 downloads cache logs
```

---

## 🎯 **Performance Tips:**

1. **Compile with optimizations:**
   ```bash
   go build -ldflags="-s -w" -o shizumusic main.go
   ```

2. **Use production MongoDB:**
   - MongoDB Atlas (free tier available)
   - Local optimized instance

3. **Enable caching:**
   - Downloads cache
   - User data cache

4. **Monitor resources:**
   - Set up alerts
   - Regular log rotation

---

## 📝 **Update Process:**

```bash
# Pull latest code
git pull

# Rebuild
go build -ldflags="-s -w" -o shizumusic main.go

# Restart service
sudo systemctl restart shizumusic
# Or
pm2 restart shizumusic
# Or
docker restart shizumusic
```

---

**Your bot is ready to deploy!** 🚀
