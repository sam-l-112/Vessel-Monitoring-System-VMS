# VMS API 文檔

## 概述
VMS (Virtual Monitoring System) API 提供完整的養殖場監控和管理功能。

## 認證
所有受保護的端點都需要在請求標頭中包含 `Authorization` 字段。

### 登入
```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

響應：
```json
{
  "success": true,
  "message": "Login successful",
  "token": "vms_token_admin_1775643198",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@vms.com",
    "role": "administrator",
    "is_active": true,
    "created_at": "2026-04-08T18:12:57+08:00",
    "updated_at": "2026-04-08T18:12:57+08:00"
  }
}
```

## API 端點

### 認證相關
- `POST /api/auth/login` - 用戶登入
- `POST /api/auth/signin` - 用戶登入（別名）

### 用戶管理
- `GET /api/users/list` - 獲取所有用戶列表（需要授權）
- `GET /api/users/profile` - 獲取當前用戶資料（需要授權）

### 魚類數據
- `GET /api/fish/data` - 獲取所有魚類養殖數據
- `POST /api/fish/data` - 添加新的魚類數據

### 天氣數據
- `GET /api/weather/data` - 獲取環境監控數據

### 餵食記錄
- `GET /api/feed/data` - 獲取餵食記錄

### 系統狀態
- `GET /api/health` - 系統健康檢查

## 數據格式

### 用戶對象
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@vms.com",
  "role": "administrator",
  "is_active": true,
  "created_at": "2026-04-08T18:12:57+08:00",
  "updated_at": "2026-04-08T18:12:57+08:00"
}
```

### 魚類數據對象
```json
{
  "id": 1,
  "user_id": 1,
  "fish_type": "tilapia",
  "quantity": 100,
  "weight": 25.5,
  "health_status": "excellent",
  "created_at": "2026-04-08T18:12:57+08:00",
  "updated_at": "2026-04-08T18:12:57+08:00",
  "user": {
    "username": "admin"
  }
}
```

### 天氣數據對象
```json
{
  "id": 1,
  "temperature": 28.5,
  "humidity": 65.2,
  "ph_level": 7.2,
  "dissolved_oxygen": 6.8,
  "location": "pond_1",
  "recorded_at": "2026-04-08T18:12:57+08:00"
}
```

### 餵食記錄對象
```json
{
  "id": 1,
  "user_id": 1,
  "feed_type": "pellets",
  "quantity": 5.0,
  "unit": "kg",
  "feed_time": "2026-04-08T18:12:57+08:00",
  "user": {
    "username": "admin"
  }
}
```

## 錯誤響應
```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message"
}
```

## 測試示例

### 登入並獲取用戶列表
```bash
# 1. 登入獲取 token
TOKEN=$(curl -X POST http://192.168.50.75/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' \
  | jq -r '.token')

# 2. 使用 token 獲取用戶列表
curl -X GET http://192.168.50.75/api/users/list \
  -H "Authorization: $TOKEN"
```

### 添加魚類數據
```bash
curl -X POST http://192.168.50.75/api/fish/data \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "fish_type": "tilapia",
    "quantity": 100,
    "weight": 25.5,
    "health_status": "excellent"
  }'
```

## 開發說明
- API 使用 JSON 格式
- 支持 CORS
- 使用 gorilla/mux 路由器
- 數據庫使用 MariaDB/MySQL
- 認證使用簡單 token（生產環境建議使用 JWT）