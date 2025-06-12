-- name: CreateUser :one
INSERT INTO users (
    id,
    created_at,
    updated_at,
    email
) VALUES (
    gen_random_uuid(), 
    NOW(), 
    NOW(), 
    $1
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUsers :many
SELECT * FROM users;

-- name: DeleteAllUsers :exec
DELETE FROM users;