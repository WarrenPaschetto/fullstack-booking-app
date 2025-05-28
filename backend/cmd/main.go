package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/handlers"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
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

	// instantiate your business-logic services around queries
	bookingSvc := service.NewBookingService(queries)
	//adminSvc := service.NewAdminService(queries)
	//availabilitySvc := service.NewAvailabilityService(queries)

	// now build your handlers.Handler
	h := handlers.NewHandler(bookingSvc)
	// —or if you don’t have NewHandler, just:
	// h := &handlers.Handler{
	//   BookingService:     bookingSvc,
	//   AdminService:       adminSvc,
	//   AvailabilityService: availabilitySvc,
	// }

	mux := http.NewServeMux()

	// public auth routes
	mux.HandleFunc("/api/register", handlers.RegisterHandler(queries))
	mux.HandleFunc("/api/login", handlers.LoginHandler(queries))
	mux.HandleFunc("api/bookings/create", h.CreateBookingHandler())

	// protected booking route
	//mux.Handle(
	//	"/api/bookings/create",
	//	middleware.AuthMiddleware(queries)(
	//		h.CreateBookingHandler(),
	//	),
	//)

	// …and so on for availability, admin, etc.

	srv := &http.Server{
		Addr:         ":" + os.Getenv("PORT"),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
