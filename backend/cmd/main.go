package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/handlers"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	dbConn, err := db.ConnectDB(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbConn.Close()

	queries := db.New(dbConn)
	bookingSvc := service.NewBookingService(queries)
	h := handlers.NewHandler(bookingSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", handlers.RegisterHandler(queries))
	mux.HandleFunc("/api/login", handlers.LoginHandler(queries))
	mux.Handle("/api/bookings/create", middleware.AuthMiddleware(h.CreateBookingHandler()))

	srv := &http.Server{
		Addr:         ":" + os.Getenv("PORT"),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
