-- name: GetOrder :one
SELECT * FROM orders WHERE id = ?;

-- name: ListOrders :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: GetUserOrders :many
WITH order_data AS (
    SELECT 
        o.id,
        o.user_id,
        o.total_amount,
        o.status,
        o.payment_method,
        o.created_at,
        o.completed_at,
        json_group_array(json_object(
            'id', oi.id,
            'course_id', oi.course_id,
            'course_slug', COALESCE(c.slug, ''),
            'course_title', COALESCE(c.title, ''),
            'price', oi.price
        )) as items
    FROM orders o
    LEFT JOIN order_items oi ON oi.order_id = o.id
    LEFT JOIN courses c ON c.id = oi.course_id
    WHERE o.user_id = ?
    GROUP BY o.id
)
SELECT * FROM order_data
ORDER BY created_at DESC;

-- name: GetOrderWithItems :one
WITH items_data AS (
    SELECT 
        oi.order_id,
        json_group_array(json_object(
            'id', oi.id,
            'course_id', oi.course_id,
            'course_slug', COALESCE(c.slug, ''),
            'course_title', COALESCE(c.title, ''),
            'price', oi.price,
            'created_at', oi.created_at
        )) as items
    FROM order_items oi
    LEFT JOIN courses c ON c.id = oi.course_id
    WHERE oi.order_id = ?
)
SELECT 
    o.id,
    o.user_id,
    o.total_amount,
    o.status,
    o.payment_method,
    o.created_at,
    o.completed_at,
    COALESCE(i.items, '[]') as items
FROM orders o
LEFT JOIN items_data i ON i.order_id = o.id
WHERE o.id = ?;

-- name: CheckUserOwnsCourse :one
-- Check if user owns a course by looking at enrollments
-- This includes both purchased courses (with order_id) and free courses (without order_id)
SELECT EXISTS(
    SELECT 1
    FROM enrollments e
    WHERE e.user_id = ?
      AND e.course_id = ?
      AND e.status = 'active'
) as owns_course;

-- name: GetCourseRevenue :one
SELECT 
    c.id,
    c.slug,
    c.title,
    COUNT(oi.id) as sales_count,
    COALESCE(SUM(oi.price), 0) as total_revenue
FROM courses c
LEFT JOIN order_items oi ON oi.course_id = c.id
LEFT JOIN orders o ON o.id = oi.order_id AND o.status = 'completed'
WHERE c.id = ?
GROUP BY c.id;
