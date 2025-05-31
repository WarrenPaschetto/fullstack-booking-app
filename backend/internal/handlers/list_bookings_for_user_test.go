package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/google/uuid"
)

func TestListBookingsForUserHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	bookingID2 := uuid.New()
	bookingID3 := uuid.New()
	now := time.Now()

	fakeBookings := []db.Booking{
		{
			ID:               bookingID,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 24),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 336),
			UpdatedAt:        now.Add(-time.Minute * 336),
		},
		{
			ID:               bookingID2,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 84),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 168),
			UpdatedAt:        now.Add(-time.Minute * 168),
		},
		{
			ID:               bookingID3,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 168),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 48),
			UpdatedAt:        now.Add(-time.Minute * 48),
		},
	}

	tests := []struct {
		name           string
		ctxUserID      interface{}
		mockList       func(ctx context.Context, id uuid.UUID) ([]db.Booking, error)
		expectStatus   int
		expectResponse []db.Booking
	}{
		{

			name:      "Success",
			ctxUserID: userID,
			mockList: func(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
				if id != userID {
					t.Errorf("ListBookingsForUser called with wrong userID: got %v, want %v", id, userID)
				}
				return fakeBookings, nil
			},
			expectStatus:   http.StatusOK,
			expectResponse: fakeBookings,
		},
		{

			name:      "User ID missing",
			ctxUserID: nil,
			mockList: func(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
				return fakeBookings, nil
			},
			expectStatus:   http.StatusUnauthorized,
			expectResponse: nil,
		},
		{

			name:      "User not found",
			ctxUserID: uuid.New(),
			mockList: func(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
				return fakeBookings, nil
			},
			expectStatus:   http.StatusInternalServerError,
			expectResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{
				ListBookingsForUserFn: tt.mockList,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.ListBookingsForUserHandler()

			req := httptest.NewRequest(http.MethodGet, "/api/bookings", nil)
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
				var got []db.Booking
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				for index, booking := range got {
					if got[index].UserID != booking.UserID {
						t.Errorf("unexpected user booking returned: %+v", got)
					}
				}
			}
		})
	}
}
