-- name: CreateUser :exec
INSERT INTO users (id, first_name, last_name, created_at, updated_at, email, password_hash, role)
VALUES (
    uuid(), 
    ?,
    ?,
    CURRENT_TIMESTAMP, 
    CURRENT_TIMESTAMP, 
    ?,
    ?,
    ?
);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: UpdateUser :exec
UPDATE users 
SET first_name = ?, last_name = ?, email = ?, password_hash = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT * FROM users ORDER BY last_name ASC, first_name ASC;