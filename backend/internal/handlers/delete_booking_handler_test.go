package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/google/uuid"
)

func TestDeleteBookingHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	validBody := DeleteBookingRequest{
		ID: bookingID,
	}
	jsonBody, _ := json.Marshal(validBody)

	invalidBody := DeleteBookingRequest{
		ID: uuid.Nil,
	}
	invalidJsonBody, _ := json.Marshal(invalidBody)

	tests := []struct {
		name         string
		ctxUserID    any
		body         []byte
		mockDelete   func(ctx context.Context, arg db.DeleteBookingParams) error
		expectStatus int
	}{
		{
			name:      "Delete booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				if arg.ID != bookingID || arg.UserID != userID {
					t.Errorf("expected ID=%v and UserID=%v, got ID=%v and UserID=%v", bookingID, userID, arg.ID, arg.UserID)
				}
				return nil
			},
			expectStatus: http.StatusNoContent,
		},
		{
			name:      "DB error",
			ctxUserID: userID,
			body:      jsonBody,
			mockDelete: func(_ context.Context, _ db.DeleteBookingParams) error {
				return errors.New("simulated DB error")
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:         "No user ID",
			ctxUserID:    nil,
			body:         jsonBody,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "No booking ID",
			ctxUserID:    userID,
			body:         invalidJsonBody,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Malformed request body",
			ctxUserID:    userID,
			body:         []byte(`{invalid json`),
			mockDelete:   nil,
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockBookingQueries{
				DeleteBookingFn: tt.mockDelete,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.DeleteBookingHandler()

			req := httptest.NewRequest(http.MethodDelete, "/bookings", bytes.NewReader(tt.body))
			if tt.ctxUserID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID))
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
		})
	}
}
