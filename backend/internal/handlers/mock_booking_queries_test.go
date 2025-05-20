package handlers

import (
	"context"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

type mockBookingQueries struct {
	CreateBookingFn           func(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error)
	GetOverlappingBookingsFn  func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error)
	DeleteBookingFn           func(ctx context.Context, arg db.DeleteBookingParams) error
	RescheduleBookingFn       func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
	GetBookingByIDFn          func(ctx context.Context, bookingID uuid.UUID) (db.Booking, error)
	ListBookingsForUserFn     func(ctx context.Context, id uuid.UUID) ([]db.Booking, error)
	ListAllBookingsForAdminFn func(ctx context.Context) ([]db.Booking, error)
}

func (m *mockBookingQueries) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {
	return m.CreateBookingFn(ctx, arg)
}
func (m *mockBookingQueries) GetOverlappingBookings(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
	return m.GetOverlappingBookingsFn(ctx, arg)
}

func (m *mockBookingQueries) DeleteBooking(ctx context.Context, arg db.DeleteBookingParams) error {
	return m.DeleteBookingFn(ctx, arg)
}

func (m *mockBookingQueries) RescheduleBooking(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
	return m.RescheduleBookingFn(ctx, arg)
}

func (m *mockBookingQueries) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (db.Booking, error) {
	return m.GetBookingByIDFn(ctx, bookingID)
}
func (m *mockBookingQueries) ListBookingsForUser(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
	return m.ListBookingsForUserFn(ctx, id)
}
func (m *mockBookingQueries) ListAllBookingsForAdmin(ctx context.Context) ([]db.Booking, error) {
	return m.ListAllBookingsForAdminFn(ctx)
}
