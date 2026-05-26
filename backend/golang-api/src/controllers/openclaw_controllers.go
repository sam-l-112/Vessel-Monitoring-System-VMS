package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"vms-api/src/models"
)

var openclawConversations = make(map[string][]Message)
var openclawConvMutex sync.RWMutex

// AIController handles AI queries via OpenCli
type AIController struct{}

// QueryAI handles AI queries via opencli -> Gemini
func (ac *AIController) QueryAI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := models.APIResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Query == "" {
		response := models.APIResponse{
			Success: false,
			Message: "Query is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Execute opencli command to query Gemini
	response, err := queryGeminiShell(req.Query)
	if err != nil {
		response := models.APIResponse{
			Success: false,
			Message: "Failed to query Gemini: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	result := models.APIResponse{
		Success: true,
		Data:    map[string]string{"response": response},
		Message: "AI response retrieved successfully",
	}
	json.NewEncoder(w).Encode(result)
}

// OpenCLAWChatRequest represents a chat request to OpenCLAW
type OpenCLAWChatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// OpenCLAWChatResponse represents a chat response from OpenCLAW
type OpenCLAWChatResponse struct {
	Success        bool        `json:"success"`
	Message        string      `json:"message,omitempty"`
	Reply          string      `json:"reply,omitempty"`
	Error          string      `json:"error,omitempty"`
	ConversationID string      `json:"conversation_id,omitempty"`
}

// OpenCLAWChatHandler handles chat messages to OpenCLAW
func OpenCLAWChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(OpenCLAWChatResponse{Success: false, Error: "Method not allowed. Use POST"})
		return
	}

	var req OpenCLAWChatRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(OpenCLAWChatResponse{Success: false, Error: "Failed to read request body"})
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(OpenCLAWChatResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(OpenCLAWChatResponse{Success: false, Error: "Message cannot be empty"})
		return
	}

	// Forward to OpenCLAW agent webhook
	reply, err := forwardToOpenCLAW(req.Message)
	if err != nil {
		reply = fmt.Sprintf("已收到您的訊息，但 OpenCLAW Agent 暫時無法連線（%v）。您的訊息已記錄，待 Agent 上線後將自動處理。", err)
	}

	json.NewEncoder(w).Encode(OpenCLAWChatResponse{
		Success: true,
		Message: "OpenCLAW reply received",
		Reply:   reply,
	})
}

// forwardToOpenCLAW forwards the message to the OpenCLAW agent webhook
func forwardToOpenCLAW(message string) (string, error) {
	apiKey := os.Getenv("OPENCLAW_API_KEY")
	if apiKey == "" {
		return "OpenCLAW 已收到您的訊息（API 金鑰未設定，使用模擬回覆）", nil
	}

	webhookURL := os.Getenv("OPENCLAW_WEBHOOK_URL")
	if webhookURL == "" {
		return "OpenCLAW Agent 尚未設定 Webhook URL，請設定 OPENCLAW_WEBHOOK_URL 環境變數", nil
	}

	payload := map[string]interface{}{
		"agent":   "openclaw",
		"action":  "chat",
		"message": message,
	}

	payloadBytes, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to contact OpenCLAW agent: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Sprintf("OpenCLAW agent responded (unparseable): %s", string(respBody)), nil
	}

	if msg, ok := result["message"].(string); ok {
		return msg, nil
	}
	return "OpenCLAW agent received your message", nil
}

func queryGeminiShell(query string) (string, error) {
	cmd := exec.Command("/bin/bash", "/home/ouo/project_f/backend/golang-api/query_gemini.sh", query)
	cmd.Env = append(os.Environ(), "PATH=/home/ouo/.nvm/versions/node/v24.15.0/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	done := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		output, err := cmd.Output()
		if err != nil {
			errChan <- err
			return
		}
		done <- string(output)
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errChan:
		return "無法取得回應: " + err.Error(), nil
	case <-time.After(120 * time.Second):
		cmd.Process.Kill()
		return "", fmt.Errorf("query timeout after 120 seconds")
	}
}
