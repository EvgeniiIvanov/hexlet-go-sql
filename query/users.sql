-- name: CreateUser :one
INSERT INTO users (email, name, age)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;

-- name: UpdateUser :exec
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    age = COALESCE(sqlc.narg('age'), age)
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;

-- name: UpsertUser :exec
INSERT INTO users (email, name, age)
VALUES (?, ?, ?)
ON CONFLICT(email) DO UPDATE SET
    name = excluded.name,
    age = excluded.age;

-- name: DeleteUsersByIDs :execrows
DELETE FROM users
WHERE id IN (sqlc.slice('ids'));
