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

type mockPatternDeleter struct {
	called      bool
	gotID       uuid.UUID
	gotProvider uuid.UUID
	returnErr   error
}

func (m *mockPatternDeleter) DeleteAvailabilityPattern(ctx context.Context, id, providerID uuid.UUID) error {
	m.called = true
	m.gotID = id
	m.gotProvider = providerID
	return m.returnErr
}

func TestDeleteAvailabilityPatternHandler(t *testing.T) {
	patternID := uuid.New()
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
			url:        "/availability/pattern/" + patternID.String(),
			vars:       map[string]string{"id": patternID.String()},
			isAdmin:    true,
			injectUser: true,
			dbErr:      nil,
			wantStatus: http.StatusNoContent,
		},
		{
			name:        "Not admin",
			url:         "/availability/pattern/" + patternID.String(),
			vars:        map[string]string{"id": patternID.String()},
			isAdmin:     false,
			injectUser:  true,
			dbErr:       nil,
			wantStatus:  http.StatusForbidden,
			wantBodySub: "Forbidden",
		},
		{
			name:        "Missing user",
			url:         "/availability/pattern/" + patternID.String(),
			vars:        map[string]string{"id": patternID.String()},
			isAdmin:     true,
			injectUser:  false,
			dbErr:       nil,
			wantStatus:  http.StatusUnauthorized,
			wantBodySub: "Authentication required",
		},
		{
			name:        "DB error",
			url:         "/availability/pattern/" + patternID.String(),
			vars:        map[string]string{"id": patternID.String()},
			isAdmin:     true,
			injectUser:  true,
			dbErr:       errors.New("oops"),
			wantStatus:  http.StatusInternalServerError,
			wantBodySub: "Unable to delete availability pattern",
		},
		{
			name:        "Missing pattern ID param",
			url:         "/availability/pattern/",
			vars:        map[string]string{},
			isAdmin:     true,
			injectUser:  true,
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "Missing pattern ID",
		},
		{
			name:        "Invalid pattern ID param",
			url:         "/availability/pattern/not-a-uuid",
			vars:        map[string]string{"id": "not-a-uuid"},
			isAdmin:     true,
			injectUser:  true,
			wantStatus:  http.StatusBadRequest,
			wantBodySub: "Invalid pattern ID",
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

			mock := &mockPatternDeleter{returnErr: tt.dbErr}
			handler := DeleteAvailabilityPatternHandler(mock)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if tt.wantStatus == http.StatusNoContent {
				if rr.Code != http.StatusNoContent {
					t.Fatalf("expected 204; got %d", rr.Code)
				}
				if rr.Body.Len() != 0 {
					t.Errorf("expected empty body; got %q", rr.Body.String())
				}
				if !mock.called {
					t.Error("expected DeleteAvailabilityPattern to be called")
				}
				if mock.gotID != patternID || mock.gotProvider != providerID {
					t.Errorf("called with (%v,%v); want (%v,%v)", mock.gotID, mock.gotProvider, patternID, providerID)
				}
				return
			}

			if tt.wantStatus != http.StatusNoContent {
				if rr.Code != tt.wantStatus {
					t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
				}
				if tt.wantBodySub != "" && !strings.Contains(rr.Body.String(), tt.wantBodySub) {
					t.Errorf("expected response to contain %q, got %q", tt.wantBodySub, rr.Body.String())
				}
			}
		})
	}
}
