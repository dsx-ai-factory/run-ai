ALTER TABLE approved_deletions ADD COLUMN IF NOT EXISTS avoid_deletions_of_obsolete_policies BOOL;
