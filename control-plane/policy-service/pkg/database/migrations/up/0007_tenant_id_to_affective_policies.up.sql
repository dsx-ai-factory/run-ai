ALTER TABLE effective_policies 
  ADD COLUMN IF NOT EXISTS tenant_id INT;

UPDATE effective_policies ep
  SET tenant_id = p.tenant_id
  FROM policies p
  WHERE ep.policy_id = p.id;

CREATE INDEX IF NOT EXISTS ep_by_tenant_and_id 
  ON effective_policies (tenant_id, policy_id);
