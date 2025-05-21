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

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
)

type mockAdminRegisterQueries struct {
	db.Queries
	shouldFailInsert bool
	shouldFailFetch  bool
}

func (m *mockAdminRegisterQueries) CreateUser(_ context.Context, user db.CreateUserParams) error {
	if user.Email == usedEmail {
		return errors.New("UNIQUE constraint failed: users.email")
	}
	if m.shouldFailInsert {
		return errInsertFailed
	}
	return nil
}

func (m *mockAdminRegisterQueries) GetUserByEmail(_ context.Context, email string) (db.User, error) {
	if m.shouldFailFetch {
		return db.User{}, errors.New("Unable to fetch new user")
	}
	return db.User{
		ID:           uuid.New(),
		FirstName:    "John",
		LastName:     "Doe",
		Email:        email,
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Role:         "admin",
	}, nil
}

type AdminRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
}

var repeatEmail = "usedEmail@email.com"

func TestCreateAdminHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		mockQuery        *mockAdminRegisterQueries
		expectedCode     int
		expectedContains string
		shouldFailHash   bool
		injectAdmin      bool
	}{
		{
			name: "Valid registration",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
				Role:      "admin",
			},
			mockQuery:      &mockAdminRegisterQueries{},
			expectedCode:   http.StatusCreated,
			shouldFailHash: false,
			injectAdmin:    true,
		},
		{
			name: "Does not have role of admin",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
				Role:      "admin",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusForbidden,
			expectedContains: "Forbidden",
			shouldFailHash:   false,
			injectAdmin:      false,
		},
		{
			name:           "Invalid request body",
			requestBody:    "{ this is an invalid request body",
			mockQuery:      &mockAdminRegisterQueries{},
			expectedCode:   http.StatusBadRequest,
			shouldFailHash: false,
			injectAdmin:    true,
		},
		{
			name: "Missing email",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
				Password:  "strongpassword",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
			shouldFailHash:   false,
			injectAdmin:      true,
		},
		{
			name: "Missing password",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
			shouldFailHash:   false,
			injectAdmin:      true,
		},
		{
			name: "Missing first name",
			requestBody: AdminRequest{
				FirstName: "",
				LastName:  "Doe",
				Email:     "user@example.com",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
			shouldFailHash:   false,
			injectAdmin:      true,
		},
		{
			name: "Missing last name",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "",
				Email:     "user@example.com",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
			shouldFailHash:   false,
			injectAdmin:      true,
		},
		{
			name: "Hash failure",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Could not hash password",
			shouldFailHash:   true,
			injectAdmin:      true,
		},
		{
			name: "Insert failure",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery: &mockAdminRegisterQueries{
				shouldFailInsert: true,
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Failed to create user",
			shouldFailHash:   false,
			injectAdmin:      true,
		},
		{
			name: "Email already registered",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     repeatEmail,
				Password:  "strongpassword",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email already registered",
			shouldFailHash:   false,
			injectAdmin:      true,
		},
		{
			name: "Unable to retrieve new user",
			requestBody: AdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery: &mockAdminRegisterQueries{
				shouldFailFetch: true,
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Unable to fetch new admin",
			shouldFailHash:   false,
			injectAdmin:      true,
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

			handler := CreateAdminHandler(tt.mockQuery)

			jsonData, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/admin", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			if tt.injectAdmin {
				ctx := context.WithValue(req.Context(), middleware.IsAdminKey, true)
				req = req.WithContext(ctx)
			}

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
