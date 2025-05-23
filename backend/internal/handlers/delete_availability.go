package handlers

import (
	"context"
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type availabilityDeleter interface {
	DeleteAvailability(ctx context.Context, id, providerID uuid.UUID) error
}

func DeleteAvailabilityHandler(q availabilityDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		providerID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
			return
		}

		vars := mux.Vars(r)
		slotIDStr, ok := vars["id"]
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing slot ID", nil)
			return
		}
		slotID, err := uuid.Parse(slotIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid slot ID", err)
			return
		}

		if err := q.DeleteAvailability(r.Context(), slotID, providerID); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to delete availability", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}
