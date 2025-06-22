package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockStore struct {
	failPattern  bool
	failSlot     bool
	createdSlots int
}

func (m *mockStore) CreateAvailabilityPattern(ctx context.Context, arg db.CreateAvailabilityPatternParams) error {
	if m.failPattern {
		return errors.New("pattern insert failed")
	}
	return nil
}

func (m *mockStore) CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error {
	m.createdSlots++
	if m.failSlot {
		return errors.New("slot insert failed")
	}
	return nil
}

func TestCreatePatternAndSlots(t *testing.T) {
	providerID := uuid.New()
	// Pattern for every Tuesday 9–11 AM from June 3 to June 17, 2025
	start := time.Date(2025, 6, 3, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 17, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name                  string
		failPattern, failSlot bool
		expectErr             bool
		minSlots              int
	}{
		{"Success", false, false, false, 3}, // 3 Tuesdays * 2 hourly slots = 6
		{"Pattern fails", true, false, true, 0},
		{"Slot fails mid-way", false, true, true, 1}, // at least one created before error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStore{
				failPattern: tt.failPattern,
				failSlot:    tt.failSlot,
			}

			svc := NewAvailabilityService(mock)
			err := svc.CreatePatternAndSlots(
				context.Background(),
				providerID,
				int32(start.Weekday()),
				start,
				end,
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.failPattern {
				// No slots should be created
				assert.Zero(t, mock.createdSlots)
			} else if !tt.failSlot {
				// Expect all slots: 3 days * 2 hours = 6
				assert.Equal(t, 6, mock.createdSlots)
			} else {
				// On slot failure, at least one slot was created before error
				assert.Greater(t, mock.createdSlots, 0)
			}
		})
	}
}
