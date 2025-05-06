package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

func TestAuthMiddleware(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user-id",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("testsecret"))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectStatus   int
		expectContains string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + tokenString,
			expectStatus:   http.StatusOK,
			expectContains: "user_id is: test-user-id",
		},
		{
			name:         "Missing Authorization header",
			authHeader:   "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "Malformed Authorization header",
			authHeader:   "BadFormatToken",
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				val := r.Context().Value(UserIDKey)
				userID, ok := val.(string)
				if !ok || userID == "" {
					http.Error(w, "user_id missing or invalid", http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("user_id is: " + userID))
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
