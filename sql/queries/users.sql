-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: SelectUserByMail :one
SELECT id,created_at,updated_at,email, hashed_password
FROM users
WHERE users.email = $1;

-- name: UpdateUserMailPassByUUID :one

UPDATE users
SET 
email = $1, 
hashed_password = $2,
updated_at = NOW()
WHERE id = $3
RETURNING *;


-- name: DeleteAllUsers :exec

DELETE FROM users;