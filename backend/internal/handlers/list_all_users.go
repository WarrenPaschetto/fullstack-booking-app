package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
)

type userLister interface {
	ListUsers(ctx context.Context) ([]db.User, error)
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserRole  string    `json:"user_role"`
}

func ListAllUsersHandler(u userLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		users, err := u.ListUsers(r.Context())
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to list users", err)
			return
		}

		resp := make([]UserResponse, 0, len(users))
		for _, u := range users {
			resp = append(resp, UserResponse{
				ID:        u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Email:     u.Email,
				CreatedAt: u.CreatedAt,
				UpdatedAt: u.UpdatedAt,
				UserRole:  u.UserRole,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, resp)

	}
}
