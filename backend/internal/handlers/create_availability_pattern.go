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
)

type AvailabilityPatternCreator interface {
	CreateAvailabilityPattern(ctx context.Context, arg db.CreateAvailabilityPatternParams) error
}

func CreateAvailabilityPatternHandler(q AvailabilityPatternCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		var req struct {
			DayOfWeek int       `json:"day_of_week"`
			StartTime time.Time `json:"start_time"`
			EndTime   time.Time `json:"end_time"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
			utils.RespondWithError(w, http.StatusBadRequest,
				"day_of_week must be 0 (Sunday) through 6 (Saturday)", nil)
			return
		}
		if !req.EndTime.After(req.StartTime) {
			utils.RespondWithError(w, http.StatusBadRequest,
				"end_time must be after start_time", nil)
			return
		}

		providerID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
			return
		}

		arg := db.CreateAvailabilityPatternParams{
			ID:         uuid.New(),
			ProviderID: providerID,
			DayOfWeek:  int64(req.DayOfWeek),
			StartTime:  req.StartTime,
			EndTime:    req.EndTime,
		}

		if err := q.CreateAvailabilityPattern(r.Context(), arg); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to create availability pattern", err)
			return
		}

		resp := struct {
			ID        uuid.UUID `json:"id"`
			DayOfWeek int       `json:"day_of_week"`
			StartTime time.Time `json:"start_time"`
			EndTime   time.Time `json:"end_time"`
		}{
			ID:        arg.ID,
			DayOfWeek: req.DayOfWeek,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
		}
		utils.RespondWithJSON(w, http.StatusCreated, resp)

	}
}
