package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
)

type AvailabilityPatternService interface {
	CreatePatternAndSlots(ctx context.Context, providerID uuid.UUID, dayOfWeek int32, start, end time.Time) error
}

func CreateAvailabilityPatternHandler(svc AvailabilityPatternService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		var req struct {
			DayOfWeek int32     `json:"day_of_week"`
			StartTime time.Time `json:"start_time"`
			EndTime   time.Time `json:"end_time"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}
		if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
			utils.RespondWithError(w, http.StatusBadRequest, "day_of_week must be 0–6", nil)
			return
		}
		if !req.EndTime.After(req.StartTime) {
			utils.RespondWithError(w, http.StatusBadRequest, "end_time must be after start_time", nil)
			return
		}

		providerID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
			return
		}

		err := svc.CreatePatternAndSlots(r.Context(), providerID, req.DayOfWeek, req.StartTime, req.EndTime)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create pattern and slots", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, map[string]string{
			"message": "Availability pattern and slots created successfully",
		})
	}
}
