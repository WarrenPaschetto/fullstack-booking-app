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

func (s *AvailabilityService) CreatePatternAndSlots(ctx context.Context, providerID uuid.UUID, dayOfWeek int32, start, end time.Time) error {
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

	slotParams := db.CreateAvailabilityParams{
		ProviderID: providerID,
		StartTime:  start,
		EndTime:    end,
	}

	return generateSlots(ctx, slotParams, s.store)
}

func generateSlots(ctx context.Context, arg db.CreateAvailabilityParams, q interface {
	CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error
}) error {
	weekdayToMatch := arg.StartTime.Weekday()
	startHour, startMin := arg.StartTime.Hour(), arg.StartTime.Minute()
	endHour, endMin := arg.EndTime.Hour(), arg.EndTime.Minute()

	current := time.Date(arg.StartTime.Year(), arg.StartTime.Month(), arg.StartTime.Day(), 0, 0, 0, 0, arg.StartTime.Location())

	for !current.After(arg.EndTime) {
		if current.Weekday() == weekdayToMatch {
			start := time.Date(current.Year(), current.Month(), current.Day(), startHour, startMin, 0, 0, current.Location())
			end := time.Date(current.Year(), current.Month(), current.Day(), endHour, endMin, 0, 0, current.Location())

			for slotStart := start; !slotStart.Add(60 * time.Minute).After(end); slotStart = slotStart.Add(60 * time.Minute) {
				slotEnd := slotStart.Add(60 * time.Minute)
				err := q.CreateAvailability(ctx, db.CreateAvailabilityParams{
					ID:         uuid.New(),
					ProviderID: arg.ProviderID,
					StartTime:  slotStart,
					EndTime:    slotEnd,
				})
				if err != nil {
					return fmt.Errorf("create availability on %s: %w", slotStart.Format("2006-01-02 15:04"), err)
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return nil
}
