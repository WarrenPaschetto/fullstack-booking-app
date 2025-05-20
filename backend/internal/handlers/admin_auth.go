package handlers

import (
	"encoding/json"
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

type RegisterAdminRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type RegisterAdminResponse struct {
	ID           uuid.UUID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginAdminResponse struct {
	Token string `json:"token"`
}

func RegisterAdminHandler(queries db.AdminQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type response struct {
			RegisterAdminResponse
		}

		decoder := json.NewDecoder(r.Body)
		req := RegisterAdminRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid request body", err)
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

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
			return
		}

		err = queries.CreateAdmin(r.Context(), db.CreateAdminParams{
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: admin.email") {
				utils.RespondWithError(w, http.StatusBadRequest, "Email already registered", nil)
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create admin", err)
			return
		}

		admin, err := queries.GetAdminByEmail(r.Context(), req.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to fetch admin", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, response{
			RegisterAdminResponse: RegisterAdminResponse{
				ID:        admin.ID,
				FirstName: admin.FirstName,
				LastName:  admin.LastName,
				Email:     admin.Email,
				CreatedAt: admin.CreatedAt,
				UpdatedAt: admin.UpdatedAt,
			},
		})
	}
}

func LoginAdminHandler(queries db.AdminQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		req := LoginAdminRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid request body", err)
			return
		}

		if req.Email == "" || req.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email and password required", nil)
			return
		}

		admin, err := queries.GetAdminByEmail(r.Context(), req.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
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
			"sub": admin.ID,
			"iat": jwt.NewNumericDate(time.Now()),
			"exp": jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		})

		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to sign token", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, LoginAdminResponse{Token: tokenString})
	}
}
