package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"vms-api/src/controllers"
	"vms-api/src/database"
	routerpkg "vms-api/src/router"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	// Load environment variables
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		dotenvPath := filepath.Join(execDir, ".env")
		if err := godotenv.Load(dotenvPath); err != nil {
			log.Printf("Warning: .env file not found at %s, trying current working directory", dotenvPath)
			if err := godotenv.Load(); err != nil {
				log.Println("Warning: .env file not found in current working directory, using system environment variables")
			}
		}
	} else {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file not found, using system environment variables")
		}
	}

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
	} else {
		defer database.CloseDB()
	}

	if err := database.CreateTables(); err != nil {
		log.Printf("Warning: Failed to create database tables: %v", err)
	}

	if err := database.SeedData(); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	r.Use(loggingMiddleware)

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"success":true,"message":"VMS API is running","data":{"timestamp":"%s","version":"1.0.0"}}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	r.HandleFunc("/api/opencli/gemini/status", controllers.GeminiStatusHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/opencli/gemini/chat", controllers.GeminiHandler).Methods("POST", "OPTIONS")

	routerpkg.LoginRoutes(r)
	routerpkg.DataRoutes(r)
	routerpkg.AIRoutes(r)

	r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped gracefully")
}
