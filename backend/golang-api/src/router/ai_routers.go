package router

import (
	"vms-api/src/controllers"

	"github.com/gorilla/mux"
)

// AIRoutes 配置所有 AI 相關的 API 路由
// ================================
// AI 服務：OpenCLI (Gemini) / OpenCLAW
// ================================
func AIRoutes(router *mux.Router) {

	// ═══════════════════════════════════════════════════════════════
	// 🤖 OpenCLI + Gemini AI API (/api/ai)
	// ═══════════════════════════════════════════════════════════════

	// POST /api/ai/query
	// 功能：透過 OpenCLI 發送問題到 Gemini AI
	// 用途：養殖管理員詢問天氣、養殖知識等
	router.HandleFunc("/api/ai/query", controllers.GeminiHandler).Methods("POST", "OPTIONS")

	// GET /api/ai/query
	// 功能：測試 AI 連線狀態
	router.HandleFunc("/api/ai/query", controllers.GeminiStatusHandler).Methods("GET", "OPTIONS")

	// POST /api/ai/conversation
	// 功能：創建新對話
	router.HandleFunc("/api/ai/conversation", controllers.CreateConversationHandler).Methods("POST", "OPTIONS")

	// GET /api/ai/conversation
	// 功能：獲取對話歷史
	router.HandleFunc("/api/ai/conversation", controllers.GetConversationHandler).Methods("GET", "OPTIONS")

	// ═══════════════════════════════════════════════════════════════
	// 🦞 OpenCLAW AI API (/api/openclaw)
	// ═══════════════════════════════════════════════════════════════

	// POST /api/openclaw/chat
	// 功能：透過 OpenCLAW 與 AI 對話
	// 用途：進階 AI 功能（需要認證）
	router.HandleFunc("/api/openclaw/chat", controllers.OpenCLAWChatHandler).Methods("POST", "OPTIONS")

	// ═══════════════════════════════════════════════════════════════
	// 📊 每日報告 API (/api/daily-report)
	// ═══════════════════════════════════════════════════════════════

	// GET /api/daily-report
	// 功能：生成當日養殖數據報告（結合 AI 總結）
	// router.HandleFunc("/api/daily-report", controllers.DailyReportHandler).Methods("GET", "OPTIONS")
}
