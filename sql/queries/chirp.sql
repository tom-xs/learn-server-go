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

-- name: GetChirpsFrom :many
SELECT * from chirp
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetAllChirps :many
SELECT * from chirp ORDER BY created_at ASC;

-- name: DeleteChirp :exec
DELETE from chirp
WHERE id = $1;

-- name: DeleteAllChirps :exec
DELETE FROM chirp;
