ALTER TABLE IF EXISTS crud_templates ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS
    (to_tsvector(
            'english',
            COALESCE(name::text, '') || ' ' ||
            COALESCE(description::text, '') || ' ' ||
            COALESCE(workload_type::text, '') || ' ' ||
            COALESCE(created_by::text, ''))) STORED;

CREATE INDEX IF NOT EXISTS crud_templates_search_vector_idx ON crud_templates using gin(search_vector); 