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

type mockFreeSlotsLister struct {
	called      bool
	gotArg      db.ListAllFreeSlotsParams
	returnSlots []db.ListAllFreeSlotsRow
	returnErr   error
}

func (m *mockFreeSlotsLister) ListAllFreeSlots(ctx context.Context, arg db.ListAllFreeSlotsParams) ([]db.ListAllFreeSlotsRow, error) {
	m.called = true
	m.gotArg = arg
	return m.returnSlots, m.returnErr
}

func TestListAllFreeSlotsHandler(t *testing.T) {
	providerID := uuid.New()
	start := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 1, 17, 0, 0, 0, time.UTC)
	sample := db.ListAllFreeSlotsRow{
		ID:        uuid.New(),
		StartTime: start,
		EndTime:   end,
	}

	tests := []struct {
		name            string
		query           string
		injectUser      bool
		mockSlots       []db.ListAllFreeSlotsRow
		mockErr         error
		wantStatus      int
		wantErrContains string
		wantSlots       []listResponse
	}{
		{
			name:       "Success",
			query:      "?start=" + start.Format(time.RFC3339) + "&end=" + end.Format(time.RFC3339),
			injectUser: true,
			mockSlots:  []db.ListAllFreeSlotsRow{sample},
			mockErr:    nil,
			wantStatus: http.StatusOK,
			wantSlots:  []listResponse{{ID: sample.ID, StartTime: sample.StartTime, EndTime: sample.EndTime}},
		},
		{
			name:            "Invalid start time",
			query:           "?start=" + "not-a-valid-time" + "&end=" + end.Format(time.RFC3339),
			injectUser:      true,
			mockSlots:       nil,
			mockErr:         nil,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "Invalid start time",
		},
		{
			name:            "Invalid end time",
			query:           "?start=" + start.Format(time.RFC3339) + "&end=" + "not-a-valid-time",
			injectUser:      true,
			mockSlots:       nil,
			mockErr:         nil,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "Invalid end time",
		},
		{
			name:            "End time before start time",
			query:           "?start=" + end.Format(time.RFC3339) + "&end=" + start.Format(time.RFC3339),
			injectUser:      true,
			mockSlots:       nil,
			mockErr:         nil,
			wantStatus:      http.StatusBadRequest,
			wantErrContains: "End must be after start",
		},
		{
			name:            "Missing user",
			query:           "?start=" + start.Format(time.RFC3339) + "&end=" + end.Format(time.RFC3339),
			injectUser:      false,
			mockSlots:       nil,
			mockErr:         nil,
			wantStatus:      http.StatusInternalServerError,
			wantErrContains: "Could not get user ID",
		},
		{
			name:            "DB error",
			query:           "?start=" + start.Format(time.RFC3339) + "&end=" + end.Format(time.RFC3339),
			injectUser:      true,
			mockSlots:       nil,
			mockErr:         errors.New("boom"),
			wantStatus:      http.StatusInternalServerError,
			wantErrContains: "Unable to retrieve available time slots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/availabilities/free"+tt.query, nil)

			ctx := req.Context()
			if tt.injectUser {
				ctx = context.WithValue(ctx, middleware.UserIDKey, providerID)
			}
			req = req.WithContext(ctx)

			mock := &mockFreeSlotsLister{returnSlots: tt.mockSlots, returnErr: tt.mockErr}
			handler := ListAllFreeSlotsHandler(mock)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d; got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantErrContains != "" {
				if !strings.Contains(rr.Body.String(), tt.wantErrContains) {
					t.Errorf("expected error containing %q; got %q", tt.wantErrContains, rr.Body.String())
				}
				return
			}

			var got []listResponse
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode JSON response: %v", err)
			}

			if len(got) != len(tt.wantSlots) {
				t.Fatalf("expected %d slots; got %d", len(tt.wantSlots), len(got))
			}
			for i := range got {
				if got[i] != tt.wantSlots[i] {
					t.Errorf("slot %d: expected %+v; got %+v", i, tt.wantSlots[i], got[i])
				}
			}
		})
	}
}
