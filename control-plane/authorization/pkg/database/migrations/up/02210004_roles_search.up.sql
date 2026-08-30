ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS
    (to_tsvector(
            'english',
            COALESCE(created_by::text, '') || ' ' ||
            COALESCE(name::text, ''))) STORED;
CREATE INDEX IF NOT EXISTS roles_search_vector_idx ON roles using gin(search_vector);