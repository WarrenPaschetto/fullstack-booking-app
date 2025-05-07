package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
)

type mockRegisterQueries struct {
	db.Queries
	shouldFailInsert bool
}

func (m *mockRegisterQueries) CreateUser(_ context.Context, _ db.CreateUserParams) error {
	if m.shouldFailInsert {
		return errInsertFailed
	}
	return nil
}

func (m *mockRegisterQueries) GetUserByEmail(_ context.Context, email string) (db.User, error) {
	return db.User{
		ID:           uuid.New(),
		FirstName:    "John",
		LastName:     "Doe",
		Email:        email,
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

var errInsertFailed = &customError{msg: "insert failed"}

type customError struct{ msg string }

func (e *customError) Error() string { return e.msg }

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      RegisterRequest
		mockQuery        *mockRegisterQueries
		expectedCode     int
		expectedContains string
	}{
		{
			name: "valid registration",
			requestBody: RegisterRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery:    &mockRegisterQueries{},
			expectedCode: http.StatusCreated,
		},
		{
			name: "missing email",
			requestBody: RegisterRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
				Password:  "strongpassword",
			},
			mockQuery:        &mockRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
		},
		{
			name: "missing password",
			requestBody: RegisterRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "",
			},
			mockQuery:        &mockRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
		},
		{
			name: "missing first name",
			requestBody: RegisterRequest{
				FirstName: "",
				LastName:  "Doe",
				Email:     "user@example.com",
			},
			mockQuery:        &mockRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
		},
		{
			name: "missing last name",
			requestBody: RegisterRequest{
				FirstName: "John",
				LastName:  "",
				Email:     "user@example.com",
			},
			mockQuery:        &mockRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
		},
		{
			name: "insert failure",
			requestBody: RegisterRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery: &mockRegisterQueries{
				shouldFailInsert: true,
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RegisterHandler(tt.mockQuery)

			jsonData, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

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
