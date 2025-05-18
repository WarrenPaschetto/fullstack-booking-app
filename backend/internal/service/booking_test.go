package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

type fakeBookingRepo struct {
	overlaps              []db.Booking
	overlapErr            error
	created               db.Booking
	createErr             error
	DeleteBookingFn       func(ctx context.Context, arg db.DeleteBookingParams) error
	RescheduleBookingFn   func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
	GetBookingByIDFn      func(ctx context.Context, bookingID uuid.UUID) (db.Booking, error)
	ListBookingsForUserFn func(ctx context.Context, id uuid.UUID) ([]db.Booking, error)
}

func (f *fakeBookingRepo) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {
	return f.created, f.createErr
}

func (f *fakeBookingRepo) GetOverlappingBookings(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
	return f.overlaps, f.overlapErr
}

func (f *fakeBookingRepo) DeleteBooking(ctx context.Context, arg db.DeleteBookingParams) error {
	return f.DeleteBookingFn(ctx, arg)
}

func (f *fakeBookingRepo) RescheduleBooking(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
	return f.RescheduleBookingFn(ctx, arg)
}

func (f *fakeBookingRepo) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (db.Booking, error) {
	return f.GetBookingByIDFn(ctx, bookingID)
}
func (f *fakeBookingRepo) ListBookingsForUser(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
	return f.ListBookingsForUserFn(ctx, id)
}

var errSimulatedOverlap = errors.New("simulated error")
var errSimulatedreate = errors.New("could not create booking")
var errDeleting = errors.New("could not delete booking")
var errReschedule = errors.New("could not reschedule booking")

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

func TestBookingService_DeleteBooking(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	tests := []struct {
		name       string
		mockDelete func(ctx context.Context, arg db.DeleteBookingParams) error
		wantErr    error
	}{
		{
			name: "Successful deletion",
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				if arg.ID != bookingID || arg.UserID != userID {
					t.Errorf("expected ID=%v and UserID=%v, got ID=%v and UserID=%v", bookingID, userID, arg.ID, arg.UserID)
				}
				return nil
			},
			wantErr: nil,
		},
		{
			name: "Unsuccessful deletion",
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				return errDeleting
			},
			wantErr: errDeleting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookingRepo{
				DeleteBookingFn: tt.mockDelete,
			}

			svc := NewBookingService(repo)
			err := svc.DeleteBooking(context.Background(), bookingID, userID)

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

		})
	}
}

func TestBookingService_RescheduleBooking(t *testing.T) {
	now := time.Date(2025, 5, 14, 10, 0, 0, 0, time.UTC)
	bookingID := uuid.New()
	newStart := now.Add(2 * time.Hour)

	tests := []struct {
		name           string
		mockReschedule func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
		overlaps       []db.Booking
		overlapErr     error
		wantBooking    db.Booking
		wantErr        error
	}{
		{
			name: "Successful reschedule",
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				if arg.ID != bookingID {
					t.Errorf("expected bookingID=%v, got bookingID=%v", bookingID, arg.ID)
				}
				if !arg.AppointmentStart.Equal(newStart) {
					t.Errorf("expected new start=%v, got %v", newStart, arg.AppointmentStart)
				}

				return db.Booking{
					ID:               arg.ID,
					AppointmentStart: arg.AppointmentStart,
					DurationMinutes:  int64(30),
				}, nil
			},
			wantBooking: db.Booking{
				ID:               bookingID,
				AppointmentStart: newStart,
				DurationMinutes:  30,
			},
			wantErr: nil,
		},
		{
			name: "Unsuccessful deletion",
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, errReschedule
			},
			wantErr: errReschedule,
		},
		{
			name: "DB error fetching overlaps",
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				t.Fatalf("reschedule could not be called on overlap error")
				return db.Booking{}, nil
			},
			overlapErr: errSimulatedOverlap,
			wantErr:    errSimulatedOverlap,
		},
		{
			name: "Overlap booking",
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				t.Fatalf("reschedule could not be called due to conflict of scheduling")
				return db.Booking{}, nil
			},
			overlaps: []db.Booking{{ID: uuid.New()}},
			wantErr:  ErrBookingConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookingRepo{
				RescheduleBookingFn: tt.mockReschedule,
				overlaps:            tt.overlaps,
				overlapErr:          tt.overlapErr,
			}

			svc := NewBookingService(repo)
			got, err := svc.RescheduleBooking(context.Background(), bookingID, newStart, 30)

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

			if got.ID != tt.wantBooking.ID ||
				!got.AppointmentStart.Equal(tt.wantBooking.AppointmentStart) ||
				got.DurationMinutes != tt.wantBooking.DurationMinutes {
				t.Errorf("got %+v, want %+v", got, tt.wantBooking)
			}

		})
	}
}

func TestBookingService_GetBookingByID(t *testing.T) {
	now := time.Date(2025, 5, 14, 10, 0, 0, 0, time.UTC)
	bookingID := uuid.New()
	userID := uuid.New()
	wrongUser := uuid.New()

	tests := []struct {
		name        string
		mockGetBook func(ctx context.Context, id uuid.UUID) (db.Booking, error)
		wantBooking db.Booking
		wantErr     error
	}{
		{
			name: "Success",
			mockGetBook: func(_ context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{
					ID:               id,
					UserID:           userID,
					AppointmentStart: now,
					DurationMinutes:  int64(45),
				}, nil
			},
			wantBooking: db.Booking{ID: bookingID, UserID: userID, AppointmentStart: now, DurationMinutes: 45},
			wantErr:     nil,
		},
		{
			name: "Booking not found",
			mockGetBook: func(_ context.Context, _ uuid.UUID) (db.Booking, error) {
				return db.Booking{}, sql.ErrNoRows
			},
			wantErr: ErrBookingNotFound,
		},
		{
			name: "Not an authorized user",
			mockGetBook: func(_ context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{
					ID:               id,
					UserID:           wrongUser,
					AppointmentStart: now,
					DurationMinutes:  int64(45),
				}, nil
			},
			wantErr: ErrNotAuthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookingRepo{
				GetBookingByIDFn: tt.mockGetBook,
			}

			svc := NewBookingService(repo)
			got, err := svc.GetBookingByID(context.Background(), bookingID, userID)

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
			if got.ID != tt.wantBooking.ID ||
				got.UserID != tt.wantBooking.UserID ||
				!got.AppointmentStart.Equal(tt.wantBooking.AppointmentStart) ||
				got.DurationMinutes != tt.wantBooking.DurationMinutes {
				t.Errorf("got %+v, want %+v", got, tt.wantBooking)
			}

		})
	}
}
