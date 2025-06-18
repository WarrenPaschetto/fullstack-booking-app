-- +goose Up

CREATE TABLE IF NOT EXISTS availability (
  id            UUID PRIMARY KEY NOT NULL,
  provider_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  start_time    TIMESTAMP NOT NULL,
  end_time      TIMESTAMP NOT NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT now(),
  updated_at    TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_availability_provider_is_admin()
  RETURNS trigger AS $check_avail$
BEGIN
  IF (SELECT user_role FROM users WHERE id = NEW.provider_id) IS DISTINCT FROM 'admin' THEN
    RAISE EXCEPTION 'provider is not an admin';
  END IF;
  RETURN NEW;
END;
$check_avail$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER check_admin
  BEFORE INSERT ON availability
  FOR EACH ROW
  EXECUTE FUNCTION check_availability_provider_is_admin();

-- +goose Down

DROP TRIGGER IF EXISTS check_admin;
DROP TRIGGER IF EXISTS no_overlap;

DROP TABLE IF EXISTS availability;
