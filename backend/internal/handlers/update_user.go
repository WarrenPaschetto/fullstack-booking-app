package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userUpdater interface {
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func UpdateUserHandler(u userUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type update struct {
			RegisterResponse
		}
		userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing user in context", nil)
			return
		}
		decoder := json.NewDecoder(r.Body)
		req := UpdateUserRequest{}
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

		params := db.UpdateUserParams{
			ID:           userID,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
		}
		err = u.UpdateUser(r.Context(), params)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				utils.RespondWithError(w, http.StatusNotFound, "User not found", nil)
			case strings.Contains(err.Error(), "UNIQUE constraint failed: users.email"):
				utils.RespondWithError(w, http.StatusBadRequest, "Email already in use", nil)
			default:
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user", err)
			}
			return
		}

		updated, err := u.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not fetch updated user", err)
			return
		}
		if updated.ID != userID {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, update{
			RegisterResponse: RegisterResponse{
				ID:        updated.ID,
				FirstName: updated.FirstName,
				LastName:  updated.LastName,
				Email:     updated.Email,
				CreatedAt: updated.CreatedAt,
				UpdatedAt: updated.UpdatedAt,
			},
		})
	}
}
