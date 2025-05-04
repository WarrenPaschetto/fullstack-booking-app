package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func RegisterHandler(queries db.UserQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type response struct {
			RegisterResponse
		}

		decoder := json.NewDecoder(r.Body)
		req := RegisterRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
			return
		}

		// validate input
		if req.Email == "" || req.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email and password required", nil)
			return
		}

		// hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
			return
		}

		// create user
		err = queries.CreateUser(r.Context(), db.CreateUserParams{
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
			return
		}

		// fetch created user
		user, err := queries.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to fetch new user", err)
			return
		}

		// convert UUID to string
		idStr, ok := user.ID.(string)
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not convert user ID", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, response{
			RegisterResponse: RegisterResponse{
				ID:        idStr,
				Email:     user.Email,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		})
	}
}
