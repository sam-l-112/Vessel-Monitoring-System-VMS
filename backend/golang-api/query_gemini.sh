#!/bin/bash
# AI Query Script - 自動檢索歷史對話，保持在同一個 Gemini 頁面

QUERY="$1"
MAX_RETRIES=2

# 嘗試連接到現有的 Gemini 頁面
for i in $(seq 1 $MAX_RETRIES); do
    CURRENT=$(opencli operate get url 2>/dev/null || echo "")
    
    if echo "$CURRENT" | grep -q "gemini.google.com"; then
        break
    fi
    
    # 嘗試打開
    opencli operate open https://gemini.google.com/app/7aeec6192d00009f > /dev/null 2>&1
    sleep 6
    
    CURRENT=$(opencli operate get url 2>/dev/null || echo "")
    if echo "$CURRENT" | grep -q "gemini.google.com"; then
        break
    fi
done

# 等待頁面準備
sleep 2

# === 自動檢索歷史對話 ===
HISTORY=$(opencli operate eval "(function() {
  const history = [];
  const userQueries = document.querySelectorAll('user-query-content p');
  const responses = document.querySelectorAll('response-container');
  
  for (let i = 0; i < Math.min(userQueries.length, responses.length); i++) {
    const q = userQueries[i].textContent?.slice(0, 100);
    const r = responses[i].textContent?.slice(0, 200);
    if (q) history.push('Q: ' + q + ' | A: ' + r);
  }
  return history.slice(-3).join('\\n');
})()" 2>/dev/null)

# 如果有歷史，附加到問題前面
ENRICHED_QUERY="$QUERY"
if [ -n "$HISTORY" ] && [ "$HISTORY" != "undefined" ] && [ "$HISTORY" != "null" ]; then
    ENRICHED_QUERY="[參考最近對話]:
$HISTORY

[新問題]: $QUERY"
fi

# 輸入問題
opencli operate eval "document.querySelector('rich-textarea').querySelector('div[contenteditable]').innerText = '$ENRICHED_QUERY'" > /dev/null 2>&1
sleep 1

# 按 Enter 傳送
opencli operate keys Enter > /dev/null 2>&1

# 等待 Gemini 回覆
sleep 10

# 提取最新回應
RESULT=$(opencli operate eval "(function() {
  const msgs = document.querySelectorAll('response-container');
  if (msgs && msgs.length > 0) {
    return msgs[msgs.length - 1].textContent.slice(0,2500);
  }
  return 'no response';
})()" 2>/dev/null)

echo "$RESULT"
# 不要關閉瀏覽器