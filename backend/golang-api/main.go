package main

import (
"fmt"
"log"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"vms-api/src/database"
"vms-api/src/router"

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
if err := godotenv.Load(); err != nil {
log.Println("Warning: .env file not found, using system environment variables")
}

// Initialize database connection
if err := database.InitDB(); err != nil {
log.Printf("Warning: Failed to connect to database: %v", err)
}
defer database.CloseDB()

// Create database tables
if err := database.CreateTables(); err != nil {
log.Printf("Warning: Failed to create database tables: %v", err)
}

// Seed initial data
if err := database.SeedData(); err != nil {
log.Printf("Warning: Failed to seed data: %v", err)
}

// Create router
r := mux.NewRouter()

// Apply global middleware
r.Use(corsMiddleware)
r.Use(loggingMiddleware)

// Health check route
r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
fmt.Fprintf(w, `{"success":true,"message":"VMS API is running","data":{"timestamp":"%s","version":"1.0.0"}}`,
time.Now().Format(time.RFC3339))
}).Methods("GET")

// Setup routes from packages
router.LoginRoutes(r)
router.DataRoutes(r)

// Static file server for documentation (optional)
r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))

// Get server configuration
port := os.Getenv("SERVER_PORT")
if port == "" {
port = "8080"
}

host := os.Getenv("SERVER_HOST")
if host == "" {
host = "0.0.0.0"
}

serverAddr := fmt.Sprintf("%s:%s", host, port)

// Create server
server := &http.Server{
Addr:         serverAddr,
Handler:      r,
ReadTimeout:  15 * time.Second,
WriteTimeout: 15 * time.Second,
IdleTimeout:  60 * time.Second,
}

// Channel to listen for interrupt signal
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)

// Start server in a goroutine
go func() {
fmt.Printf("🚀 VMS API Server starting on http://%s\n", serverAddr)
fmt.Println("📋 Available endpoints:")
fmt.Println("  POST /api/login - User login")
fmt.Println("  GET  /api/health - Health check")
fmt.Println("  GET  /api/users - Get users (requires auth)")
fmt.Println("👤 Test credentials:")
fmt.Println("  admin:admin123 (Administrator)")
fmt.Println("  user:user123 (Regular User)")
fmt.Println("🛑 Press Ctrl+C to stop the server")

if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
log.Fatalf("Server failed to start: %v", err)
}
}()

// Wait for interrupt signal
<-c
fmt.Println("\n🛑 Shutting down server...")

// Graceful shutdown
shutdownTimeout := 30 * time.Second
shutdownChan := make(chan bool, 1)

go func() {
if err := server.Close(); err != nil {
log.Printf("Server forced to shutdown: %v", err)
}
shutdownChan <- true
}()

select {
case <-shutdownChan:
fmt.Println("✅ Server shutdown gracefully")
case <-time.After(shutdownTimeout):
fmt.Println("⚠️  Server shutdown timeout, forcing shutdown")
}

fmt.Println("👋 Goodbye!")
}
