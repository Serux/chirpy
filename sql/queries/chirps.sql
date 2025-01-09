-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body,user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: SelectAllChirps :many
Select * FROM chirps ORDER BY chirps.created_at;

-- name: DeleteAllChirps :exec

DELETE FROM chirps;