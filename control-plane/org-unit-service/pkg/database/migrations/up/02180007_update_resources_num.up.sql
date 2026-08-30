ALTER TABLE resources ALTER COLUMN cpu_deserved TYPE DECIMAL(12,3);
ALTER TABLE resources ALTER COLUMN cpu_limit TYPE DECIMAL(12,3);
ALTER TABLE resources ALTER COLUMN cpu_over_quota_weight TYPE DECIMAL(12,3);
ALTER TABLE resources ALTER COLUMN memory_deserved TYPE DECIMAL(12,3);
ALTER TABLE resources ALTER COLUMN memory_limit TYPE DECIMAL(12,3);
ALTER TABLE resources ALTER COLUMN memory_over_quota_weight TYPE DECIMAL(12,3);
