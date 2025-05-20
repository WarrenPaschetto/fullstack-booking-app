package handlers

import (
	"testing"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
)

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name string
		svc  *service.BookingService
	}{
		{
			name: "Nil service",
			svc:  nil,
		},
		{
			name: "Real service",
			svc:  service.NewBookingService(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.svc)
			if h == nil {
				t.Fatal("got nil handler")
			}
			if h.BookingService != tt.svc {
				t.Errorf("got %v; want %v", h.BookingService, tt.svc)
			}
		})
	}
}
