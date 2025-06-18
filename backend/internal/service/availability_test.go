package service

import (
	"context"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAvailabilityQueries struct {
	patternCalled bool
	slotCalls     int
	errPattern    error
	errSlot       error
}

func (m *mockAvailabilityQueries) CreateAvailabilityPattern(ctx context.Context, arg db.CreateAvailabilityPatternParams) error {
	m.patternCalled = true
	return m.errPattern
}

func (m *mockAvailabilityQueries) CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error {
	m.slotCalls++
	return m.errSlot
}

func TestCreatePatternAndSlots(t *testing.T) {
	providerID := uuid.New()
	start := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)

	t.Run("happy path", func(t *testing.T) {
		mock := &mockAvailabilityQueries{}
		svc := NewAvailabilityService(mock)

		err := svc.CreatePatternAndSlots(context.Background(), providerID, int32(start.Weekday()), start, end)

		assert.NoError(t, err)
		assert.True(t, mock.patternCalled)
		assert.Greater(t, mock.slotCalls, 0)
	})
}
