package handlers

import (
	"context"
	"database/sql"
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

func TestGetBookingByIDHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	fakeBooking := db.Booking{
		ID:               bookingID,
		UserID:           userID,
		AppointmentStart: now,
		DurationMinutes:  30,
		CreatedAt:        now.Add(-time.Hour),
		UpdatedAt:        now.Add(-time.Minute),
	}

	tests := []struct {
		name             string
		routeID          string
		ctxUserID        interface{}
		mockGet          func(ctx context.Context, id uuid.UUID) (db.Booking, error)
		expectStatus     int
		expectResponse   *db.Booking
		expectedContains string
	}{
		{

			name:      "Success",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				if id != bookingID {
					t.Errorf("expected GetBookingByID called with %s, got %s", bookingID, id)
				}
				return fakeBooking, nil
			},
			expectStatus:   http.StatusOK,
			expectResponse: &fakeBooking,
		},
		{
			name:      "Not found",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{}, sql.ErrNoRows
			},
			expectStatus:     http.StatusNotFound,
			expectedContains: "Booking not found",
		},
		{
			name:      "Forbidden",
			routeID:   bookingID.String(),
			ctxUserID: uuid.New(),
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return fakeBooking, nil
			},
			expectStatus:     http.StatusForbidden,
			expectedContains: "Not allowed",
		},
		{
			name:      "DB error",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{}, errors.New("some db failure")
			},
			expectStatus:     http.StatusInternalServerError,
			expectedContains: "Error fetching booking",
		},
		{
			name:             "Missing booking id",
			routeID:          "",
			ctxUserID:        userID,
			mockGet:          nil,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Missing booking ID",
		},
		{
			name:             "Booking id is nil",
			routeID:          uuid.Nil.String(),
			ctxUserID:        userID,
			mockGet:          nil,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Booking ID is required",
		},
		{
			name:             "Bad id param",
			routeID:          "not-a-uuid",
			ctxUserID:        userID,
			mockGet:          nil,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Invalid booking ID",
		},
		{
			name:             "No auth",
			routeID:          bookingID.String(),
			ctxUserID:        nil,
			mockGet:          nil,
			expectStatus:     http.StatusUnauthorized,
			expectedContains: "User ID missing or not a UUID in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{
				GetBookingByIDFn: tt.mockGet,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.GetBookingByIDHandler()

			urlPath := "/api/bookings"
			if tt.routeID != "" {
				urlPath += "/" + tt.routeID
			}

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)

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
				t.Fatalf("expected status %d, got %d; body=%s", tt.expectStatus, rr.Code, rr.Body.String())
			}

			if tt.expectStatus == http.StatusOK {
				var got db.Booking
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				if got.ID != fakeBooking.ID || got.UserID != fakeBooking.UserID {
					t.Errorf("unexpected booking returned: %+v", got)
				}
			}
			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected response to contain %q, got %s", tt.expectedContains, rr.Body.String())
			}
		})
	}
}
