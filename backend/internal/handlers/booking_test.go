package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
)

type mockBookingQueries struct {
	CreateBookingFn          func(ctx context.Context, arg db.CreateBookingParams) error
	GetOverlappingBookingsFn func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error)
}

func (m *mockBookingQueries) CreateBooking(ctx context.Context, arg db.CreateBookingParams) error {
	return m.CreateBookingFn(ctx, arg)
}

func (m *mockBookingQueries) GetOverlappingBookings(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
	return m.GetOverlappingBookingsFn(ctx, arg)
}

func TestCreateBookingHandler(t *testing.T) {
	userID := uuid.New()
	validBody := BookingRequest{
		AppointmentStart: time.Now().Add(time.Hour),
		DurationMinutes:  60,
	}
	jsonBody, _ := json.Marshal(validBody)

	tests := []struct {
		name             string
		ctxUserID        any
		body             []byte
		mockOverlap      func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error)
		mockCreate       func(ctx context.Context, arg db.CreateBookingParams) error
		expectStatus     int
		expectedContains string
	}{
		{
			name:      "valid booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockOverlap: func(_ context.Context, _ db.GetOverlappingBookingsParams) ([]db.Booking, error) {
				return nil, nil
			},
			mockCreate: func(_ context.Context, _ db.CreateBookingParams) error {
				return nil
			},
			expectStatus: http.StatusCreated,
		},
		{
			name:         "missing auth context",
			ctxUserID:    nil,
			body:         jsonBody,
			mockOverlap:  nil,
			mockCreate:   nil,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:      "overlapping booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockOverlap: func(_ context.Context, _ db.GetOverlappingBookingsParams) ([]db.Booking, error) {
				return []db.Booking{{ID: uuid.New()}}, nil
			},
			mockCreate:   nil,
			expectStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBookingQueries{
				CreateBookingFn:          tt.mockCreate,
				GetOverlappingBookingsFn: tt.mockOverlap,
			}

			handler := CreateBookingHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(tt.body))
			if tt.ctxUserID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID))
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}
		})
	}
}
