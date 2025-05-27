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
)

type mockRangeLister struct {
	called        bool
	gotProviderID uuid.UUID
	gotParams     db.ListAvailabilityInRangeParams
	returnSlots   []db.ListAvailabilityInRangeRow
	returnError   error
}

func (m *mockRangeLister) ListAvailabilityInRange(ctx context.Context, params db.ListAvailabilityInRangeParams) ([]db.ListAvailabilityInRangeRow, error) {
	m.called = true
	m.gotProviderID = params.ProviderID
	m.gotParams = params
	return m.returnSlots, m.returnError
}

func TestListAvailabilityInRangeHandler(t *testing.T) {
	providerID := uuid.New()
	start := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	sample := db.ListAvailabilityInRangeRow{
		ID:         uuid.New(),
		ProviderID: providerID,
		StartTime:  start,
		EndTime:    end,
	}

	tests := []struct {
		name         string
		query        string
		injectUser   bool
		mockSlots    []db.ListAvailabilityInRangeRow
		mockErr      error
		wantStatus   int
		wantContains string
		wantSlots    []InRangeResponse
	}{
		{
			name:       "Success",
			query:      "?start=" + start.Format(time.RFC3339) + "&end=" + end.Format(time.RFC3339),
			injectUser: true,
			mockSlots:  []db.ListAvailabilityInRangeRow{sample},
			mockErr:    nil,
			wantStatus: http.StatusOK,
			wantSlots:  []InRangeResponse{{ID: sample.ID, ProviderID: sample.ProviderID, StartTime: sample.StartTime, EndTime: sample.EndTime}},
		},
		{
			name:         "Missing end time query params",
			query:        "?start=" + start.Format(time.RFC3339) + "&end=",
			injectUser:   true,
			mockErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantContains: "Start and end query params are required",
		},
		{
			name:         "Invalid start time",
			query:        "?start=" + "not-a-valid-time" + "&end=" + end.Format(time.RFC3339),
			injectUser:   true,
			mockErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantContains: "Invalid start time",
		},
		{
			name:         "Invalid end time",
			query:        "?start=" + start.Format(time.RFC3339) + "&end=" + "not-a-valid-time",
			injectUser:   true,
			mockErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantContains: "Invalid end time",
		},
		{
			name:         "End time before start time",
			query:        "?start=" + end.Format(time.RFC3339) + "&end=" + start.Format(time.RFC3339),
			injectUser:   true,
			mockErr:      nil,
			wantStatus:   http.StatusBadRequest,
			wantContains: "End time must be after start time",
		},
		{
			name:         "User not authenticated",
			query:        "?start=" + start.Format(time.RFC3339) + "&end=" + end.Format(time.RFC3339),
			injectUser:   false,
			mockErr:      nil,
			wantStatus:   http.StatusUnauthorized,
			wantContains: "Authentication required",
		},
		{
			name:         "DB error",
			query:        "?start=" + start.Format(time.RFC3339) + "&end=" + end.Format(time.RFC3339),
			injectUser:   true,
			mockSlots:    []db.ListAvailabilityInRangeRow{sample},
			mockErr:      errors.New("boom"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Unable to retrieve availability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/availability"+tt.query, nil)

			ctx := req.Context()
			if tt.injectUser {
				ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			}
			req = req.WithContext(ctx)

			mock := &mockRangeLister{returnSlots: tt.mockSlots, returnError: tt.mockErr}
			handler := ListAvailabilityInRangeHandler(mock)

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

			var got []InRangeResponse
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode JSON response: %v", err)
			}

			if len(got) != len(tt.wantSlots) {
				t.Fatalf("expected %d slots; got %d", len(tt.wantSlots), len(got))
			}
			for i := range got {
				if got[i] != tt.wantSlots[i] {
					t.Fatalf("slot %d: expected %+v; got %+v", i, tt.wantSlots[i], got[i])
				}
			}
		})
	}
}
