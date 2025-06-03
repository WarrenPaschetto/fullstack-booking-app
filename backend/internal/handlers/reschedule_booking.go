package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RescheduleBookingRequest struct {
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int32     `json:"duration_minutes"`
}

func (h *Handler) RescheduleBookingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "User ID missing or not a UUID in context", nil)
			return
		}
		isAdmin := middleware.IsAdminFromContext(r.Context())

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

		var req RescheduleBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		if req.DurationMinutes <= 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "Duration minutes must be greater than zero", nil)
			return
		}

		updated, err := h.BookingService.RescheduleBooking(
			r.Context(),
			bookingID,
			userID,
			req.AppointmentStart,
			req.DurationMinutes,
			isAdmin,
		)
		if errors.Is(err, service.ErrBookingConflict) {
			utils.RespondWithError(w, http.StatusConflict, "Time slot already booked", nil)
			return
		}
		if errors.Is(err, service.ErrNotAuthorized) {
			utils.RespondWithError(w, http.StatusForbidden, "You are not authorized to reschedule this booking", nil)
			return
		}
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't reschedule", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, updated)
	}
}
