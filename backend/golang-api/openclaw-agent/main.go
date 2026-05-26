package main

import (
	"bytes"
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
)

var vmsAPI = "http://127.0.0.1:8080"

func main() {
	port := os.Getenv("AGENT_PORT")
	if port == "" {
		port = "9090"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookHandler)
	mux.HandleFunc("/health", healthHandler)

	addr := ":" + port
	log.Printf("OpenCLAW Agent starting on %s", addr)
	log.Printf("Webhook endpoint: http://localhost:%s/webhook", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"agent":   "openclaw",
		"version": "1.0.0",
	})
}

type webhookReq struct {
	Agent      string      `json:"agent"`
	Action     string      `json:"action"`
	Message    string      `json:"message"`
	SessionKey string      `json:"session_key,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
	Timestamp  string      `json:"timestamp,omitempty"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	r.Body.Close()

	var req webhookReq
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %v"}`, err), http.StatusBadRequest)
		return
	}

	log.Printf("[webhook] agent=%s action=%s message=%s", req.Agent, req.Action, req.Message)

	reply := processMessage(req.Message)

	resp := map[string]interface{}{
		"success": true,
		"message": reply,
		"data": map[string]string{
			"agent":  req.Agent,
			"action": req.Action,
			"reply":  reply,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func processMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "您好！我是 OpenCLAW 養殖助手，請問有什麼可以幫助您的？"
	}

	lower := strings.ToLower(msg)

	if reply, matched := handleAddFish(msg, lower); matched {
		return reply
	}
	if reply, matched := handleAddFeed(msg, lower); matched {
		return reply
	}
	if reply := handleQueryFish(lower); reply != "" {
		return reply
	}
	if strings.Contains(lower, "天氣") || strings.Contains(lower, "weather") {
		return fetchWeatherData()
	}
	if strings.Contains(lower, "水溫") || strings.Contains(lower, "水質") || strings.Contains(lower, "ph") {
		return "目前水質數據需要從感測器取得。建議您在儀表板查看即時數據。"
	}
	if strings.Contains(lower, "飼料") || strings.Contains(lower, "feed") {
		return "飼料管理功能在「飼料管理」頁面。您可以查看飼料使用記錄或新增飼料。也可以直接對我說「新增飼料」來操作。"
	}
	if strings.Contains(lower, "hello") || strings.Contains(lower, "hi") || strings.Contains(lower, "你好") {
		return "你好！我是 OpenCLAW 養殖助手。我可以協助您新增魚類資料、飼料記錄、查詢天氣等。試試對我說「新增吳郭魚 500 條 1 公斤」！"
	}

	answer, err := queryOpenCLIGemini(msg)
	if err == nil && answer != "" {
		return answer
	}

	return fmt.Sprintf("已收到您的訊息。您也可以試試：\n- 「新增吳郭魚 500 條 1 公斤」\n- 「新增飼料 魚飼料 50 公斤」\n- 「查詢魚類資料」")
}

var chnNum = map[string]int{
	"零": 0, "一": 1, "二": 2, "兩": 2, "三": 3, "四": 4,
	"五": 5, "六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
}

func parseChineseNum(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if n, ok := chnNum[s]; ok {
		return n, true
	}
	if strings.Contains(s, "十") {
		parts := strings.Split(s, "十")
		left, right := 1, 0
		if parts[0] != "" {
			if n, ok := chnNum[parts[0]]; ok {
				left = n
			}
		}
		if len(parts) > 1 && parts[1] != "" {
			if n, ok := chnNum[parts[1]]; ok {
				right = n
			}
		}
		return left*10 + right, true
	}
	return 0, false
}

func parseFlexibleNum(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if n, err := strconv.Atoi(s); err == nil {
		return n, true
	}
	return parseChineseNum(s)
}

func extractWeight(s string) float64 {
	re := regexp.MustCompile(`(\d+\.?\d*)\s*公斤`)
	if m := re.FindStringSubmatch(s); len(m) > 1 {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil {
			return f
		}
	}
	re2 := regexp.MustCompile(`([一二兩三四五六七八九十\d]+)\s*公斤`)
	if m := re2.FindStringSubmatch(s); len(m) > 1 {
		if n, ok := parseChineseNum(m[1]); ok {
			return float64(n)
		}
	}
	return 0
}

func extractQuantity(s string) (int, bool) {
	re := regexp.MustCompile(`(\d+)\s*條`)
	if m := re.FindStringSubmatch(s); len(m) > 1 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n, true
		}
	}
	re2 := regexp.MustCompile(`([一二兩三四五六七八九十\d]+)\s*條`)
	if m := re2.FindStringSubmatch(s); len(m) > 1 {
		return parseChineseNum(m[1])
	}
	return 0, false
}

func handleAddFish(msg string, lower string) (string, bool) {
	if !strings.Contains(lower, "新增") && !strings.Contains(lower, "添加") && !strings.Contains(lower, "加入") {
		return "", false
	}
	if !strings.Contains(lower, "魚") || strings.Contains(lower, "飼料") {
		return "", false
	}

	re := regexp.MustCompile(`(?:新增|添加|加入)\s*(\S+?魚)\s*`)
	matches := re.FindStringSubmatch(msg)
	var fishType string
	if len(matches) > 1 {
		fishType = matches[1]
	} else {
		re2 := regexp.MustCompile(`(?:新增|添加|加入)\s*(.+?)(?:\s*魚)`)
		matches2 := re2.FindStringSubmatch(msg)
		if len(matches2) > 1 {
			fishType = matches2[1] + "魚"
		} else {
			re3 := regexp.MustCompile(`(?:新增|添加|加入)\s*(\S+)`)
			matches3 := re3.FindStringSubmatch(msg)
			if len(matches3) > 1 {
				fishType = matches3[1]
				if !strings.Contains(fishType, "魚") {
					fishType += "魚"
				}
			}
		}
	}
	if fishType == "" || fishType == "魚" || fishType == "新增魚" {
		return "請使用格式：新增 [魚種] [數量] 條 [重量]。例如：新增吳郭魚 500 條 1 公斤", true
	}

	quantity, ok := extractQuantity(msg)
	if !ok || quantity <= 0 {
		quantity = 1
	}

	weight := extractWeight(msg)
	health := "good"

	return callVMSAddFish(fishType, quantity, weight, health), true
}

func handleAddFeed(msg string, lower string) (string, bool) {
	if !strings.Contains(lower, "新增") && !strings.Contains(lower, "添加") && !strings.Contains(lower, "加入") {
		return "", false
	}
	if !strings.Contains(lower, "飼料") {
		return "", false
	}

	re := regexp.MustCompile(`(?:新增|添加|加入)\s*(?:飼料)?\s*(\S+?)\s*(\d+\.?\d*)\s*公斤`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) < 3 {
		re2 := regexp.MustCompile(`(?:新增|添加|加入)\s*(?:飼料)?\s*(\S+?)\s*([一二兩三四五六七八九十\d]+\.?\d*)\s*公斤`)
		matches = re2.FindStringSubmatch(msg)
	}
	if len(matches) < 3 {
		return "請使用格式：新增飼料 [類型] [數量] 公斤。例如：新增飼料 魚飼料 50 公斤", true
	}

	feedType := matches[1]
	quantity, _ := strconv.ParseFloat(matches[2], 64)

	return callVMSAddFeed(feedType, quantity), true
}

func handleQueryFish(lower string) string {
	if !strings.Contains(lower, "查詢") && !strings.Contains(lower, "顯示") && !strings.Contains(lower, "查看") && !strings.Contains(lower, "列出") {
		return ""
	}
	if !strings.Contains(lower, "魚") {
		return ""
	}

	resp, err := http.Get(vmsAPI + "/api/fish/data")
	if err != nil {
		return "無法查詢魚類資料，API 連線失敗。"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Sprintf("查詢結果：%s", string(body))
	}

	if !result["success"].(bool) {
		return "查詢失敗：" + result["message"].(string)
	}

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return "目前沒有任何魚類資料。"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("目前共有 %d 筆魚類資料：", len(data)))
	for _, item := range data {
		fish := item.(map[string]interface{})
		ft := fish["fish_type"]
		q := fish["quantity"]
		w := fish["weight"]
		h := fish["health_status"]
		lines = append(lines, fmt.Sprintf("- %s：%v 條，體重 %v kg，健康狀況：%v", ft, q, w, h))
	}
	return strings.Join(lines, "\n")
}

func callVMSAddFish(fishType string, quantity int, weight float64, health string) string {
	payload := map[string]interface{}{
		"fish_type":     fishType,
		"quantity":      quantity,
		"health_status": health,
	}
	if weight > 0 {
		payload["weight"] = weight
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(vmsAPI+"/api/fish/data", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Sprintf("無法新增魚類資料：%v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["success"] == true {
		return fmt.Sprintf("✅ 已成功新增 %s %d 條！您可以在「魚類數據」頁面查看。", fishType, quantity)
	}
	msg := "新增失敗"
	if m, ok := result["message"].(string); ok {
		msg = m
	}
	return fmt.Sprintf("❌ 新增 %s 失敗：%s", fishType, msg)
}

func callVMSAddFeed(feedType string, quantity float64) string {
	payload := map[string]interface{}{
		"feed_type": feedType,
		"quantity":  quantity,
		"unit":      "kg",
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(vmsAPI+"/api/feed/data", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Sprintf("無法新增飼料記錄：%v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["success"] == true {
		return fmt.Sprintf("✅ 已成功新增飼料記錄：%s %.1f 公斤！", feedType, quantity)
	}
	msg := "新增失敗"
	if m, ok := result["message"].(string); ok {
		msg = m
	}
	return fmt.Sprintf("❌ 新增飼料記錄失敗：%s", msg)
}

func queryOpenCLIGemini(query string) (string, error) {
	openCLI := "/home/ouo/OpenCLI/dist/src/main.js"
	nodePath := "/home/ouo/.nvm/versions/node/v24.15.0/bin/node"

	if _, err := os.Stat(openCLI); os.IsNotExist(err) {
		return "", fmt.Errorf("OpenCLI not found")
	}

	cmd := exec.Command(nodePath, openCLI, "gemini", "ask", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("OpenCLI error: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("empty response")
	}

	return result, nil
}

func fetchWeatherData() string {
	resp, err := http.Get(vmsAPI + "/api/weather/data")
	if err != nil {
		return "無法取得天氣數據，請確認 API 服務是否正常運作。"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Sprintf("天氣數據查詢結果：%s", string(body))
	}

	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		record := data[0].(map[string]interface{})
		return fmt.Sprintf("最新環境數據：溫度 %.1f°C，濕度 %.1f%%，pH值 %.1f，溶解氧 %.1f mg/L",
			toFloat(record["temperature"]), toFloat(record["humidity"]),
			toFloat(record["ph_level"]), toFloat(record["dissolved_oxygen"]))
	}

	return "目前暫無天氣數據。"
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	}
	return 0
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
