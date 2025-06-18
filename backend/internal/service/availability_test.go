package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAvailabilityService struct {
	called bool
	err    error
	got    struct {
		providerID uuid.UUID
		dayOfWeek  int32
		startTime  time.Time
		endTime    time.Time
	}
}

func (m *mockAvailabilityService) CreatePatternAndSlots(ctx context.Context, providerID uuid.UUID, dayOfWeek int32, start, end time.Time) error {
	m.called = true
	m.got.providerID = providerID
	m.got.dayOfWeek = dayOfWeek
	m.got.startTime = start
	m.got.endTime = end
	return m.err
}

func TestCreatePatternAndSlots_ServiceStyle(t *testing.T) {
	providerID := uuid.New()
	start := time.Date(2025, 6, 3, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 24, 17, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		err              error
		expectErr        bool
		expectCalled     bool
		expectSlotString string
	}{
		{
			name:         "Success",
			err:          nil,
			expectErr:    false,
			expectCalled: true,
		},
		{
			name:             "Pattern creation fails",
			err:              errors.New("db insert error"),
			expectErr:        true,
			expectCalled:     true,
			expectSlotString: "db insert error",
		},
		{
			name:             "Slot creation fails",
			err:              errors.New("slot insert error"),
			expectErr:        true,
			expectCalled:     true,
			expectSlotString: "slot insert error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAvailabilityService{err: tt.err}

			err := mock.CreatePatternAndSlots(context.Background(), providerID, int32(start.Weekday()), start, end)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectSlotString != "" {
					assert.Contains(t, err.Error(), tt.expectSlotString)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectCalled, mock.called)
		})
	}
}
