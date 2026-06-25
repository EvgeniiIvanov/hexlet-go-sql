-- name: CreateEnrollment :one
INSERT INTO enrollments (user_id, course_id, status)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetEnrollment :one
SELECT * FROM enrollments
WHERE id = ?;

-- name: GetEnrollmentByUserAndCourse :one
SELECT * FROM enrollments
WHERE user_id = ? AND course_id = ?;

-- name: ListEnrollments :many
SELECT 
    e.id,
    e.user_id,
    u.email as user_email,
    u.name as user_name,
    e.course_id,
    c.slug as course_slug,
    c.title as course_title,
    e.enrolled_at,
    e.status
FROM enrollments e
JOIN users u ON e.user_id = u.id
JOIN courses c ON e.course_id = c.id
ORDER BY e.enrolled_at DESC;

-- name: ListEnrollmentsByUser :many
SELECT 
    e.id,
    e.user_id,
    u.email as user_email,
    u.name as user_name,
    e.course_id,
    c.slug as course_slug,
    c.title as course_title,
    e.enrolled_at,
    e.status
FROM enrollments e
JOIN users u ON e.user_id = u.id
JOIN courses c ON e.course_id = c.id
WHERE e.user_id = ?
ORDER BY e.enrolled_at DESC;

-- name: ListEnrollmentsByCourse :many
SELECT 
    e.id,
    e.user_id,
    u.email as user_email,
    u.name as user_name,
    e.course_id,
    c.slug as course_slug,
    c.title as course_title,
    e.enrolled_at,
    e.status
FROM enrollments e
JOIN users u ON e.user_id = u.id
JOIN courses c ON e.course_id = c.id
WHERE e.course_id = ?
ORDER BY e.enrolled_at DESC;

-- name: UpdateEnrollmentStatus :exec
UPDATE enrollments
SET status = ?
WHERE user_id = ? AND course_id = ? AND status = 'active';

-- name: CheckUserExists :one
SELECT 1 FROM users WHERE id = ?;

-- name: CheckCourseExists :one
SELECT 1 FROM courses WHERE id = ?;
