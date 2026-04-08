package models

import (
"time"
)

// User represents a user in the system
type User struct {
ID           int       `json:"id" db:"id"`
Username     string    `json:"username" db:"username"`
Email        string    `json:"email" db:"email"`
PasswordHash string    `json:"-" db:"password_hash"` // Don't expose password in JSON
Role         string    `json:"role" db:"role"`
IsActive     bool      `json:"is_active" db:"is_active"`
CreatedAt    time.Time `json:"created_at" db:"created_at"`
UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// FishData represents fish farming data
type FishData struct {
ID           int       `json:"id" db:"id"`
UserID       int       `json:"user_id" db:"user_id"`
FishType     string    `json:"fish_type" db:"fish_type"`
Quantity     int       `json:"quantity" db:"quantity"`
Weight       float64   `json:"weight" db:"weight"`
HealthStatus string    `json:"health_status" db:"health_status"`
CreatedAt    time.Time `json:"created_at" db:"created_at"`
UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
User         *User     `json:"user,omitempty" db:"-"` // Joined data
}

// WeatherData represents environmental data
type WeatherData struct {
ID               int       `json:"id" db:"id"`
Temperature      float64   `json:"temperature" db:"temperature"`
Humidity         float64   `json:"humidity" db:"humidity"`
PhLevel          float64   `json:"ph_level" db:"ph_level"`
DissolvedOxygen  float64   `json:"dissolved_oxygen" db:"dissolved_oxygen"`
Location         string    `json:"location" db:"location"`
RecordedAt       time.Time `json:"recorded_at" db:"recorded_at"`
}

// FeedRecord represents feeding data
type FeedRecord struct {
ID       int       `json:"id" db:"id"`
UserID   int       `json:"user_id" db:"user_id"`
FeedType string    `json:"feed_type" db:"feed_type"`
Quantity float64   `json:"quantity" db:"quantity"`
Unit     string    `json:"unit" db:"unit"`
FeedTime time.Time `json:"feed_time" db:"feed_time"`
User     *User     `json:"user,omitempty" db:"-"` // Joined data
}

// LoginRequest represents login request payload
type LoginRequest struct {
Username string `json:"username"`
Password string `json:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
Success bool    `json:"success"`
Message string  `json:"message"`
Token   string  `json:"token,omitempty"`
User    *User   `json:"user,omitempty"`
}

// APIResponse represents generic API response
type APIResponse struct {
Success bool        `json:"success"`
Message string      `json:"message"`
Data    interface{} `json:"data,omitempty"`
Error   string      `json:"error,omitempty"`
}

// Pagination represents pagination info
type Pagination struct {
Page       int `json:"page"`
Limit      int `json:"limit"`
Total      int `json:"total"`
TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents paginated API response
type PaginatedResponse struct {
Success    bool        `json:"success"`
Message    string      `json:"message"`
Data       interface{} `json:"data"`
Pagination Pagination  `json:"pagination"`
}
