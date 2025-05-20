package handlers

import (
	"errors"
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) GetBookingByIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookingId, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid booking ID", err)
			return
		}

		raw := r.Context().Value(middleware.UserIDKey)
		userID, ok := raw.(uuid.UUID)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "User ID missing or not a UUID in context", nil)
			return
		}

		booking, err := h.BookingService.GetBookingByID(r.Context(), bookingId, userID)
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			utils.RespondWithError(w, http.StatusNotFound, "booking not found", nil)
		case errors.Is(err, service.ErrNotAuthorized):
			utils.RespondWithError(w, http.StatusForbidden, "not allowed", nil)
		case err != nil:
			utils.RespondWithError(w, http.StatusInternalServerError, "error fetching booking", err)
		default:
			utils.RespondWithJSON(w, http.StatusOK, booking)
		}
	}
}
