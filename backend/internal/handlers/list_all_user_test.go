package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
)

func (m *mockUserQuerier) ListUsers(_ context.Context) ([]db.User, error) {

	return []db.User{}, nil
}

func TestListAllUsersHandler(t *testing.T) {
	tests := []struct {
		name             string
		mockQuery        *mockUserQuerier
		injectAdmin      bool
		expectedCode     int
		expectedContains string
	}{
		{
			name:             "Forbidden if not admin",
			mockQuery:        &mockUserQuerier{},
			injectAdmin:      false,
			expectedCode:     http.StatusForbidden,
			expectedContains: `"Forbidden"`,
		},
		{
			name:             "Success returns empty array",
			mockQuery:        &mockUserQuerier{},
			injectAdmin:      true,
			expectedCode:     http.StatusOK,
			expectedContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := ListAllUsersHandler(tt.mockQuery)

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			rr := httptest.NewRecorder()

			if tt.injectAdmin {
				ctx := context.WithValue(req.Context(), middleware.IsAdminKey, true)
				req = req.WithContext(ctx)
			}

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Fatalf("want status %d; got %d", tt.expectedCode, rr.Code)
			}

			if tt.expectedContains != "" {
				if !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedContains)) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedContains, rr.Body.String())
				}
			} else {
				var users []UserResponse
				if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
					t.Fatalf("invalid JSON response: %v", err)
				}
				if len(users) != 0 {
					t.Errorf("expected empty slice, got %v", users)
				}
			}
		})
	}
}
