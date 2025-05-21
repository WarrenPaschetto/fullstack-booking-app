package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uuid.New().String(),
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte("testsecret"))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	noSubToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	noSubTokenString, err := noSubToken.SignedString([]byte("testsecret"))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	badSubToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "not-uuid",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	badSubTokenString, err := badSubToken.SignedString([]byte("testsecret"))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	rsPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	rsToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": uuid.New().String(),
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	rsString, err := rsToken.SignedString(rsPriv)
	if err != nil {
		t.Fatalf("Failed to sign RS256 token: %v", err)
	}

	tests := []struct {
		name             string
		authHeader       string
		secret           string
		expectStatus     int
		expectContains   string
		shouldFailClaims bool
	}{
		{
			name:             "Valid token",
			authHeader:       "Bearer " + tokenString,
			secret:           "testsecret",
			expectStatus:     http.StatusOK,
			expectContains:   "user_id is: ",
			shouldFailClaims: false,
		},
		{
			name:             "Missing JWT_SECRET",
			authHeader:       "Bearer " + tokenString,
			secret:           "",
			expectStatus:     http.StatusInternalServerError,
			expectContains:   "Missing JWT_SECRET",
			shouldFailClaims: false,
		},
		{
			name:             "Wrong signing method",
			authHeader:       "Bearer " + rsString,
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			expectContains:   "Invalid or expired token",
			shouldFailClaims: false,
		},
		{
			name:             "Invalid token claims",
			authHeader:       "Bearer " + tokenString,
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			expectContains:   "Invalid token claims",
			shouldFailClaims: true,
		},
		{
			name:             "Token missing subject claims",
			authHeader:       "Bearer " + noSubTokenString,
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			expectContains:   "Missing subject claim",
			shouldFailClaims: false,
		},
		{
			name:             "Invalid user ID format",
			authHeader:       "Bearer " + badSubTokenString,
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			expectContains:   "Invalid user ID format",
			shouldFailClaims: false,
		},
		{
			name:             "Missing token",
			authHeader:       "",
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			expectContains:   "Missing or malformed token",
			shouldFailClaims: false,
		},
		{
			name:             "Invalid token",
			authHeader:       "Bearer " + "",
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			expectContains:   "Invalid or expired token",
			shouldFailClaims: false,
		},
		{
			name:             "Missing Authorization header",
			authHeader:       "",
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			shouldFailClaims: false,
		},
		{
			name:             "Malformed Authorization header",
			authHeader:       "BadFormatToken",
			secret:           "testsecret",
			expectStatus:     http.StatusUnauthorized,
			shouldFailClaims: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Getenv("JWT_SECRET")
			t.Cleanup(func() { os.Setenv("JWT_SECRET", old) })
			os.Setenv("JWT_SECRET", tt.secret)

			oldParse := ParseTokenFn
			if tt.shouldFailClaims {
				ParseTokenFn = func(tokenString string, keyFunc jwt.Keyfunc) (*jwt.Token, error) {
					return &jwt.Token{
						Valid:  true,
						Claims: &jwt.RegisteredClaims{},
					}, nil
				}
			}
			defer func() { ParseTokenFn = oldParse }()

			handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				val := r.Context().Value(UserIDKey)
				userID, ok := val.(uuid.UUID)
				if !ok || userID == uuid.Nil {
					http.Error(w, "user_id missing or invalid", http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("user_id is: " + userID.String()))
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}
			if tt.expectContains != "" && !strings.Contains(rr.Body.String(), tt.expectContains) {
				t.Errorf("expected response to contain %q, got %q", tt.expectContains, rr.Body.String())
			}
		})
	}
}

func TestIsAdminFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			name:     "No value in context",
			ctx:      context.Background(),
			expected: false,
		},
		{
			name:     "Value is true",
			ctx:      context.WithValue(context.Background(), IsAdminKey, true),
			expected: true,
		},
		{
			name:     "Value is false",
			ctx:      context.WithValue(context.Background(), IsAdminKey, false),
			expected: false,
		},
		{
			name:     "Value wrong type",
			ctx:      context.WithValue(context.Background(), IsAdminKey, "not a bool"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAdminFromContext(tt.ctx)
			if got != tt.expected {
				t.Fatalf("IsAdminFromContext returned %v; expected %v", got, tt.expected)
			}
		})
	}
}
