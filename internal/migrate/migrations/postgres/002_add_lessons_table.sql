-- +goose Up
CREATE TABLE lessons (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_lessons_course_id ON lessons(course_id);

-- +goose Down
DROP INDEX IF EXISTS idx_lessons_course_id;
DROP TABLE IF EXISTS lessons;
