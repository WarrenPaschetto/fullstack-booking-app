package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
)

type BookingRequest struct {
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int       `json:"duration_minutes"`
}

type BookingResponse struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int       `json:"duration_minutes"`
	UserID           uuid.UUID `json:"user_id"`
}

func CreateBookingHandler(queries db.BookingQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.Context().Value(middleware.UserIDKey)
		userID, ok := raw.(uuid.UUID)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "User ID missing or not a UUID in context", nil)
			return
		}

		decoder := json.NewDecoder(r.Body)
		req := BookingRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid request body", err)
			return
		}

		// Check for overlapping bookings
		overlaps, err := queries.GetOverlappingBookings(r.Context(), db.GetOverlappingBookingsParams{
			NewStart: req.AppointmentStart,
			NewEnd:   req.AppointmentStart.Add(time.Duration(req.DurationMinutes) * time.Minute),
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check for overlapping bookings", err)
			return
		}
		if len(overlaps) > 0 {
			utils.RespondWithError(w, http.StatusConflict, "Booking overlaps with an existing one", nil)
			return
		}

		id := uuid.New()
		err = queries.CreateBooking(r.Context(), db.CreateBookingParams{
			AppointmentStart: req.AppointmentStart,
			DurationMinutes:  int64(req.DurationMinutes),
			UserID:           userID,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create booking", err)
			return
		}

		now := time.Now()
		utils.RespondWithJSON(w, http.StatusCreated, BookingResponse{
			ID:               id,
			CreatedAt:        now,
			UpdatedAt:        now,
			AppointmentStart: req.AppointmentStart,
			DurationMinutes:  req.DurationMinutes,
			UserID:           userID,
		})
	}
}
