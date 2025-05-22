-- +goose Up
CREATE TABLE bookings (
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    updated_at TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    appointment_start TIMESTAMP NOT NULL,
    duration_minutes INTEGER NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot_id UUID REFERENCES availability(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE bookings;
