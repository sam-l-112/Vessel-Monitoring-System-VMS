package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	openCLIHomePath = "/home/ouo/OpenCLI/dist/src/main.js"
	nodePathStatic  = "/home/ouo/.nvm/versions/node/v24.15.0/bin/node"
	geminiURL       = "https://gemini.google.com/app/7aeec6192d00009f?hl=zh-tw"
)

// EnsureOpenCLIReady checks whether OpenCLI daemon is running.
func EnsureOpenCLIReady() error {
	openCLI, nodePath, err := resolveOpenCLIPaths()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, nodePath, openCLI, "daemon", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("daemon status failed: %v: %s", err, strings.TrimSpace(string(output)))
	}

	outStr := string(output)
	if !strings.Contains(outStr, "Daemon: running") {
		return fmt.Errorf("daemon not running: %s", outStr)
	}

	return nil
}

// GeminiQueryService sends a message through OpenCLI to Gemini and returns the answer.
func GeminiQueryService(ctx context.Context, message string) (string, string, error) {
	openCLI, nodePath, err := resolveOpenCLIPaths()
	if err != nil {
		return "", "", err
	}

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	if err := openGeminiPage(ctx, nodePath, openCLI); err != nil {
		return "", "", err
	}

	if err := focusGeminiInput(ctx, nodePath, openCLI); err != nil {
		return "", "", err
	}

	if err := sendGeminiMessage(ctx, nodePath, openCLI, message); err != nil {
		return "", "", err
	}

	_, fullHTML, err := waitForGeminiResponse(ctx, nodePath, openCLI)
	if err != nil {
		return "", fullHTML, err
	}

	return extractLastAIResponse(fullHTML), fullHTML, nil
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

func openGeminiPage(ctx context.Context, nodePath, openCLI string) error {
	cmd := exec.CommandContext(ctx, nodePath, openCLI, "browser", "open", geminiURL)
	cmd.Dir = "/home/ouo/OpenCLI"
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to open Gemini page: %v: %s", err, string(out))
	}
	time.Sleep(3 * time.Second)
	return nil
}

func focusGeminiInput(ctx context.Context, nodePath, openCLI string) error {
	cmd := exec.CommandContext(ctx, nodePath, openCLI, "browser", "eval", `
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
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to focus Gemini input: %v", err)
	}
	time.Sleep(2 * time.Second)
	return nil
}

func sendGeminiMessage(ctx context.Context, nodePath, openCLI, message string) error {
	escaped := strings.ReplaceAll(message, "'", "\\'")
	cmd := exec.CommandContext(ctx, nodePath, openCLI, "browser", "eval", fmt.Sprintf(`
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
		return fmt.Errorf("failed to send message: %v: %s", err, string(out))
	}
	return nil
}

func waitForGeminiResponse(ctx context.Context, nodePath, openCLI string) (string, string, error) {
	cmd := exec.CommandContext(ctx, nodePath, openCLI, "browser", "eval", `
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
	fullText := strings.TrimSpace(string(replyOutput))
	if err != nil {
		return "", fullText, fmt.Errorf("failed to read Gemini response: %v: %s", err, fullText)
	}
	if strings.HasPrefix(fullText, "ERROR") || strings.Contains(fullText, "INPUT_NOT_FOUND") {
		return "", fullText, fmt.Errorf("Gemini browser eval error: %s", fullText)
	}
	return fullText, fullText, nil
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
