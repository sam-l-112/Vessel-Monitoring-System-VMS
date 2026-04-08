# Air 熱重載工具安裝與使用指南

## 概述
Air 是一個 Go 語言的熱重載工具，能夠在開發過程中自動監控程式碼變化並重新編譯、重啟應用，提升開發效率。

## 安裝步驟

### 1. 安裝 Air
```bash
go install github.com/air-verse/air@latest
```

### 2. 驗證安裝
```bash
~/go/bin/air -v
```
應該顯示類似以下輸出：
```
air v1.65.0
```

## 配置設置

### 1. 初始化配置
進入後端目錄並初始化 air 配置：
```bash
cd /home/ouo/project_f/backend/golang-api
~/go/bin/air init
```

### 2. 手動創建 .air.toml 配置
如果初始化失敗，可以手動創建 `.air.toml` 文件：

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

## 啟動指令

### 基本啟動
```bash
cd /home/ouo/project_f/backend/golang-api
~/go/bin/air
```

### 背景運行
```bash
cd /home/ouo/project_f/backend/golang-api
~/go/bin/air &
```

## 使用說明

### 1. 啟動後
- Air 會自動編譯並啟動 Go 應用
- 監聽文件變化（.go 文件）
- 當檢測到變化時，自動重新編譯並重啟

### 2. 停止服務
- 在終端按 `Ctrl+C` 停止 air
- 如果背景運行，使用 `pkill air` 或找到進程 ID 並殺死

### 3. 測試 API
啟動後，可以測試登入 API：
```bash
curl -X POST http://192.168.50.75/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

### 4. 開發流程
1. 修改 Go 程式碼
2. 保存文件
3. Air 自動重載應用
4. 測試更改

## 故障排除

### 端口衝突
如果遇到 "address already in use" 錯誤：
```bash
# 檢查使用端口的進程
lsof -i :8080

# 殺死進程（替換 PID）
kill <PID>
```

### Air 命令找不到
確保 air 已正確安裝並在 PATH 中，或使用完整路徑：
```bash
~/go/bin/air
```

### 配置問題
如果 .air.toml 配置有問題，刪除並重新初始化：
```bash
rm .air.toml
~/go/bin/air init
```

## 注意事項
- Air 僅監控 .go 文件變化
- 對於靜態文件或配置變化，需要手動重啟
- 生產環境不建議使用 air，使用編譯後的二進制文件