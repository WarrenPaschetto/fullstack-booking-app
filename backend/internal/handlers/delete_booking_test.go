package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestDeleteBookingHandler(t *testing.T) {

	userID := uuid.New()
	bookingID := uuid.New()

	tests := []struct {
		name             string
		ctxUserID        any
		routeID          string
		mockDelete       func(ctx context.Context, arg db.DeleteBookingParams) error
		expectStatus     int
		expectedContains string
	}{
		{
			name:      "Delete booking",
			ctxUserID: userID,
			routeID:   bookingID.String(),
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				if arg.ID != bookingID {
					t.Errorf("expected booking ID %v, got %v", bookingID, arg.ID)
				}
				if arg.UserID != userID {
					t.Errorf("expected user ID %v, got %v", userID, arg.UserID)
				}
				return nil
			},
			expectStatus: http.StatusNoContent,
		},
		{
			name:      "DB error",
			ctxUserID: userID,
			routeID:   bookingID.String(),
			mockDelete: func(_ context.Context, _ db.DeleteBookingParams) error {
				return errors.New("simulated DB error")
			},
			expectStatus:     http.StatusInternalServerError,
			expectedContains: "Failed to delete booking",
		},
		{
			name:             "No user ID in context",
			ctxUserID:        nil,
			routeID:          bookingID.String(),
			mockDelete:       nil,
			expectStatus:     http.StatusUnauthorized,
			expectedContains: "User ID missing or not a UUID in context",
		},
		{
			name:             "Missing booking ID",
			ctxUserID:        userID,
			routeID:          "",
			mockDelete:       nil,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Missing slot ID",
		},
		{
			name:             "Invalid booking ID",
			ctxUserID:        userID,
			routeID:          "4595",
			mockDelete:       nil,
			expectStatus:     http.StatusBadRequest,
			expectedContains: "Invalid slot ID",
		},
		{
			name:      "Booking not found",
			ctxUserID: userID,
			routeID:   bookingID.String(),
			mockDelete: func(_ context.Context, _ db.DeleteBookingParams) error {
				return service.ErrBookingNotFound
			},
			expectStatus:     http.StatusNotFound,
			expectedContains: "Booking not found",
		},
		{
			name:      "User not allowed to delete this booking",
			ctxUserID: userID,
			routeID:   bookingID.String(),
			mockDelete: func(_ context.Context, _ db.DeleteBookingParams) error {
				return service.ErrNotAuthorized
			},
			expectStatus:     http.StatusForbidden,
			expectedContains: "Not allowed to delete this booking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{}
			if tt.mockDelete != nil {
				mockQ.DeleteBookingFn = tt.mockDelete
			} else {
				mockQ.DeleteBookingFn = func(_ context.Context, _ db.DeleteBookingParams) error {
					t.Fatalf("DeleteBooking should not have been called in test %q", tt.name)
					return nil
				}
			}

			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.DeleteBookingHandler()

			urlPath := "/api/bookings"
			if tt.routeID != "" {
				urlPath += "/" + tt.routeID
			}
			req := httptest.NewRequest(http.MethodDelete, urlPath, nil)

			if tt.ctxUserID != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID),
				)
			}

			if tt.routeID != "" {
				req = mux.SetURLVars(req, map[string]string{"id": tt.routeID})
			}

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, but got %d. Body: %q", tt.name, tt.expectStatus, rr.Code, rr.Body.String())
			}
			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected response to contain %q, got %s", tt.expectedContains, rr.Body.String())
			}
		})
	}
}
