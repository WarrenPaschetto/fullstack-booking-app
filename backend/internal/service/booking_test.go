package service

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

type fakeBookingRepo struct {
	overlaps                  []db.Booking
	overlapErr                error
	created                   db.Booking
	createErr                 error
	DeleteBookingFn           func(ctx context.Context, arg db.DeleteBookingParams) error
	RescheduleBookingFn       func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
	GetBookingByIDFn          func(ctx context.Context, bookingID uuid.UUID) (db.Booking, error)
	ListBookingsForUserFn     func(ctx context.Context, id uuid.UUID) ([]db.Booking, error)
	CreateAvailabilityFn      func(ctx context.Context, arg db.CreateAvailabilityParams) error
	ListAllBookingsForAdminFn func(ctx context.Context) ([]db.Booking, error)
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
func (f *fakeBookingRepo) CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error {
	return nil
}
func (f *fakeBookingRepo) ListAllBookingsForAdmin(ctx context.Context) ([]db.Booking, error) {
	return f.ListAllBookingsForAdminFn(ctx)
}

var errSimulatedOverlap = errors.New("simulated error")
var errSimulatedCreate = errors.New("could not create booking")
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
			createErr: errSimulatedCreate,
			wantErr:   errSimulatedCreate,
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
		ctxUserID  uuid.UUID
		isAdmin    bool
		wantErr    error
	}{
		{
			name: "Successful user deletion",
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				if arg.ID != bookingID || arg.UserID != userID {
					t.Errorf("expected ID=%v and UserID=%v, got ID=%v and UserID=%v", bookingID, userID, arg.ID, arg.UserID)
				}
				return nil
			},
			ctxUserID: userID,
			isAdmin:   false,
			wantErr:   nil,
		},
		{
			name: "Successful admin deletion",
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				if arg.ID != bookingID {
					t.Errorf("expected ID=%v and UserID=%v, got ID=%v and UserID=%v", bookingID, userID, arg.ID, arg.UserID)
				}
				return nil
			},
			ctxUserID: uuid.Nil,
			isAdmin:   true,
			wantErr:   nil,
		},
		{
			name: "Booking not found",
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				return sql.ErrNoRows
			},
			ctxUserID: userID,
			isAdmin:   false,
			wantErr:   ErrBookingNotFound,
		},
		{
			name: "Unsuccessful deletion",
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				return errDeleting
			},
			ctxUserID: userID,
			isAdmin:   false,
			wantErr:   errDeleting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookingRepo{
				DeleteBookingFn: tt.mockDelete,
			}

			svc := NewBookingService(repo)
			err := svc.DeleteBooking(context.Background(), bookingID, tt.ctxUserID, tt.isAdmin)

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
	userID := uuid.New()
	newStart := now.Add(2 * time.Hour)

	tests := []struct {
		name           string
		ctxUser        uuid.UUID
		ctxAdmin       bool
		mockReschedule func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
		overlaps       []db.Booking
		overlapErr     error
		wantBooking    db.Booking
		wantErr        error
	}{
		{
			name:     "Successful reschedule",
			ctxUser:  userID,
			ctxAdmin: false,
			mockReschedule: func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				if arg.ID != bookingID {
					t.Errorf("expected bookingID=%v, got bookingID=%v", bookingID, arg.ID)
				}
				if !arg.AppointmentStart.Equal(newStart) {
					t.Errorf("expected new start=%v, got %v", newStart, arg.AppointmentStart)
				}

				return db.Booking{
					ID:               arg.ID,
					AppointmentStart: arg.AppointmentStart,
					DurationMinutes:  int32(30),
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
			name:     "Successful admin reschedule of another user",
			ctxUser:  uuid.Nil,
			ctxAdmin: true,
			mockReschedule: func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				if arg.ID != bookingID {
					t.Errorf("expected bookingID=%v, got bookingID=%v", bookingID, arg.ID)
				}
				if !arg.AppointmentStart.Equal(newStart) {
					t.Errorf("expected new start=%v, got %v", newStart, arg.AppointmentStart)
				}

				return db.Booking{
					ID:               arg.ID,
					AppointmentStart: arg.AppointmentStart,
					DurationMinutes:  int32(30),
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
			name:     "Reschedule returns DB error",
			ctxUser:  userID,
			ctxAdmin: false,
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, errReschedule
			},
			wantErr: errReschedule,
		},
		{
			name:     "DB error fetching overlaps",
			ctxUser:  userID,
			ctxAdmin: false,
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				t.Fatalf("reschedule could not be called on overlap error")
				return db.Booking{}, nil
			},
			overlapErr: errSimulatedOverlap,
			wantErr:    errSimulatedOverlap,
		},
		{
			name:     "Overlap booking",
			ctxUser:  userID,
			ctxAdmin: false,
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
			got, err := svc.RescheduleBooking(context.Background(), bookingID, tt.ctxUser, newStart, 30, tt.ctxAdmin)

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
	otherErr := errors.New("db fail")

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
					DurationMinutes:  int32(45),
				}, nil
			},
			wantBooking: db.Booking{ID: bookingID, UserID: userID, AppointmentStart: now, DurationMinutes: 45},
			wantErr:     nil,
		},
		{
			name: "Booking not found",
			mockGetBook: func(_ context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{}, sql.ErrNoRows
			},
			wantErr: ErrBookingNotFound,
		},
		{
			name: "DB error",
			mockGetBook: func(_ context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{}, otherErr
			},
			wantErr: otherErr,
		},
		{
			name: "Not an authorized user",
			mockGetBook: func(_ context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{
					ID:               id,
					UserID:           wrongUser,
					AppointmentStart: now,
					DurationMinutes:  int32(45),
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

func TestListUserBookings(t *testing.T) {
	now := time.Date(2025, 5, 14, 10, 0, 0, 0, time.UTC)
	bookingID := uuid.New()
	bookingID2 := uuid.New()
	bookingID3 := uuid.New()
	userID := uuid.New()
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
	}

	tests := []struct {
		name         string
		mockList     func(ctx context.Context, id uuid.UUID) ([]db.Booking, error)
		wantBookings []db.Booking
		wantErr      error
	}{
		{
			name: "Success",
			mockList: func(_ context.Context, id uuid.UUID) ([]db.Booking, error) {
				return fakeBookings, nil
			},
			wantBookings: fakeBookings,
			wantErr:      nil,
		},
		{
			name: "DB error",
			mockList: func(_ context.Context, id uuid.UUID) ([]db.Booking, error) {
				return []db.Booking{}, otherErr
			},
			wantBookings: nil,
			wantErr:      otherErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookingRepo{
				ListBookingsForUserFn: tt.mockList,
			}
			svc := NewBookingService(repo)

			got, err := svc.ListUserBookings(context.Background(), userID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !reflect.DeepEqual(got, tt.wantBookings) {
				t.Errorf("got = %#v, want %#v", got, tt.wantBookings)
			}

		})
	}
}

func TestListAllBookings(t *testing.T) {
	now := time.Now().UTC()

	b1 := db.Booking{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		AppointmentStart: now.Add(24 * time.Hour),
		DurationMinutes:  30,
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

	tests := []struct {
		name    string
		mockFn  func(ctx context.Context) ([]db.Booking, error)
		want    []db.Booking
		wantErr bool
	}{
		{
			name: "success returns bookings",
			mockFn: func(ctx context.Context) ([]db.Booking, error) {
				return []db.Booking{b1, b2}, nil
			},
			want:    []db.Booking{b1, b2},
			wantErr: false,
		},
		{
			name: "error from queries bubbles up",
			mockFn: func(ctx context.Context) ([]db.Booking, error) {
				return nil, errors.New("database failure")
			},
			want:    []db.Booking{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &fakeBookingRepo{
				ListAllBookingsForAdminFn: tt.mockFn,
			}
			svc := NewBookingService(mockQ)

			got, err := svc.ListAllBookings(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListAllBookings() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("ListAllBookings() returned %d bookings, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("ListAllBookings()[%d].ID = %v; want %v", i, got[i].ID, tt.want[i].ID)
				}
			}
		})
	}
}
