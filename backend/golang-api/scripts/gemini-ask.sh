#!/bin/bash

# Gemini OpenCLI Chat Script - Working Version
# Path: /home/ouo/project_f/backend/golang-api/scripts/gemini-ask.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd ~/OpenCLI

MSG="$1"
if [ -z "$MSG" ]; then
  echo '{"success":false,"error":"Message is empty"}'
  exit 1
fi

output_json() {
  echo "{\"success\":\"$1\",\"reply\":\"$2\",\"error\":\"$3\"}"
}

# Step 1: Open Gemini if not already open
node dist/src/main.js browser open https://gemini.google.com/app > /dev/null 2>&1
node dist/src/main.js browser wait time 4 > /dev/null 2>&1

# Create temp JS file
TEMP_JS=$(mktemp /tmp/gemini.XXXXXX.js)
cat > "$TEMP_JS" << 'JSEOF'
(async function() {
  const msg = "MSG_PLACEHOLDER";
  
  // Find input
  const input = document.querySelector('div[contenteditable=true][role=textbox], div[contenteditable=true][aria-label*=請輸入]');
  if (!input) return JSON.stringify({success: false, error: 'Input not found'});
  
  input.textContent = msg;
  input.dispatchEvent(new Event('input', {bubbles: true}));
  
  // Wait for button
  await new Promise(r => setTimeout(r, 2500));
  
  // Click send
  const btn = document.querySelector('button[aria-label*=傳送]');
  if (!btn || btn.disabled) return JSON.stringify({success: false, error: 'Send button not ready'});
  btn.click();
  
  // Wait for response
  await new Promise(r => setTimeout(r, 25000));
  
  // Get response - use title as summary
  const title = document.title.split(' - Google Gemini')[0];
  
  return JSON.stringify({success: true, reply: title || 'Response received', url: window.location.href});
})()
JSEOF

# Replace placeholder
sed -i "s/MSG_PLACEHOLDER/$MSG/" "$TEMP_JS"

# Execute
RESULT=$(node dist/src/main.js browser eval "$(cat "$TEMP_JS")" 2>&1)
rm -f "$TEMP_JS"

# Parse result
if echo "$RESULT" | grep -q '"success":true'; then
  REPLY=$(echo "$RESULT" | grep -oP '"reply":\s*"\K[^"]+' | head -1)
  if [ -n "$REPLY" ]; then
    output_json true "$REPLY" ""
  else
    output_json true "Response received" ""
  fi
else
  ERROR=$(echo "$RESULT" | grep -oP '"error":\s*"\K[^"]+' | head -1)
  output_json false "" "${ERROR:-Unknown error}"
fi

exit 0