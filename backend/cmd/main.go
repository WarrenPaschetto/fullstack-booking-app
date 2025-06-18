package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
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
	availabilitySvc := service.NewAvailabilityService(queries)
	r := mux.NewRouter()

	r.HandleFunc("/api/register", handlers.RegisterHandler(queries)).Methods("POST")
	r.HandleFunc("/api/login", handlers.LoginHandler(queries)).Methods("POST")
	r.HandleFunc("/api/availabilities/free", handlers.ListAllFreeSlotsHandler(queries)).Methods("GET")

	bookings := r.PathPrefix("/api/bookings").Subrouter()
	bookings.Use(middleware.AuthMiddleware)

	bookings.Handle("/create", h.CreateBookingHandler()).Methods("POST")
	//bookings.Handle("", h.ListBookingsForUserHandler()).Methods("GET")
	//bookings.Handle("/{id}", h.GetBookingByIDHandler()).Methods("GET")
	bookings.Handle("/{id}", h.RescheduleBookingHandler()).Methods("PUT")
	bookings.Handle("/{id}", h.DeleteBookingHandler()).Methods("DELETE")

	admins := r.PathPrefix("/api").Subrouter()
	admins.Use(middleware.AuthMiddleware)

	admins.Handle("/bookings/all", h.ListAllBookingsHandler()).Methods("GET")
	admins.Handle("/users/all", handlers.ListAllUsersHandler(queries)).Methods("GET")
	admins.Handle("/avail-pattern/create", handlers.CreateAvailabilityPatternHandler(availabilitySvc)).Methods("POST")

	// Logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
		})
	})

	// Add explicit 404 logger
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("404 Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
	})

	// Wrap router in CORS AFTER all routes
	handlerWithCORS := middleware.CORS(r)

	// Serve
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handlerWithCORS,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Listening on port %s…\n", port)
	log.Fatal(srv.ListenAndServe())
}
