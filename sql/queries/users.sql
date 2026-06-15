-- name: CreateUser :one
INSERT INTO users(
    id,
    created_at,
    updated_at,
    email,
    hashed_password
)
VALUES(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET email = $2, hashed_password = $3, updated_at = $4
WHERE id = $1
RETURNING id, email, created_at, updated_at;

-- name: SearchUser :one
SELECT * from users
WHERE email = $1;

-- name: DeleteUsers :exec
TRUNCATE TABLE users CASCADE;
