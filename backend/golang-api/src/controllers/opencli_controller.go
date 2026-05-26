package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"vms-api/src/services"
)

type Conversation struct {
	ID       string    `json:"id"`
	Messages []Message `json:"messages"`
	Created  time.Time `json:"created"`
}

type Message struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type GeminiRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
}

type GeminiResponse struct {
	Success        bool          `json:"success"`
	Message        string        `json:"message,omitempty"`
	Reply          string        `json:"reply,omitempty"`
	Error          string        `json:"error,omitempty"`
	FullHTML       string        `json:"full_html,omitempty"`
	ConversationID string        `json:"conversation_id,omitempty"`
	Conversation   *Conversation `json:"conversation,omitempty"`
}

type SystemStatus struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

var conversations = make(map[string]*Conversation)
var conversationMutex sync.RWMutex

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

	userMessage := Message{
		Role:    "user",
		Content: req.Message,
		Time:    time.Now(),
	}
	conversationMutex.Lock()
	conversation.Messages = append(conversation.Messages, userMessage)
	conversationMutex.Unlock()

	if err := services.EnsureOpenCLIReady(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: fmt.Sprintf("OpenCLI not ready: %v", err)})
		return
	}

	reply, fullHTML, err := services.GeminiQueryService(r.Context(), req.Message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GeminiResponse{Success: false, Error: fmt.Sprintf("Failed to get Gemini response: %v", err), FullHTML: fullHTML})
		return
	}

	assistantMessage := Message{
		Role:    "assistant",
		Content: reply,
		Time:    time.Now(),
	}
	conversationMutex.Lock()
	conversation.Messages = append(conversation.Messages, assistantMessage)
	conversationMutex.Unlock()

	json.NewEncoder(w).Encode(GeminiResponse{
		Success:        true,
		Message:        "Gemini reply received",
		Reply:          reply,
		FullHTML:       fullHTML,
		ConversationID: conversation.ID,
		Conversation:   conversation,
	})
}

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
	if err := services.EnsureOpenCLIReady(); err != nil {
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
			"chat":   "POST /api/opencli/gemini/chat",
			"status": "GET /api/opencli/gemini/status",
		},
	}
}
