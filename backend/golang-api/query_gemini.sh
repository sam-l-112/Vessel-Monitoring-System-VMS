#!/bin/bash
# AI Query Script - uses OpenCLI gemini CLI command (robust version)

QUERY="$1"
OPENCLI="/home/ouo/OpenCLI/dist/src/main.js"

# Kill any existing browser page and start fresh
cd /home/ouo/OpenCLI && node "$OPENCLI" browser close 2>/dev/null
sleep 1

# Start new conversation
cd /home/ouo/OpenCLI && node "$OPENCLI" gemini new 2>&1
sleep 5

# Ask the question
RESULT=$(cd /home/ouo/OpenCLI && node "$OPENCLI" gemini ask "$QUERY" 2>&1)

echo "$RESULT"