package handlers

import (
	"errors"
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) GetBookingByIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "User ID missing or not a UUID in context", nil)
			return
		}

		vars := mux.Vars(r)
		bookingIDStr, ok := vars["id"]
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing booking ID", nil)
			return
		}
		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid booking ID", err)
			return
		}
		if bookingID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Booking ID is required", nil)
			return
		}

		booking, err := h.BookingService.GetBookingByID(r.Context(), bookingID, userID)
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "Booking not found", nil)
		case errors.Is(err, service.ErrNotAuthorized):
			utils.RespondWithError(w, http.StatusForbidden, "Not allowed", nil)
		case err != nil:
			utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching booking", err)
		default:
			utils.RespondWithJSON(w, http.StatusOK, booking)
		}
	}
}
