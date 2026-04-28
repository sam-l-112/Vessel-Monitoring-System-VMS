#!/bin/bash
# AI Query Script

QUERY="$1"

# 打開頁面，確保完全載入
opencli operate open https://gemini.google.com/app > /dev/null 2>&1
sleep 15

# 確保頁面載入完成
opencli operate eval "document.querySelector('rich-textarea') ? 'ready' : 'not ready'" > /dev/null 2>&1
sleep 2

# 取得目前回覆數量
BEFORE=$(opencli operate eval "c=document.querySelectorAll('response-container'); c.length" 2>/dev/null)
echo "Before: $BEFORE"

# 輸入文字
opencli operate eval "x=document.querySelector('rich-textarea').querySelector('div[contenteditable]'); x.textContent='$QUERY'; x.dispatchEvent(new Event('input',{bubbles:true}))" > /dev/null 2>&1

sleep 3

# 送出
opencli operate keys Enter > /dev/null 2>&1

sleep 2

# 等待新回覆生成
sleep 25

# 取得新回覆 (倒數第1個)
RESULT=$(opencli operate eval "c=document.querySelectorAll('response-container'); c[c.length-1].textContent.slice(0,3500)" 2>/dev/null)
echo "$RESULT"