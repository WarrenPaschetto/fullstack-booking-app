package handlers

import "github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"

type Handler struct {
	BookingService *service.BookingService
}

func NewHandler(bs *service.BookingService) *Handler {
	return &Handler{BookingService: bs}
}
