
CREATE TABLE IF NOT EXISTS policy_compliance (
                                                 workload_type text NOT NULL,
                                                 tenant_id integer NOT NULL,
                                                 project_id integer NOT NULL,
                                                 template_id uuid NOT NULL,
                                                 aligned boolean,
                                                 reason JSONB,
                                                 policy_time timestamp with time zone,
                                                 imposed_asset boolean
);

CREATE UNIQUE INDEX IF NOT EXISTS policy_compliance_pkey
        ON policy_compliance (tenant_id,project_id,template_id,workload_type);
