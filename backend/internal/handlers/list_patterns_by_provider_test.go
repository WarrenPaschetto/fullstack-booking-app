package handlers

import (
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

type mockPatternLister struct {
	called        bool
	gotProviderID uuid.UUID
	returnSlots   []db.ListPatternsByProviderRow
	returnError   error
}

func (m *mockPatternLister) ListPatternsByProvider(ctx context.Context, providerID uuid.UUID) ([]db.ListPatternsByProviderRow, error) {
	m.called = true
	m.gotProviderID = providerID
	return m.returnSlots, m.returnError
}

func TestListPatternsByProviderHandler(t *testing.T) {
	providerID := uuid.New()
	start := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	day := int64(3)
	now := time.Now()

	sample := db.ListPatternsByProviderRow{
		ID:        uuid.New(),
		DayOfWeek: day,
		StartTime: start,
		EndTime:   end,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name              string
		injectUser        bool
		noProviderID      bool
		invalidProviderID bool
		mockSlots         []db.ListPatternsByProviderRow
		mockErr           error
		wantStatus        int
		wantContains      string
		wantSlots         []PatternsResponse
	}{
		{
			name:              "Success",
			injectUser:        true,
			noProviderID:      false,
			invalidProviderID: false,
			mockSlots:         []db.ListPatternsByProviderRow{sample},
			mockErr:           nil,
			wantStatus:        http.StatusOK,
			wantSlots: []PatternsResponse{{
				ID:        sample.ID,
				DayOfWeek: sample.DayOfWeek,
				StartTime: sample.StartTime.(time.Time),
				EndTime:   sample.EndTime.(time.Time),
				CreatedAt: sample.CreatedAt,
				UpdatedAt: sample.UpdatedAt,
			}},
		},
		{
			name:              "Missing provider id",
			injectUser:        true,
			noProviderID:      true,
			invalidProviderID: false,
			mockErr:           nil,
			wantStatus:        http.StatusBadRequest,
			wantContains:      "Missing provider_id",
		},
		{
			name:              "Invalid provider id",
			injectUser:        true,
			noProviderID:      false,
			invalidProviderID: true,
			mockErr:           nil,
			wantStatus:        http.StatusBadRequest,
			wantContains:      "Invalid provider_id",
		},
		{
			name:              "User not authenticated",
			injectUser:        false,
			noProviderID:      false,
			invalidProviderID: false,
			mockErr:           nil,
			wantStatus:        http.StatusUnauthorized,
			wantContains:      "Authentication required",
		},
		{
			name:              "DB error",
			injectUser:        true,
			noProviderID:      false,
			invalidProviderID: false,
			mockSlots:         []db.ListPatternsByProviderRow{sample},
			mockErr:           errors.New("boom"),
			wantStatus:        http.StatusInternalServerError,
			wantContains:      "Unable to retrieve availability patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var req *http.Request

			switch {
			case tt.noProviderID:
				req = httptest.NewRequest(http.MethodGet, "/availability/patterns/", nil)

			case tt.invalidProviderID:
				req = httptest.NewRequest(http.MethodGet, "/availability/patterns/not-a-uuid", nil)
				req = mux.SetURLVars(req, map[string]string{"provider_id": "not-a-uuid"})

			default:
				req = httptest.NewRequest(http.MethodGet, "/availability/patterns/"+providerID.String(), nil)
				req = mux.SetURLVars(req, map[string]string{"provider_id": providerID.String()})
			}

			ctx := req.Context()
			if tt.injectUser {
				ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			}
			req = req.WithContext(ctx)

			mock := &mockPatternLister{returnSlots: tt.mockSlots, returnError: tt.mockErr}
			handler := ListPatternsByProviderHandler(mock)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d; got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantContains != "" {
				if !strings.Contains(rr.Body.String(), tt.wantContains) {
					t.Errorf("expected error containing %q; got %q", tt.wantContains, rr.Body.String())
				}
				return
			}

			var got []PatternsResponse
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode JSON response: %v", err)
			}

			if len(got) != len(tt.wantSlots) {
				t.Fatalf("expected %d slots; got %d", len(tt.wantSlots), len(got))
			}

			for i, want := range tt.wantSlots {
				gotItem := got[i]

				if gotItem.ID != want.ID {
					t.Errorf("slot %d ID = %v; want %v", i, gotItem.ID, want.ID)
				}
				if gotItem.DayOfWeek != want.DayOfWeek {
					t.Errorf("slot %d day = %d; want %d", i, gotItem.DayOfWeek, want.DayOfWeek)
				}
				if !gotItem.StartTime.Equal(want.StartTime) {
					t.Errorf("slot %d StartTime = %v; want %v", i, gotItem.StartTime, want.StartTime)
				}
				if !gotItem.EndTime.Equal(want.EndTime) {
					t.Errorf("slot %d EndTime = %v; want %v", i, gotItem.EndTime, want.EndTime)
				}
			}

		})
	}
}
