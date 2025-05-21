package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
)

type mockUpdateQueries struct {
	userNotFound     bool
	shouldFailUpdate bool
	shouldFailFetch  bool
	ReturnUser       db.User
}

func (m *mockUpdateQueries) UpdateUser(ctx context.Context, arg db.UpdateUserParams) error {
	if m.userNotFound {
		return sql.ErrNoRows
	}
	if m.shouldFailUpdate {
		return errors.New("Update failed")
	}
	if arg.Email == usedEmail {
		return errors.New("UNIQUE constraint failed: users.email")
	}
	return nil
}
func (m *mockUpdateQueries) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.shouldFailFetch {
		return db.User{}, errors.New("Could not fetch updated user")
	}
	return m.ReturnUser, nil
}
func (m *mockUpdateQueries) CreateUser(_ context.Context, _ db.CreateUserParams) error {
	return nil
}
func (m *mockUpdateQueries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockUpdateQueries) ListUsers(ctx context.Context) ([]db.User, error) {
	return []db.User{}, nil
}

func TestUpdateUserHandler(t *testing.T) {
	fakeUserId := uuid.New()
	fakeUser := db.User{
		ID:           fakeUserId,
		FirstName:    "John",
		LastName:     "Doe",
		Email:        "user@example.com",
		PasswordHash: "password",
	}
	wrongUser := db.User{
		ID:           uuid.Nil,
		FirstName:    "Mark",
		LastName:     "Smith",
		Email:        "M12email@example.com",
		PasswordHash: "secret",
	}

	tests := []struct {
		name             string
		requestBody      interface{}
		mockUpdate       *mockUpdateQueries
		expectedCode     int
		expectedContains string
		injectUserID     bool
		shouldFailHash   bool
	}{
		{
			name: "Successful update",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockUpdate: &mockUpdateQueries{
				ReturnUser: fakeUser,
			},
			expectedCode:   http.StatusOK,
			injectUserID:   true,
			shouldFailHash: false,
		},
		{
			name: "Wrong user",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockUpdate: &mockUpdateQueries{
				ReturnUser: wrongUser,
			},
			expectedCode:   http.StatusUnauthorized,
			injectUserID:   true,
			shouldFailHash: false,
		},
		{
			name: "Missing user in context",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockUpdate: &mockUpdateQueries{
				ReturnUser: fakeUser,
			},
			expectedCode:   http.StatusUnauthorized,
			injectUserID:   false,
			shouldFailHash: false,
		},
		{
			name:           "Invalid request body",
			requestBody:    "{ this is an invalid request body",
			mockUpdate:     &mockUpdateQueries{},
			expectedCode:   http.StatusBadRequest,
			injectUserID:   true,
			shouldFailHash: false,
		},
		{
			name: "Missing email",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
				Password:  "strongpassword",
			},
			mockUpdate:       &mockUpdateQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "Missing password",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "",
			},
			mockUpdate:       &mockUpdateQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "Missing first name",
			requestBody: UpdateUserRequest{
				FirstName: "",
				LastName:  "Doe",
				Email:     "user@example.com",
			},
			mockUpdate:       &mockUpdateQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "Missing last name",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "",
				Email:     "user@example.com",
			},
			mockUpdate:       &mockUpdateQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "User not found",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockUpdate: &mockUpdateQueries{
				userNotFound: true,
			},
			expectedCode:     http.StatusNotFound,
			expectedContains: "User not found",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "Insert failure",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockUpdate: &mockUpdateQueries{
				shouldFailUpdate: true,
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Failed to update user",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "Email already registered",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     usedEmail,
				Password:  "strongpassword",
			},
			mockUpdate:       &mockUpdateQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email already in use",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "Email not found",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "new-email@email.com",
				Password:  "strongpassword",
			},
			mockUpdate:       &mockUpdateQueries{shouldFailFetch: true},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Could not fetch updated user",
			injectUserID:     true,
			shouldFailHash:   false,
		},
		{
			name: "Hash failure",
			requestBody: UpdateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "newEmail@email.com",
				Password:  "strongpassword",
			},
			mockUpdate:       &mockUpdateQueries{},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Could not hash password",
			injectUserID:     true,
			shouldFailHash:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			oldHash := HashPasswordFn
			if tt.shouldFailHash {
				HashPasswordFn = func(_ []byte, _ int) ([]byte, error) {
					return nil, errors.New("simulated hash error")
				}
			}

			defer func() { HashPasswordFn = oldHash }()

			handler := UpdateUserHandler(tt.mockUpdate)

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/user", &buf)
			req.Header.Set("Content-Type", "application/json")
			if tt.injectUserID {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, fakeUserId)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, rr.Code)
			}

			if tt.expectedContains != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedContains)) {
				t.Errorf("expected response to contain %q, got %s", tt.expectedContains, rr.Body.String())
			}
		})
	}
}
