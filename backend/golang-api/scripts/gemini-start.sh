#!/bin/bash
# Gemini OpenCLI Startup Script
# Path: /home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh

cd ~/OpenCLI

echo "Starting Gemini OpenCLI Service..."

# Start OpenCLI daemon if not running
if ! pgrep -f "opencli" > /dev/null; then
    echo "Starting OpenCLI daemon..."
    node dist/src/main.js daemon start 2>/dev/null
    sleep 3
fi

# Ensure Chrome is running with remote debugging
if ! pgrep -f "chrome.*--remote-debugging-port" > /dev/null; then
    echo "Starting Chrome with remote debugging..."
    google-chrome \
        --remote-debugging-port=9222 \
        --user-data-dir=$HOME/.config/google-chrome \
        --no-first-run \
        --no-default-browser-check \
        --disable-gpu \
        --disable-software-rasterizer \
        --disable-dev-shm-usage \
        --disable-background-networking \
        --disable-default-apps \
        --disable-extensions \
        --disable-sync \
        --hide-scrollbars \
        --mute-audio \
        --window-size=1920,1080 \
        > /dev/null 2>&1 &
    sleep 5
fi

# Verify setup
if node dist/src/main.js doctor 2>&1 | grep -q "Everything looks good"; then
    echo "✓ Gemini OpenCLI service is ready"
else
    echo "⚠ Warning: Some checks failed"
fi