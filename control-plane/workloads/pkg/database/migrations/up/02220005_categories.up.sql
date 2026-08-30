-- Create categories table
CREATE TABLE IF NOT EXISTS workloads_categories (
    id uuid NOT NULL PRIMARY KEY,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    created_by varchar NOT NULL,
    updated_at timestamp with time zone DEFAULT now(),
    updated_by varchar,
    deleted_at timestamp with time zone,
    deleted_by varchar,
    name varchar NOT NULL,
    description varchar,
    tenant_id bigint NOT NULL,
    scope_type varchar NOT NULL DEFAULT 'tenant'
);

-- Create workload type to category mappings table
CREATE TABLE IF NOT EXISTS external_workloads_type_config (
    id uuid NOT NULL PRIMARY KEY,
    tenant_id bigint NOT NULL DEFAULT -1,
    type varchar NOT NULL,
    category_id uuid NOT NULL,
    scope_type varchar NOT NULL DEFAULT 'tenant',
    updated_at timestamp with time zone DEFAULT now(),
    updated_by varchar,
    FOREIGN KEY (category_id) REFERENCES workloads_categories(id)
);

-- Add category_id to workloads table
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS category_id uuid;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_workloads_category'
    ) THEN
        ALTER TABLE workloads ADD CONSTRAINT fk_workloads_category
            FOREIGN KEY (category_id) REFERENCES workloads_categories(id);
    END IF;
END $$;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_categories_tenant ON workloads_categories (tenant_id);
CREATE INDEX IF NOT EXISTS idx_workloads_category ON workloads (category_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_tenant_name_active ON workloads_categories (tenant_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_external_workloads_type_config_tenant_type ON external_workloads_type_config (tenant_id, type);

-- Insert default categories
INSERT INTO workloads_categories (id, name, description, tenant_id, created_by, updated_by, scope_type, created_at, updated_at) VALUES
    (gen_random_uuid(), 'Build', 'Build workloads', -1, 'system', 'system', 'system', now(), now()),
    (gen_random_uuid(), 'Train', 'Training workloads', -1, 'system', 'system', 'system', now(), now()),
    (gen_random_uuid(), 'Deploy', 'Deployment and inference workloads', -1, 'system', 'system', 'system', now(), now())
ON CONFLICT (tenant_id, name) WHERE deleted_at IS NULL DO NOTHING;

-- Insert default workload type to category mappings
WITH category_ids AS (
    SELECT id, name
    FROM workloads_categories
    WHERE scope_type = 'system'
)
INSERT INTO external_workloads_type_config (id, tenant_id, type, category_id, scope_type, updated_at, updated_by)
SELECT
    gen_random_uuid(),
    -1,
    type,
    c.id,
    'system',
    now(),
    'system'
FROM (
    VALUES
        ('Deployment', 'Deploy'),
        ('SeldonDeployment', 'Deploy'),
        ('VirtualMachineInstance', 'Train'),
        ('TFJob', 'Train'),
        ('PyTorchJob', 'Train'),
        ('XGBoostJob', 'Train'),
        ('MPIJob', 'Train'),
        ('Notebook', 'Build'),
        ('Job', 'Build'),
        ('StatefulSet', 'Deploy'),
        ('ReplicaSet', 'Deploy'),
        ('Pod', 'Deploy'),
        ('AmlJob', 'Train'),
        ('Service', 'Deploy'),
        ('CronJob', 'Deploy'),
        ('DevWorkspace', 'Build'),
        ('RayCluster', 'Train'),
        ('RayService', 'Deploy'),
        ('RayJob', 'Train'),
        ('TaskRun', 'Build'),
        ('PipelineRun', 'Deploy'),
        ('Workflow', 'Deploy'),
        ('ScheduledWorkflow', 'Deploy')
) AS mappings(type, category_name)
JOIN category_ids c ON c.name = mappings.category_name
ON CONFLICT (tenant_id, type)
DO UPDATE SET
    category_id = EXCLUDED.category_id,
    updated_at = now(),
    updated_by = 'system';
