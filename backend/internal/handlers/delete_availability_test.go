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
		name        string
		url         string
		vars        map[string]string
		isAdmin     bool
		injectUser  bool
		dbErr       error
		wantStatus  int
		wantBodySub string
	}{
		{
			name:       "Success",
			url:        "/availability/" + slotID.String(),
			vars:       map[string]string{"id": slotID.String()},
			isAdmin:    true,
			injectUser: true,
			dbErr:      nil,
			wantStatus: http.StatusNoContent,
		},
		{
			name:        "Not admin",
			url:         "/availability/" + slotID.String(),
			vars:        map[string]string{"id": slotID.String()},
			isAdmin:     false,
			injectUser:  true,
			dbErr:       nil,
			wantStatus:  http.StatusForbidden,
			wantBodySub: "Forbidden",
		},
		{
			name:        "Missing user",
			url:         "/availability/" + slotID.String(),
			vars:        map[string]string{"id": slotID.String()},
			isAdmin:     true,
			injectUser:  false,
			dbErr:       nil,
			wantStatus:  http.StatusInternalServerError,
			wantBodySub: "Could not get user ID",
		},
		{
			name:        "DB error",
			url:         "/availability/" + slotID.String(),
			vars:        map[string]string{"id": slotID.String()},
			isAdmin:     true,
			injectUser:  true,
			dbErr:       errors.New("oops"),
			wantStatus:  http.StatusInternalServerError,
			wantBodySub: "Unable to delete availability",
		},
		{
			name:        "Missing slot ID param",
			url:         "/availability",
			vars:        map[string]string{},
			isAdmin:     true,
			injectUser:  true,
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "Missing slot ID",
		},
		{
			name:        "Invalid slot ID param",
			url:         "/availability/not-a-uuid",
			vars:        map[string]string{"id": "not-a-uuid"},
			isAdmin:     true,
			injectUser:  true,
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "Invalid slot ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			req = mux.SetURLVars(req, tt.vars)

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
			if tt.wantBodySub != "" && !strings.Contains(rr.Body.String(), tt.wantBodySub) {
				t.Errorf("expected response to contain %q, got %q", tt.wantBodySub, rr.Body.String())
			}
		})
	}
}
