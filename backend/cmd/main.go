package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/handlers"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

func main() {
	// load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}

	// connect to Turso DB
	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Create sqlc Queries object
	queries := db.New(database)

	// Instantiate the service layer
	bookingSvc := service.NewBookingService(queries)

	// Instantiate HTTP handlers with the service
	h := handlers.NewHandler(bookingSvc)

	// Build the router
	r := chi.NewRouter()

	// Global middleware

	// 7. Define your routes
	r.Post("/bookings", h.CreateBookingHandler())
	r.Delete("/bookings/{id}", h.DeleteBookingHandler())
	r.Put("/bookings/{id}/reschedule", h.RescheduleBookingHandler())

	// 8. Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	// ⚙️ Create a custom server with timeouts
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Start server
	log.Printf("Starting server on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
