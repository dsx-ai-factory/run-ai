DROP INDEX IF EXISTS roles_search_vector_idx;
ALTER TABLE IF EXISTS roles DROP COLUMN IF EXISTS search_vector;
