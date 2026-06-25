-- name: CreateCourse :one
INSERT INTO courses (slug, title, price)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetCourse :one
SELECT * FROM courses
WHERE id = ?;

-- name: GetCourseBySlug :one
SELECT * FROM courses
WHERE slug = ?;

-- name: ListCourses :many
SELECT * FROM courses
ORDER BY id;

-- name: FindCoursesByIDs :many
SELECT * FROM courses
WHERE id IN (sqlc.slice('ids'));

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
