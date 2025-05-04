-- name: CreateUser :exec
INSERT INTO users (id, created_at, updated_at, email, password_hash)
VALUES (
    uuid(), 
    CURRENT_TIMESTAMP, 
    CURRENT_TIMESTAMP, 
    ?,
    ?
);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: UpdateUser :exec
UPDATE users 
SET email = ?, password_hash = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;