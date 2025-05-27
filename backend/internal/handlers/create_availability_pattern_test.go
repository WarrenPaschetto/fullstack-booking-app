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
	db.Queries
	failCreate bool
}

func (m *mockAvailabilityPatternQueries) CreateAvailabilityPattern(ctx context.Context, arg db.CreateAvailabilityPatternParams) error {
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
			injectAdmin:      false,
			invalidReqBody:   false,
			failCreate:       false,
		},
		{
			name:             "Not a valid request body",
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid request body",
			injectAdmin:      true,
			invalidReqBody:   true,
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
			injectAdmin:      true,
			invalidReqBody:   false,
			failCreate:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var body io.Reader
			if tt.invalidReqBody {
				body = strings.NewReader("{ this is an invalid request body")
			} else {
				b, _ := json.Marshal(tt.requestBody)
				body = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/availability", body)
			req.Header.Set("Content-Type", "application/json")

			ctx := req.Context()
			ctx = context.WithValue(ctx, middleware.IsAdminKey, tt.injectAdmin)
			ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			req = req.WithContext(ctx)

			mock := &mockAvailabilityPatternQueries{failCreate: tt.failCreate}
			handler := CreateAvailabilityPatternHandler(mock)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Fatalf("expected %d, got %d", tt.expectedCode, rr.Code)
			}
			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("body %q does not contain %q", rr.Body.String(), tt.expectedContains)
			}
		})
	}
}
