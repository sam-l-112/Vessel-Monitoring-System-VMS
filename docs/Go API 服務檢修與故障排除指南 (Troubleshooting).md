# Go API 服務檢修與故障排除指南 (Troubleshooting)

當你的 `myapi.service` (Systemd + Air) 出現問題，或是 API 無法正常啟動時，請依照此流程進行檢查。

---

## 第一步：檢查服務狀態與錯誤訊息

首先確認服務是否正在運行，並觀察錯誤日誌。

### 1. 檢查 Systemd 狀態

```bash
sudo systemctl status myapi

```

* **Active: active (running)**: 正常運行中。
* **Active: failed / activating**: 代表啟動失敗，正在嘗試重啟。

### 2. 查看即時輸出日誌 (重要)

這是診斷問題最快的方法，可以看到程式內部的 `fmt.Println` 或報錯。

```bash
journalctl -u myapi -f

```

> **常見報錯訊息：**
> * `bind: address already in use`: 代表 Port (8080) 被佔用了。
> * `.env file not found`: 代表環境變數檔案路徑不正確。
> 
> 

---

## 🚧 第二步：處理 Port 8080 佔用問題

當你看到 `address already in use`，代表舊的進程沒關乾淨。

### 1. 找出是誰在佔用 Port

使用 `lsof` 指令查看目前是哪個 PID (進程編號) 在使用 8080。

```bash
sudo lsof -i :8080

```

* **PID**: 記錄下顯示的數字（例如 `1386091`）。

### 2. 強制結束該進程

使用 `kill -9` 指令將該「孤兒進程」徹底關閉。

```bash
# 語法：sudo kill -9 [PID]
sudo kill -9 1386091

```

### 3. 一鍵清理 (進階指令)

如果你不想查 PID，想直接清空所有佔用 8080 的程式：

```bash
sudo fuser -k 8080/tcp

```

---

## 🔄 第三步：重新啟動服務

清理完環境後，讓 Systemd 重新接管服務。

```bash
# 重新載入設定 (若有修改 .service 檔才需要)
sudo systemctl daemon-reload

# 重啟服務
sudo systemctl restart myapi

# 確認是否啟動成功
sudo systemctl status myapi

```

---

## 💡 常見問題檢查清單 (Checklist)

| 檢查項目 | 解決方法 |
| --- | --- |
| **`.env` 讀不到** | 確認檔案位於 `WorkingDirectory` 下，且檔名有 `.` 開頭。 |
| **Air 沒反應** | 檢查 `ExecStart` 路徑是否正確指向 `/home/ouo/go/bin/air`。 |
| **權限不足** | 確保專案資料夾擁有者為 `ouo`，指令：`sudo chown -R ouo:ouo 專案路徑`。 |
| **改了 Code 沒更新** | 檢查 `journalctl` 是否顯示 `building...`，若無，可能是 `air` 監控路徑設定錯誤。 |

---

## 🚀 總結必殺技

如果服務亂掉，直接執行這串「三部曲」通常能解決 90% 的問題：

```bash
sudo fuser -k 8080/tcp && sudo systemctl restart myapi && journalctl -u myapi -f

```
