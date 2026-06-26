-- name: CreateOrder :one
INSERT INTO orders (user_id, total_amount, status, payment_method)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, course_id, price)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = ?, completed_at = CASE WHEN ? = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END
WHERE id = ?;

-- name: CompleteOrder :exec
UPDATE orders
SET status = 'completed', completed_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: FailOrder :exec
UPDATE orders
SET status = 'failed'
WHERE id = ?;

-- name: RefundOrder :exec
UPDATE orders
SET status = 'refunded'
WHERE id = ?;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = ?;
