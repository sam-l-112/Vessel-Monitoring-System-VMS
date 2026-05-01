#!/bin/bash

# Simple test script - just verify OpenCLI can communicate with Chrome
# Path: /home/ouo/project_f/backend/golang-api/scripts/test-opencli.sh

cd ~/OpenCLI

echo "Testing OpenCLI connection..."

# Test 1: Check doctor
echo "1. Checking OpenCLI doctor..."
node dist/src/main.js doctor 2>&1 | head -5

# Test 2: Open a simple page
echo ""
echo "2. Opening Google..."
node dist/src/main.js browser open https://www.google.com 2>&1

# Test 3: Wait and get title
echo ""
echo "3. Getting page title..."
node dist/src/main.js browser wait time 2
node dist/src/main.js browser get title

# Test 4: Check Chrome process
echo ""
echo "4. Chrome processes:"
ps aux | grep chrome | grep -v grep | wc -l

echo ""
echo "=== Test Complete ==="