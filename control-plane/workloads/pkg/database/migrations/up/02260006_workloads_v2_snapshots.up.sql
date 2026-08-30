CREATE TABLE IF NOT EXISTS workloads_v2_snapshots (
        id uuid NOT NULL,

        created_by text        NOT NULL,
        created_at timestamptz NOT NULL,
        updated_by text        NOT NULL,
        updated_at timestamptz NOT NULL,

        tenant_id int4 NOT NULL,
        cluster_id uuid NOT NULL,
        department_id text NOT NULL DEFAULT '',
        project_id text NOT NULL,

        workload_type text NOT NULL,

        requested_name text NOT NULL DEFAULT '',
        name text NOT NULL,
        namespace text NOT NULL DEFAULT '',

        spec jsonb NOT NULL,
        master_spec jsonb,
        master_same_as_worker bool,
        assets_mapping jsonb,

        PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS workloads_v2_snapshots_lookup ON workloads_v2_snapshots (tenant_id, workload_type);
