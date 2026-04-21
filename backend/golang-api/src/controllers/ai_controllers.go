package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

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
	// Run a shell script to query Gemini
	script := fmt.Sprintf(`
opencli operate open https://gemini.google.com/app > /dev/null 2>&1
sleep 5
opencli operate eval "$(cat <<'SCRIPT'
(function(){ 
  const ta = document.querySelector('rich-textarea')?.querySelector('div[contenteditable]'); 
  if(ta){ ta.innerText = %q; ta.dispatchEvent(new Event('input', { bubbles: true })); return 'typed'; } 
  return 'fail'; 
})()
SCRIPT
)" > /dev/null 2>&1
opencli operate keys "Enter" > /dev/null 2>&1
sleep 12
opencli operate eval "(function(){ const c = document.querySelector('chat-window-content'); return c ? c.textContent.slice(0,3000) : 'no content'; })()"
opencli operate close > /dev/null 2>&1
`, query)

	cmd := exec.Command("bash", "-c", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("shell script failed: %v", err)
	}

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

// Deprecated: use queryGeminiShell instead
func queryGemini(query string) (string, error) {
	return queryGeminiShell(query)
}

func runOpenCliCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		return fmt.Errorf("invalid command")
	}

	command := exec.Command(parts[0], parts[1:]...)
	return command.Run()
}
