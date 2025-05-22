-- +goose Up

CREATE TABLE availability (
  id             UUID    PRIMARY KEY NOT NULL,
  provider_id    UUID    NOT NULL,
  start_time     TIMESTAMP NOT NULL,
  end_time       TIMESTAMP NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  updated_at     TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY(provider_id) REFERENCES users(id) ON DELETE CASCADE
);

-- only admins may insert availability
CREATE TRIGGER check_admin
  BEFORE INSERT ON availability
  FOR EACH ROW
  WHEN (SELECT role FROM users WHERE id = NEW.provider_id) != 'admin'
BEGIN
  SELECT RAISE(ABORT, 'provider is not an admin');
END;

-- prevent overlapping slots per provider
CREATE TRIGGER no_overlap
  BEFORE INSERT ON availability
  FOR EACH ROW
  WHEN EXISTS (
    SELECT 1
      FROM availability
     WHERE provider_id = NEW.provider_id
       AND NOT (
         NEW.end_time   <= start_time
         OR NEW.start_time >= end_time
       )
  )
BEGIN
  SELECT RAISE(ABORT, 'time slot overlaps existing one');
END;
-- +goose Down

DROP TRIGGER IF EXISTS check_admin;
DROP TRIGGER IF EXISTS no_overlap;

DROP TABLE IF EXISTS availability;
