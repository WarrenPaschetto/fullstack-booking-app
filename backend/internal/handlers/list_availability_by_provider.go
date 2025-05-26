package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProviderAvailabilityLister interface {
	ListAvailabilityByProvider(ctx context.Context, providerID uuid.UUID) ([]db.Availability, error)
}

type AvailabilityResponse struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ListAvailabilityByProviderHandler(q ProviderAvailabilityLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		providerIdStr, ok := vars["provider_id"]
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing provider_id in URL", nil)
			return
		}

		providerID, err := uuid.Parse(providerIdStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid provider_id", err)
			return
		}

		freeSlots, err := q.ListAvailabilityByProvider(r.Context(), providerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to retrieve availability", err)
			return
		}

		resp := make([]AvailabilityResponse, 0, len(freeSlots))
		for _, slot := range freeSlots {
			resp = append(resp, AvailabilityResponse{
				ID:         slot.ID,
				ProviderID: slot.ProviderID,
				StartTime:  slot.StartTime,
				EndTime:    slot.EndTime,
				CreatedAt:  slot.CreatedAt,
				UpdatedAt:  slot.UpdatedAt,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}
