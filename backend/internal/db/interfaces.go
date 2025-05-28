package db

import (
	"context"

	"github.com/google/uuid"
)

type BookingQuerier interface {
	CreateBooking(ctx context.Context, arg CreateBookingParams) (Booking, error)
	GetOverlappingBookings(ctx context.Context, arg GetOverlappingBookingsParams) ([]Booking, error)
	DeleteBooking(ctx context.Context, arg DeleteBookingParams) error
	RescheduleBooking(ctx context.Context, arg RescheduleBookingParams) (Booking, error)
	GetBookingByID(ctx context.Context, bookingID uuid.UUID) (Booking, error)
	ListBookingsForUser(ctx context.Context, userID uuid.UUID) ([]Booking, error)
}
