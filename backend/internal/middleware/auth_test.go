package middleware

import (
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
		name           string
		authHeader     string
		secret         string
		expectStatus   int
		expectContains string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + tokenString,
			secret:         "testsecret",
			expectStatus:   http.StatusOK,
			expectContains: "user_id is: ",
		},
		{
			name:           "Missing JWT_SECRET",
			authHeader:     "Bearer " + tokenString,
			secret:         "",
			expectStatus:   http.StatusInternalServerError,
			expectContains: "Missing JWT_SECRET",
		},
		{
			name:           "Wrong signing method",
			authHeader:     "Bearer " + rsString,
			secret:         "testsecret",
			expectStatus:   http.StatusUnauthorized,
			expectContains: "Invalid or expired token",
		},
		{
			name:           "Token missing subject claims",
			authHeader:     "Bearer " + noSubTokenString,
			secret:         "testsecret",
			expectStatus:   http.StatusUnauthorized,
			expectContains: "Missing subject claim",
		},
		{
			name:           "Invalid user ID format",
			authHeader:     "Bearer " + badSubTokenString,
			secret:         "testsecret",
			expectStatus:   http.StatusUnauthorized,
			expectContains: "Invalid user ID format",
		},
		{
			name:           "Missing token",
			authHeader:     "",
			secret:         "testsecret",
			expectStatus:   http.StatusUnauthorized,
			expectContains: "Missing or malformed token",
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer " + "",
			secret:         "testsecret",
			expectStatus:   http.StatusUnauthorized,
			expectContains: "Invalid or expired token",
		},
		{
			name:         "Missing Authorization header",
			authHeader:   "",
			secret:       "testsecret",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "Malformed Authorization header",
			authHeader:   "BadFormatToken",
			secret:       "testsecret",
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tt.secret)
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
