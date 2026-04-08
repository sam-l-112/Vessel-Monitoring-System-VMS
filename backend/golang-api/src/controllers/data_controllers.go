package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"vms-api/src/database"
	"vms-api/src/models"
)

// FishController handles fish farming data operations
type FishController struct{}

// GetFishData returns all fish data
func (fc *FishController) GetFishData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query(`
		SELECT f.id, f.user_id, f.fish_type, f.quantity, f.weight, f.health_status,
			   f.created_at, f.updated_at, u.username
		FROM fish_data f
		LEFT JOIN users u ON f.user_id = u.id
		ORDER BY f.created_at DESC`)

	if err != nil {
		response := models.APIResponse{
			Success: false,
			Message: "Database error",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	var fishData []models.FishData
	for rows.Next() {
		var fish models.FishData
		var username sql.NullString
		err := rows.Scan(
			&fish.ID, &fish.UserID, &fish.FishType, &fish.Quantity,
			&fish.Weight, &fish.HealthStatus, &fish.CreatedAt, &fish.UpdatedAt, &username)
		if err != nil {
			continue
		}
		if username.Valid {
			fish.User = &models.User{Username: username.String}
		}
		fishData = append(fishData, fish)
	}

	response := models.APIResponse{
		Success: true,
		Data:    fishData,
		Message: "Fish data retrieved successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// AddFishData adds new fish data
func (fc *FishController) AddFishData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var fishReq struct {
		UserID       int     `json:"user_id"`
		FishType     string  `json:"fish_type"`
		Quantity     int     `json:"quantity"`
		Weight       float64 `json:"weight"`
		HealthStatus string  `json:"health_status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&fishReq); err != nil {
		response := models.APIResponse{
			Success: false,
			Message: "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	_, err := database.DB.Exec(`
		INSERT INTO fish_data (user_id, fish_type, quantity, weight, health_status)
		VALUES (?, ?, ?, ?, ?)`,
		fishReq.UserID, fishReq.FishType, fishReq.Quantity, fishReq.Weight, fishReq.HealthStatus)

	if err != nil {
		response := models.APIResponse{
			Success: false,
			Message: "Failed to add fish data",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Fish data added successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// WeatherController handles environmental data operations
type WeatherController struct{}

// GetWeatherData returns weather/environmental data
func (wc *WeatherController) GetWeatherData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query(`
		SELECT id, temperature, humidity, ph_level, dissolved_oxygen, location, recorded_at
		FROM weather_data
		ORDER BY recorded_at DESC
		LIMIT 100`)

	if err != nil {
		response := models.APIResponse{
			Success: false,
			Message: "Database error",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	var weatherData []models.WeatherData
	for rows.Next() {
		var weather models.WeatherData
		err := rows.Scan(
			&weather.ID, &weather.Temperature, &weather.Humidity,
			&weather.PhLevel, &weather.DissolvedOxygen, &weather.Location, &weather.RecordedAt)
		if err != nil {
			continue
		}
		weatherData = append(weatherData, weather)
	}

	response := models.APIResponse{
		Success: true,
		Data:    weatherData,
		Message: "Weather data retrieved successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// FeedController handles feeding data operations
type FeedController struct{}

// GetFeedData returns feeding records
func (fc *FeedController) GetFeedData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query(`
		SELECT f.id, f.user_id, f.feed_type, f.quantity, f.unit, f.feed_time, u.username
		FROM feed_data f
		LEFT JOIN users u ON f.user_id = u.id
		ORDER BY f.feed_time DESC`)

	if err != nil {
		response := models.APIResponse{
			Success: false,
			Message: "Database error",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	var feedData []models.FeedRecord
	for rows.Next() {
		var feed models.FeedRecord
		var username sql.NullString
		err := rows.Scan(
			&feed.ID, &feed.UserID, &feed.FeedType, &feed.Quantity,
			&feed.Unit, &feed.FeedTime, &username)
		if err != nil {
			continue
		}
		if username.Valid {
			feed.User = &models.User{Username: username.String}
		}
		feedData = append(feedData, feed)
	}

	response := models.APIResponse{
		Success: true,
		Data:    feedData,
		Message: "Feed data retrieved successfully",
	}
	json.NewEncoder(w).Encode(response)
}
