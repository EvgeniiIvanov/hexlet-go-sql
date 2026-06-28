-- name: CreateCourse :one
INSERT INTO courses (slug, title, price)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateCourse :exec
UPDATE courses
SET title = COALESCE(sqlc.narg('title'), title),
    price = COALESCE(sqlc.narg('price'), price)
WHERE id = ?;

-- name: DeleteCourse :exec
DELETE FROM courses
WHERE id = ?;

-- name: UpsertCourse :exec
INSERT INTO courses (slug, title, price)
VALUES (?, ?, ?)
ON CONFLICT(slug) DO UPDATE SET
    title = excluded.title,
    price = excluded.price;
