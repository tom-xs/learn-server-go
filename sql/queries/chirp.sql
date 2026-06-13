-- name: CreateChirp :one
INSERT INTO chirp(
    id,
    created_at,
    updated_at,
    body,
    user_id
)
VALUES(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetChirp :one
SELECT * from chirp
WHERE id = $1;

-- name: GetAllChirps :many
SELECT * from chirp ORDER BY created_at;

-- name: DeleteChirps :exec
DELETE FROM chirp;
