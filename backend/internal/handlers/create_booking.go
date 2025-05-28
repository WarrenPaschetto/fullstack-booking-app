package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
)

type BookingRequest struct {
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int       `json:"duration_minutes"`
}

func (h *Handler) CreateBookingHandler() http.HandlerFunc {
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
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		id := uuid.New()
		booking, err := h.BookingService.CreateBooking(r.Context(), id, userID, req.AppointmentStart, req.DurationMinutes)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create booking", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, booking)
	}
}
