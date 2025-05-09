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
	"github.com/google/uuid"
)

type mockBookingQueries struct {
	CreateBookingFn          func(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error)
	GetOverlappingBookingsFn func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error)
	DeleteBookingFn          func(ctx context.Context, arg db.DeleteBookingParams) error
	RescheduleBookingFn      func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
}

func (m *mockBookingQueries) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {
	return m.CreateBookingFn(ctx, arg)
}

func (m *mockBookingQueries) GetOverlappingBookings(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
	return m.GetOverlappingBookingsFn(ctx, arg)
}

func (m *mockBookingQueries) DeleteBooking(ctx context.Context, arg db.DeleteBookingParams) error {
	return m.DeleteBookingFn(ctx, arg)
}

func (m *mockBookingQueries) RescheduleBooking(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
	return m.RescheduleBookingFn(ctx, arg)
}

func TestCreateBookingHandler(t *testing.T) {
	userID := uuid.New()
	validBody := BookingRequest{
		AppointmentStart: time.Now().Add(time.Hour),
		DurationMinutes:  60,
	}
	jsonBody, _ := json.Marshal(validBody)

	tests := []struct {
		name         string
		ctxUserID    any
		body         []byte
		mockOverlap  func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error)
		mockCreate   func(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error)
		expectStatus int
	}{
		{
			name:      "valid booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockOverlap: func(_ context.Context, _ db.GetOverlappingBookingsParams) ([]db.Booking, error) {
				return nil, nil
			},
			mockCreate: func(_ context.Context, _ db.CreateBookingParams) (db.Booking, error) {
				return db.Booking{ID: uuid.New()}, nil
			},
			expectStatus: http.StatusCreated,
		},
		{
			name:         "missing auth context",
			ctxUserID:    nil,
			body:         jsonBody,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "malformed request body",
			ctxUserID:    userID,
			body:         []byte(`{invalid json`),
			expectStatus: http.StatusBadRequest,
		},
		{
			name:      "overlapping booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockOverlap: func(_ context.Context, _ db.GetOverlappingBookingsParams) ([]db.Booking, error) {
				return []db.Booking{{ID: uuid.New()}}, nil
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockBookingQueries{
				CreateBookingFn:          tt.mockCreate,
				GetOverlappingBookingsFn: tt.mockOverlap,
			}

			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.CreateBookingHandler()

			req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(tt.body))
			if tt.ctxUserID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID))
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
		})
	}
}

func TestDeleteBookingHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	validBody := DeleteBookingRequest{
		ID: bookingID,
	}
	jsonBody, _ := json.Marshal(validBody)

	tests := []struct {
		name         string
		ctxUserID    any
		body         []byte
		mockDelete   func(ctx context.Context, arg db.DeleteBookingParams) error
		expectStatus int
	}{
		{
			name:      "delete booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				if arg.ID != bookingID || arg.UserID != userID {
					t.Errorf("expected ID=%v and UserID=%v, got ID=%v and UserID=%v", bookingID, userID, arg.ID, arg.UserID)
				}
				return nil
			},
			expectStatus: http.StatusNoContent,
		},
		{
			name:      "DB error",
			ctxUserID: userID,
			body:      jsonBody,
			mockDelete: func(_ context.Context, _ db.DeleteBookingParams) error {
				return errors.New("simulated DB error")
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:         "No user ID",
			ctxUserID:    nil,
			body:         jsonBody,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "malformed request body",
			ctxUserID:    userID,
			body:         []byte(`{invalid json`),
			mockDelete:   nil,
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockBookingQueries{
				DeleteBookingFn: tt.mockDelete,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.DeleteBookingHandler()

			req := httptest.NewRequest(http.MethodDelete, "/bookings", bytes.NewReader(tt.body))
			if tt.ctxUserID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID))
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
		})
	}
}
