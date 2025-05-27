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

type AvailabilityInRangeLister interface {
	ListAvailabilityInRange(ctx context.Context, params db.ListAvailabilityInRangeParams) ([]db.ListAvailabilityInRangeRow, error)
}

type InRangeResponse struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

func ListAvailabilityInRangeHandler(q AvailabilityInRangeLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

		if startStr == "" || endStr == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Start and end query params are required", nil)
			return
		}

		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid start time", err)
			return
		}
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid end time", err)
			return
		}
		if end.Before(start) {
			utils.RespondWithError(w, http.StatusBadRequest, "End time must be after start time", nil)
			return
		}

		providerID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
			return
		}

		freeSlots, err := q.ListAvailabilityInRange(r.Context(), db.ListAvailabilityInRangeParams{
			ProviderID: providerID,
			StartTime:  start,
			EndTime:    end,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to retrieve availability", err)
			return
		}

		resp := make([]InRangeResponse, len(freeSlots))
		for i, slot := range freeSlots {
			resp[i] = InRangeResponse{
				ID:         slot.ID,
				ProviderID: providerID,
				StartTime:  slot.StartTime,
				EndTime:    slot.EndTime,
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}
