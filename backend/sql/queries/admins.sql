-- name: CreateAdmin :exec
INSERT INTO admins (id, first_name, last_name, created_at, updated_at, email, password_hash)
VALUES (
    uuid(), 
    ?,
    ?,
    CURRENT_TIMESTAMP, 
    CURRENT_TIMESTAMP, 
    ?,
    ?
);

-- name: GetAdminByEmail :one
SELECT * FROM admins
WHERE email = ?;

-- name: GetAdminByName :many
SELECT * FROM admins
WHERE LOWER(first_name) = LOWER(?) AND LOWER(last_name) = LOWER(?);

-- name: UpdateAdmin :exec
UPDATE admins 
SET first_name = ?, last_name = ?, email = ?, password_hash = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteAdmin :exec
DELETE FROM admins WHERE id = ?;

-- name: ListAdmins :many
SELECT * FROM admins ORDER BY last_name ASC, first_name ASC;