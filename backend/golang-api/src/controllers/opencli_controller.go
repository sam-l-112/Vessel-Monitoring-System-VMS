package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Conversation represents a chat session
type Conversation struct {
	ID       string    `json:"id"`
	Messages []Message `json:"messages"`
	Created  time.Time `json:"created"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string    `json:"role"` // "user" or "assistant"
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

// GeminiRequest is the input payload for Gemini chat requests.
type GeminiRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// GeminiResponse is the output payload returned to the frontend.
type GeminiResponse struct {
	Success        bool          `json:"success"`
	Message        string        `json:"message,omitempty"`
	Reply          string        `json:"reply,omitempty"`
	Error          string        `json:"error,omitempty"`
	FullHTML       string        `json:"full_html,omitempty"`
	ConversationID string        `json:"conversation_id,omitempty"`
	Conversation   *Conversation `json:"conversation,omitempty"`
}

// SystemStatus is the response payload for the status endpoint.
type SystemStatus struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

const (
	openCLIHomePath = "/home/ouo/OpenCLI/dist/src/main.js"
	nodePathStatic  = "/home/ouo/.nvm/versions/node/v24.15.0/bin/node"
	geminiURL       = "https://gemini.google.com/app/7aeec6192d00009f?hl=zh-tw"
)

// Global conversation store (in production, use a database)
var conversations = make(map[string]*Conversation)
var conversationMutex sync.RWMutex

// CreateConversationHandler creates a new conversation
func CreateConversationHandler(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Method not allowed. Use POST"})
		return
	}

	conversationID := fmt.Sprintf("conv_%d", time.Now().UnixNano())
	conversation := &Conversation{
		ID:       conversationID,
		Messages: []Message{},
		Created:  time.Now(),
	}

	conversationMutex.Lock()
	conversations[conversationID] = conversation
	conversationMutex.Unlock()

	json.NewEncoder(w).Encode(GeminiResponse{
		Success:        true,
		Message:        "Conversation created",
		ConversationID: conversationID,
		Conversation:   conversation,
	})
}

// GetConversationHandler retrieves a conversation by ID
func GetConversationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	conversationID := r.URL.Query().Get("id")
	if conversationID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Conversation ID required"})
		return
	}

	conversationMutex.RLock()
	conversation, exists := conversations[conversationID]
	conversationMutex.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Conversation not found"})
		return
	}

	json.NewEncoder(w).Encode(GeminiResponse{
		Success:        true,
		Message:        "Conversation retrieved",
		ConversationID: conversationID,
		Conversation:   conversation,
	})
}

// GeminiHandler receives a message and forwards it to Gemini via OpenCLI.
func GeminiHandler(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Method not allowed. Use POST"})
		return
	}

	var req GeminiRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Failed to read request body"})
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Invalid JSON format. Expected {\"message\": \"...\"}"})
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Message cannot be empty"})
		return
	}

	// Get or create conversation
	var conversation *Conversation
	if req.ConversationID != "" {
		conversationMutex.RLock()
		conv, exists := conversations[req.ConversationID]
		conversationMutex.RUnlock()
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: "Conversation not found"})
			return
		}
		conversation = conv
	} else {
		// Create new conversation
		conversationID := fmt.Sprintf("conv_%d", time.Now().UnixNano())
		conversation = &Conversation{
			ID:       conversationID,
			Messages: []Message{},
			Created:  time.Now(),
		}
		conversationMutex.Lock()
		conversations[conversationID] = conversation
		conversationMutex.Unlock()
	}

	// Add user message to conversation
	userMessage := Message{
		Role:    "user",
		Content: req.Message,
		Time:    time.Now(),
	}
	conversationMutex.Lock()
	conversation.Messages = append(conversation.Messages, userMessage)
	conversationMutex.Unlock()

	if err := ensureOpenCLIReady(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: fmt.Sprintf("OpenCLI not ready: %v", err)})
		return
	}

	reply, fullHTML, err := callGeminiViaOpenCLI(req.Message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: fmt.Sprintf("Failed to get Gemini response: %v", err)})
		return
	}

	// Add assistant message to conversation
	assistantMessage := Message{
		Role:    "assistant",
		Content: cleanGeminiResponse(reply),
		Time:    time.Now(),
	}
	conversationMutex.Lock()
	conversation.Messages = append(conversation.Messages, assistantMessage)
	conversationMutex.Unlock()

	json.NewEncoder(w).Encode(GeminiResponse{
		Success:        true,
		Message:        "Gemini reply received",
		Reply:          cleanGeminiResponse(reply),
		FullHTML:       fullHTML,
		ConversationID: conversation.ID,
		Conversation:   conversation,
	})
}

// GeminiStatusHandler returns the current OpenCLI/Gemini readiness state.
func GeminiStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	status := checkSystemStatus()
	w.WriteHeader(status.Code)
	json.NewEncoder(w).Encode(status)
}

func checkSystemStatus() SystemStatus {
	if err := ensureOpenCLIReady(); err != nil {
		return SystemStatus{
			Code:    http.StatusServiceUnavailable,
			Success: false,
			Status:  "not_ready",
			Message: fmt.Sprintf("OpenCLI unavailable: %v", err),
		}
	}

	return SystemStatus{
		Code:    http.StatusOK,
		Success: true,
		Status:  "ready",
		Message: "OpenCLI and Gemini are ready",
		Details: map[string]string{
			"chat":                "POST /api/ai/query",
			"status":              "GET  /api/ai/query",
			"create_conversation": "POST /api/ai/conversation",
			"get_conversation":    "GET  /api/ai/conversation",
		},
	}
}

func ensureOpenCLIReady() error {
	openCLI, nodePath, err := resolveOpenCLIPaths()
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-lc", fmt.Sprintf("%s %s doctor 2>&1", nodePath, openCLI))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("doctor failed: %v: %s", err, strings.TrimSpace(string(output)))
	}

	if !strings.Contains(string(output), "[OK]") {
		return fmt.Errorf("doctor output not healthy: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

func resolveOpenCLIPaths() (string, string, error) {
	openCLI := openCLIHomePath
	nodePath := nodePathStatic

	if _, err := os.Stat(openCLI); os.IsNotExist(err) {
		home := os.Getenv("HOME")
		openCLI = fmt.Sprintf("%s/OpenCLI/dist/src/main.js", home)
		nodePath = fmt.Sprintf("%s/.nvm/versions/node/v24.15.0/bin/node", home)
		if _, err := os.Stat(openCLI); os.IsNotExist(err) {
			return "", "", fmt.Errorf("OpenCLI not found at %s or %s", openCLIHomePath, openCLI)
		}
	}

	return openCLI, nodePath, nil
}

func callGeminiViaOpenCLI(message string) (string, string, error) {
	openCLI, nodePath, err := resolveOpenCLIPaths()
	if err != nil {
		return "", "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, nodePath, openCLI, "browser", "open", geminiURL)
	cmd.Dir = "/home/ouo/OpenCLI"
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("failed to open Gemini page: %v: %s", err, string(out))
	}
	time.Sleep(3 * time.Second)

	cmd = exec.CommandContext(ctx, nodePath, openCLI, "browser", "eval", `
        (async function() {
            const newChat = document.querySelector("div[aria-label='新對話'], div[aria-label='New chat'], a[href*='new']");
            if (newChat) {
                newChat.click();
                await new Promise(r => setTimeout(r, 1500));
            }
            return document.body.innerText;
        })()
    `)
	cmd.Dir = "/home/ouo/OpenCLI"
	cmd.CombinedOutput()
	time.Sleep(2 * time.Second)

	escaped := strings.ReplaceAll(message, "'", "\\'")
	cmd = exec.CommandContext(ctx, nodePath, openCLI, "browser", "eval", fmt.Sprintf(`
        (async function() {
            const text = '%s';
            const selectors = [
                "div.ql-editor",
                "div[contenteditable='true'][role='textbox']",
                "div[contenteditable='true']",
                "textarea",
            ];
            let input = null;
            for (const sel of selectors) {
                input = document.querySelector(sel);
                if (input) break;
            }
            if (!input) return "INPUT_NOT_FOUND";
            input.focus();
            const range = document.createRange();
            range.selectNodeContents(input);
            range.collapse(false);
            const sel = window.getSelection();
            sel.removeAllRanges();
            sel.addRange(range);
            document.execCommand('insertText', false, text);
            const buttons = Array.from(document.querySelectorAll('button'));
            for (const btn of buttons) {
                const label = btn.getAttribute('aria-label') || '';
                if (label.includes('傳送') || label.includes('Send') || label.includes('Submit')) {
                    btn.click();
                    break;
                }
            }
            return 'SENT';
        })()
    `, escaped))
	cmd.Dir = "/home/ouo/OpenCLI"
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("failed to send message: %v: %s", err, string(out))
	}

	time.Sleep(2 * time.Second)
	cmd = exec.CommandContext(ctx, nodePath, openCLI, "browser", "eval", `
        (async function() {
            for (let i = 0; i < 60; i++) {
                await new Promise(r => setTimeout(r, 1000));
                const stopBtn = document.querySelector('[aria-label="停止生成"], [aria-label="Stop generating"]');
                if (!stopBtn) {
                    await new Promise(r => setTimeout(r, 500));
                    return document.body.innerText;
                }
            }
            return document.body.innerText;
        })()
    `)
	cmd.Dir = "/home/ouo/OpenCLI"
	replyOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to read Gemini response: %v: %s", err, string(replyOutput))
	}

	fullText := strings.TrimSpace(string(replyOutput))
	if strings.HasPrefix(fullText, "ERROR") || strings.Contains(fullText, "INPUT_NOT_FOUND") {
		return "", fullText, fmt.Errorf("Gemini browser eval error: %s", fullText)
	}

	return extractLastAIResponse(fullText), fullText, nil
}

func extractLastAIResponse(fullText string) string {
	lines := strings.Split(fullText, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.Contains(line, "Gemini 說了") || strings.Contains(line, "Gemini 说了") {
			continue
		}
		return line
	}
	return strings.TrimSpace(fullText)
}

func cleanGeminiResponse(response string) string {
	lines := strings.Split(response, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "Update available") || strings.Contains(trimmed, "Run: npm") || strings.Contains(trimmed, "你說了") || strings.Contains(trimmed, "You said") {
			continue
		}
		out = append(out, trimmed)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
