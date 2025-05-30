-- name: CreateAvailability :exec
INSERT INTO availability (id, provider_id, start_time, end_time)
VALUES (
    $1,
    $2,
    $3,
    $4
);

-- name: DeleteAvailability :exec
DELETE FROM availability WHERE id = $1
AND provider_id = $2;

-- name: ListAvailabilityByProvider :many
SELECT
  id,
  provider_id,
  start_time,
  end_time,
  created_at,
  updated_at
FROM availability
WHERE provider_id = $1 ORDER BY start_time;

-- name: ListAvailabilityInRange :many
SELECT
  id,
  provider_id,
  start_time,
  end_time
FROM availability
WHERE provider_id = $1
AND start_time >= $2
AND end_time <= $3
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
  AND s.provider_id = $1
  AND s.start_time >= $2
  AND s.end_time <= $3
ORDER BY s.start_time;


