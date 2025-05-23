-- name: CreateAvailability :exec
INSERT INTO availability (id, provider_id, start_time, end_time)
VALUES (
    ?,
    ?,
    ?,
    ?
);

-- name: DeleteAvailability :exec
DELETE FROM availability WHERE id = ?
AND provider_id = ?;

-- name: ListAvailabilityByProvider :many
SELECT
  id,
  provider_id,
  start_time,
  end_time,
  created_at,
  updated_at
FROM availability
WHERE provider_id = ? ORDER BY start_time;

-- name: ListAvailabilityInRange :many
SELECT
  id,
  provider_id,
  start_time,
  end_time
FROM availability
WHERE provider_id = ?
AND start_time >= ?
AND end_time <= ?
ORDER BY start_time;

-- name: ListAllFreeSlots :many
SELECT
s.id,
s.start_time,
s.end_time
FROM availability AS s
LEFT JOIN bookings AS b
  ON b.slot_id = s.id
WHERE b.slot_id is NULL
  AND s.provider_id = ?
  AND s.start_time >= ?
  AND s.end_time <= ?
ORDER BY s.start_time;


