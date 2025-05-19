package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/db"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/middleware"
	"github.com/WarrenPaschetto/fullstack-booking-app/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockBookingQueries struct {
	CreateBookingFn           func(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error)
	GetOverlappingBookingsFn  func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error)
	DeleteBookingFn           func(ctx context.Context, arg db.DeleteBookingParams) error
	RescheduleBookingFn       func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
	GetBookingByIDFn          func(ctx context.Context, bookingID uuid.UUID) (db.Booking, error)
	ListBookingsForUserFn     func(ctx context.Context, id uuid.UUID) ([]db.Booking, error)
	ListAllBookingsForAdminFn func(ctx context.Context) ([]db.Booking, error)
}

func (m *mockBookingQueries) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {
	return m.CreateBookingFn(ctx, arg)
}

func (m *mockBookingQueries) GetOverlappingBookings(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
	return m.GetOverlappingBookingsFn(ctx, arg)
}

func (m *mockBookingQueries) DeleteBooking(ctx context.Context, arg db.DeleteBookingParams) error {
	return m.DeleteBookingFn(ctx, arg)
}

func (m *mockBookingQueries) RescheduleBooking(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
	return m.RescheduleBookingFn(ctx, arg)
}

func (m *mockBookingQueries) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (db.Booking, error) {
	return m.GetBookingByIDFn(ctx, bookingID)
}
func (m *mockBookingQueries) ListBookingsForUser(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
	return m.ListBookingsForUserFn(ctx, id)
}
func (m *mockBookingQueries) ListAllBookingsForAdmin(ctx context.Context) ([]db.Booking, error) {
	return m.ListAllBookingsForAdminFn(ctx)
}

func TestCreateBookingHandler(t *testing.T) {
	userID := uuid.New()
	validBody := BookingRequest{
		AppointmentStart: time.Now().Add(time.Hour),
		DurationMinutes:  60,
	}
	jsonBody, _ := json.Marshal(validBody)

	tests := []struct {
		name         string
		ctxUserID    any
		body         []byte
		mockOverlap  func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error)
		mockCreate   func(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error)
		expectStatus int
	}{
		{
			name:      "Valid booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockOverlap: func(_ context.Context, _ db.GetOverlappingBookingsParams) ([]db.Booking, error) {
				return nil, nil
			},
			mockCreate: func(_ context.Context, _ db.CreateBookingParams) (db.Booking, error) {
				return db.Booking{ID: uuid.New()}, nil
			},
			expectStatus: http.StatusCreated,
		},
		{
			name:         "Missing auth context",
			ctxUserID:    nil,
			body:         jsonBody,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "Malformed request body",
			ctxUserID:    userID,
			body:         []byte(`{invalid json`),
			expectStatus: http.StatusBadRequest,
		},
		{
			name:      "Overlapping booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockOverlap: func(_ context.Context, _ db.GetOverlappingBookingsParams) ([]db.Booking, error) {
				return []db.Booking{{ID: uuid.New()}}, nil
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockBookingQueries{
				CreateBookingFn:          tt.mockCreate,
				GetOverlappingBookingsFn: tt.mockOverlap,
			}

			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.CreateBookingHandler()

			req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(tt.body))
			if tt.ctxUserID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID))
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
		})
	}
}

func TestDeleteBookingHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	validBody := DeleteBookingRequest{
		ID: bookingID,
	}
	jsonBody, _ := json.Marshal(validBody)

	invalidBody := DeleteBookingRequest{
		ID: uuid.Nil,
	}
	invalidJsonBody, _ := json.Marshal(invalidBody)

	tests := []struct {
		name         string
		ctxUserID    any
		body         []byte
		mockDelete   func(ctx context.Context, arg db.DeleteBookingParams) error
		expectStatus int
	}{
		{
			name:      "Delete booking",
			ctxUserID: userID,
			body:      jsonBody,
			mockDelete: func(_ context.Context, arg db.DeleteBookingParams) error {
				if arg.ID != bookingID || arg.UserID != userID {
					t.Errorf("expected ID=%v and UserID=%v, got ID=%v and UserID=%v", bookingID, userID, arg.ID, arg.UserID)
				}
				return nil
			},
			expectStatus: http.StatusNoContent,
		},
		{
			name:      "DB error",
			ctxUserID: userID,
			body:      jsonBody,
			mockDelete: func(_ context.Context, _ db.DeleteBookingParams) error {
				return errors.New("simulated DB error")
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:         "No user ID",
			ctxUserID:    nil,
			body:         jsonBody,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "No booking ID",
			ctxUserID:    userID,
			body:         invalidJsonBody,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Malformed request body",
			ctxUserID:    userID,
			body:         []byte(`{invalid json`),
			mockDelete:   nil,
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockBookingQueries{
				DeleteBookingFn: tt.mockDelete,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.DeleteBookingHandler()

			req := httptest.NewRequest(http.MethodDelete, "/bookings", bytes.NewReader(tt.body))
			if tt.ctxUserID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID))
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
		})
	}
}

func TestRescheduleBookingHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	reqBody := RescheduleBookingRequest{
		AppointmentStart: now.Add(time.Hour),
		DurationMinutes:  30,
		ID:               bookingID,
	}
	jsonBody, _ := json.Marshal(reqBody)

	fakeBooking := db.Booking{
		ID:               bookingID,
		UserID:           userID,
		AppointmentStart: now.Add(time.Hour),
		DurationMinutes:  30,
		CreatedAt:        now.Add(-time.Hour),
		UpdatedAt:        now.Add(-time.Minute),
	}

	tests := []struct {
		name           string
		routeID        string
		body           []byte
		mockReschedule func(ctx context.Context, arg db.RescheduleBookingParams) (db.Booking, error)
		expectStatus   int
	}{
		{
			name:    "Rescheduled booking success",
			routeID: bookingID.String(),
			body:    jsonBody,
			mockReschedule: func(_ context.Context, arg db.RescheduleBookingParams) (db.Booking, error) {
				if arg.ID != bookingID {
					t.Errorf("expected ID %s, got %s", bookingID, arg.ID)
				}
				if !arg.AppointmentStart.Equal(fakeBooking.AppointmentStart) {
					t.Errorf("expected start %s, got %s", fakeBooking.AppointmentStart, arg.AppointmentStart)
				}
				return fakeBooking, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "Bad UUID",
			routeID:      "not-a-uuid",
			body:         jsonBody,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Malformed JSON",
			routeID:      bookingID.String(),
			body:         []byte(`{invalid json`),
			expectStatus: http.StatusBadRequest,
		},
		{
			name:    "Booking conflict",
			routeID: bookingID.String(),
			body:    jsonBody,
			mockReschedule: func(_ context.Context, _ db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, service.ErrBookingConflict
			},
			expectStatus: http.StatusConflict,
		},
		{
			name:    "Internal error",
			routeID: bookingID.String(),
			body:    jsonBody,
			mockReschedule: func(_ context.Context, _ db.RescheduleBookingParams) (db.Booking, error) {
				return db.Booking{}, errors.New("boom")
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQ := &mockBookingQueries{
				GetBookingByIDFn: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
					return fakeBooking, nil
				},
				GetOverlappingBookingsFn: func(ctx context.Context, arg db.GetOverlappingBookingsParams) ([]db.Booking, error) {
					return nil, nil
				},
				RescheduleBookingFn: tt.mockReschedule,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.RescheduleBookingHandler()

			req := httptest.NewRequest(http.MethodPatch, "/bookings/"+tt.routeID, bytes.NewReader(tt.body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.routeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectStatus, rr.Code)
			}
		})
	}
}

func TestGetBookingByIDHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	now := time.Now()

	fakeBooking := db.Booking{
		ID:               bookingID,
		UserID:           userID,
		AppointmentStart: now,
		DurationMinutes:  30,
		CreatedAt:        now.Add(-time.Hour),
		UpdatedAt:        now.Add(-time.Minute),
	}

	tests := []struct {
		name           string
		routeID        string
		ctxUserID      interface{}
		mockGet        func(ctx context.Context, id uuid.UUID) (db.Booking, error)
		expectStatus   int
		expectResponse *db.Booking
	}{
		{

			name:      "Success",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				if id != bookingID {
					t.Errorf("expected GetBookingByID called with %s, got %s", bookingID, id)
				}
				return fakeBooking, nil
			},
			expectStatus:   http.StatusOK,
			expectResponse: &fakeBooking,
		},
		{
			name:      "Not found",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{}, sql.ErrNoRows
			},
			expectStatus: http.StatusNotFound,
		},
		{
			name:      "Forbidden",
			routeID:   bookingID.String(),
			ctxUserID: uuid.New(),
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return fakeBooking, nil
			},
			expectStatus: http.StatusForbidden,
		},
		{
			name:      "DB error",
			routeID:   bookingID.String(),
			ctxUserID: userID,
			mockGet: func(ctx context.Context, id uuid.UUID) (db.Booking, error) {
				return db.Booking{}, errors.New("some db failure")
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:         "Bad id param",
			routeID:      "not-a-uuid",
			ctxUserID:    userID,
			mockGet:      nil,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "No auth",
			routeID:      bookingID.String(),
			ctxUserID:    nil,
			mockGet:      nil,
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{
				GetBookingByIDFn: tt.mockGet,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.GetBookingByIDHandler()

			req := httptest.NewRequest(http.MethodGet, "/bookings/"+tt.routeID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.routeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if tt.ctxUserID != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID),
				)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Fatalf("expected status %d, got %d; body=%s", tt.expectStatus, rr.Code, rr.Body.String())
			}

			if tt.expectStatus == http.StatusOK {
				var got db.Booking
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				if got.ID != fakeBooking.ID || got.UserID != fakeBooking.UserID {
					t.Errorf("unexpected booking returned: %+v", got)
				}
			}
		})
	}
}

func TestListBookingsForUserHandler(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()
	bookingID2 := uuid.New()
	bookingID3 := uuid.New()
	now := time.Now()

	fakeBookings := []db.Booking{
		{
			ID:               bookingID,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 24),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 336),
			UpdatedAt:        now.Add(-time.Minute * 336),
		},
		{
			ID:               bookingID2,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 84),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 168),
			UpdatedAt:        now.Add(-time.Minute * 168),
		},
		{
			ID:               bookingID3,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 168),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 48),
			UpdatedAt:        now.Add(-time.Minute * 48),
		},
	}

	tests := []struct {
		name           string
		ctxUserID      interface{}
		mockList       func(ctx context.Context, id uuid.UUID) ([]db.Booking, error)
		expectStatus   int
		expectResponse []db.Booking
	}{
		{

			name:      "Success",
			ctxUserID: userID,
			mockList: func(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
				if id != userID {
					t.Errorf("ListBookingsForUser called with wrong userID: got %v, want %v", id, userID)
				}
				return fakeBookings, nil
			},
			expectStatus:   http.StatusOK,
			expectResponse: fakeBookings,
		},
		{

			name:      "User ID missing",
			ctxUserID: nil,
			mockList: func(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
				return fakeBookings, nil
			},
			expectStatus:   http.StatusUnauthorized,
			expectResponse: nil,
		},
		{

			name:      "User not found",
			ctxUserID: uuid.New(),
			mockList: func(ctx context.Context, id uuid.UUID) ([]db.Booking, error) {
				return fakeBookings, nil
			},
			expectStatus:   http.StatusInternalServerError,
			expectResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{
				ListBookingsForUserFn: tt.mockList,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.ListBookingsForUserHandler()

			req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
			if tt.ctxUserID != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID),
				)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Fatalf("expected status %d, got %d; body=%s", tt.expectStatus, rr.Code, rr.Body.String())
			}

			if tt.expectStatus == http.StatusOK {
				var got []db.Booking
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				for index, booking := range got {
					if got[index].UserID != booking.UserID {
						t.Errorf("unexpected user booking returned: %+v", got)
					}
				}
			}
		})
	}
}

func TestListAllBookingsForAdminHandler(t *testing.T) {
	now := time.Date(2025, 5, 14, 10, 0, 0, 0, time.UTC)
	bookingID := uuid.New()
	bookingID2 := uuid.New()
	bookingID3 := uuid.New()
	userID := uuid.New()
	anotherUser := uuid.New()
	otherErr := errors.New("db fail")

	fakeBookings := []db.Booking{
		{
			ID:               bookingID,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 24),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 336),
			UpdatedAt:        now.Add(-time.Minute * 336),
		},
		{
			ID:               bookingID2,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 84),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 168),
			UpdatedAt:        now.Add(-time.Minute * 168),
		},
		{
			ID:               bookingID3,
			UserID:           userID,
			AppointmentStart: now.Add(time.Hour * 168),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 48),
			UpdatedAt:        now.Add(-time.Minute * 48),
		},
		{
			ID:               bookingID,
			UserID:           anotherUser,
			AppointmentStart: now.Add(time.Hour * 2),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 336),
			UpdatedAt:        now.Add(-time.Minute * 336),
		},
		{
			ID:               bookingID2,
			UserID:           anotherUser,
			AppointmentStart: now.Add(time.Hour * 50),
			DurationMinutes:  30,
			CreatedAt:        now.Add(-time.Hour * 168),
			UpdatedAt:        now.Add(-time.Minute * 168),
		},
	}

	tests := []struct {
		name           string
		mockList       func(ctx context.Context) ([]db.Booking, error)
		expectStatus   int
		expectResponse []db.Booking
	}{
		{

			name: "Success",
			mockList: func(ctx context.Context) ([]db.Booking, error) {
				return fakeBookings, nil
			},
			expectStatus:   http.StatusOK,
			expectResponse: fakeBookings,
		},
		{

			name: "DB error",
			mockList: func(ctx context.Context) ([]db.Booking, error) {
				return []db.Booking{}, otherErr
			},
			expectStatus:   http.StatusInternalServerError,
			expectResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockQ := &mockBookingQueries{
				ListAllBookingsForAdminFn: tt.mockList,
			}
			bookingSvc := service.NewBookingService(mockQ)

			h := &Handler{BookingService: bookingSvc}
			handler := h.ListAllBookingsForAdminHandler()

			req := httptest.NewRequest(http.MethodGet, "/bookings", nil)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Fatalf("expected status %d, got %d; body=%s", tt.expectStatus, rr.Code, rr.Body.String())
			}

			if tt.expectStatus == http.StatusOK {
				var got []db.Booking
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				for index, booking := range got {
					if got[index].UserID != booking.UserID {
						t.Errorf("unexpected user booking returned: %+v", got)
					}
				}
			}
		})
	}
}
