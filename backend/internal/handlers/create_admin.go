package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func CreateAdminHandler(q db.UserQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type response struct {
			RegisterResponse
		}

		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
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

		hashedPassword, err := HashPasswordFn([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
			return
		}

		err = q.CreateUser(r.Context(), db.CreateUserParams{
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			Role:         "admin",
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
				utils.RespondWithError(w, http.StatusBadRequest, "Email already registered", nil)
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
			return
		}

		admin, err := q.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to fetch new admin", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, response{
			RegisterResponse: RegisterResponse{
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
