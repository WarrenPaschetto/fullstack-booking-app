package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RescheduleBookingRequest struct {
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int       `json:"duration_minutes"`
	ID               uuid.UUID `json:"id"`
}

func (h *Handler) RescheduleBookingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid booking ID", err)
			return
		}

		var req RescheduleBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid request body", err)
			return
		}

		updated, err := h.BookingService.RescheduleBooking(
			r.Context(),
			id,
			req.AppointmentStart,
			req.DurationMinutes,
		)
		if errors.Is(err, service.ErrBookingConflict) {
			utils.RespondWithError(w, http.StatusConflict, "time slot already booked", nil)
			return
		}
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "couldn't reschedule", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, updated)
	}
}
