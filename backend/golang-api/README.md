# VMS Golang API

養殖監控系統 (Vessel Monitoring System) 的後端 API，使用 Golang 開發。

## 功能特色

- 🔐 **用戶認證**: JWT-based 登入系統
- 📊 **RESTful API**: 標準 REST API 設計
- 🛡️ **安全中間件**: CORS 和日誌中間件
- 📝 **完整文檔**: 清晰的 API 文檔

## 快速開始

### 環境需求

- Go 1.19+ (推薦 1.21+)
- Git

### 安裝依賴

```bash
# 初始化 Go 模組 (如果尚未初始化)
go mod init vms-api

# 下載依賴
go mod tidy
```

### 運行伺服器

```bash
# 開發模式
go run main.go

# 或使用 Air 進行熱重載 (推薦)
air
```

伺服器將在 `http://localhost:8080` 啟動。

## API 端點

### 公開端點

#### 健康檢查
```http
GET /api/health
```

回應:
```json
{
  "success": true,
  "message": "VMS API is running",
  "data": {
    "timestamp": "2024-01-01T12:00:00Z",
    "version": "1.0.0"
  }
}
```

#### 用戶登入
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

成功回應:
```json
{
  "success": true,
  "token": "vms_token_admin_1704110400",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@vms.com",
    "role": "administrator"
  },
  "message": "Login successful"
}
```

失敗回應:
```json
{
  "success": false,
  "message": "Invalid username or password"
}
```

### 受保護端點

#### 獲取用戶列表
```http
GET /api/users
Authorization: Bearer <token>
```

## 測試帳號

| 用戶名稱 | 密碼     | 角色          |
|----------|----------|---------------|
| admin    | admin123 | administrator |
| user     | user123  | user          |

## 開發工具

### Air (熱重載)

安裝 Air:
```bash
go install github.com/air-verse/air@latest
```

創建 `.air.toml` 配置檔案:
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

### 測試 API

使用 curl 測試登入:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 專案結構

```
golang-api/
├── main.go           # 主程式入口
├── go.mod           # Go 模組定義
├── go.sum           # 依賴校驗和
├── internal/        # 內部套件
├── src/            # 源碼目錄
└── README.md       # 專案文檔
```

## 安全注意事項

⚠️ **生產環境注意**:

1. **密碼加密**: 目前使用明文密碼，生產環境請使用 bcrypt 或 argon2
2. **JWT Token**: 使用正式的 JWT 實作替代簡單 token
3. **資料庫**: 使用真實資料庫替代記憶體存儲
4. **HTTPS**: 生產環境必須使用 HTTPS
5. **環境變數**: 敏感資訊應從環境變數讀取

## 貢獻

1. Fork 此專案
2. 建立功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交變更 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 開啟 Pull Request

## 授權

此專案採用 MIT 授權 - 查看 [LICENSE](../LICENSE) 檔案了解詳情。
