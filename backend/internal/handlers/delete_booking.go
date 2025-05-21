package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
)

type DeleteBookingRequest struct {
	ID uuid.UUID `json:"id"`
}

func (h *Handler) DeleteBookingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.Context().Value(middleware.UserIDKey)
		userID, ok := raw.(uuid.UUID)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "User ID missing or not a UUID in context", nil)
			return
		}

		decoder := json.NewDecoder(r.Body)
		req := DeleteBookingRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		if req.ID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Booking ID is required", nil)
			return
		}

		if err = h.BookingService.DeleteBooking(r.Context(), req.ID, userID); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete booking", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
