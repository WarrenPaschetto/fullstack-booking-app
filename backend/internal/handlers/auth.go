package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userQuerier interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	UserRole  string `json:"user_role"`
}

type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserRole  string    `json:"user_role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

var HashPasswordFn = bcrypt.GenerateFromPassword
var SignTokenFn = func(tok *jwt.Token, secret []byte) (string, error) {
	return tok.SignedString(secret)
}

func RegisterHandler(q userQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type response struct {
			RegisterResponse
		}

		decoder := json.NewDecoder(r.Body)
		req := RegisterRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		if req.FirstName == "" || req.LastName == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "First and last name required", nil)
			return
		}

		if req.Email == "" || req.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email and password required", nil)
			return
		}
		if req.UserRole == "" {
			req.UserRole = "user"
		}

		hashedPassword, err := HashPasswordFn([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
			return
		}

		err = q.CreateUser(r.Context(), db.CreateUserParams{
			ID:           uuid.New(),
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			UserRole:     req.UserRole,
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
				utils.RespondWithError(w, http.StatusBadRequest, "Email already registered", nil)
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
			return
		}

		user, err := q.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to fetch new user", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, response{
			RegisterResponse: RegisterResponse{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				UserRole:  user.UserRole,
			},
		})
	}
}

func LoginHandler(q userQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		req := LoginRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		if req.Email == "" || req.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email and password required", nil)
			return
		}

		user, err := q.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}
		fmt.Println("Login user:", user.Email, "with role:", user.UserRole)

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid credentials", err)
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			utils.RespondWithError(w, http.StatusInternalServerError, "Missing JWT_SECRET", nil)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":       user.ID.String(),
			"user_role": user.UserRole,
			"firstName": user.FirstName,
			"iat":       jwt.NewNumericDate(time.Now()),
			"exp":       jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		})

		tokenString, err := SignTokenFn(token, []byte(secret))
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to sign token", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, LoginResponse{Token: tokenString})
	}
}
