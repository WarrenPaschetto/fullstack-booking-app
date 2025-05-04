package main

import (
	"log"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}

	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()
}
