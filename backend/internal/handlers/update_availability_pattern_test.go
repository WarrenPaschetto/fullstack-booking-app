package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type mockPatternUpdater struct {
	getPattern   db.AvailabilityPattern
	getErr       error
	updateArg    db.UpdateAvailabilityPatternParams
	updateErr    error
	calledGet    bool
	calledUpdate bool
}

func (m *mockPatternUpdater) GetAvailabilityPatternByID(ctx context.Context, id uuid.UUID) (db.AvailabilityPattern, error) {
	m.calledGet = true
	return m.getPattern, m.getErr
}

func (m *mockPatternUpdater) UpdateAvailabilityPattern(ctx context.Context, arg db.UpdateAvailabilityPatternParams) error {
	m.calledUpdate = true
	m.updateArg = arg
	return m.updateErr
}

func TestUpdateAvailabilityPatternHandler(t *testing.T) {
	patternID := uuid.New()
	ownerID := uuid.New()
	otherID := uuid.New()

	existing := db.AvailabilityPattern{
		ID:         patternID,
		ProviderID: ownerID,
		DayOfWeek:  2,
		StartTime:  time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	validBody := map[string]interface{}{
		"day_of_week": 4,
		"start_time":  "2025-06-01T10:00:00Z",
		"end_time":    "2025-06-01T12:00:00Z",
	}
	bodyBytes, _ := json.Marshal(validBody)

	tests := []struct {
		name            string
		setupRequest    func() *http.Request
		setupContext    func(*http.Request) *http.Request
		mock            *mockPatternUpdater
		wantStatus      int
		wantBodyContain string
		checkUpdate     func(t *testing.T, m *mockPatternUpdater)
	}{
		{
			name: "Success",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/"+patternID.String(), bytes.NewReader(bodyBytes))
				req = mux.SetURLVars(req, map[string]string{"id": patternID.String()})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, ownerID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, true)
				return req.WithContext(ctx)
			},
			mock: &mockPatternUpdater{
				getPattern: existing,
			},
			wantStatus: http.StatusOK,
			checkUpdate: func(t *testing.T, m *mockPatternUpdater) {
				if !m.calledUpdate {
					t.Error("expected UpdateAvailabilityPattern to be called")
				}
				if m.updateArg.ID != patternID {
					t.Errorf("got update ID %v; want %v", m.updateArg.ID, patternID)
				}
				if m.updateArg.DayOfWeek != 4 {
					t.Errorf("got day %d; want 4", m.updateArg.DayOfWeek)
				}
			},
		},
		{
			name: "Missing auth",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/"+patternID.String(), bytes.NewReader(bodyBytes))
				req = mux.SetURLVars(req, map[string]string{"id": patternID.String()})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				return req
			},
			mock:            &mockPatternUpdater{},
			wantStatus:      http.StatusUnauthorized,
			wantBodyContain: "Authentication required",
		},
		{
			name: "User not admin",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/"+patternID.String(), bytes.NewReader(bodyBytes))
				req = mux.SetURLVars(req, map[string]string{"id": patternID.String()})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, ownerID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, false)
				return req.WithContext(ctx)
			},
			mock:            &mockPatternUpdater{},
			wantStatus:      http.StatusForbidden,
			wantBodyContain: "Forbidden",
		},
		{
			name: "Missing pattern id",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPut, "/availability/patterns/", bytes.NewReader(bodyBytes))
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, ownerID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, true)
				return req.WithContext(ctx)
			},
			mock:            &mockPatternUpdater{},
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "Missing id",
		},
		{
			name: "Invalid id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/not-a-uuid", bytes.NewReader(bodyBytes))
				req = mux.SetURLVars(req, map[string]string{"id": "not-a-uuid"})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, ownerID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, true)
				return req.WithContext(ctx)
			},
			mock:            &mockPatternUpdater{},
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "Invalid id",
		},
		{
			name: "Pattern not found",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/"+patternID.String(), bytes.NewReader(bodyBytes))
				req = mux.SetURLVars(req, map[string]string{"id": patternID.String()})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, ownerID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, true)
				return req.WithContext(ctx)
			},
			mock: &mockPatternUpdater{
				getErr: errors.New("sql: no rows"),
			},
			wantStatus:      http.StatusNotFound,
			wantBodyContain: "Pattern not found",
		},
		{
			name: "Not owner",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/"+patternID.String(), bytes.NewReader(bodyBytes))
				req = mux.SetURLVars(req, map[string]string{"id": patternID.String()})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, otherID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, true)
				return req.WithContext(ctx)
			},
			mock: &mockPatternUpdater{
				getPattern: existing,
			},
			wantStatus:      http.StatusForbidden,
			wantBodyContain: "You do not own",
		},
		{
			name: "Invalid body",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/"+patternID.String(),
					strings.NewReader("{ invalid json"))
				req = mux.SetURLVars(req, map[string]string{"id": patternID.String()})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, ownerID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, true)
				return req.WithContext(ctx)
			},
			mock:            &mockPatternUpdater{getPattern: existing},
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "Invalid request body",
		},
		{
			name: "Update error",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/availability/patterns/"+patternID.String(), bytes.NewReader(bodyBytes))
				req = mux.SetURLVars(req, map[string]string{"id": patternID.String()})
				return req
			},
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, ownerID)
				ctx = context.WithValue(ctx, middleware.IsAdminKey, true)
				return req.WithContext(ctx)
			},
			mock: &mockPatternUpdater{
				getPattern: existing,
				updateErr:  errors.New("boom"),
			},
			wantStatus:      http.StatusInternalServerError,
			wantBodyContain: "Unable to update availability pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := tt.setupRequest()
			req = tt.setupContext(req)

			rr := httptest.NewRecorder()
			h := UpdateAvailabilityPatternHandler(tt.mock)
			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d; want %d; body=%q", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantBodyContain != "" {
				if !strings.Contains(rr.Body.String(), tt.wantBodyContain) {
					t.Errorf("body %q missing %q", rr.Body.String(), tt.wantBodyContain)
				}
			}
			if tt.checkUpdate != nil {
				tt.checkUpdate(t, tt.mock)
			}
		})
	}
}
