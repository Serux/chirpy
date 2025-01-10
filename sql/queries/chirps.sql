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
SELECT * FROM chirps ORDER BY chirps.created_at;

-- name: SelectOneChirps :one
SELECT * FROM chirps WHERE chirps.id = $1;

-- name: DeleteByIdChirps :exec
DELETE FROM chirps
WHERE chirps.id = $1
AND chirps.user_id = $2
;

-- name: DeleteAllChirps :exec
DELETE FROM chirps;