package handlers

import (
	"context"
	"database/sql"
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
		name           string
		routeID        string
		ctxUserID      interface{}
		mockGet        func(ctx context.Context, id uuid.UUID) (db.Booking, error)
		expectStatus   int
		expectResponse *db.Booking
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
			expectStatus: http.StatusNotFound,
		},
		{
			name:      "Forbidden",
			routeID:   bookingID.String(),
			ctxUserID: uuid.New(),
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return fakeBooking, nil
			},
			expectStatus: http.StatusForbidden,
		},
		{
			name:      "DB error",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{}, errors.New("some db failure")
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:         "Bad id param",
			routeID:      "not-a-uuid",
			ctxUserID:    userID,
			mockGet:      nil,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "No auth",
			routeID:      bookingID.String(),
			ctxUserID:    nil,
			mockGet:      nil,
			expectStatus: http.StatusUnauthorized,
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

			req := httptest.NewRequest(http.MethodGet, "/bookings/"+tt.routeID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.routeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if tt.ctxUserID != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID),
				)
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
		})
	}
}
