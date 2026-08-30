ALTER TABLE node
    ADD COLUMN IF NOT EXISTS nv_link_clique_id TEXT;
ALTER TABLE node
    ADD COLUMN IF NOT EXISTS nv_link_domain_uid TEXT;