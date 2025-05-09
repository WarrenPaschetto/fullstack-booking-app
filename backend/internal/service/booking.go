package service

import (
	"context"
	"errors"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

var ErrBookingConflict = errors.New("booking time slot conflict")

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
