package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"vms-api/src/models"
)

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

func queryGeminiShell(query string) (string, error) {
	cmd := exec.Command("/bin/bash", "/home/ouo/project_f/backend/golang-api/query_gemini.sh", query)
	cmd.Env = append(os.Environ(), "PATH=/home/ouo/.nvm/versions/node/v24.14.1/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

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
