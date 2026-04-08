package router

import (
	"vms-api/src/controllers"

	"github.com/gorilla/mux"
)

// DataRoutes sets up data management routes
func DataRoutes(router *mux.Router) {
	fishController := &controllers.FishController{}
	weatherController := &controllers.WeatherController{}
	feedController := &controllers.FeedController{}

	// Fish data routes
	router.HandleFunc("/api/fish/data", fishController.GetFishData).Methods("GET")
	router.HandleFunc("/api/fish/data", fishController.AddFishData).Methods("POST", "OPTIONS")

	// Weather data routes
	router.HandleFunc("/api/weather/data", weatherController.GetWeatherData).Methods("GET")

	// Feed data routes
	router.HandleFunc("/api/feed/data", feedController.GetFeedData).Methods("GET")
}
