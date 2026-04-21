package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
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
	// Set PATH to include nvm node
	envPath := "/home/ouo/.nvm/versions/node/v24.14.1/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

	// Step 1: Open Gemini
	cmd1 := exec.Command("opencli", "operate", "open", "https://gemini.google.com/app")
	cmd1.Env = append(os.Environ(), "PATH="+envPath)
	cmd1.Run()
	time.Sleep(5 * time.Second)

	// Step 2: Type query
	typeScript := "opencli operate eval \"(function(){ const ta = document.querySelector('rich-textarea')?.querySelector('div[contenteditable]'); if(ta){ ta.innerText = '" + query + "'; ta.dispatchEvent(new Event('input', { bubbles: true })); return 'typed'; } return 'fail'; })()\""
	cmd2 := exec.Command("sh", "-c", typeScript)
	cmd2.Run()

	// Step 3: Press Enter
	cmd3 := exec.Command("opencli", "operate", "keys", "Enter")
	cmd3.Run()
	time.Sleep(12 * time.Second)

	// Step 4: Extract response
	cmd4 := exec.Command("opencli", "operate", "eval", "(function(){ const c = document.querySelector('chat-window-content'); return c ? c.textContent.slice(0,3000) : 'no content'; })")
	output, err := cmd4.Output()
	if err != nil {
		return "", fmt.Errorf("failed to extract: %v", err)
	}

	// Step 5: Close
	cmd5 := exec.Command("opencli", "operate", "close")
	cmd5.Run()

	// Clean up the response
	lines := strings.Split(string(output), "\n")
	var responseLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.Contains(trimmed, "Update available") && !strings.HasPrefix(trimmed, "[") {
			responseLines = append(responseLines, trimmed)
		}
	}

	return strings.Join(responseLines, "\n"), nil
}
