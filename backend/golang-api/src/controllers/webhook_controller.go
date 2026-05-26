package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type AgentWebhookRequest struct {
	Agent      string      `json:"agent,omitempty"`
	Action     string      `json:"action,omitempty"`
	Message    string      `json:"message,omitempty"`
	SessionKey string      `json:"session_key,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
	Timestamp  string      `json:"timestamp,omitempty"`
}

type AgentWebhookResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	RunID   string      `json:"run_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func AgentWebhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AgentWebhookResponse{Success: false, Message: "Method not allowed. Use POST"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AgentWebhookResponse{Success: false, Message: "Failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req AgentWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AgentWebhookResponse{Success: false, Message: "Invalid JSON format"})
		return
	}

	if req.Timestamp == "" {
		req.Timestamp = time.Now().Format(time.RFC3339)
	}

	log.Printf("[OpenClaw Webhook] agent=%s action=%s message=%s session=%s payload=%+v metadata=%+v timestamp=%s",
		req.Agent, req.Action, req.Message, req.SessionKey, req.Payload, req.Metadata, req.Timestamp)

	runID := time.Now().Format("20060102150405") + "-" + req.Agent

	/*
		TODO: MariaDB integration — persist AI query result or pond status
		Example:
		if req.Action == "ai_query" {
			result := map[string]interface{}{...}
			_, err := database.DB.Exec(`INSERT INTO ai_logs (run_id, agent, action, request, response, created_at) VALUES (?, ?, ?, ?, ?, NOW())`,
				runID, req.Agent, req.Action, string(body), resultJSON, ...)
		}
		if req.Action == "update_pond" {
			_, err := database.DB.Exec(`UPDATE fish_data SET health_status = ? WHERE id = ?`, ...)
		}
	*/

	json.NewEncoder(w).Encode(AgentWebhookResponse{
		Success: true,
		Message: "Webhook received successfully",
		RunID:   runID,
		Data: map[string]string{
			"agent":     req.Agent,
			"action":    req.Action,
			"timestamp": req.Timestamp,
		},
	})
}
