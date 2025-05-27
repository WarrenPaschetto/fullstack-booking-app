package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
)

type mockAvailabilityPatternQueries struct {
	failCreate bool
	called     bool
	gotParams  db.CreateAvailabilityPatternParams
}

func (m *mockAvailabilityPatternQueries) CreateAvailabilityPattern(ctx context.Context, arg db.CreateAvailabilityPatternParams) error {
	m.called = true
	m.gotParams = arg
	if m.failCreate {
		return errors.New("failure")
	}
	return nil
}

type AvailPatternRequest struct {
	DayOfWeek int64     `json:"day_of_week"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func TestCreateAvailabilityPatternHandler(t *testing.T) {
	providerID := uuid.New()

	tests := []struct {
		name             string
		requestBody      AvailPatternRequest
		expectedCode     int
		expectedContains string
		injectUserID     bool
		injectAdmin      bool
		invalidReqBody   bool
		failCreate       bool
	}{
		{
			name: "Successful availability pattern creation",
			requestBody: AvailPatternRequest{
				DayOfWeek: int64(1),
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedCode:   http.StatusCreated,
			injectUserID:   true,
			injectAdmin:    true,
			invalidReqBody: false,
			failCreate:     false,
		},
		{
			name: "Does not have role of admin",
			requestBody: AvailPatternRequest{
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			},
			expectedCode:     http.StatusForbidden,
			expectedContains: "Forbidden",
			injectUserID:     true,
			injectAdmin:      false,
			invalidReqBody:   false,
			failCreate:       false,
		},
		{
			name:             "Not a valid request body",
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid request body",
			injectUserID:     true,
			injectAdmin:      true,
			invalidReqBody:   true,
			failCreate:       false,
		},
		{
			name: "Invalid day of week request",
			requestBody: AvailPatternRequest{
				DayOfWeek: int64(7),
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "day_of_week must be 0 (Sunday) through 6 (Saturday)",
			injectUserID:     true,
			injectAdmin:      true,
			invalidReqBody:   false,
			failCreate:       false,
		},
		{
			name: "End time before start time",
			requestBody: AvailPatternRequest{
				DayOfWeek: int64(5),
				StartTime: time.Now().Add(2 * time.Hour),
				EndTime:   time.Now().Add(time.Hour),
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "end_time must be after start_time",
			injectUserID:     true,
			injectAdmin:      true,
			invalidReqBody:   false,
			failCreate:       false,
		},
		{
			name: "Failed availability creation",
			requestBody: AvailPatternRequest{
				DayOfWeek: int64(1),
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Unable to create availability pattern",
			injectUserID:     true,
			injectAdmin:      true,
			invalidReqBody:   false,
			failCreate:       true,
		},
		{
			name: "No user ID",
			requestBody: AvailPatternRequest{
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Could not get user ID",
			injectUserID:     false,
			injectAdmin:      true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			var body io.Reader
			if tt.invalidReqBody {
				body = strings.NewReader("{ this is an invalid request body")
			} else {
				b, _ := json.Marshal(tt.requestBody)
				body = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/availability/patterns", body)
			req.Header.Set("Content-Type", "application/json")

			ctx := req.Context()
			ctx = context.WithValue(ctx, middleware.IsAdminKey, tt.injectAdmin)
			if tt.injectUserID {
				ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			}
			req = req.WithContext(ctx)

			mock := &mockAvailabilityPatternQueries{failCreate: tt.failCreate}
			handler := CreateAvailabilityPatternHandler(mock)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Fatalf("expected %d, got %d", tt.expectedCode, rr.Code)
			}
			if tt.expectedContains != "" {
				if !strings.Contains(rr.Body.String(), tt.expectedContains) {
					t.Errorf("want error %q; got %q", tt.expectedContains, rr.Body.String())
				}
				return
			}

			if !mock.called {
				t.Error("expected CreateAvailabilityPattern to be called")
			}

			var got struct {
				ID        uuid.UUID `json:"id"`
				DayOfWeek int       `json:"day_of_week"`
				StartTime time.Time `json:"start_time"`
				EndTime   time.Time `json:"end_time"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("decode success body: %v", err)
			}
			if got.DayOfWeek != int(tt.requestBody.DayOfWeek) ||
				!got.StartTime.Equal(tt.requestBody.StartTime) ||
				!got.EndTime.Equal(tt.requestBody.EndTime) {
				t.Errorf("unexpected response %+v", got)
			}
		})
	}
}
