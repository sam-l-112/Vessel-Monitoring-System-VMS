package models

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

var DB *sqlx.DB

func init() {
	err := godotenv.Load()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
}
