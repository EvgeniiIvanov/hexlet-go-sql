-- +goose Up
CREATE TABLE lessons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE INDEX idx_lessons_course_id ON lessons(course_id);

-- +goose Down
DROP INDEX IF EXISTS idx_lessons_course_id;
DROP TABLE IF EXISTS lessons;
