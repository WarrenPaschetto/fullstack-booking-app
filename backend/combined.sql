PRAGMA foreign_keys = ON;

-- clean up
DROP TRIGGER IF EXISTS check_admin;
DROP TRIGGER IF EXISTS no_overlap;

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS availability_pattern;
DROP TABLE IF EXISTS availability;
DROP TABLE IF EXISTS users;

-- users
CREATE TABLE users (
    id UUID PRIMARY KEY NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    updated_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user'
);

-- availability
CREATE TABLE availability (
  id             UUID    PRIMARY KEY NOT NULL,
  provider_id    UUID    NOT NULL,
  start_time     TIMESTAMP NOT NULL,
  end_time       TIMESTAMP NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  updated_at     TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY(provider_id) REFERENCES users(id) ON DELETE CASCADE
);

-- availability patterns
CREATE TABLE availability_pattern (
  id             UUID    PRIMARY KEY NOT NULL,
  provider_id    UUID    NOT NULL REFERENCES users(id),
  day_of_week INTEGER NOT NULL CHECK(day_of_week BETWEEN 0 AND 6),
  start_time     TIME NOT NULL,
  end_time       TIME NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT (datetime('now')),
  updated_at     TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);

-- bookings
CREATE TABLE bookings (
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    updated_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    appointment_start TIMESTAMP NOT NULL,
    duration_minutes INTEGER NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot_id UUID REFERENCES availability(id) ON DELETE CASCADE
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
