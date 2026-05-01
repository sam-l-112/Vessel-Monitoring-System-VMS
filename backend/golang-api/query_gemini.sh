#!/bin/bash
# AI Query Script - uses OpenCLI gemini CLI

QUERY="$1"
OPENCLI="/home/ouo/OpenCLI/dist/src/main.js"
cd /home/ouo/OpenCLI

# 直接問問題 不關閉瀏覽器
node "$OPENCLI" gemini ask "$QUERY" 2>&1