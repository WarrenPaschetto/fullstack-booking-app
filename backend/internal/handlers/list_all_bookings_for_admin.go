package handlers

import (
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
)

func (h *Handler) ListAllBookingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookings, err := h.BookingService.ListAllBookings(r.Context())
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError,
				"Failed to list all bookings", err)
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, bookings)
	}
}
