package db

import (
	"context"
)

type UserQuerier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (User, error)
}

type BookingQuerier interface {
	CreateBooking(ctx context.Context, arg CreateBookingParams) error
	GetOverlappingBookings(ctx context.Context, arg GetOverlappingBookingsParams) ([]Booking, error)
}
