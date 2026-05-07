package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"vms-api/src/database"
)

// ============== Request/Response Types ==============

type GeminiRequest struct {
	Message string `json:"message"`
}

type GeminiResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Reply    string `json:"reply,omitempty"`
	Error    string `json:"error,omitempty"`
	FullHTML string `json:"full_html,omitempty"`
}

type StatusResponse struct {
	Success   bool        `json:"success"`
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Commands  []string    `json:"setup_commands,omitempty"`
}

// ============== Main Handler ==============

func GeminiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS for CORS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(GeminiResponse{
			Success: false,
			Error:   "Method not allowed. Use POST",
		})
		return
	}

	// Parse request body
	var req GeminiRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GeminiResponse{
			Success: false,
			Error:   "Failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GeminiResponse{
			Success: false,
			Error:   "Invalid JSON format. Expected: {\"message\": \"your question\"}",
		})
		return
	}

	// Validate message
	if strings.TrimSpace(req.Message) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GeminiResponse{
			Success: false,
			Error:   "Message cannot be empty",
		})
		return
	}

	// Ensure prerequisites are running
	if err := ensureOpenCLIAndChrome(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(GeminiResponse{
			Success: false,
			Error:   fmt.Sprintf("System not ready: %v. Run: /home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh", err),
		})
		return
	}

	// Call Gemini via OpenCLI
	reply, fullHTML, err := callGeminiViaOpenCLI(req.Message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GeminiResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get Gemini response: %v", err),
		})
		return
	}

	// Return successful response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GeminiResponse{
		Success:  true,
		Message:  "Message sent to Gemini successfully",
		Reply:    reply,
		FullHTML: fullHTML,
	})
}

// ============== Status Handler ==============

func GeminiStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check all prerequisites
	status := checkSystemStatus()

	w.WriteHeader(status.Code)
	json.NewEncoder(w).Encode(status)
}

// ============== System Status Check ==============

type SystemStatus struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func checkSystemStatus() SystemStatus {
	// Check 1: OpenCLI daemon - use doctor command
	cmd := exec.Command("bash", "-c", "cd ~/OpenCLI && node dist/src/main.js doctor 2>&1")
	output, err := cmd.Output()
	if err != nil || !strings.Contains(string(output), "Everything looks good") {
		return SystemStatus{
			Code:    http.StatusServiceUnavailable,
			Success: false,
			Status:  "daemon_not_running",
			Message: "OpenCLI daemon is not running or not responding",
			Details: map[string]interface{}{
				"doctor_output": string(output),
				"solution":      "Run: /home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh",
				"script_path":   "/home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh",
			},
		}
	}

	// Check 2: Chrome process
	cmd = exec.Command("pgrep", "-f", "chrome")
	if err = cmd.Run(); err != nil {
		return SystemStatus{
			Code:    http.StatusServiceUnavailable,
			Success: false,
			Status:  "chrome_not_running",
			Message: "Chrome is not running",
			Details: map[string]interface{}{
				"solution": "Run: google-chrome --remote-debugging-port=9222 &",
				"check":    "pgrep -f chrome",
			},
		}
	}

	// Check 3: gemini-ask.sh exists
	if _, err := os.Stat("/home/ouo/project_f/backend/golang-api/scripts/gemini-ask.sh"); os.IsNotExist(err) {
		return SystemStatus{
			Code:    http.StatusServiceUnavailable,
			Success: false,
			Status:  "script_not_found",
			Message: "gemini-ask.sh script not found",
			Details: map[string]interface{}{
				"path": "/home/ouo/project_f/backend/golang-api/scripts/gemini-ask.sh",
			},
		}
	}

	return SystemStatus{
		Code:    http.StatusOK,
		Success: true,
		Status:  "ready",
		Message: "Gemini API is ready to use",
		Details: map[string]interface{}{
			"endpoints": map[string]string{
				"chat":   "POST /api/opencli/gemini/chat",
				"status": "GET  /api/opencli/gemini/status",
			},
			"example": map[string]string{
				"method": "POST",
				"url":    "/api/opencli/gemini/chat",
				"body":   "{\"message\": \"你好，請介紹你自己\"}",
			},
			"standby_note": "System is configured to work during screen-off/standby",
		},
	}
}

// ============== Ensure Prerequisites ==============

func ensureOpenCLIAndChrome() error {
	// Check if OpenCLI daemon port is listening (19825)
	cmd := exec.Command("bash", "-c", "lsof -i :19825 2>/dev/null | grep LISTEN")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("OpenCLI daemon not responding. Run: /home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh")
	}

	// Check Chrome - just check if it's running
	cmd = exec.Command("pgrep", "-f", "chrome")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Chrome not running. Run: google-chrome --remote-debugging-port=9222 &")
	}

	return nil
}

// ============== Call Gemini via OpenCLI (核心函數) =============
// 功能：透過 OpenCLI 控制瀏覽器發送訊息到 Gemini 並取得回覆

func callGeminiViaOpenCLI(message string) (string, string, error) {
	// 設定路徑
	openCLIPath := "/home/ouo/OpenCLI/dist/src/main.js"
	nodePath := "/home/ouo/.nvm/versions/node/v24.15.0/bin/node"
	
	// 檢查檔案是否存在
	if _, err := os.Stat(openCLIPath); os.IsNotExist(err) {
		homeDir := os.Getenv("HOME")
		openCLIPath = homeDir + "/OpenCLI/dist/src/main.js"
		nodePath = homeDir + "/.nvm/versions/node/v24.15.0/bin/node"
		if _, err := os.Stat(openCLIPath); os.IsNotExist(err) {
			return "", "", fmt.Errorf("OpenCLI not found")
		}
	}

	// 設定超時時間 90 秒
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	
	// ========== Step 1: 檢查並打開 Gemini 頁面 ==========
	cmd = exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "get", "url")
	urlOut, _ := cmd.CombinedOutput()
	urlStr := strings.TrimSpace(string(urlOut))
	log.Printf("Current URL: %s", urlStr)

	// 如果不是 Gemini 頁面，打開新頁面
	if !strings.Contains(urlStr, "gemini.google.com") || strings.Contains(urlStr, "about:blank") {
		log.Println("Opening Gemini page...")
		cmd = exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "open", "https://gemini.google.com/app?hl=zh-tw")
		cmd.Dir = "/home/ouo/OpenCLI"
		cmd.CombinedOutput()
		time.Sleep(6 * time.Second) // 等待頁面加載
	}

	// ========== Step 2: 清除輸入框 ==========
	log.Println("Clearing input box...")
	cmd = exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "eval", `
		(() => {
			const inp = document.querySelector("div.ql-editor");
			if (inp) { inp.innerText = ""; inp.textContent = ""; }
			return inp ? "ready" : "not_found";
		})()
	`)
	cmd.CombinedOutput()
	time.Sleep(1 * time.Second)

	// ========== Step 3: 發送訊息並等待回覆 ==========
	log.Println("Sending message to Gemini...")
	cmd = exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "eval", fmt.Sprintf(`
		(async function() {
			const userQuery = %q;
			try {
				// 找到輸入框
				let input = null;
				for(let i=0; i<5; i++) {
					input = document.querySelector("div.ql-editor");
					if(input) break;
					await new Promise(r => setTimeout(r, 500));
				}
				if (!input) return "INPUT_NOT_FOUND";
				
				// 輸入文字並發送
				input.innerText = userQuery;
				input.dispatchEvent(new Event("input", {bubbles: true}));
				await new Promise(r => setTimeout(r, 500));
				input.dispatchEvent(new KeyboardEvent("keydown", {key: "Enter", code: "Enter", bubbles: true}));
				
				// 等待回覆 (最多 45 秒)
				for (let i = 0; i < 45; i++) {
					await new Promise(r => setTimeout(r, 1000));
					const stopBtn = document.querySelector('[aria-label="停止生成"], [aria-label="Stop generating"]');
					if (stopBtn) continue;
					break;
				}
				
				// 獲取整個頁面文字
				return document.body.innerText;
			} catch(e) {
				return "ERROR: " + e.message;
			}
		})()
	`, message))

	replyOutput, _ := cmd.CombinedOutput()
	fullText := strings.TrimSpace(string(replyOutput))
	log.Printf("Raw response: %s", fullText[:min(200, len(fullText))])

	// 錯誤處理
	if fullText == "TIMEOUT" || fullText == "INPUT_NOT_FOUND" || strings.HasPrefix(fullText, "ERROR") {
		return "", "", fmt.Errorf("Gemini response failed: %s", fullText)
	}

	// 從完整文字中提取最後的 AI 回覆
	lines := strings.Split(fullText, "\n")
	var reply string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		// 跳過 UI 元素、導航、空白行
		if len(line) < 15 {
			continue
		}
		if strings.Contains(line, "Update available") ||
			strings.Contains(line, "Run: npm") ||
			strings.Contains(line, "我的內容") ||
			strings.Contains(line, "設定") ||
			strings.Contains(line, "工具") ||
			strings.Contains(line, "快捷鍵") ||
			strings.Contains(line, "Gemini") ||
			strings.Contains(line, "登入") ||
			strings.Contains(line, "新的對話") ||
			strings.Contains(line, "關於") ||
			strings.Contains(line, "企業用途") ||
			strings.Contains(line, "新視窗") ||
			strings.Contains(line, "訂閱") {
			continue
		}
		reply = line
		break
	}

	if reply == "" {
		reply = lines[len(lines)-1]
	}

	log.Printf("Extracted reply: %s", reply[:min(100, len(reply))])
	return reply, "", nil
}

// ============== Setup Handler ==============

func GeminiSetupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Use POST"})
		return
	}

	// Run setup commands
	var results []map[string]string

	// 1. Create scripts directory
	mkdirCmd := exec.Command("mkdir", "-p", "/home/ouo/scripts")
	if err := mkdirCmd.Run(); err != nil {
		results = append(results, map[string]string{"step": "mkdir", "status": "failed", "error": err.Error()})
	} else {
		results = append(results, map[string]string{"step": "mkdir", "status": "ok"})
	}

	// 2. Create startup script
	setupScript := `#!/bin/bash
cd ~/OpenCLI
node dist/src/main.js daemon start 2>/dev/null
sleep 2
google-chrome --remote-debugging-port=9222 --user-data-dir=$HOME/.config/google-chrome --no-first-run --no-default-browser-check > /dev/null 2>&1 &
echo "OpenCLI Gemini service started"
`

	scriptPath := "/home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh"
	if err := os.WriteFile(scriptPath, []byte(setupScript), 0755); err != nil {
		results = append(results, map[string]string{"step": "write_startup_script", "status": "failed", "error": err.Error()})
	} else {
		results = append(results, map[string]string{"step": "write_startup_script", "status": "ok"})
	}

	// 3. Create autostart
	autostartDir := os.Getenv("HOME") + "/.config/autostart"
	autostartFile := autostartDir + "/gemini-opencli.desktop"
	desktopEntry := `[Desktop Entry]
Type=Application
Name=Gemini OpenCLI Service
Exec=/home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`

	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		results = append(results, map[string]string{"step": "mkdir_autostart", "status": "failed", "error": err.Error()})
	} else if err := os.WriteFile(autostartFile, []byte(desktopEntry), 0644); err != nil {
		results = append(results, map[string]string{"step": "write_autostart", "status": "failed", "error": err.Error()})
	} else {
		results = append(results, map[string]string{"step": "write_autostart", "status": "ok"})
	}

	// 4. Disable sleep
	gsettingsCmds := []string{
		"gsettings set org.gnome.desktop.session idle-delay 0",
		"gsettings set org.gnome.desktop.lockdown disable-lock-screen true",
	}

	for _, g := range gsettingsCmds {
		cmd := exec.Command("bash", "-c", g)
		if err := cmd.Run(); err != nil {
			results = append(results, map[string]string{"step": g, "status": "warning", "error": err.Error()})
		} else {
			results = append(results, map[string]string{"step": g, "status": "ok"})
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"message":     "Setup complete. System will work during standby.",
		"results":     results,
		"next_steps":  []string{"1. Restart system or run: /home/ouo/project_f/backend/golang-api/scripts/gemini-start.sh", "2. Lock screen and test: curl -X POST http://localhost:8080/api/opencli/gemini/chat -d '{\"message\":\"test\"}'"},
	})
}

// ============== AI Query Handler (for frontend) ==============

type AIQueryRequest struct {
	Query string `json:"query"`
}

type AIQueryResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Reply   string      `json:"reply,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// AIQueryHandler handles AI chat requests from frontend
func AIQueryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AIQueryHandler: started")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS for CORS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AIQueryResponse{
			Success: false,
			Message: "Method not allowed. Use POST",
		})
		return
	}

	// Parse request body
	var req AIQueryRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AIQueryResponse{
			Success: false,
			Message: "Failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AIQueryResponse{
			Success: false,
			Message: "Invalid JSON format",
		})
		return
	}

	// Validate query
	if strings.TrimSpace(req.Query) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AIQueryResponse{
			Success: false,
			Message: "Query cannot be empty",
		})
		return
	}

	// Check if OpenCLI exists
	openCLIPath := "/home/ouo/OpenCLI/dist/src/main.js"
	if _, err := os.Stat(openCLIPath); os.IsNotExist(err) {
		openCLIPath = os.Getenv("HOME") + "/OpenCLI/dist/src/main.js"
		if _, err := os.Stat(openCLIPath); os.IsNotExist(err) {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(AIQueryResponse{
				Success: false,
				Message: "OpenCLI not found. Please install OpenCLI at ~/OpenCLI",
			})
			return
		}
	}

	// 直接传递用户问题，不加任何系统提示词
	prompt := req.Query

	// Call Gemini via OpenCLI
	reply, _, err := callGeminiViaOpenCLI(prompt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AIQueryResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get AI response: %v", err),
		})
		return
	}

	// 清理回复中的历史记录和多余内容
	cleanReply := cleanGeminiResponse(reply)
	if cleanReply == "" {
		cleanReply = reply
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AIQueryResponse{
		Success: true,
		Message: "Query processed successfully",
		Reply:   cleanReply,
		Data: map[string]interface{}{
			"response": cleanReply,
		},
	})
}

// buildSmartPrompt creates a prompt that tells Gemini what actions it can perform
func buildSmartPrompt(query string) string {
	return fmt.Sprintf(`你是一個水產養殖管理系統 AI 助手。用戶說：「%s」

你可以執行的操作：
1. 魚類管理：查看、新增、修改、刪除魚類數據
2. 飼料管理：查看、新增、修改飼料記錄
3. 天氣查詢：查詢現在天氣和預報
4. 數據分析：提供養殖建議

請先用繁體中文回覆用戶，然後如果你判斷需要執行操作，請在回覆最後加上以下格式的標記：
[ACTION]
{"action": "操作類型", "type": "魚類/飼料", "field": "欄位", "value": "新值", "target": "目標名稱"}
[/ACTION]

例如：
- 要把金目鱸數量改成200：[ACTION]{"action":"update","type":"魚類","field":"quantity","value":"200","target":"金目鱸"}[/ACTION]
- 新增魚類：[ACTION]{"action":"add","type":"魚類"}[/ACTION]

如果不需要執行操作，只回覆對話即可。`, query)
}

// callGeminiWithAction sends prompt and detects if action needs to be executed
func callGeminiWithAction(prompt string) (string, map[string]string, error) {
	reply, err := callGeminiForReport(prompt)
	if err != nil {
		return "", nil, err
	}

	// Parse action from response
	var action map[string]string
	if strings.Contains(reply, "[ACTION]") {
		start := strings.Index(reply, "[ACTION]")
		end := strings.Index(reply, "[/ACTION]")
		if end > start {
			actionJSON := reply[start+8 : end]
			// Simple parsing - in production use proper JSON unmarshal
			action = parseAction(actionJSON)
			// Remove action tag from reply
			reply = strings.Replace(reply, "[ACTION]"+actionJSON+"[/ACTION]", "", -1)
		}
	}

	return strings.TrimSpace(reply), action, nil
}

// parseAction parses the action JSON string
func parseAction(actionJSON string) map[string]string {
	action := make(map[string]string)
	action["action"] = extractJSONValue(actionJSON, "action")
	action["type"] = extractJSONValue(actionJSON, "type")
	action["field"] = extractJSONValue(actionJSON, "field")
	action["value"] = extractJSONValue(actionJSON, "value")
	action["target"] = extractJSONValue(actionJSON, "target")
	return action
}

func extractJSONValue(jsonStr, key string) string {
	pattern := `"` + key + `"[^:]*:([^,}]+)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(jsonStr)
	if len(matches) > 1 {
		return strings.Trim(strings.TrimSpace(matches[1]), `"`)
	}
	return ""
}

// executeAction performs the requested action
func executeAction(action map[string]string) string {
	objType := action["type"]

	if objType == "魚類" || objType == "鱼类" {
		return executeFishAction(action)
	} else if objType == "飼料" || objType == "饲料" {
		return executeFeedAction(action)
	}

	return ""
}

// executeFishAction handles fish-related actions
func executeFishAction(action map[string]string) string {
	target := action["target"]
	field := action["field"]
	value := action["value"]

	if target == "" {
		return ""
	}

	// Find fish by name
	var fishID int
	err := database.DB.QueryRow("SELECT id FROM fish_data WHERE fish_type = ? LIMIT 1", target).Scan(&fishID)
	if err != nil {
		return fmt.Sprintf("找不到魚類「%s」，請先新增", target)
	}

	if action["action"] == "update" || action["action"] == "修改" {
		if field == "quantity" {
			_, err := database.DB.Exec("UPDATE fish_data SET quantity = ?, updated_at = NOW() WHERE id = ?", value, fishID)
			if err != nil {
				return "更新失敗：" + err.Error()
			}
			return fmt.Sprintf("已將「%s」的數量更新為 %s 尾", target, value)
		} else if field == "weight" {
			_, err := database.DB.Exec("UPDATE fish_data SET weight = ?, updated_at = NOW() WHERE id = ?", value, fishID)
			if err != nil {
				return "更新失敗：" + err.Error()
			}
			return fmt.Sprintf("已將「%s」的體重更新為 %s kg", target, value)
		} else if field == "health_status" || field == "健康狀況" {
			statusMap := map[string]string{"良好": "good", "优良": "excellent", "一般": "fair", "差": "poor"}
			status := statusMap[value]
			if status == "" {
				status = "good"
			}
			_, err := database.DB.Exec("UPDATE fish_data SET health_status = ?, updated_at = NOW() WHERE id = ?", status, fishID)
			if err != nil {
				return "更新失敗：" + err.Error()
			}
			return fmt.Sprintf("已將「%s」的健康狀況更新為 %s", target, value)
		}
	}

	return "支援的操作：數量(quantity)、體重(weight)、健康狀況(health_status)"
}

// executeFeedAction handles feed-related actions
func executeFeedAction(action map[string]string) string {
	actionType := action["action"]

	if actionType == "add" || actionType == "新增" {
		feedType := action["field"]
		quantity := action["value"]

		if feedType == "" {
			feedType = "一般飼料"
		}
		if quantity == "" {
			quantity = "10"
		}

		qty, _ := strconv.ParseFloat(quantity, 64)
		_, err := database.DB.Exec("INSERT INTO feed_data (user_id, feed_type, quantity, unit) VALUES (1, ?, ?, 'kg')", feedType, qty)
		if err != nil {
			return "新增飼料失敗：" + err.Error()
		}
		return fmt.Sprintf("已新增飼料記錄：%s %s kg", feedType, quantity)
	}

	return "目前支援：新增飼料"
}

// ============== Daily Report Handler ==============

type DailyReportRequest struct {
	Area string `json:"area"`
}

type DailyReportResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Report   string      `json:"report,omitempty"`
	GeneratedAt string   `json:"generated_at"`
}

// DailyReportHandler generates automatic daily report for fishermen
func DailyReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ensure OpenCLI is ready
	if err := ensureOpenCLIAndChrome(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(DailyReportResponse{
			Success: false,
			Message: fmt.Sprintf("System not ready: %v", err),
		})
		return
	}

	// Collect all data from various sources
	reportData := collectAllDataForReport()

	// Create prompt for Gemini to generate daily report
	prompt := fmt.Sprintf(`你是養殖管理系統 AI 助手。請根據以下數據，為澎湖地區的漁民生成一份簡潔的「今日養殖作業報告」，格式如下：

## 📊 今日概覽
- 天氣狀況：
- 水溫：
- 建議：

## 🐟 魚群狀況
- 總數量：
- 健康狀況：

## 🦐 飼料建議
- 今日投餵量：
- 建議：

## ⚠️ 注意事項
- 任何異常狀況或建議

數據來源：
%s

請用繁體中文回覆，重點保持在 200 字以內，方便漁民快速閱讀。`, reportData)

	// Call Gemini to generate report
	reply, err := callGeminiForReport(prompt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(DailyReportResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to generate report: %v", err),
		})
		return
	}

	// Return the generated report
	w.WriteHeader(http.StatusOK)
	fishData := getFishDataForReport()
	weather := getLatestWeather()
	json.NewEncoder(w).Encode(DailyReportResponse{
		Success:    true,
		Message:    "Report generated successfully",
		Report:     reply,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Data: map[string]interface{}{
			"fishData": fishData,
			"weather": weather,
			"feedAmount": getTodayFeedAmount(),
		},
	})
}

// collectAllDataForReport gathers all data from database and APIs
func collectAllDataForReport() string {
	var sb strings.Builder

	// Get fish data
	sb.WriteString("【魚群數據】\n")
	fishData := getFishDataForReport()
	sb.WriteString(fmt.Sprintf("- 總數量: %d 尾\n", fishData["totalCount"]))
	sb.WriteString(fmt.Sprintf("- 總重量: %.1f kg\n", fishData["totalWeight"]))
	sb.WriteString(fmt.Sprintf("- 健康狀況: %s\n", fishData["healthStatus"]))

	// Get weather data
	sb.WriteString("\n【天氣數據】\n")
	weather := getLatestWeather()
	if weather != nil {
		sb.WriteString(fmt.Sprintf("- 氣溫: %.1f°C\n", weather["temperature"]))
		sb.WriteString(fmt.Sprintf("- 濕度: %.1f%%\n", weather["humidity"]))
		sb.WriteString(fmt.Sprintf("- 預估水溫: %.1f°C\n", weather["water_temp"]))
	}

	// Get feed data
	sb.WriteString("\n【飼料數據】\n")
	sb.WriteString(fmt.Sprintf("- 今日投餵量: %.1f kg\n", getTodayFeedAmount()))

	return sb.String()
}

// getFishCount returns total fish count from database
func getFishCount() int {
	if database.DB == nil {
		log.Println("Database not connected, returning 0")
		return 0
	}

	var totalQuantity int
	err := database.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM fish_data").Scan(&totalQuantity)
	if err != nil {
		log.Printf("Failed to get fish count: %v", err)
		return 0
	}

	log.Printf("Total fish count from DB: %d", totalQuantity)
	return totalQuantity
}

// getFishData returns fish data for report
func getFishDataForReport() map[string]interface{} {
	if database.DB == nil {
		return map[string]interface{}{
			"totalCount":  0,
			"totalWeight": 0.0,
			"healthStatus": "無數據",
		}
	}

	var totalCount int
	var totalWeight float64
	var healthStats map[string]int = make(map[string]int)

	rows, err := database.DB.Query("SELECT COALESCE(SUM(quantity), 0), COALESCE(SUM(weight), 0), health_status FROM fish_data GROUP BY health_status")
	if err != nil {
		log.Printf("Failed to get fish data: %v", err)
		return map[string]interface{}{"totalCount": 0, "totalWeight": 0.0, "healthStatus": "無數據"}
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		var weight float64
		var health string
		if err := rows.Scan(&count, &weight, &health); err == nil {
			totalCount += count
			totalWeight += weight
			healthStats[health] = healthStats[health] + count
		}
	}

	healthStatus := "正常"
	if excellent, ok := healthStats["excellent"]; ok && excellent > 0 {
		healthStatus = "优良"
	} else if good, ok := healthStats["good"]; ok && good > 0 {
		healthStatus = "良好"
	} else if fair, ok := healthStats["fair"]; ok && fair > 0 {
		healthStatus = "一般"
	} else if poor, ok := healthStats["poor"]; ok && poor > 0 {
		healthStatus = "需改善"
	}

	if totalCount == 0 {
		healthStatus = "無數據"
	}

	return map[string]interface{}{
		"totalCount":  totalCount,
		"totalWeight": totalWeight,
		"healthStatus": healthStatus,
		"healthStats": healthStats,
	}
}

// getLatestWeather returns latest weather data from CWA API
func getLatestWeather() map[string]float64 {
	area := "penghu"
	apiKey := getCWAAPIKey(area)
	if apiKey == "" {
		log.Println("CWA API key not found, using default values")
		return map[string]float64{
			"temperature": 28.5,
			"humidity": 75.0,
			"water_temp": 26.0,
		}
	}

	stationID := getCWAStationID("", area)
	baseURL := getCWABaseURL()
	apiURL := fmt.Sprintf("%s/O-A0001-001?StationID=%s", baseURL, stationID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Failed to create CWA request: %v", err)
		return map[string]float64{"temperature": 28.5, "humidity": 75.0, "water_temp": 26.0}
	}

	req.Header.Set("Authorization", apiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch CWA weather: %v", err)
		return map[string]float64{"temperature": 28.5, "humidity": 75.0, "water_temp": 26.0}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("CWA API returned status %d", resp.StatusCode)
		return map[string]float64{"temperature": 28.5, "humidity": 75.0, "water_temp": 26.0}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read CWA response: %v", err)
		return map[string]float64{"temperature": 28.5, "humidity": 75.0, "water_temp": 26.0}
	}

	var stationResp CWAWeatherStationResponse
	if err := json.Unmarshal(body, &stationResp); err != nil {
		log.Printf("Failed to parse CWA response: %v", err)
		return map[string]float64{"temperature": 28.5, "humidity": 75.0, "water_temp": 26.0}
	}

	if len(stationResp.Records.Station) > 0 {
		station := stationResp.Records.Station[0]
		we := station.WeatherElement

		temp := 28.5
		humidity := 75.0

		if we.AirTemperature != "" {
			fmt.Sscanf(we.AirTemperature, "%f", &temp)
		}
		if we.RelativeHumidity != "" {
			fmt.Sscanf(we.RelativeHumidity, "%f", &humidity)
		}

		log.Printf("CWA Weather: temp=%.1f, humidity=%.1f", temp, humidity)

		return map[string]float64{
			"temperature": temp,
			"humidity":     humidity,
			"water_temp":   temp - 2.0,
		}
	}

	return map[string]float64{"temperature": 28.5, "humidity": 75.0, "water_temp": 26.0}
}

// getTodayFeedAmount returns today's feed amount
func getTodayFeedAmount() float64 {
	return 150.0 // Placeholder
}

// callGeminiForReport generates report using Gemini (simplified version)
func callGeminiForReport(prompt string) (string, error) {
	openCLIPath := "/home/ouo/OpenCLI/dist/src/main.js"
	nodePath := "/home/ouo/.nvm/versions/node/v24.15.0/bin/node"

	if _, err := os.Stat(openCLIPath); os.IsNotExist(err) {
		homeDir := os.Getenv("HOME")
		openCLIPath = homeDir + "/OpenCLI/dist/src/main.js"
		nodePath = homeDir + "/.nvm/versions/node/v24.15.0/bin/node"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Check current URL
	cmd := exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "get", "url")
	urlOut, _ := cmd.CombinedOutput()
	urlStr := strings.TrimSpace(string(urlOut))

	if urlStr == "" || urlStr == "about:blank" || (!strings.Contains(urlStr, "gemini.google.com") && !strings.Contains(urlStr, "google.com")) {
		log.Println("Opening Gemini page...")
		cmd = exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "open", "https://gemini.google.com/app")
		cmd.Dir = "/home/ouo/OpenCLI"
		cmd.CombinedOutput()
		time.Sleep(5 * time.Second)
	}

	// Wait for page to fully load
	cmd = exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "eval", `
		(async function() {
			await new Promise(r => setTimeout(r, 2000));
			return document.readyState;
		})()
	`)
	cmd.CombinedOutput()

	// Send prompt to Gemini
	cmd = exec.CommandContext(ctx, nodePath, openCLIPath, "browser", "eval", fmt.Sprintf(`
		(async function() {
			try {
				// Try multiple selectors for the input box
				const selectors = [
					'div.ql-editor',
					'div[contenteditable="true"][role="textbox"]',
					'div[contenteditable="true"]',
					'textarea',
					'rich-textarea div[contenteditable="true"]',
					'.input-area textarea'
				];
				let input = null;
				for (const sel of selectors) {
					input = document.querySelector(sel);
					if (input) break;
				}
				if (!input) {
					// Log page content for debugging
					console.log('Page title:', document.title);
					console.log('Body content:', document.body.innerText.substring(0, 500));
					return "INPUT_NOT_FOUND";
				}
				input.focus();
				const sel = window.getSelection();
				const range = document.createRange();
				range.selectNodeContents(input);
				range.collapse(false);
				sel.removeAllRanges();
				sel.addRange(range);
				document.execCommand('insertText', false, %q);

				await new Promise(r => setTimeout(r, 500));

				const btns = document.querySelectorAll('button');
				for (const b of btns) {
					const aria = b.getAttribute('aria-label') || '';
					if ((aria.includes('傳送') || aria.includes('Send') || aria.includes('Submit')) && !b.disabled) {
						b.click();
						break;
					}
				}

				for (let i = 0; i < 50; i++) {
					await new Promise(r => setTimeout(r, 1000));
					const stopBtn = document.querySelector('[aria-label="停止生成"], [aria-label="Stop generating"]');
					if (!stopBtn) {
						const chat = document.querySelector('chat-history, [id="chat-history"]');
						if (chat) {
							const texts = chat.innerText.split('Gemini 說了');
							if (texts.length > 1) {
								const reply = texts[texts.length - 1].trim();
								if (reply.length > 20) return reply;
							}
						}
					}
				}
				return "TIMEOUT";
			} catch(e) {
				return "ERROR: " + e.message;
			}
		})()
	`, prompt))

	replyOutput, _ := cmd.CombinedOutput()
	reply := strings.TrimSpace(string(replyOutput))

	if reply == "TIMEOUT" || reply == "INPUT_NOT_FOUND" || strings.HasPrefix(reply, "ERROR") {
		return "", fmt.Errorf("Gemini response failed: %s", reply)
	}

	return reply, nil
}

// ============== OpenCLAW Chat Handler ==============

type OpenCLAWRequest struct {
	Message string `json:"message"`
}

type OpenCLAWResponse struct {
	Success  bool   `json:"success"`
	Reply    string `json:"reply"`
	Message  string `json:"message,omitempty"`
}

func OpenCLAWChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(OpenCLAWResponse{Success: false, Message: "Method not allowed"})
		return
	}

	var req OpenCLAWRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(OpenCLAWResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	if req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(OpenCLAWResponse{Success: false, Message: "Empty message"})
		return
	}

	// 使用 OpenCLAW CLI 命令
	openCLAWPath := "/home/ouo/openclaw/openclaw.mjs"
	nodePath := "/home/ouo/.nvm/versions/node/v24.15.0/bin/node"

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 构建提示词，包含操作指令
	prompt := fmt.Sprintf(`你是水產養殖管理系統 AI 助手。用戶說：「%s」

你可以執行的操作：
1. 魚類管理：查看、新增、修改、刪除魚類數據
2. 飼料管理：查看、新增、修改飼料記錄  
3. 天氣查詢：查詢現在天氣和預報

如果需要執行操作，請用以下格式標記：
[ACTION]{"action":"操作","type":"魚類/飼料","field":"欄位","value":"值","target":"目標"}[/ACTION]

用戶問題：%s

請用繁體中文回覆用戶的問題。`, req.Message, req.Message)

	// 通过 CLI 执行 agent 命令 (使用 --local 模式)
	cmd := exec.CommandContext(ctx, nodePath, openCLAWPath, "agent", "--local", "-m", prompt)
	cmd.Dir = "/home/ouo/openclaw"

	output, err := cmd.CombinedOutput()
	reply := strings.TrimSpace(string(output))

	if err != nil {
		log.Printf("OpenCLAW error: %v, output: %s", err, string(output))
		reply = "抱歉，目前無法連線到 AI 服務。請稍後再試。"
	}

	// 清理回复中的 ANSI 颜色代码
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	reply = ansiRegex.ReplaceAllString(reply, "")

	// 提取 ACTION 并执行
	if strings.Contains(reply, "[ACTION]") {
		actionResult := extractAndExecuteAction(reply)
		if actionResult != "" {
			reply = reply + "\n\n" + actionResult
		}
	}

	json.NewEncoder(w).Encode(OpenCLAWResponse{
		Success: true,
		Reply:   reply,
	})
}

// extractAndExecuteAction 从回复中提取 ACTION 并执行
func extractAndExecuteAction(response string) string {
	start := strings.Index(response, "[ACTION]")
	end := strings.Index(response, "[/ACTION]")
	if end > start && start >= 0 {
		actionJSON := response[start+8:end]
		action := parseAction(actionJSON)
		return executeAction(action)
	}
	return ""
}

// cleanGeminiResponse 清理 Gemini 回复中的多余内容
func cleanGeminiResponse(response string) string {
	lines := strings.Split(response, "\n")
	var cleanLines []string
	skipRest := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 遇到对话历史标记，跳过后面所有内容
		if strings.Contains(line, "你說了") || strings.Contains(line, "你说了") || 
		   strings.Contains(line, "You said") || strings.Contains(line, "用戶問題") {
			skipRest = true
			continue
		}

		if skipRest {
			continue
		}

		// 跳过更新提示
		if strings.Contains(line, "Update available:") || strings.Contains(line, "Run: npm install") {
			continue
		}

		// 跳过空行
		if trimmed == "" {
			continue
		}

		cleanLines = append(cleanLines, line)
	}

	result := strings.TrimSpace(strings.Join(cleanLines, "\n"))

	// 如果结果太短，取前几行
	if len(result) < 20 {
		result = ""
		count := 0
		for _, line := range lines {
			if count >= 5 {
				break
			}
			if !strings.Contains(line, "Update available") && !strings.Contains(line, "你說了") {
				result += line + "\n"
				count++
			}
		}
		result = strings.TrimSpace(result)
	}

	return result
}