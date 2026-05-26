package router

import (
	"vms-api/src/controllers"
	"vms-api/src/middleware"

	"github.com/gorilla/mux"
)

func WebhookRoutes(router *mux.Router) {
	sub := router.PathPrefix("/api/webhook").Subrouter()
	sub.Use(middleware.OpenClawAuthMiddleware)
	sub.HandleFunc("/agent", controllers.AgentWebhookHandler).Methods("POST", "OPTIONS")
}
