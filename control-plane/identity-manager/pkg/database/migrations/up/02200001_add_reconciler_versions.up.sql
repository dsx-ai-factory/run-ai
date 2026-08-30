-- Create the new table 'reconciler_versions'
CREATE TABLE IF NOT EXISTS reconciler_versions (
    id                      UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id               TEXT NOT NULL,
    applications            VARCHAR(10) NOT NULL,
    internal_applications   VARCHAR(10) NOT NULL,
    created_by              TEXT,
    updated_by              TEXT,
    created_at              TIMESTAMPTZ DEFAULT now(),
    updated_at              TIMESTAMPTZ DEFAULT now(),

    FOREIGN KEY (tenant_id) REFERENCES realms(tenant_id)
    );

-- Create index on table
create unique index if not exists reconciler_versions_tenant_id_idx on reconciler_versions (tenant_id);

-- Insert data into 'reconciler_versions' from the existing 'realms' table
INSERT INTO reconciler_versions (tenant_id, applications, internal_applications)
SELECT tenant_id, 'v1', 'v1' FROM identity_manager.realms ON CONFLICT DO NOTHING;