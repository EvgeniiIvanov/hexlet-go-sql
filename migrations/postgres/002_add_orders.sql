-- Migration 002: Add order items and indexes
-- Note: The orders table is already created in 001_schema.sql
-- This migration adds order_items table and performance indexes

-- Order items: what was purchased in each order (PostgreSQL)
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL,
    course_id INTEGER NOT NULL,
    price INTEGER NOT NULL,  -- price at time of purchase (in cents)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses(id),
    UNIQUE(order_id, course_id)  -- prevent duplicate course in same order
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_course_id ON order_items(course_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_order_id ON enrollments(order_id);
