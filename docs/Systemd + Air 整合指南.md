# Go API 開發環境自動化：Systemd + Air 整合指南

本文件說明如何透過 **Systemd** 管理 **Air**，在 Linux 伺服器上建立一個「自動重啟、熱重載、背景執行」的專業開發環境。

---

## 📋 環境資訊

* **使用者名稱**: `ouo`
* **專案路徑**: `/home/ouo/project_f/backend/golang-api`
* **Air 執行檔路徑**: `/home/ouo/go/bin/air`
* **Go 執行檔路徑**: `/usr/local/go/bin/go`

---

## 🛠 步驟一：建立 Systemd 服務檔案

在 Linux 系統中建立一個新的服務設定檔。

```bash
sudo nano /etc/systemd/system/vms-api.service

```

### 貼入以下設定內容

```ini
[Unit]
Description=VMS API Service (with Air Hot Reload)
After=network.target

[Service]
# 基本設定
Type=simple
User=ouo
WorkingDirectory=/home/ouo/project_f/backend/golang-api

# 核心指令：啟動 Air
ExecStart=/home/ouo/go/bin/air

# 自動重啟機制 (確保開發過程不中斷)
Restart=always
RestartSec=5

# 環境變數設定
Environment="SERVER_PORT=8080"
Environment="SERVER_HOST=0.0.0.0"
# 確保 Service 能找到 Go 與 Air 的路徑
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin:/home/ouo/go/bin"

[Install]
WantedBy=multi-user.target

```

---

## ⚙️ 步驟二：啟動與啟用服務

執行以下指令讓設定生效並啟動服務：

1. **重新載入系統服務清單**
```bash
sudo systemctl daemon-reload

```


2. **啟動服務並設定開機自啟**

```bash
   sudo systemctl enable --now vms-api

```

3. **檢查服務狀態**
```bash
sudo systemctl status vms-api

```



---

## 🔍 步驟三：日誌監控 (開發必備)

因為服務在背景執行，你需要透過 `journalctl` 來查看 Air 的即時編譯訊息與程式輸出：

```bash
# 查看即時日誌 (按 Ctrl+C 退出)
journalctl -u vms-api -f

```

---

## 🔄 常見管理指令

| 動作 | 指令 |
| --- | --- |
| **啟動**服務 | `sudo systemctl start vms-api` |
| **停止**服務 | `sudo systemctl stop vms-api` |
| **重啟**服務 | `sudo systemctl restart vms-api` |
| **停用**開機自啟 | `sudo systemctl disable vms-api` |
| **修改設定後**重新載入 | `sudo systemctl daemon-reload` |

---

## 💡 注意事項

1. **檔案權限**：請確保 `/home/ouo/project_f/backend/golang-api` 資料夾的權限屬於 `ouo` 使用者，否則 Air 無法寫入暫存檔。
2. **.env 檔案**：Air 會讀取 `WorkingDirectory` 下的 `.env` 檔案，請確認檔案位置正確。
3. **生產環境**：正式上線 (Production) 時，建議將 `ExecStart` 改為直接執行編譯好的 Binary 檔案（如 `./myapi`），並關閉 Air 以節省資源。

---

> **Tip:** 如果發現 Air 沒有反應，可以嘗試在專案目錄下執行 `air init` 生成 `.air.toml` 設定檔進行微調。

