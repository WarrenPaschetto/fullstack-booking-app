package main

import (
	"log"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("DATABASE_URL")

	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()
}
