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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockAdminRegisterQueries struct {
	db.AdminQuerier
	shouldFailInsert bool
	shouldFailFetch  bool
}

func (m *mockAdminRegisterQueries) CreateAdmin(_ context.Context, admin db.CreateAdminParams) error {
	if admin.Email == fakeEmail {
		return errors.New("UNIQUE constraint failed: admin.email")
	}
	if m.shouldFailInsert {
		return errorInsertFailed
	}
	return nil
}

var fakeEmail = "usedEmail@email.com"

func (m *mockAdminRegisterQueries) GetAdminByEmail(_ context.Context, email string) (db.Admin, error) {
	if m.shouldFailFetch {
		return db.Admin{}, errors.New("Unable to fetch admin")
	}
	return db.Admin{
		ID:           uuid.New(),
		FirstName:    "John",
		LastName:     "Doe",
		Email:        email,
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

var errorInsertFailed = &customizedError{msg: "insert failed"}

type customizedError struct{ msg string }

func (e *customizedError) Error() string { return e.msg }

func TestRegisterAdminHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		mockQuery        *mockAdminRegisterQueries
		expectedCode     int
		expectedContains string
		shouldFailHash   bool
	}{
		{
			name: "Valid registration",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery:      &mockAdminRegisterQueries{},
			expectedCode:   http.StatusCreated,
			shouldFailHash: false,
		},
		{
			name:           "Invalid request body",
			requestBody:    "{ this is an invalid request body",
			mockQuery:      &mockAdminRegisterQueries{},
			expectedCode:   http.StatusInternalServerError,
			shouldFailHash: false,
		},
		{
			name: "Missing email",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
				Password:  "strongpassword",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
			shouldFailHash:   false,
		},
		{
			name: "Missing password",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
			shouldFailHash:   false,
		},
		{
			name: "Missing first name",
			requestBody: RegisterAdminRequest{
				FirstName: "",
				LastName:  "Doe",
				Email:     "user@example.com",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
			shouldFailHash:   false,
		},
		{
			name: "Missing last name",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "",
				Email:     "user@example.com",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "First and last name required",
			shouldFailHash:   false,
		},
		{
			name: "Hash failure",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "Rambo",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Could not hash password",
			shouldFailHash:   true,
		},
		{
			name: "Insert failure",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery: &mockAdminRegisterQueries{
				shouldFailInsert: true,
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Failed to create admin",
			shouldFailHash:   false,
		},
		{
			name: "Email already registered",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     fakeEmail,
				Password:  "strongpassword",
			},
			mockQuery:        &mockAdminRegisterQueries{},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email already registered",
			shouldFailHash:   false,
		},
		{
			name: "Unable to retrieve new admin",
			requestBody: RegisterAdminRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "user@example.com",
				Password:  "strongpassword",
			},
			mockQuery: &mockAdminRegisterQueries{
				shouldFailFetch: true,
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Unable to fetch admin",
			shouldFailHash:   false,
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

			handler := RegisterAdminHandler(tt.mockQuery)

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

type mockAdminQuerier struct {
	GetAdminByEmailFn func(ctx context.Context, email string) (db.Admin, error)
}

func (m *mockAdminQuerier) CreateAdmin(ctx context.Context, p db.CreateAdminParams) error {
	return nil
}

func (m *mockAdminQuerier) GetAdminByEmail(ctx context.Context, email string) (db.Admin, error) {
	return m.GetAdminByEmailFn(ctx, email)
}

func TestLoginAdminHandler(t *testing.T) {
	plain := "plain-password"
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("could not hash password: %v", err)
	}
	mockAdmin := db.Admin{
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
		mockGet          func(ctx context.Context, email string) (db.Admin, error)
		expectedCode     int
		expectedContains string
		shouldFailSign   bool
	}{
		{
			name:   "Successful login",
			secret: "testsecret",
			body:   LoginAdminRequest{Email: mockAdmin.Email, Password: plain},
			mockGet: func(_ context.Context, email string) (db.Admin, error) {
				return mockAdmin, nil
			},
			expectedCode:     http.StatusOK,
			expectedContains: `"token":`,
			shouldFailSign:   false,
		},
		{
			name:   "Wrong password",
			secret: "testsecret",
			body:   LoginAdminRequest{Email: mockAdmin.Email, Password: "wrongpass"},
			mockGet: func(_ context.Context, email string) (db.Admin, error) {
				return mockAdmin, nil
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Invalid credentials",
			shouldFailSign:   false,
		},
		{
			name:   "Unknown email",
			secret: "testsecret",
			body:   LoginAdminRequest{Email: "no@one.com", Password: "irrelevant"},
			mockGet: func(_ context.Context, email string) (db.Admin, error) {
				return db.Admin{}, errors.New("sql: no rows")
			},
			expectedCode:     http.StatusUnauthorized,
			expectedContains: "Invalid credentials",
			shouldFailSign:   false,
		},
		{
			name:   "Missing email",
			secret: "testsecret",
			body:   LoginAdminRequest{Email: "", Password: plain},
			mockGet: func(_ context.Context, email string) (db.Admin, error) {
				return mockAdmin, nil
			},
			expectedCode:     http.StatusBadRequest,
			expectedContains: "Email and password required",
			shouldFailSign:   false,
		},
		{
			name:           "Malformed JSON",
			secret:         "testsecret",
			body:           `{ not json }`,
			mockGet:        nil,
			expectedCode:   http.StatusInternalServerError,
			shouldFailSign: false,
		},
		{
			name:   "Missing JWT_SECRET",
			secret: "",
			body:   LoginAdminRequest{Email: mockAdmin.Email, Password: plain},
			mockGet: func(_ context.Context, email string) (db.Admin, error) {
				return mockAdmin, nil
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Missing JWT_SECRET",
			shouldFailSign:   false,
		},
		{
			name:   "Sign token failure",
			secret: "testsecret",
			body:   LoginAdminRequest{Email: mockAdmin.Email, Password: plain},
			mockGet: func(_ context.Context, email string) (db.Admin, error) {
				return mockAdmin, nil
			},
			expectedCode:     http.StatusInternalServerError,
			expectedContains: "Failed to sign token",
			shouldFailSign:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			oldSign := SignTokenFn
			if tt.shouldFailSign {
				SignTokenFn = func(_ *jwt.Token, _ []byte) (string, error) {
					return "", errors.New("simulated sign error")
				}
			}
			defer func() { SignTokenFn = oldSign }()

			old := os.Getenv("JWT_SECRET")
			t.Cleanup(func() { os.Setenv("JWT_SECRET", old) })
			os.Setenv("JWT_SECRET", tt.secret)

			mockQ := &mockAdminQuerier{GetAdminByEmailFn: tt.mockGet}
			handler := LoginAdminHandler(mockQ)

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
