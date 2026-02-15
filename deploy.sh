#!/bin/bash
# Deploy as systemd service

if [ "$EUID" -eq 0 ]; then
    echo "Don't run this script as root!"
    exit 1
fi

SERVICE_NAME="shizumusic"
WORK_DIR=$(pwd)
USER=$(whoami)

echo "📦 Creating systemd service..."

# Build the binary first
go build -ldflags="-s -w" -o shizumusic main.go

# Create service file
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=ShizuMusic Telegram Bot
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$WORK_DIR
ExecStart=$WORK_DIR/shizumusic
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
sudo systemctl daemon-reload

# Enable and start service
sudo systemctl enable $SERVICE_NAME
sudo systemctl start $SERVICE_NAME

echo ""
echo "✅ Service deployed!"
echo ""
echo "Useful commands:"
echo "  sudo systemctl status $SERVICE_NAME   # Check status"
echo "  sudo systemctl restart $SERVICE_NAME  # Restart"
echo "  sudo systemctl stop $SERVICE_NAME     # Stop"
echo "  sudo journalctl -u $SERVICE_NAME -f   # View logs"
