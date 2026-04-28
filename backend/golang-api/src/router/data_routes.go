package router

import (
	"vms-api/src/controllers"

	"github.com/gorilla/mux"
)

func DataRoutes(router *mux.Router) {
	fishController := &controllers.FishController{}
	weatherController := &controllers.WeatherController{}
	feedController := &controllers.FeedController{}
	aiController := &controllers.AIController{}
	cwaController := &controllers.CWAOpenDataController{}

	router.HandleFunc("/api/fish/data", fishController.GetFishData).Methods("GET")
	router.HandleFunc("/api/fish/data", fishController.AddFishData).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/weather/data", weatherController.GetWeatherData).Methods("GET")
	router.HandleFunc("/api/weather/data/cwa", cwaController.GetCWAWeatherData).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/weather/forecast", cwaController.GetCWAForecast).Methods("GET", "OPTIONS")

	router.HandleFunc("/api/feed/data", feedController.GetFeedData).Methods("GET")

	router.HandleFunc("/api/ai/query", aiController.QueryAI).Methods("POST", "OPTIONS")
}
