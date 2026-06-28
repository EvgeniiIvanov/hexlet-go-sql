-- PostgreSQL-specific queries for orders
-- Uses json_build_object and json_agg instead of json_object and json_group_array

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
        COALESCE(
            json_agg(
                json_build_object(
                    'id', oi.id,
                    'course_id', oi.course_id,
                    'course_slug', COALESCE(c.slug, ''),
                    'course_title', COALESCE(c.title, ''),
                    'price', oi.price
                )
            ) FILTER (WHERE oi.id IS NOT NULL),
            '[]'::json
        ) as items
    FROM orders o
    LEFT JOIN order_items oi ON oi.order_id = o.id
    LEFT JOIN courses c ON c.id = oi.course_id
    WHERE o.user_id = $1
    GROUP BY o.id, o.user_id, o.total_amount, o.status, o.payment_method, o.created_at, o.completed_at
)
SELECT * FROM order_data
ORDER BY created_at DESC;

-- name: GetOrderWithItems :one
WITH items_data AS (
    SELECT 
        oi.order_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'id', oi.id,
                    'course_id', oi.course_id,
                    'course_slug', COALESCE(c.slug, ''),
                    'course_title', COALESCE(c.title, ''),
                    'price', oi.price,
                    'created_at', oi.created_at
                )
            ) FILTER (WHERE oi.id IS NOT NULL),
            '[]'::json
        ) as items
    FROM order_items oi
    LEFT JOIN courses c ON c.id = oi.course_id
    WHERE oi.order_id = $1
)
SELECT 
    o.id,
    o.user_id,
    o.total_amount,
    o.status,
    o.payment_method,
    o.created_at,
    o.completed_at,
    COALESCE(i.items, '[]'::json) as items
FROM orders o
LEFT JOIN items_data i ON i.order_id = o.id
WHERE o.id = $2;
