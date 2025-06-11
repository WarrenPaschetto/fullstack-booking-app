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

func (m *mockBookingQueries) ListAllBookings(ctx context.Context) ([]db.Booking, error) {
	return m.ListAllBookingsForAdminFn(ctx)
}

func TestListAllBookingsHandler(t *testing.T) {

	now := time.Now()
	b1 := db.Booking{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		AppointmentStart: now.Add(24 * time.Hour),
		DurationMinutes:  45,
		CreatedAt:        now.Add(-48 * time.Hour),
		UpdatedAt:        now.Add(-24 * time.Hour),
	}
	b2 := db.Booking{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		AppointmentStart: now.Add(48 * time.Hour),
		DurationMinutes:  60,
		CreatedAt:        now.Add(-72 * time.Hour),
		UpdatedAt:        now.Add(-36 * time.Hour),
	}
	fakeList := []db.Booking{b1, b2}

	tests := []struct {
		name           string
		mockListFn     func(ctx context.Context) ([]db.Booking, error)
		expectStatus   int
		expectResponse []db.Booking
	}{
		{
			name: "success returns 200 and all bookings",
			mockListFn: func(ctx context.Context) ([]db.Booking, error) {
				return fakeList, nil
			},
			expectStatus:   http.StatusOK,
			expectResponse: fakeList,
		},
		{
			name: "service error returns 500",
			mockListFn: func(ctx context.Context) ([]db.Booking, error) {
				return nil, errors.New("db failure")
			},
			expectStatus:   http.StatusInternalServerError,
			expectResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{
				ListAllBookingsForAdminFn: tt.mockListFn,
			}
			bookingSvc := service.NewBookingService(mockQ)
			h := &Handler{BookingService: bookingSvc}
			handler := h.ListAllBookingsHandler()

			req := httptest.NewRequest(http.MethodGet, "/api/bookings/all", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Fatalf("expected status %d, got %d; body=%q", tt.expectStatus, rr.Code, rr.Body.String())
			}

			if tt.expectStatus == http.StatusOK {
				var got []db.Booking
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				if len(got) != len(tt.expectResponse) {
					t.Fatalf("expected %d bookings, got %d", len(tt.expectResponse), len(got))
				}
				for i, want := range tt.expectResponse {
					if got[i].ID != want.ID {
						t.Errorf("booking[%d].ID = %v; want %v", i, got[i].ID, want.ID)
					}
				}
			} else {
				var errResp map[string]interface{}
				if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
					t.Fatalf("expected JSON error body, got %q", rr.Body.String())
				}
				if _, ok := errResp["error"]; !ok {
					t.Errorf("expected top-level \"error\" key in response, got %v", errResp)
				}
			}
		})
	}
}
