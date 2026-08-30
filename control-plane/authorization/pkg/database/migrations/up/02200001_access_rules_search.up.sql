ALTER TABLE IF EXISTS access_rules ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS
    (to_tsvector(
            'english',
            COALESCE(created_by::text, '') || ' ' ||
            COALESCE(subject_id::text, '') || ' ' ||
            COALESCE(subject_type::text, '') || ' ' ||
            COALESCE(scope_type::text, '') || ' ' ||
            COALESCE(scope_name::text, ''))) STORED;
CREATE INDEX IF NOT EXISTS access_rules_search_vector_idx ON access_rules using gin(search_vector);