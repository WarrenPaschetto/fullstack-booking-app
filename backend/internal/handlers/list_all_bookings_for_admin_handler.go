package handlers

import (
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
)

func (h *Handler) ListAllBookingsForAdminHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookings, err := h.BookingService.ListAdminBookings(r.Context())
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list bookings", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, bookings)
	}
}
