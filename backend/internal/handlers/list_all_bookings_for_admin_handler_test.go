package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/google/uuid"
)

func TestListAllBookingsForAdminHandler(t *testing.T) {
	now := time.Date(2025, 5, 14, 10, 0, 0, 0, time.UTC)
	bookingID := uuid.New()
	bookingID2 := uuid.New()
	bookingID3 := uuid.New()
	userID := uuid.New()
	anotherUser := uuid.New()
	otherErr := errors.New("db fail")

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
		{
			ID:               bookingID,
			UserID:           anotherUser,
			AppointmentStart: now.Add(time.Hour * 2),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 336),
			UpdatedAt:        now.Add(-time.Minute * 336),
		},
		{
			ID:               bookingID2,
			UserID:           anotherUser,
			AppointmentStart: now.Add(time.Hour * 50),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 168),
			UpdatedAt:        now.Add(-time.Minute * 168),
		},
	}

	tests := []struct {
		name           string
		mockList       func(ctx context.Context) ([]db.Booking, error)
		expectStatus   int
		expectResponse []db.Booking
	}{
		{

			name: "Success",
			mockList: func(ctx context.Context) ([]db.Booking, error) {
				return fakeBookings, nil
			},
			expectStatus:   http.StatusOK,
			expectResponse: fakeBookings,
		},
		{

			name: "DB error",
			mockList: func(ctx context.Context) ([]db.Booking, error) {
				return []db.Booking{}, otherErr
			},
			expectStatus:   http.StatusInternalServerError,
			expectResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{
				ListAllBookingsForAdminFn: tt.mockList,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.ListAllBookingsForAdminHandler()

			req := httptest.NewRequest(http.MethodGet, "/bookings", nil)

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
