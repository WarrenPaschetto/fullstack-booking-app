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

type FreeSlotsLister interface {
	ListAllFreeSlots(ctx context.Context, arg db.ListAllFreeSlotsParams) ([]db.ListAllFreeSlotsRow, error)
}

type listResponse struct {
	ID        uuid.UUID `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func ListAllFreeSlotsHandler(l FreeSlotsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

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
			utils.RespondWithError(w, http.StatusBadRequest, "End must be after start", nil)
			return
		}

		providerID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
			return
		}

		freeSlots, err := l.ListAllFreeSlots(r.Context(), db.ListAllFreeSlotsParams{
			ProviderID: providerID,
			StartTime:  start,
			EndTime:    end,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to retrieve available time slots", err)
			return
		}

		resp := make([]listResponse, 0, len(freeSlots))
		for _, slot := range freeSlots {
			resp = append(resp, listResponse{
				ID:        slot.ID,
				StartTime: slot.StartTime,
				EndTime:   slot.EndTime,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}
