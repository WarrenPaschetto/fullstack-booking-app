package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

func TestDeleteUserHandler(t *testing.T) {
	missingUser := uuid.New()

	mockUser := db.User{
		ID: uuid.New(),
	}

	tests := []struct {
		name             string
		body             any
		mockDelete       func(ctx context.Context, id uuid.UUID) error
		expectedCode     int
		expectedContains string
	}{
		{
			name: "Successful deletion",
			body: DeleteRequest{UserId: mockUser.ID},
			mockDelete: func(_ context.Context, id uuid.UUID) error {
				return nil
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name: "Invalid user id",
			body: DeleteRequest{UserId: uuid.Nil},
			mockDelete: func(_ context.Context, id uuid.UUID) error {
				return nil
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid user ID",
		},
		{
			name: "No user in db",
			body: DeleteRequest{UserId: missingUser},
			mockDelete: func(_ context.Context, id uuid.UUID) error {
				return sql.ErrNoRows
			},
			expectedCode:     http.StatusNotFound,
			expectedContains: "User not found",
		},
		{
			name: "Invalid request body",
			body: `{ not json }`,
			mockDelete: func(_ context.Context, id uuid.UUID) error {
				return nil
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid request body",
		},
		{
			name: "Unsuccessful delete",
			body: DeleteRequest{UserId: mockUser.ID},
			mockDelete: func(_ context.Context, id uuid.UUID) error {
				return errors.New("failed deletion")
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Unable to delete user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockUserQuerier{DeleteUserFn: tt.mockDelete}
			handler := DeleteUserHandler(mockQ)

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodDelete, "/", &buf)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d; body=%q", tt.expectedCode, rr.Code, rr.Body.String())
			}

			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected response to contain %q, got %s", tt.expectedContains, rr.Body.String())
			}
		})
	}
}
