package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestRescheduleBookingHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	reqBody := RescheduleBookingRequest{
		AppointmentStart: now.Add(time.Hour),
		DurationMinutes:  30,
		ID:               bookingID,
	}
	jsonBody, _ := json.Marshal(reqBody)

	fakeBooking := db.Booking{
		ID:               bookingID,
		UserID:           userID,
		AppointmentStart: now.Add(time.Hour),
		DurationMinutes:  30,
		CreatedAt:        now.Add(-time.Hour),
		UpdatedAt:        now.Add(-time.Minute),
	}

	tests := []struct {
		name           string
		routeID        string
		body           []byte
		mockReschedule func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
		expectStatus   int
	}{
		{
			name:    "Rescheduled booking success",
			routeID: bookingID.String(),
			body:    jsonBody,
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				if arg.ID != bookingID {
					t.Errorf("expected ID %s, got %s", bookingID, arg.ID)
				}
				if !arg.AppointmentStart.Equal(fakeBooking.AppointmentStart) {
					t.Errorf("expected start %s, got %s", fakeBooking.AppointmentStart, arg.AppointmentStart)
				}
				return fakeBooking, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "Bad UUID",
			routeID:      "not-a-uuid",
			body:         jsonBody,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Malformed JSON",
			routeID:      bookingID.String(),
			body:         []byte(`{invalid json`),
			expectStatus: http.StatusBadRequest,
		},
		{
			name:    "Booking conflict",
			routeID: bookingID.String(),
			body:    jsonBody,
			mockReschedule: func(_ context.Context, _ db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, service.ErrBookingConflict
			},
			expectStatus: http.StatusConflict,
		},
		{
			name:    "Internal error",
			routeID: bookingID.String(),
			body:    jsonBody,
			mockReschedule: func(_ context.Context, _ db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, errors.New("boom")
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockBookingQueries{
				GetBookingByIDFn: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
					return fakeBooking, nil
				},
				GetOverlappingBookingsFn: func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
					return nil, nil
				},
				RescheduleBookingFn: tt.mockReschedule,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.RescheduleBookingHandler()

			req := httptest.NewRequest(http.MethodPatch, "/bookings/"+tt.routeID, bytes.NewReader(tt.body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.routeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
		})
	}
}
