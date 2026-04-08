package router

import (
	"vms-api/src/controllers"

	"github.com/gorilla/mux"
)

// LoginRoutes sets up authentication and user management routes
func LoginRoutes(router *mux.Router) {
	loginController := &controllers.LoginController{}

	// Authentication routes
	router.HandleFunc("/api/auth/login", loginController.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/signin", loginController.Login).Methods("POST", "OPTIONS") // Alternative endpoint

	// User management routes
	router.HandleFunc("/api/users/list", loginController.GetUsers).Methods("GET")          // Get all users
	router.HandleFunc("/api/users/profile", loginController.GetUserProfile).Methods("GET") // Get current user profile (placeholder)
}
