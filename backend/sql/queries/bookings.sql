-- name: CreateBooking :exec
INSERT INTO bookings (id, created_at, updated_at, appointment_start, duration_minutes, user_id)
VALUES (
    uuid(),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    ?,
    ?,
    ?
);
-- name: DeleteBooking :exec
DELETE FROM bookings 
WHERE id = ? AND user_id = ?;

-- name: ListBookingsForUser :many
SELECT * FROM bookings
WHERE user_id = ?
ORDER BY appointment_start;

-- name: ListAllBookingsForAdmin :many
SELECT * From bookings
ORDER BY appointment_start;

-- name: GetOverlappingBookings :many
SELECT * FROM bookings
WHERE appointment_start < DATETIME(?, '+' || ? || ' minutes')
AND DATETIME(appointment_start, '+' || duration_minutes || ' minutes') > ?;

-- name: GetBookingByID :one
SELECT * FROM bookings
WHERE id = ?;