package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"

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
	// 使用 shell script 在後台查詢
	cmd := exec.Command("/bin/bash", "/home/ouo/project_f/backend/golang-api/query_gemini.sh", query)
	cmd.Env = append(os.Environ(), "PATH=/home/ouo/.nvm/versions/node/v24.14.1/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	output, err := cmd.Output()
	if err != nil {
		return "無法取得回應: " + err.Error(), nil
	}

	return string(output), nil
}
