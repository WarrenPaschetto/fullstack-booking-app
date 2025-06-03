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

func (h *Handler) DeleteBookingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "User ID missing or not a UUID in context", nil)
			return
		}
		isAdmin := middleware.IsAdminFromContext(r.Context())

		vars := mux.Vars(r)
		slotIDStr, ok := vars["id"]
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing slot ID", nil)
			return
		}
		slotID, err := uuid.Parse(slotIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid slot ID", err)
			return
		}
		if slotID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Booking ID is required", nil)
			return
		}

		err = h.BookingService.DeleteBooking(r.Context(), slotID, userID, isAdmin)
		if errors.Is(err, service.ErrBookingNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "Booking not found", nil)
			return
		}
		if errors.Is(err, service.ErrNotAuthorized) {
			utils.RespondWithError(w, http.StatusForbidden, "Not allowed to delete this booking", nil)
			return
		}
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete booking", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
