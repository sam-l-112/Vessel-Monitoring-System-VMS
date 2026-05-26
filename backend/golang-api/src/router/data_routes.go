package router

import (
	"vms-api/src/controllers"

	"github.com/gorilla/mux"
)

// DataRoutes 配置所有數據相關的 API 路由
// ================================
// 數據流向：前端 (Vue.js) → Nginx → Go API → MariaDB
// ================================
func DataRoutes(router *mux.Router) {
	// 初始化控制器
	fishController := &controllers.FishController{}
	weatherController := &controllers.WeatherController{}
	feedController := &controllers.FeedController{}
	cwaController := &controllers.CWAOpenDataController{}

	// ═══════════════════════════════════════════════════════════════
	// 🐟 魚類數據 API (/api/fish/data)
	// ═══════════════════════════════════════════════════════════════

	// GET /api/fish/data - 取得所有魚類
	router.HandleFunc("/api/fish/data", fishController.GetFishData).Methods("GET")

	// POST /api/fish/data - 新增魚類
	router.HandleFunc("/api/fish/data", fishController.AddFishData).Methods("POST")

	// PUT /api/fish/data - 更新魚類 (ID 在 body 中)
	router.HandleFunc("/api/fish/data", fishController.UpdateFishData).Methods("PUT")

	// PUT /api/fish/data/{id} - 更新魚類 (ID 在 URL 中)
	router.HandleFunc("/api/fish/data/{id}", fishController.UpdateFishDataByID).Methods("PUT")

	// DELETE /api/fish/data/{id} - 刪除魚類
	router.HandleFunc("/api/fish/data/{id}", fishController.DeleteFishData).Methods("DELETE")

	// ═══════════════════════════════════════════════════════════════
	// 🌾 飼料管理 API (/api/feed/data)
	// ═══════════════════════════════════════════════════════════════

	// GET /api/feed/data - 取得所有飼料記錄
	router.HandleFunc("/api/feed/data", feedController.GetFeedData).Methods("GET")

	// POST /api/feed/data - 新增飼料
	router.HandleFunc("/api/feed/data", feedController.AddFeedData).Methods("POST")

	// PUT /api/feed/data - 更新飼料 (ID 在 body 中)
	router.HandleFunc("/api/feed/data", feedController.UpdateFeedData).Methods("PUT")

	// PUT /api/feed/data/{id} - 更新飼料 (ID 在 URL 中)
	router.HandleFunc("/api/feed/data/{id}", feedController.UpdateFeedDataByID).Methods("PUT")

	// DELETE /api/feed/data/{id} - 刪除飼料
	router.HandleFunc("/api/feed/data/{id}", feedController.DeleteFeedData).Methods("DELETE")

	// ═══════════════════════════════════════════════════════════════
	// 🌤️ 天氣/環境數據 API (/api/weather/data)不
	// ═══════════════════════════════════════════════════════════════

	// GET /api/weather/data - 取得本地天氣數據
	router.HandleFunc("/api/weather/data", weatherController.GetWeatherData).Methods("GET")

	// GET /api/weather/data/cwa - 取得中央氣象局數據
	router.HandleFunc("/api/weather/data/cwa", cwaController.GetCWAWeatherData).Methods("GET")

	// GET /api/weather/forecast - 取得天氣預報
	router.HandleFunc("/api/weather/forecast", cwaController.GetCWAForecast).Methods("GET")

	// ═══════════════════════════════════════════════════════════════
	// 🚨 示警系統 API (/api/dataset)
	// ═══════════════════════════════════════════════════════════════

	// GET /api/dataset
	// 功能：取得示警類型清單
	// 參數：apikey (必填), format (必填, json/xml), limit (選填), offset (選填)
	router.HandleFunc("/api/dataset", controllers.GetAlertTypes).Methods("GET", "OPTIONS")

	// ═══════════════════════════════════════════════════════════════
	// 📡 NCDR 示警訊息 API (/api/datastore)
	// ═══════════════════════════════════════════════════════════════

	// GET /api/datastore
	// 功能：查詢 NCDR 國家災害防救中心即時示警訊息
	// 參數：apikey (必填), format (必填, json/xml), capcode, govcode, countycode,
	//       effectivetime, expirestime, limit, offset (選填)
	router.HandleFunc("/api/datastore", controllers.GetAlertMessages).Methods("GET", "OPTIONS")

}
