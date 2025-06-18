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

type mockAvailabilityQueries struct {
	createPatternCalled     bool
	createAvailabilityCalls int
	errOnPattern            error
	errOnSlot               error
}

func (m *mockAvailabilityQueries) CreateAvailabilityPattern(ctx context.Context, arg db.CreateAvailabilityPatternParams) error {
	m.createPatternCalled = true
	return m.errOnPattern
}

func (m *mockAvailabilityQueries) CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error {
	m.createAvailabilityCalls++
	return m.errOnSlot
}

func TestCreatePatternAndSlots(t *testing.T) {
	providerID := uuid.New()
	start := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 29, 17, 0, 0, 0, time.UTC) // Includes several matching weekdays

	t.Run("successfully creates pattern and slots", func(t *testing.T) {
		mock := &mockAvailabilityQueries{}
		svc := NewAvailabilityService(mock)

		err := svc.CreatePatternAndSlots(context.Background(), providerID, int32(start.Weekday()), start, end)

		assert.NoError(t, err)
		assert.True(t, mock.createPatternCalled)
		assert.Greater(t, mock.createAvailabilityCalls, 0)
	})

	t.Run("fails to create pattern", func(t *testing.T) {
		mock := &mockAvailabilityQueries{errOnPattern: errors.New("fail pattern")}
		svc := NewAvailabilityService(mock)

		err := svc.CreatePatternAndSlots(context.Background(), providerID, 1, start, end)

		assert.Error(t, err)
		assert.True(t, mock.createPatternCalled)
		assert.Equal(t, 0, mock.createAvailabilityCalls)
	})

	t.Run("fails to create availability slot", func(t *testing.T) {
		mock := &mockAvailabilityQueries{errOnSlot: errors.New("fail slot")}
		svc := NewAvailabilityService(mock)

		err := svc.CreatePatternAndSlots(context.Background(), providerID, int32(start.Weekday()), start, end)

		assert.Error(t, err)
		assert.True(t, mock.createPatternCalled)
		assert.Greater(t, mock.createAvailabilityCalls, 0)
	})
}
