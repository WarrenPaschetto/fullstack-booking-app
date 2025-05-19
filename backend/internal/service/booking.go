package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

var ErrBookingConflict = errors.New("booking time slot conflict")
var ErrBookingNotFound = errors.New("booking not found")
var ErrNotAuthorized = errors.New("not authorized")
var ErrNoBookingsFound = errors.New("no bookings found")

type BookingService struct {
	queries db.BookingQuerier
}

func NewBookingService(q db.BookingQuerier) *BookingService {
	return &BookingService{queries: q}
}

func (s *BookingService) CreateBooking(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
	start time.Time,
	durationMinutes int,
) (db.Booking, error) {

	overlaps, err := s.queries.GetOverlappingBookings(ctx, db.GetOverlappingBookingsParams{
		NewStart: start,
		NewEnd:   start.Add(time.Duration(durationMinutes) * time.Minute),
	})
	if err != nil {
		return db.Booking{}, err
	}
	if len(overlaps) > 0 {
		return db.Booking{}, ErrBookingConflict
	}

	appointment, err := s.queries.CreateBooking(ctx, db.CreateBookingParams{
		ID:               id,
		AppointmentStart: start,
		DurationMinutes:  int64(durationMinutes),
		UserID:           userID,
	})
	if err != nil {
		return db.Booking{}, err
	}

	return appointment, nil
}

func (s *BookingService) DeleteBooking(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	err := s.queries.DeleteBooking(ctx, db.DeleteBookingParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *BookingService) RescheduleBooking(
	ctx context.Context,
	bookingID uuid.UUID,
	newStart time.Time,
	durationMinutes int,
) (db.Booking, error) {

	overlaps, err := s.queries.GetOverlappingBookings(ctx, db.GetOverlappingBookingsParams{
		NewStart: newStart,
		NewEnd:   newStart.Add(time.Duration(durationMinutes) * time.Minute),
	})
	if err != nil {
		return db.Booking{}, err
	}
	if len(overlaps) > 0 {
		return db.Booking{}, ErrBookingConflict
	}

	updated, err := s.queries.RescheduleBooking(ctx, db.RescheduleBookingParams{
		ID:               bookingID,
		AppointmentStart: newStart,
	})
	if err != nil {
		return db.Booking{}, err
	}

	return updated, nil
}

func (s *BookingService) GetBookingByID(
	ctx context.Context,
	bookingID, userID uuid.UUID,
) (db.Booking, error) {
	appt, err := s.queries.GetBookingByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Booking{}, ErrBookingNotFound
		}
		return db.Booking{}, err
	}

	if appt.UserID != userID {
		return db.Booking{}, ErrNotAuthorized
	}

	return appt, nil
}

func (s *BookingService) ListUserBookings(
	ctx context.Context,
	userID uuid.UUID,
) ([]db.Booking, error) {
	bookings, err := s.queries.ListBookingsForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []db.Booking{}, ErrNoBookingsFound
		}
		return []db.Booking{}, err
	}
	if len(bookings) == 0 {
		return nil, ErrNoBookingsFound
	}
	if bookings[0].UserID != userID {
		return []db.Booking{}, ErrNotAuthorized
	}

	return bookings, nil
}

func (s *BookingService) ListAdminBookings(ctx context.Context) ([]db.Booking, error) {
	bookings, err := s.queries.ListAllBookingsForAdmin(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []db.Booking{}, ErrNoBookingsFound
		}
		return []db.Booking{}, err
	}
	if len(bookings) == 0 {
		return nil, ErrNoBookingsFound
	}
	return bookings, nil
}
