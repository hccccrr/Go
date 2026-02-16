#!/bin/bash
# Add 2GB swap memory

echo "🔧 Adding swap memory..."

# Check if swap already exists
if [ $(swapon --show | wc -l) -gt 0 ]; then
    echo "⚠️ Swap already exists"
    swapon --show
    exit 0
fi

# Create swap file
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make it permanent
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Verify
echo "✅ Swap added successfully!"
free -h
