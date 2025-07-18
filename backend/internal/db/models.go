// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"time"

	"github.com/google/uuid"
)

type Availability struct {
	ID         uuid.UUID
	ProviderID uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AvailabilityPattern struct {
	ID         uuid.UUID
	ProviderID uuid.UUID
	DayOfWeek  int32
	StartTime  time.Time
	EndTime    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Booking struct {
	ID               uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AppointmentStart time.Time
	DurationMinutes  int32
	UserID           uuid.UUID
	SlotID           uuid.UUID
}

type User struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Email        string
	PasswordHash string
	UserRole     string
}
