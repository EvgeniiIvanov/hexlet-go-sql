-- name: GetCourse :one
SELECT * FROM courses
WHERE id = ?;

-- name: GetCourseBySlug :one
SELECT * FROM courses
WHERE slug = ?;

-- name: ListCourses :many
SELECT id, slug, title, price
FROM courses
ORDER BY id
LIMIT ? OFFSET ?;

-- name: FindCoursesByIDs :many
SELECT * FROM courses
WHERE id IN (sqlc.slice('ids'));

-- name: CountCourses :one
SELECT COUNT(*) FROM courses;
