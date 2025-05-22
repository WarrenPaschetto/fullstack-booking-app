package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
)

type mockAvailabilityQueries struct {
	db.Queries
}

func (m *mockAvailabilityQueries) CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error {
	return nil
}

type AvailRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func TestCreateAvailabilityHandler(t *testing.T) {
	providerID := uuid.New()

	tests := []struct {
		name             string
		requestBody      AvailRequest
		expectedCode     int
		expectedContains string
		injectAdmin      bool
	}{
		{
			name: "Successful availability creation",
			requestBody: AvailRequest{
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedCode: http.StatusCreated,
			injectAdmin:  true,
		},
		{
			name: "Does not have role of admin",
			requestBody: AvailRequest{
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			},
			expectedCode:     http.StatusForbidden,
			expectedContains: "Forbidden",
			injectAdmin:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			b, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/availability", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			// 2) inject both admin‐flag and providerID into context
			ctx := req.Context()
			ctx = context.WithValue(ctx, middleware.IsAdminKey, tt.injectAdmin)
			ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := CreateAvailabilityHandler(&mockAvailabilityQueries{})
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
