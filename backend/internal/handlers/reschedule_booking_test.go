package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestRescheduleBookingHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	reqBody := RescheduleBookingRequest{
		AppointmentStart: now.Add(time.Hour),
		DurationMinutes:  30,
	}
	jsonBody, _ := json.Marshal(reqBody)

	invalidReq := RescheduleBookingRequest{
		AppointmentStart: now.Add(time.Hour),
		DurationMinutes:  -10,
	}
	invalidBody, _ := json.Marshal(invalidReq)

	fakeBooking := db.Booking{
		ID:               bookingID,
		UserID:           userID,
		AppointmentStart: now.Add(time.Hour),
		DurationMinutes:  30,
		CreatedAt:        now.Add(-time.Hour),
		UpdatedAt:        now.Add(-time.Minute),
	}

	tests := []struct {
		name             string
		routeID          string
		ctxUserID        any
		body             []byte
		mockReschedule   func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
		expectStatus     int
		expectedContains string
	}{
		{
			name:      "Rescheduled booking success",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			body:      jsonBody,
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
			name:             "Missing user ID",
			routeID:          bookingID.String(),
			ctxUserID:        nil,
			body:             jsonBody,
			expectStatus:     http.StatusUnauthorized,
			expectedContains: "User ID missing or not a UUID in context",
		},
		{
			name:             "Missing booking ID in URL",
			routeID:          "",
			ctxUserID:        userID,
			body:             jsonBody,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Missing booking ID",
		},
		{
			name:             "Bad UUID",
			routeID:          "not-a-uuid",
			ctxUserID:        userID,
			body:             jsonBody,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Invalid booking ID",
		},
		{
			name:             "No booking ID",
			routeID:          uuid.Nil.String(),
			ctxUserID:        userID,
			body:             jsonBody,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Booking ID is required",
		},
		{
			name:             "Malformed JSON",
			routeID:          bookingID.String(),
			ctxUserID:        userID,
			body:             []byte(`{invalid json`),
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Invalid request body",
		},

		{
			name:             "Duration minutes invalid",
			routeID:          bookingID.String(),
			ctxUserID:        userID,
			body:             invalidBody,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Duration minutes must be greater than zero",
		},
		{
			name:      "Booking conflict",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			body:      jsonBody,
			mockReschedule: func(_ context.Context, _ db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, service.ErrBookingConflict
			},
			expectStatus:     http.StatusConflict,
			expectedContains: "Time slot already booked",
		},
		{
			name:      "Not authorized",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			body:      jsonBody,
			mockReschedule: func(_ context.Context, _ db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, service.ErrNotAuthorized
			},
			expectStatus:     http.StatusForbidden,
			expectedContains: "You are not authorized to reschedule this booking",
		},
		{
			name:      "Internal error",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			body:      jsonBody,
			mockReschedule: func(_ context.Context, _ db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, errors.New("boom")
			},
			expectStatus:     http.StatusInternalServerError,
			expectedContains: "Couldn't reschedule",
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

			req := httptest.NewRequest(http.MethodPut, "/api/bookings/"+tt.routeID, bytes.NewReader(tt.body))

			if tt.ctxUserID != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID),
				)
			}

			if tt.routeID != "" {
				req = mux.SetURLVars(req, map[string]string{"id": tt.routeID})
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected response to contain %q, got %s", tt.expectedContains, rr.Body.String())
			}

			if tt.expectStatus == http.StatusOK {
				var returned db.Booking
				if err := json.NewDecoder(rr.Body).Decode(&returned); err != nil {
					t.Fatalf("couldn't decode success response: %v", err)
				}
				if returned.ID != fakeBooking.ID {
					t.Errorf("got ID %v, want %v", returned.ID, fakeBooking.ID)
				}
				if !returned.AppointmentStart.Equal(fakeBooking.AppointmentStart) {
					t.Errorf("got AppointmentStart %v, want %v", returned.AppointmentStart, fakeBooking.AppointmentStart)
				}
				if returned.DurationMinutes != fakeBooking.DurationMinutes {
					t.Errorf("got DurationMinutes %v, want %v", returned.DurationMinutes, fakeBooking.DurationMinutes)
				}
				if returned.UserID != fakeBooking.UserID {
					t.Errorf("got user ID %v, want %v", returned.UserID, fakeBooking.UserID)
				}
			}
		})
	}
}
