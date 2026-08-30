DROP INDEX IF EXISTS crud_templates_search_vector_idx;
ALTER TABLE IF EXISTS crud_templates DROP COLUMN IF EXISTS search_vector; 