ALTER TABLE node ADD COLUMN IF NOT EXISTS tenant_id INTEGER;

-- Create a function to add tenant id for existing rows
CREATE OR REPLACE FUNCTION nodes_tenant_id_setup_trigger_for_existing_rows()
RETURNS VOID AS $$
BEGIN
    -- Fetch tenant_id from clusters table
UPDATE node
SET
    tenant_id = c.tenant_id
    FROM clusters AS c
WHERE
    node.cluster_uuid = c.id::text and node.tenant_id IS NULL;
RETURN;
END;
$$ LANGUAGE plpgsql;

-- Execute the function to set up triggers for existing rows
SELECT nodes_tenant_id_setup_trigger_for_existing_rows();