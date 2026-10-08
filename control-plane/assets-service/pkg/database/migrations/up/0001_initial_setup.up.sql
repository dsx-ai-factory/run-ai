-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

----------------------------------------------------------------------------------
-- Create Timestamp Trigger
----------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

----------------------------------------------------------------------------------
-- Create Asset Name Validation Trigger
----------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION asset_name_validation()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.name IN (select name FROM assets_metadata WHERE NEW.kind = assets_metadata.kind AND
                                 NEW.tenant_id = assets_metadata.tenant_id AND NEW.scope != assets_metadata.scope)
  THEN
      RAISE EXCEPTION 'non-% % asset named % already exists', NEW.scope, NEW.kind, NEW.name;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

----------------------------------------------------------------------------------
-- Create the Tables
----------------------------------------------------------------------------------

--
-- Name: assets_access_key; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_access_key (
    asset_id uuid,
    access_key_id text,
    secret_access_key text,
    existing_secret_name text,
    key_id integer
);

--
-- Name: assets_compute; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_compute (
    asset_id uuid,
    node_type text,
    gpu_request jsonb,
    cpu_core_request numeric,
    cpu_core_limit numeric,
    cpu_memory_request text,
    cpu_memory_limit text,
    large_shm_request boolean,
    extended_resources jsonb
);

--
-- Name: assets_docker_registry; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_docker_registry (
    asset_id uuid,
    "user" text,
    password text,
    url text,
    key_id integer,
    existing_secret_name text
);

--
-- Name: assets_environment; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_environment (
    asset_id uuid,
    command text,
    args text,
    environment_variables jsonb,
    image text,
    image_pull_policy text,
    working_dir text,
    host_ipc boolean,
    host_network boolean,
    connections jsonb,
    service_type text,
    create_home_dir boolean,
    allow_privilege_escalation boolean,
    run_as_uid bigint,
    run_as_gid bigint,
    capabilities jsonb,
    shared_memory boolean,
    integrations jsonb,
    supplemental_groups text,
    override_uid_gid_in_workspace boolean,
    uid_gid_source text
);

--
-- Name: assets_git; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_git (
    asset_id uuid,
    repository text,
    branch text,
    revision text,
    password_asset_id uuid,
    path text
);

--
-- Name: assets_host_path; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_host_path (
    asset_id uuid,
    mount_path text,
    type text,
    path text,
    read_only boolean
);

--
-- Name: assets_metadata; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_metadata (
    id uuid,
    kind text,
    scope text,
    tenant_id integer,
    cluster_id uuid,
    project_id integer,
    name text,
    created_by text,
    created_at timestamp with time zone,
    updated_by text,
    updated_at timestamp with time zone,
    project_name text,
    is_policy boolean,
    department_id text,
    auto_delete boolean
);

--
-- Name: assets_nfs; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_nfs (
    asset_id uuid,
    server text,
    path text,
    read_only boolean,
    mount_path text
);

--
-- Name: assets_password; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_password (
    asset_id uuid,
    "user" text,
    password text,
    existing_secret_name text,
    key_id integer,
    private_use_only boolean
);

--
-- Name: assets_pvc; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_pvc (
    asset_id uuid,
    claim_name text,
    existing_pvc boolean,
    path text,
    read_only boolean,
    claim_info jsonb
);

--
-- Name: assets_registry; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_registry (
    asset_id uuid,
    credential_kind text,
    credential_id uuid,
    url text
);

--
-- Name: assets_s3; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_s3 (
    asset_id uuid,
    url text,
    bucket text,
    access_key_asset_id uuid,
    path text
);

--
-- Name: assets_template; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_template (
    asset_id uuid,
    specific_env text
);

--
-- Name: assets_template_content; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS assets_template_content (
    template_asset_id uuid,
    contained_asset_id uuid
);

--
-- Name: policy_compliance; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS policy_compliance (
    workload_type text,
    tenant_id integer,
    project_id integer,
    asset_id uuid,
    aligned boolean,
    reason text,
    policy_time timestamp with time zone,
    imposed_asset boolean
);

--
-- Name: workloads; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS workloads (
    id uuid,
    tenant_id integer,
    project_id integer,
    cluster_id uuid,
    name text,
    desired_state text,
    current_state text,
    created_by text,
    created_at timestamp with time zone,
    updated_by text,
    updated_at timestamp with time zone,
    workload jsonb,
    specific_env text,
    workload_type text
);

--
-- Name: workloads_assets; Type: TABLE;  
--

CREATE TABLE IF NOT EXISTS workloads_assets (
    entry_id integer,
    workload_id uuid,
    asset_id uuid,
    asset_kind text,
    auto_delete boolean
);

----------------------------------------------------------------------------------
-- Create all necessary indexes
----------------------------------------------------------------------------------
CREATE UNIQUE INDEX IF NOT EXISTS assets_metadata_pkey       ON assets_metadata (id);
CREATE        INDEX IF NOT EXISTS assets_metadata_by_project ON assets_metadata (project_id,tenant_id);

CREATE UNIQUE INDEX IF NOT EXISTS assets_metadata_by_kind_name_tenant ON assets_metadata (kind,tenant_id,name) WHERE (scope = 'tenant'::text);
CREATE UNIQUE INDEX IF NOT EXISTS assets_metadata_by_kind_name_proj   ON assets_metadata (kind,tenant_id,name,project_id) WHERE (scope = 'project'::text);
CREATE UNIQUE INDEX IF NOT EXISTS assets_metadata_by_kind_name_dept   ON assets_metadata (kind,tenant_id,name,department_id) WHERE (scope = 'department'::text);

CREATE UNIQUE INDEX IF NOT EXISTS assets_access_key_pkey     ON assets_access_key (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_password_pkey       ON assets_password (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_registry_pkey       ON assets_registry (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_docker_registry_key ON assets_docker_registry (asset_id);

CREATE UNIQUE INDEX IF NOT EXISTS assets_environment_pkey ON assets_environment (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_compute_pkey     ON assets_compute (asset_id);

CREATE UNIQUE INDEX IF NOT EXISTS assets_git_pkey         ON assets_git (asset_id);
CREATE        INDEX IF NOT EXISTS assets_git_by_password  ON assets_git (password_asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_host_path_pkey   ON assets_host_path (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_nfs_pkey         ON assets_nfs (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_pvc_pkey         ON assets_pvc (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_s3_pkey          ON assets_s3 (asset_id);
CREATE        INDEX IF NOT EXISTS assets_s3_by_access_key ON assets_s3 (access_key_asset_id);

CREATE UNIQUE INDEX IF NOT EXISTS assets_template_pkey         ON assets_template (asset_id);
CREATE UNIQUE INDEX IF NOT EXISTS assets_template_content_pkey ON assets_template_content (template_asset_id, contained_asset_id);

CREATE UNIQUE INDEX IF NOT EXISTS workloads_pkey         ON workloads (id);
CREATE UNIQUE INDEX IF NOT EXISTS workloads_by_reference ON workloads (tenant_id,cluster_id,project_id,workload_type,name);

CREATE UNIQUE INDEX IF NOT EXISTS workloads_assets_pkey        ON workloads_assets (entry_id);
CREATE        INDEX IF NOT EXISTS workloads_assets_by_asset_id ON workloads_assets (asset_id);

CREATE UNIQUE INDEX IF NOT EXISTS policy_compliance_pkey ON policy_compliance (tenant_id,project_id,asset_id,workload_type);

----------------------------------------------------------------------------------
-- Apply Timestamp Trigger
----------------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'assets_metadata'::regclass AND tgname = 'set_timestamp')
    THEN
        CREATE TRIGGER set_timestamp
            BEFORE UPDATE ON assets_metadata
            FOR EACH ROW
            EXECUTE PROCEDURE trigger_set_timestamp();
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'workloads'::regclass AND tgname = 'set_timestamp')
    THEN
        CREATE TRIGGER set_timestamp
            BEFORE UPDATE ON workloads
            FOR EACH ROW
            EXECUTE PROCEDURE trigger_set_timestamp();
    END IF;
END$$;

----------------------------------------------------------------------------------
-- Apply Asset Name Validation Trigger
----------------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'assets_metadata'::regclass AND tgname = 'validate_asset_name') THEN
        CREATE TRIGGER validate_asset_name
            BEFORE UPDATE ON assets_metadata
            FOR EACH ROW
            EXECUTE PROCEDURE asset_name_validation();
    END IF;
END$$;

----------------------------------------------------------------------------------
-- Load Data
----------------------------------------------------------------------------------

--
-- Data for Name: assets_compute; Type: TABLE DATA;  
--

INSERT INTO assets_compute (
    asset_id, node_type, gpu_request, cpu_core_request,
    cpu_core_limit, cpu_memory_request, cpu_memory_limit,
    large_shm_request, extended_resources
) VALUES
('1a261710-fcba-4c8f-8be2-03af4a8cab1d', NULL, '{"gpuPortion": {"portion": 1}}', 0.1, NULL, '100M', NULL, NULL, NULL),
('2079838e-6e3c-46de-b80a-c6d81f042cc9', NULL, '{"gpuPortion": {"portion": 0}}', 0.1, NULL, '100M', NULL, NULL, NULL),
('734550cc-b1e1-4075-987b-7a1c2e4d42d0', NULL, '{"gpuPortion": {"portion": 0.5}}', 0.1, NULL, '100M', NULL, NULL, NULL),
('d6fdb301-2515-4193-be5b-e42ca100ee09', NULL, '{"gpuPortion": {"portion": 2}}', 0.2, NULL, '200M', NULL, NULL, NULL),
('f461ea2f-f937-4387-9976-b4cc4bae6e38', NULL, '{"gpuPortion": {"portion": 0.1}}', 0.1, NULL, '100M', NULL, NULL, NULL)
ON CONFLICT (asset_id) DO NOTHING;

--
-- Data for Name: assets_environment; Type: TABLE DATA;  
--

INSERT INTO assets_environment (
    asset_id, command, args, environment_variables, image, image_pull_policy,
    working_dir, host_ipc, host_network, connections, service_type,
    create_home_dir, allow_privilege_escalation, run_as_uid, run_as_gid,
    capabilities, shared_memory, integrations, supplemental_groups,
    override_uid_gid_in_workspace, uid_gid_source
) VALUES
('01ef611d-ce9b-4bff-a29f-6ba4bb4224dc', 'tensorboard', '--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}', NULL, 'tensorflow/tensorflow:2.20.0', 'IfNotPresent', NULL, NULL, NULL, '[{"name": "TensorBoard", "internalToolInfo": {"toolType": "tensorboard", "containerPort": 6006, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]', NULL, NULL, NULL, NULL, NULL, '[]', 'f', NULL, NULL, NULL, 'fromTheImage'),
('102d4a7b-2345-47d4-bcc1-52347eef1b4e', 'start-notebook.sh', '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''''', '[]', 'jupyter/scipy-notebook', 'IfNotPresent', NULL, NULL, NULL, '[{"name": "Jupyter", "internalToolInfo": {"toolType": "jupyter-notebook", "containerPort": 8888, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]', NULL, NULL, NULL, NULL, NULL, '[]', 'f', NULL, NULL, NULL, 'fromTheImage')
-- The omitted jupyter-tensorboard environment row referenced an internal image.
ON CONFLICT (asset_id) DO NOTHING;

--
-- Data for Name: assets_metadata; Type: TABLE DATA;  
--

INSERT INTO assets_metadata (
    id, kind, scope, tenant_id, cluster_id, project_id, name, created_by,
    created_at, updated_by, updated_at, project_name, is_policy,
    department_id, auto_delete
) VALUES
('01ef611d-ce9b-4bff-a29f-6ba4bb4224dc', 'environment', 'tenant', 2147483647, NULL, NULL, 'tensorboard', 'runai', '2023-02-15 07:36:48.365317+00', 'runai', '2023-02-15 07:36:48.365317+00', NULL, 'f', NULL, NULL),
('102d4a7b-2345-47d4-bcc1-52347eef1b4e', 'environment', 'tenant', 2147483647, NULL, NULL, 'jupyter-lab', 'runai', '2023-02-15 07:25:32.294217+00', 'runai', '2023-02-15 07:25:32.294217+00', NULL, 'f', NULL, NULL),
('1a261710-fcba-4c8f-8be2-03af4a8cab1d', 'compute', 'tenant', 2147483647, NULL, NULL, 'one-gpu', 'runai', '2023-02-14 18:48:10.809341+00', 'runai', '2023-02-14 18:48:10.809341+00', NULL, 'f', NULL, NULL),
('2079838e-6e3c-46de-b80a-c6d81f042cc9', 'compute', 'tenant', 2147483647, NULL, NULL, 'cpu-only', 'runai', '2023-02-14 18:46:45.19818+00', 'runai', '2023-02-14 18:46:45.19818+00', NULL, 'f', NULL, NULL),
('6d9a2313-f447-40af-bd61-56bafa249e52', 'environment', 'tenant', 2147483647, NULL, NULL, 'jupyter-tensorboard', 'runai', '2023-02-14 18:51:14.086903+00', 'runai', '2023-02-14 18:51:14.086903+00', NULL, 'f', NULL, NULL),
('734550cc-b1e1-4075-987b-7a1c2e4d42d0', 'compute', 'tenant', 2147483647, NULL, NULL, 'half-gpu', 'runai', '2023-02-14 18:47:53.924767+00', 'runai', '2023-02-14 18:47:53.924767+00', NULL, 'f', NULL, NULL),
('d6fdb301-2515-4193-be5b-e42ca100ee09', 'compute', 'tenant', 2147483647, NULL, NULL, 'two-gpus', 'runai', '2023-02-14 18:48:28.422635+00', 'runai', '2023-02-14 18:48:28.422635+00', NULL, 'f', NULL, NULL),
('f461ea2f-f937-4387-9976-b4cc4bae6e38', 'compute', 'tenant', 2147483647, NULL, NULL, 'small-fraction', 'runai', '2023-02-14 18:47:27.732118+00', 'runai', '2023-02-14 18:47:27.732118+00', NULL, 'f', NULL, NULL)
ON CONFLICT (id) DO NOTHING;
