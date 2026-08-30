DROP INDEX IF EXISTS access_rules_search_vector_idx;
ALTER TABLE IF EXISTS access_rules DROP COLUMN IF EXISTS search_vector;
