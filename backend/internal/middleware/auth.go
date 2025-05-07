package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing or malformed token", nil)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			utils.RespondWithError(w, http.StatusInternalServerError, "Missing JWT_SECRET", nil)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token", err)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token claims", nil)
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing subject claim", nil)
			return
		}
		userUUID, err := uuid.Parse(sub)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID format", err)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userUUID)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
