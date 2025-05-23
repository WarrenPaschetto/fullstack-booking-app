// internal/handlers/delete_availability_test.go
package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type mockDeleteQueries struct {
	called      bool
	gotID       uuid.UUID
	gotProvider uuid.UUID
	returnErr   error
}

func (m *mockDeleteQueries) DeleteAvailability(ctx context.Context, id, providerID uuid.UUID) error {
	m.called = true
	m.gotID = id
	m.gotProvider = providerID
	return m.returnErr
}

func TestDeleteAvailabilityHandler(t *testing.T) {
	slotID := uuid.New()
	providerID := uuid.New()

	tests := []struct {
		name       string
		isAdmin    bool
		injectUser bool
		dbErr      error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Success",
			isAdmin:    true,
			injectUser: true,
			dbErr:      nil,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Not admin",
			isAdmin:    false,
			injectUser: true,
			dbErr:      nil,
			wantStatus: http.StatusForbidden,
			wantBody:   "Forbidden",
		},
		{
			name:       "Missing user",
			isAdmin:    true,
			injectUser: false,
			dbErr:      nil,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Could not get user ID",
		},
		{
			name:       "DB error",
			isAdmin:    true,
			injectUser: true,
			dbErr:      errors.New("oops"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Unable to delete availability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/availability/"+slotID.String(), nil)
			req = mux.SetURLVars(req, map[string]string{"id": slotID.String()})

			ctx := req.Context()
			ctx = context.WithValue(ctx, middleware.IsAdminKey, tt.isAdmin)
			if tt.injectUser {
				ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			}
			req = req.WithContext(ctx)

			mock := &mockDeleteQueries{returnErr: tt.dbErr}
			handler := DeleteAvailabilityHandler(mock)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
			if tt.wantBody != "" && !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tt.wantBody, rr.Body.String())
			}
			if tt.wantStatus == http.StatusNoContent {
				if !mock.called {
					t.Errorf("expected DeleteAvailability to be called")
				}
				if mock.gotID != slotID {
					t.Errorf("expected id %s, got %s", slotID, mock.gotID)
				}
				if mock.gotProvider != providerID {
					t.Errorf("expected provider %s, got %s", providerID, mock.gotProvider)
				}
			}
		})
	}
}
