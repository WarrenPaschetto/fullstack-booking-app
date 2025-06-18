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
const IsAdminKey contextKey = "is_admin"

var ParseTokenFn = func(tokenString string, keyFunc jwt.Keyfunc) (*jwt.Token, error) {
	return jwt.Parse(tokenString, keyFunc)
}

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

		token, err := ParseTokenFn(tokenString, func(token *jwt.Token) (interface{}, error) {
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
		UserRole, _ := claims["user_role"].(string)
		isAdmin := (UserRole == "admin")

		ctx := context.WithValue(r.Context(), UserIDKey, userUUID)
		ctx = context.WithValue(ctx, IsAdminKey, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func IsAdminFromContext(ctx context.Context) bool {
	v := ctx.Value(IsAdminKey)
	isAdmin, _ := v.(bool)
	return isAdmin
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(UserIDKey)
	id, ok := v.(uuid.UUID)
	return id, ok
}
