package db

import (
	"context"

	"github.com/google/uuid"
)

type UserQuerier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UpdateUser(ctx context.Context, arg UpdateUserParams) error
	ListUsers(ctx context.Context) ([]User, error)
}

type BookingQuerier interface {
	CreateBooking(ctx context.Context, arg CreateBookingParams) (Booking, error)
	GetOverlappingBookings(ctx context.Context, arg GetOverlappingBookingsParams) ([]Booking, error)
	DeleteBooking(ctx context.Context, arg DeleteBookingParams) error
	RescheduleBooking(ctx context.Context, arg RescheduleBookingParams) (Booking, error)
	GetBookingByID(ctx context.Context, bookingID uuid.UUID) (Booking, error)
	ListBookingsForUser(ctx context.Context, userID uuid.UUID) ([]Booking, error)
}
