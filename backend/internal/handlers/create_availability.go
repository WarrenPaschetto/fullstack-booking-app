package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/utils"
	"github.com/google/uuid"
)

type createAvailabilityRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func CreateAvailabilityHandler(queries db.BookingQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !middleware.IsAdminFromContext(r.Context()) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
			return
		}

		decoder := json.NewDecoder(r.Body)
		req := createAvailabilityRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		providerID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
			return
		}

		arg := db.CreateAvailabilityParams{
			ID:         uuid.New(),
			ProviderID: providerID,
			StartTime:  req.StartTime,
			EndTime:    req.EndTime,
		}

		if err := queries.CreateAvailability(r.Context(), arg); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Unable to create availability", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"id":         arg.ID,
			"start_time": arg.StartTime,
			"end_time":   arg.EndTime,
		})

	}
}
