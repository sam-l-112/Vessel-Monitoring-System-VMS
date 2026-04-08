package router

import (
	"vms-api/src/controllers"

	"github.com/gorilla/mux"
)

// LoginRoutes sets up login-related routes
func LoginRoutes(router *mux.Router) {
	loginController := &controllers.LoginController{}

	// Authentication routes
	router.HandleFunc("/api/login", loginController.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/login", loginController.Login).Methods("POST", "OPTIONS")

	// User management routes (protected)
	router.HandleFunc("/api/users", loginController.GetUsers).Methods("GET")
}
