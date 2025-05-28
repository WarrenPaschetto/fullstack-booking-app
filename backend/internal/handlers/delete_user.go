package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
)

type userDeleter interface {
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
type DeleteRequest struct {
	UserId uuid.UUID `json:"user_id"`
}

func DeleteUserHandler(q userDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		req := DeleteRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}
		if req.UserId == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID", nil)
			return
		}

		err = q.DeleteUser(r.Context(), req.UserId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.RespondWithError(w, http.StatusNotFound, "User not found", nil)
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to delete user", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
