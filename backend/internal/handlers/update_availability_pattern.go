package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type patternUpdater interface {
	GetAvailabilityPatternByID(ctx context.Context, id uuid.UUID) (db.AvailabilityPattern, error)
	UpdateAvailabilityPattern(ctx context.Context, arg db.UpdateAvailabilityPatternParams) error
}

type UpdateRequest struct {
	DayOfWeek int64     `json:"day_of_week"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type UpdateResponse struct {
	DayOfWeek int64     `json:"day_of_week"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	ID        uuid.UUID `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

func UpdateAvailabilityPatternHandler(q patternUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
			return
		}
		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		patternIdStr, ok := mux.Vars(r)["id"]
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing id", nil)
			return
		}

		patternID, err := uuid.Parse(patternIdStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id", err)
			return
		}

		existing, err := q.GetAvailabilityPatternByID(r.Context(), patternID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Pattern not found", err)
			return
		}
		if existing.ProviderID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "You do not own this pattern", nil)
			return
		}

		decoder := json.NewDecoder(r.Body)
		req := UpdateRequest{}
		err = decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		arg := db.UpdateAvailabilityPatternParams{
			DayOfWeek: req.DayOfWeek,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			ID:        patternID,
		}

		err = q.UpdateAvailabilityPattern(r.Context(), arg)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to update availability pattern", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, UpdateResponse{
			DayOfWeek: req.DayOfWeek,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			ID:        patternID,
			UpdatedAt: time.Now(),
		})
	}
}
