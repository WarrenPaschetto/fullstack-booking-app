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
	ID               string    `json:"id"`
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int32     `json:"duration_minutes"`
}

func (h *Handler) CreateBookingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
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

		id, err := uuid.Parse(req.ID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Unable to parse ID", err)
			return
		}
		booking, err := h.BookingService.CreateBooking(r.Context(), id, userID, req.AppointmentStart, req.DurationMinutes, id)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create booking", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, booking)
	}
}
