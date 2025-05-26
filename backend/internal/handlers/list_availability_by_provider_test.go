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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type mockProviderLister struct {
	called        bool
	gotProviderID uuid.UUID
	returnSlots   []db.Availability
	returnError   error
}

func (m *mockProviderLister) ListAvailabilityByProvider(ctx context.Context, providerID uuid.UUID) ([]db.Availability, error) {
	m.called = true
	m.gotProviderID = providerID
	return m.returnSlots, m.returnError
}

func TestListAvailabilityByProviderHandler(t *testing.T) {
	slotTime := time.Now().Truncate(time.Hour)
	sampleSlot := db.Availability{
		ID:         uuid.New(),
		ProviderID: uuid.New(),
		StartTime:  slotTime,
		EndTime:    slotTime.Add(time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	tests := []struct {
		name         string
		url          string
		vars         map[string]string
		mockSlots    []db.Availability
		mockError    error
		wantStatus   int
		wantContains string
	}{
		{
			name:       "Success",
			url:        "/providers/" + sampleSlot.ProviderID.String() + "/availability",
			vars:       map[string]string{"provider_id": sampleSlot.ProviderID.String()},
			mockSlots:  []db.Availability{sampleSlot},
			mockError:  nil,
			wantStatus: http.StatusOK,
		},
		{
			name:         "Missing provider_id",
			url:          "/providers//availability",
			vars:         map[string]string{},
			wantStatus:   http.StatusBadRequest,
			wantContains: "Missing provider_id",
		},
		{
			name:         "Invalid provider_id",
			url:          "/providers/not-a-uuid/availability",
			vars:         map[string]string{"provider_id": "not-a-uuid"},
			wantStatus:   http.StatusBadRequest,
			wantContains: "Invalid provider_id",
		},
		{
			name:         "DB error",
			url:          "/providers/" + sampleSlot.ProviderID.String() + "/availability",
			vars:         map[string]string{"provider_id": sampleSlot.ProviderID.String()},
			mockSlots:    nil,
			mockError:    errors.New("boom"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Unable to retrieve availability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req = mux.SetURLVars(req, tt.vars)

			mock := &mockProviderLister{
				returnSlots: tt.mockSlots,
				returnError: tt.mockError,
			}
			handler := ListAvailabilityByProviderHandler(mock)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.wantContains != "" {
				if !strings.Contains(rr.Body.String(), tt.wantContains) {
					t.Errorf("expected body to contain %q, got %q", tt.wantContains, rr.Body.String())
				}
				return
			}

			decoder := json.NewDecoder(rr.Body)
			var got []AvailabilityResponse
			err := decoder.Decode(&got)
			if err != nil {
				t.Fatalf("failed to decode JSON: %v", err)
			}
			if len(got) != len(tt.mockSlots) {
				t.Fatalf("expected %d items, got %d", len(tt.mockSlots), len(got))
			}
			if got[0].ID != sampleSlot.ID ||
				!got[0].StartTime.Equal(sampleSlot.StartTime) ||
				!got[0].EndTime.Equal(sampleSlot.EndTime) {
				t.Errorf("unexpected slot returned: %+v", got[0])
			}

		})
	}
}
