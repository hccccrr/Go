#!/bin/bash
# Install system dependencies

echo "📦 Installing system dependencies..."

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
else
    echo "Cannot detect OS"
    exit 1
fi

# Install based on OS
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    echo "Detected: $OS"
    
    # Update package list
    sudo apt update
    
    # Install FFmpeg
    echo "Installing FFmpeg..."
    sudo apt install -y ffmpeg
    
    # Install yt-dlp
    echo "Installing yt-dlp..."
    sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
        -o /usr/local/bin/yt-dlp
    sudo chmod a+rx /usr/local/bin/yt-dlp
    
    # Install MongoDB (optional)
    read -p "Install MongoDB locally? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo apt install -y mongodb
        sudo systemctl start mongodb
        sudo systemctl enable mongodb
    fi
    
elif [[ "$OS" == *"Arch"* ]]; then
    echo "Detected: Arch Linux"
    sudo pacman -S ffmpeg yt-dlp
    
elif [[ "$OS" == *"Fedora"* ]]; then
    echo "Detected: Fedora"
    sudo dnf install -y ffmpeg yt-dlp
    
else
    echo "Unsupported OS: $OS"
    echo "Please install FFmpeg and yt-dlp manually"
    exit 1
fi

echo ""
echo "✅ System dependencies installed!"
echo ""
echo "Next steps:"
echo "1. Install Go: https://go.dev/dl/"
echo "2. Configure: cp .env.example .env && nano .env"
echo "3. Run: bash start.sh"
