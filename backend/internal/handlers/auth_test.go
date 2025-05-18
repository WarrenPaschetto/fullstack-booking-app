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
	shouldFailFetch  bool
}

func (m *mockRegisterQueries) CreateUser(_ context.Context, user db.CreateUserParams) error {
	if user.Email == usedEmail {
		return errors.New("UNIQUE constraint failed: users.email")
	}
	if m.shouldFailInsert {
		return errInsertFailed
	}
	return nil
}

var usedEmail = "usedEmail@email.com"

func (m *mockRegisterQueries) GetUserByEmail(_ context.Context, email string) (db.User, error) {
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
	}, nil
}

var errInsertFailed = &customError{msg: "insert failed"}

type customError struct{ msg string }

func (e *customError) Error() string { return e.msg }

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
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
			name:         "Invalid request body",
			requestBody:  "{ this is an invalid request body",
			mockQuery:    &mockRegisterQueries{},
			expectedCode: http.StatusInternalServerError,
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
		{
			name: "Email already registered",
			requestBody: RegisterRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     usedEmail,
				Password:  "strongpassword",
			},
			mockQuery:        &mockRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email already registered",
		},
		{
			name: "Unable to retrieve new user",
			requestBody: RegisterRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery: &mockRegisterQueries{
				shouldFailFetch: true,
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Unable to fetch new user",
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

func (m *mockUserQuerier) CreateUser(ctx context.Context, p db.CreateUserParams) error {
	return nil // unused by LoginHandler
}

func (m *mockUserQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return m.GetUserByEmailFn(ctx, email)
}

func TestLoginHandler(t *testing.T) {
	plain := "plain-password"
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("could not hash password: %v", err)
	}
	mockUser := db.User{
		ID:           uuid.New(),
		Email:        "email@example.com",
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name             string
		secret           string
		body             any
		mockGet          func(ctx context.Context, email string) (db.User, error)
		expectedCode     int
		expectedContains string
	}{
		{
			name:   "Successful login",
			secret: "testsecret",
			body:   LoginRequest{Email: mockUser.Email, Password: plain},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return mockUser, nil
			},
			expectedCode:     http.StatusOK,
			expectedContains: `"token":`,
		},
		{
			name:   "Wrong password",
			secret: "testsecret",
			body:   LoginRequest{Email: mockUser.Email, Password: "wrongpass"},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return mockUser, nil
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid credentials",
		},
		{
			name:   "Unknown email",
			secret: "testsecret",
			body:   LoginRequest{Email: "no@one.com", Password: "irrelevant"},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return db.User{}, errors.New("sql: no rows")
			},
			expectedCode:     http.StatusUnauthorized,
			expectedContains: "Invalid credentials",
		},
		{
			name:   "Missing email",
			secret: "testsecret",
			body:   LoginRequest{Email: "", Password: plain},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return mockUser, nil
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
		},
		{
			name:         "Malformed JSON",
			secret:       "testsecret",
			body:         `{ not json }`,
			mockGet:      nil, // handler will fail before calling DB
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:   "Missing JWT_SECRET",
			secret: "",
			body:   LoginRequest{Email: mockUser.Email, Password: plain},
			mockGet: func(_ context.Context, email string) (db.User, error) {
				return mockUser, nil
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Missing JWT_SECRET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Getenv("JWT_SECRET")
			t.Cleanup(func() { os.Setenv("JWT_SECRET", old) })
			os.Setenv("JWT_SECRET", tt.secret)

			mockQ := &mockUserQuerier{GetUserByEmailFn: tt.mockGet}
			handler := LoginHandler(mockQ)

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", &buf)
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
