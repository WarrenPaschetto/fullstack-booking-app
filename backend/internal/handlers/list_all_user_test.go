package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
)

func (m *mockUserQuerier) ListUsers(ctx context.Context) ([]db.User, error) {
	return m.ListUsersFunc(ctx)
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
			name: "Unable to list users",
			mockQuery: &mockUserQuerier{ListUsersFunc: func(ctx context.Context) ([]db.User, error) {
				return []db.User{}, errors.New("simulated error")
			}},
			injectAdmin:      true,
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Unable to list users",
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

func TestListAllUsersHandler_Mapping(t *testing.T) {
	t0 := time.Date(2025, 5, 21, 12, 0, 0, 0, time.UTC)
	users := []db.User{
		{
			ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			FirstName: "Alice",
			LastName:  "Anderson",
			Email:     "alice@example.com",
			CreatedAt: t0,
			UpdatedAt: t0,
			Role:      "user",
		},
		{
			ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			FirstName: "Bob",
			LastName:  "Brown",
			Email:     "bob@example.com",
			CreatedAt: t0,
			UpdatedAt: t0,
			Role:      "admin",
		},
	}

	mock := &mockUserQuerier{
		ListUsersFunc: func(ctx context.Context) ([]db.User, error) {
			return users, nil
		},
	}

	handler := ListAllUsersHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	ctx := context.WithValue(req.Context(), middleware.IsAdminKey, true)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	var got []UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	want := []UserResponse{
		{ID: users[0].ID, FirstName: "Alice", LastName: "Anderson", Email: "alice@example.com", CreatedAt: t0, UpdatedAt: t0, Role: "user"},
		{ID: users[1].ID, FirstName: "Bob", LastName: "Brown", Email: "bob@example.com", CreatedAt: t0, UpdatedAt: t0, Role: "admin"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("mapped slice = %#v\nwant            = %#v", got, want)
	}
}
