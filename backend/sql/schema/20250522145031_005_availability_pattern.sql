-- +goose Up
CREATE TABLE availability_pattern (
  id             UUID    PRIMARY KEY NOT NULL,
  provider_id    UUID    NOT NULL REFERENCES users(id),
  day_of_week INTEGER NOT NULL CHECK(day_of_week BETWEEN 0 AND 6),
  start_time     TIME NOT NULL,
  end_time       TIME NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  updated_at     TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS availability_pattern;