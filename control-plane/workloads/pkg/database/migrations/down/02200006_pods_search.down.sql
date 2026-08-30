 DROP INDEX IF EXISTS pods_search_vector_idx;
 ALTER TABLE pods DROP COLUMN IF EXISTS search_vector;