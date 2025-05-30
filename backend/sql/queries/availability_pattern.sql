-- name: CreateAvailabilityPattern :exec
INSERT INTO availability_pattern (id, provider_id, day_of_week, start_time, end_time)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdateAvailabilityPattern :exec
UPDATE availability_pattern
SET 
  day_of_week = $1,
  start_time = $2,
  end_time = $3,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $4;

-- name: DeleteAvailabilityPattern :exec
DELETE FROM availability_pattern
WHERE id = $1
AND provider_id = $2;

-- name: ListPatternsByProvider :many
SELECT
  id,
  day_of_week,
  start_time,
  end_time,
  created_at,
  updated_at
FROM availability_pattern
WHERE provider_id = $1
ORDER BY day_of_week, start_time;

-- name: GetAvailabilityPatternByID :one
SELECT id, provider_id, day_of_week, start_time, end_time, created_at, updated_at
FROM availability_pattern
WHERE id = $1;
