DROP INDEX IF EXISTS ep_by_tenant_and_id;

ALTER TABLE effective_policies 
  DROP COLUMN IF EXISTS tenant_id;

