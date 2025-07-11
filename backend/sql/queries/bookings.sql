-- name: CreateBooking :one
INSERT INTO bookings (id, created_at, updated_at, appointment_start, duration_minutes, user_id, slot_id)
VALUES (
    $1,
    now(),
    now(),
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings 
WHERE id = $1 
AND (user_id = $2 or $3::boolean);

-- name: RescheduleBooking :one
UPDATE bookings
SET appointment_start = $2,
    duration_minutes = $3,
    updated_at = now()
WHERE id = $1
AND (user_id = $4 or $5::boolean)
RETURNING *;

-- name: ListBookingsForUser :many
SELECT * FROM bookings
WHERE user_id = $1
ORDER BY appointment_start;

-- name: ListAllBookingsForAdmin :many
SELECT * From bookings
ORDER BY appointment_start;

-- name: GetOverlappingBookings :many
SELECT *
FROM bookings
WHERE 
  appointment_start < $1
  AND appointment_start + (duration_minutes || ' minutes')::interval > $2;

-- name: GetBookingByID :one
SELECT * FROM bookings
WHERE id = $1;