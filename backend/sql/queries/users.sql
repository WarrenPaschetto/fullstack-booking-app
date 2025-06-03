-- name: CreateUser :exec
INSERT INTO users (id, first_name, last_name, created_at, updated_at, email, password_hash, user_role)
VALUES ( 
    $1,
    $2,
    $3,
    now(), 
    now(), 
    $4,
    $5,
    $6
);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :exec
UPDATE users 
SET first_name = $1, last_name = $2, email = $3, password_hash = $4, updated_at = now()
WHERE id = $5;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY last_name ASC, first_name ASC;