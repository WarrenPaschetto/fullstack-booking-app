package main

import (
	"log"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
)

func main() {
	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()
}
