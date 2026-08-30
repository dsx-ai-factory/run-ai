CREATE SCHEMA IF NOT EXISTS test_schema;

CREATE TABLE IF NOT EXISTS test_schema.items (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at DATE NOT NULL
);
