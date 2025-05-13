package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
			name: "Valid registration",
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
			name: "Missing email",
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
			name: "Missing password",
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
			name: "Missing first name",
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
			name: "Missing last name",
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
			name: "Insert failure",
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

type mockUserQuerier struct {
	GetUserByEmailFn func(ctx context.Context, email string) (db.User, error)
}

func (m *mockUserQuerier) CreateUser(_ context.Context, _ db.CreateUserParams) error {
	return nil
}

func (m *mockUserQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return m.GetUserByEmailFn(ctx, email)
}
func TestLoginHandler(t *testing.T) {
	plain := "plain-password"
	hash, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	mockUser := db.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name         string
		secret       string
		body         any
		mockGet      func(ctx context.Context, email string) (db.User, error)
		wantStatus   int
		wantContains string
	}{
		{
			name:   "successful login",
			secret: "testsecret",
			body:   LoginRequest{Email: mockUser.Email, Password: plain},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return mockUser, nil
			},
			wantStatus:   http.StatusOK,
			wantContains: `"token":`,
		},
		{
			name:   "wrong password",
			secret: "testsecret",
			body:   LoginRequest{Email: mockUser.Email, Password: "wrongpass"},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return mockUser, nil
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "Invalid credentials",
		},
		{
			name:   "unknown email",
			secret: "testsecret",
			body:   LoginRequest{Email: "no@one.com", Password: "irrelevant"},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return db.User{}, errors.New("sql: no rows")
			},
			wantStatus:   http.StatusUnauthorized,
			wantContains: "Invalid credentials",
		},
		{
			name:       "malformed JSON",
			secret:     "testsecret",
			body:       `{ not json }`,
			mockGet:    nil, // handler will fail before calling DB
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Getenv("JWT_SECRET")
			t.Cleanup(func() { os.Setenv("JWT_SECRET", old) })
			os.Setenv("JWT_SECRET", tt.secret)

			// 1) Build handler with mock
			mockQ := &mockUserQuerier{GetUserByEmailFn: tt.mockGet}
			handler := LoginHandler(mockQ)

			// 2) Marshal body (or raw string)
			var buf bytes.Buffer
			switch b := tt.body.(type) {
			case string:
				buf.WriteString(b)
			default:
				_ = json.NewEncoder(&buf).Encode(b)
			}

			// 3) Create request
			req := httptest.NewRequest(http.MethodPost, "/login", &buf)
			req.Header.Set("Content-Type", "application/json")

			// 4) Call handler
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// 5) Check status
			if rr.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d; body=%q", tt.wantStatus, rr.Code, rr.Body.String())
			}
			// 6) Check response body contains expected string
			if tt.wantContains != "" && !strings.Contains(rr.Body.String(), tt.wantContains) {
				t.Errorf("response %q does not contain %q", rr.Body.String(), tt.wantContains)
			}
		})
	}
}
