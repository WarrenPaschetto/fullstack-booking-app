-- name: CreateBooking :one
INSERT INTO bookings (id, created_at, updated_at, appointment_start, duration_minutes, user_id)
VALUES (
    ?,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings 
WHERE id = ? AND user_id = ?;

-- name: RescheduleBooking :one
UPDATE bookings
SET appointment_start = ?
WHERE id = ?
RETURNING *;

-- name: ListBookingsForUser :many
SELECT * FROM bookings
WHERE user_id = ?
ORDER BY appointment_start;

-- name: ListAllBookingsForAdmin :many
SELECT * From bookings
ORDER BY appointment_start;

-- name: GetOverlappingBookings :many
SELECT
  id,
  created_at,
  updated_at,
  appointment_start,
  duration_minutes,
  user_id
FROM bookings
WHERE 
  appointment_start < :new_end
  AND DATETIME(appointment_start, '+' || duration_minutes || ' minutes') > :new_start;

-- name: GetBookingByID :one
SELECT * FROM bookings
WHERE id = ?;