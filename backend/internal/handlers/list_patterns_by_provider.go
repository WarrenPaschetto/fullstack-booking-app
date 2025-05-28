package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type providerPatternsLister interface {
	ListPatternsByProvider(ctx context.Context, providerID uuid.UUID) ([]db.ListPatternsByProviderRow, error)
}

type PatternsResponse struct {
	ID        uuid.UUID `json:"id"`
	DayOfWeek int64     `json:"day_of_week"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ListPatternsByProviderHandler(q providerPatternsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		providerIdStr, ok := vars["provider_id"]
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing provider_id", nil)
			return
		}

		providerID, err := uuid.Parse(providerIdStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid provider_id", err)
			return
		}

		if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
			return
		}

		patterns, err := q.ListPatternsByProvider(r.Context(), providerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to retrieve availability patterns", err)
			return
		}

		resp := make([]PatternsResponse, len(patterns))
		for i, p := range patterns {
			st, ok := p.StartTime.(time.Time)
			if !ok {
				utils.RespondWithError(w, http.StatusInternalServerError,
					"Malformed start_time in DB row", nil)
				return
			}
			et, ok := p.EndTime.(time.Time)
			if !ok {
				utils.RespondWithError(w, http.StatusInternalServerError,
					"Malformed end_time in DB row", nil)
				return
			}
			resp[i] = PatternsResponse{
				ID:        p.ID,
				DayOfWeek: p.DayOfWeek,
				StartTime: st,
				EndTime:   et,
				CreatedAt: p.CreatedAt,
				UpdatedAt: p.UpdatedAt,
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}
