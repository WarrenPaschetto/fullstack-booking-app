-- name: CreateAvailabilityPattern :exec
INSERT INTO availability_pattern (id, provider_id, day_of_week, start_time, end_time)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateAvailabilityPattern :exec
UPDATE availability_pattern
SET 
  day_of_week = ?,
  start_time = ?,
  end_time = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteAvailabilityPattern :exec
DELETE FROM availability_pattern
WHERE id = ?;

-- name: ListPatternsByProvider :many
SELECT
  id,
  day_of_week,
  start_time,
  end_time,
  created_at,
  updated_at
FROM availability_pattern
WHERE provider_id = ?
ORDER BY day_of_week, start_time;