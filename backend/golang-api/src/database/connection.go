package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// DB is the global database connection
var DB *sql.DB

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// InitDB initializes the database connection
func InitDB() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	config := Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "vms_user"),
		Password: getEnv("DB_PASSWORD", "vms_password"),
		DBName:   getEnv("DB_NAME", "vms_db"),
	}

	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)

	log.Println("Database connection established successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// CreateTables creates necessary database tables
func CreateTables() error {
	// Drop existing tables if they exist
	dropQueries := []string{
		"DROP TABLE IF EXISTS feed_data",
		"DROP TABLE IF EXISTS fish_data",
		"DROP TABLE IF EXISTS weather_data",
		"DROP TABLE IF EXISTS users",
	}

	for _, query := range dropQueries {
		if _, err := DB.Exec(query); err != nil {
			log.Printf("Warning: Failed to drop table: %v", err)
		}
	}

	// Users table
	usersTable := `
	CREATE TABLE users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role ENUM('administrator', 'user') DEFAULT 'user',
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`

	// Fish data table
	fishDataTable := `
	CREATE TABLE IF NOT EXISTS fish_data (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		fish_type VARCHAR(100) NOT NULL,
		quantity INT NOT NULL,
		weight DECIMAL(10,2),
		health_status ENUM('excellent', 'good', 'fair', 'poor') DEFAULT 'good',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`

	// Weather data table
	weatherDataTable := `
	CREATE TABLE IF NOT EXISTS weather_data (
		id INT AUTO_INCREMENT PRIMARY KEY,
		temperature DECIMAL(5,2),
		humidity DECIMAL(5,2),
		ph_level DECIMAL(4,2),
		dissolved_oxygen DECIMAL(5,2),
		location VARCHAR(100),
		recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	// Feed data table
	feedDataTable := `
	CREATE TABLE IF NOT EXISTS feed_data (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		feed_type VARCHAR(100) NOT NULL,
		quantity DECIMAL(10,2) NOT NULL,
		unit VARCHAR(20) DEFAULT 'kg',
		feed_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`

	tables := []string{usersTable, fishDataTable, weatherDataTable, feedDataTable}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	log.Println("Database tables created successfully")
	return nil
}

// SeedData inserts initial data into the database
func SeedData() error {
	// Check if admin user exists
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "admin").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check admin user: %v", err)
	}

	if count == 0 {
		// Insert default admin user (password: admin123)
		// In production, use proper password hashing
		_, err := DB.Exec(`
			INSERT INTO users (username, email, password_hash, role)
			VALUES (?, ?, ?, ?)`,
			"admin", "admin@vms.com", "admin123", "administrator")
		if err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}

		// Insert default regular user (password: user123)
		_, err = DB.Exec(`
			INSERT INTO users (username, email, password_hash, role)
			VALUES (?, ?, ?, ?)`,
			"user", "user@vms.com", "user123", "user")
		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}

		log.Println("Default users created successfully")
	}

	return nil
}
