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

-- name: GetCourseWithEnrollments :one
WITH enrollment_list AS (
    SELECT
        e.course_id,
        json_group_array(json_object(
            'id', e.id,
            'user_id', e.user_id,
            'user_email', COALESCE(u.email, ''),
            'user_name', u.name,
            'enrolled_at', e.enrolled_at,
            'status', e.status
        )) AS payload
    FROM enrollments e
    LEFT JOIN users u ON e.user_id = u.id
    WHERE e.course_id = ?
)
SELECT
    c.id,
    c.slug,
    c.title,
    c.price,
    COALESCE(el.payload, '[]') AS enrollments
FROM courses c
LEFT JOIN enrollment_list el ON el.course_id = c.id
WHERE c.id = ?;
