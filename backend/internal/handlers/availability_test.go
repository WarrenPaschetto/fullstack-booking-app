package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAvailabilityStore struct {
	called bool
	err    error
	got    struct {
		providerID uuid.UUID
		dayOfWeek  int32
		startTime  time.Time
		endTime    time.Time
	}
}

func (m *mockAvailabilityStore) CreatePatternAndSlots(ctx context.Context, providerID uuid.UUID, dayOfWeek int32, start, end time.Time) error {
	m.called = true
	m.got.providerID = providerID
	m.got.dayOfWeek = dayOfWeek
	m.got.startTime = start
	m.got.endTime = end
	return m.err
}

func TestCreateAvailabilityPatternHandler_Style(t *testing.T) {
	providerID := uuid.New()
	start := time.Now().Add(time.Hour)
	end := start.Add(2 * time.Hour)

	tests := []struct {
		name             string
		reqBody          any
		expectedCode     int
		expectedContains string
		isAdmin          bool
		injectUserID     bool
		mockErr          error
	}{
		{
			name:         "Success",
			reqBody:      map[string]any{"day_of_week": int32(start.Weekday()), "start_time": start, "end_time": end},
			expectedCode: http.StatusCreated,
			isAdmin:      true,
			injectUserID: true,
			mockErr:      nil,
		},
		{
			name:             "Non-admin",
			reqBody:          map[string]any{"day_of_week": 1, "start_time": start, "end_time": end},
			expectedCode:     http.StatusForbidden,
			expectedContains: "Forbidden",
			isAdmin:          false,
			injectUserID:     true,
		},
		{
			name:             "Missing user ID",
			reqBody:          map[string]any{"day_of_week": 1, "start_time": start, "end_time": end},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Could not get user ID",
			isAdmin:          true,
			injectUserID:     false,
		},
		{
			name:             "Invalid JSON",
			reqBody:          "{ invalid json",
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid request body",
			isAdmin:          true,
			injectUserID:     true,
		},
		{
			name:             "Bad day_of_week",
			reqBody:          map[string]any{"day_of_week": 8, "start_time": start, "end_time": end},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "day_of_week must be 0–6",
			isAdmin:          true,
			injectUserID:     true,
		},
		{
			name:             "End before start",
			reqBody:          map[string]any{"day_of_week": int32(start.Weekday()), "start_time": end, "end_time": start},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "end_time must be after start_time",
			isAdmin:          true,
			injectUserID:     true,
		},
		{
			name:             "Service error",
			reqBody:          map[string]any{"day_of_week": 1, "start_time": start, "end_time": end},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Failed to create pattern",
			isAdmin:          true,
			injectUserID:     true,
			mockErr:          errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Reader
			switch b := tt.reqBody.(type) {
			case string:
				body = bytes.NewReader([]byte(b))
			default:
				buf, _ := json.Marshal(b)
				body = bytes.NewReader(buf)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/avail-pattern/create", body)
			req.Header.Set("Content-Type", "application/json")

			ctx := req.Context()
			ctx = context.WithValue(ctx, middleware.IsAdminKey, tt.isAdmin)
			if tt.injectUserID {
				ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			}
			req = req.WithContext(ctx)

			mockSvc := &mockAvailabilityStore{err: tt.mockErr}
			handler := CreateAvailabilityPatternHandler(mockSvc)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedContains != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedContains)
			}
		})
	}
}
