package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"vms-api/src/database"
	"vms-api/src/models"
)

// LoginController handles authentication-related operations
type LoginController struct{}

// Login handles user login
func (lc *LoginController) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		response := models.LoginResponse{
			Success: false,
			Message: "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		response := models.LoginResponse{
			Success: false,
			Message: "Username and password are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query user from database
	var user models.User
	var passwordHash string
	err := database.DB.QueryRow(`
		SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE username = ? AND is_active = true`,
		loginReq.Username).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			response := models.LoginResponse{
				Success: false,
				Message: "Invalid username or password",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}
		response := models.LoginResponse{
			Success: false,
			Message: "Database error",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// For now, compare plain text passwords (in production, use bcrypt)
	if passwordHash != loginReq.Password {
		response := models.LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate simple token (in production, use JWT)
	token := generateToken(user.Username)

	response := models.LoginResponse{
		Success: true,
		Token:   token,
		User:    &user,
		Message: "Login successful",
	}

	json.NewEncoder(w).Encode(response)
}

// generateToken creates a simple token (replace with JWT in production)
func generateToken(username string) string {
	return fmt.Sprintf("vms_token_%s_%d", username, time.Now().Unix())
}

// GetUsers returns all users (admin only)
func (lc *LoginController) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simple token validation (replace with proper JWT validation in production)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response := models.APIResponse{
			Success: false,
			Message: "Authorization header required",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Query all users
	rows, err := database.DB.Query(`
		SELECT id, username, email, role, is_active, created_at, updated_at
		FROM users
		WHERE is_active = true
		ORDER BY created_at DESC`)

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

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email,
			&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	response := models.APIResponse{
		Success: true,
		Data:    users,
		Message: "Users retrieved successfully",
	}
	json.NewEncoder(w).Encode(response)
}
