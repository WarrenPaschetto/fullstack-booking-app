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

type mockAvailabilityQueries struct {
	db.Queries
	failCreate bool
}

func (m *mockAvailabilityQueries) CreateAvailability(ctx context.Context, arg db.CreateAvailabilityParams) error {
	if m.failCreate {
		return errors.New("failure")
	}
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
		invalidReqBody   bool
		injectUserID     bool
		failCreate       bool
	}{
		{
			name: "Successful availability creation",
			requestBody: AvailRequest{
				StartTime: time.Now().Add(time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedCode:   http.StatusCreated,
			injectAdmin:    true,
			injectUserID:   true,
			invalidReqBody: false,
			failCreate:     false,
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
			injectUserID:     true,
			invalidReqBody:   false,
			failCreate:       false,
		},
		{
			name:             "Not a valid request body",
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid request body",
			injectAdmin:      true,
			injectUserID:     true,
			invalidReqBody:   true,
			failCreate:       false,
		},
		{
			name:             "Can't find userID",
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Could not get user ID",
			injectAdmin:      true,
			injectUserID:     false,
			invalidReqBody:   false,
			failCreate:       false,
		},
		{
			name:             "Failed availability creation",
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Unable to create availability",
			injectAdmin:      true,
			injectUserID:     true,
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
			if tt.injectUserID {
				ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			}
			req = req.WithContext(ctx)

			mock := &mockAvailabilityQueries{failCreate: tt.failCreate}
			handler := CreateAvailabilityHandler(mock)

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
