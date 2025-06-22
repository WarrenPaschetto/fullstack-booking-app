package service

import (
	"context"
	"fmt"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

type AvailabilityStore interface {
	CreateAvailabilityPattern(ctx context.Context, arg db.CreateAvailabilityPatternParams) error
	CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error
}

type AvailabilityService struct {
	store AvailabilityStore
}

func NewAvailabilityService(store AvailabilityStore) *AvailabilityService {
	return &AvailabilityService{store: store}
}

func (s *AvailabilityService) CreatePatternAndSlots(
	ctx context.Context,
	providerID uuid.UUID,
	dayOfWeek int32,
	start, end time.Time,
) error {
	pattern := db.CreateAvailabilityPatternParams{
		ID:         uuid.New(),
		ProviderID: providerID,
		DayOfWeek:  dayOfWeek,
		StartTime:  start,
		EndTime:    end,
	}
	if err := s.store.CreateAvailabilityPattern(ctx, pattern); err != nil {
		return fmt.Errorf("create pattern: %w", err)
	}

	return generateSlots(
		ctx,
		time.Weekday(dayOfWeek),
		start,
		end,
		providerID,
		s.store,
	)
}

func generateSlots(
	ctx context.Context,
	dayToMatch time.Weekday,
	startRange, endRange time.Time,
	providerID uuid.UUID,
	store AvailabilityStore,
) error {
	startHour, startMin := startRange.Hour(), startRange.Minute()
	endHour, endMin := endRange.Hour(), endRange.Minute()

	current := time.Date(startRange.Year(), startRange.Month(), startRange.Day(), 0, 0, 0, 0, startRange.Location())

	for !current.After(endRange) {
		if current.Weekday() == dayToMatch {
			dayStart := time.Date(current.Year(), current.Month(), current.Day(), startHour, startMin, 0, 0, current.Location())
			dayEnd := time.Date(current.Year(), current.Month(), current.Day(), endHour, endMin, 0, 0, current.Location())

			for slotStart := dayStart; !slotStart.Add(time.Hour).After(dayEnd); slotStart = slotStart.Add(time.Hour) {
				slotEnd := slotStart.Add(time.Hour)
				if err := store.CreateAvailability(ctx, db.CreateAvailabilityParams{
					ID:         uuid.New(),
					ProviderID: providerID,
					StartTime:  slotStart,
					EndTime:    slotEnd,
				}); err != nil {
					return fmt.Errorf("create availability on %s: %w", slotStart.Format("2006‑01‑02 15:04"), err)
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	return nil
}
