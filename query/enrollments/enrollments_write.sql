-- name: CreateEnrollment :one
INSERT INTO enrollments (user_id, course_id, status)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateEnrollmentStatus :exec
UPDATE enrollments
SET status = ?
WHERE user_id = ? AND course_id = ? AND status = 'active';
