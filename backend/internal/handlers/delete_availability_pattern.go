package handlers

import (
	"context"
	"net/http"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type availabilityPatternDeleter interface {
	DeleteAvailabilityPattern(ctx context.Context, id, providerID uuid.UUID) error
}

func DeleteAvailabilityPatternHandler(q availabilityPatternDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		providerID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
			return
		}

		vars := mux.Vars(r)
		patternIDStr, ok := vars["id"]
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing pattern ID", nil)
			return
		}
		patternID, err := uuid.Parse(patternIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid pattern ID", err)
			return
		}

		if err := q.DeleteAvailabilityPattern(r.Context(), patternID, providerID); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to delete availability pattern", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}
