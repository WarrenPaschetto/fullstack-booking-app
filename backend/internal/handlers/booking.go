package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookingRequest struct {
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int       `json:"duration_minutes"`
}

type BookingResponse struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int       `json:"duration_minutes"`
	UserID           uuid.UUID `json:"user_id"`
}

type DeleteBookingRequest struct {
	ID uuid.UUID `json:"id"`
}

type RescheduleBookingRequest struct {
	AppointmentStart time.Time `json:"appointment_start"`
	DurationMinutes  int       `json:"duration_minutes"`
	ID               uuid.UUID `json:"id"`
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

func (h *Handler) RescheduleBookingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid booking ID", err)
			return
		}

		var body struct {
			NewStart        time.Time `json:"new_start"`
			DurationMinutes int       `json:"duration_minutes"`
		}

		decoder := json.NewDecoder(r.Body)
		req := RescheduleBookingRequest{}
		err = decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		updated, err := h.BookingService.RescheduleBooking(r.Context(), id, body.NewStart, body.DurationMinutes)
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
