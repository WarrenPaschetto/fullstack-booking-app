package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

type fakeBookingRepo struct {
	overlaps   []db.Booking
	overlapErr error
	created    db.Booking
	createErr  error
}

func (f *fakeBookingRepo) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {
	return f.created, f.createErr
}

func (f *fakeBookingRepo) GetOverlappingBookings(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
	return f.overlaps, f.overlapErr
}

func (f *fakeBookingRepo) DeleteBooking(ctx context.Context, arg db.DeleteBookingParams) error {
	return nil
}

func (f *fakeBookingRepo) RescheduleBooking(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
	return db.Booking{}, nil
}

func (f *fakeBookingRepo) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (db.Booking, error) {
	return db.Booking{}, nil
}

var errSimulatedOverlap = errors.New("simulated error")
var errSimulatedreate = errors.New("could not create booking")

func TestBookingService_CreateBooking(t *testing.T) {
	now := time.Date(2025, 5, 14, 10, 0, 0, 0, time.UTC)
	userID := uuid.New()
	id := uuid.New()

	tests := []struct {
		name       string
		overlaps   []db.Booking
		overlapErr error
		created    db.Booking
		createErr  error
		wantErr    error
	}{
		{
			name:     "Valid booking",
			overlaps: nil,
			created: db.Booking{
				ID:               id,
				UserID:           userID,
				AppointmentStart: now,
				DurationMinutes:  30,
			},
			wantErr: nil,
		},
		{
			name:       "DB error fetching overlaps",
			overlapErr: errSimulatedOverlap,
			wantErr:    errSimulatedOverlap,
		},
		{
			name:     "Overlap booking",
			overlaps: []db.Booking{{ID: uuid.New()}},
			wantErr:  ErrBookingConflict,
		},
		{
			name:      "Create booking error",
			overlaps:  nil,
			createErr: errSimulatedreate,
			wantErr:   errSimulatedreate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookingRepo{
				overlaps:   tt.overlaps,
				overlapErr: tt.overlapErr,
				created:    tt.created,
				createErr:  tt.createErr,
			}

			svc := NewBookingService(repo)
			got, err := svc.CreateBooking(context.Background(), id, userID, now, 30)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.created.ID ||
				got.UserID != tt.created.UserID ||
				!got.AppointmentStart.Equal(tt.created.AppointmentStart) {
				t.Errorf("got %+v, want %+v", got, tt.created)
			}

		})
	}
}
