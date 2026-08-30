
DO $$
DECLARE
    CREATED_BY text := 'Run:ai';
    CREATED_AT timestamp := now();

    -- common spec values
    common_command text := 'jupyter-lab';
    common_path_base_args text := '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token='''' --ServerApp.allow_remote_access=true --allow-root --port=8888 --no-browser';
    common_image_pull_policy text := 'IfNotPresent';
    common_connections jsonb := '[{"name": "Jupyter", "internalToolInfo": {"toolType": "jupyter-notebook", "connectionType": "ExternalUrl", "containerPort": 8888, "externalUrlInfo": {}}}]';
    common_uid_gid_source text := 'fromTheImage';
    common_working_dir text := '/opt/NeMo';
    common_path_routing_method text := 'path-base';

    -- values for NeMo environments preset
    nemo_uuid uuid := '844edfab-940c-47a6-81aa-60d8a87dd72f';
    nemo_name text := 'nemo';
    nemo_description text := 'A framework for training and deploying LLMs and generative AI developed by NVIDIA with automated data processing, model training techniques, and flexible deployment options';
    nemo_image text:= 'nvcr.io/nvidia/nemo:25.02';

    -- values for BioNeMo environments preset
    bionemo_uuid uuid := '1312b1f8-223c-42f9-8076-64cac7103f66';
    bionemo_name text := 'bio-nemo';
    bionemo_description text := 'A framework developed by NVIDIA for large-scale biomolecular models, optimized to support drug discovery, genomics, and protein structure prediction';
    bionemo_image text:= 'nvcr.io/nvidia/clara/bionemo-framework:2.5';

    -- values for PyTorch environments preset
    pytorch_uuid uuid := '6618d68d-3aa1-43aa-ac3a-0a3f60ee5364';
    pytorch_name text := 'pytorch';
    pytorch_description text := 'An integrated deep learning framework accelerated by NVIDIA, built for dynamic training and seamless compatibility with Python tools like NumPy and SciPy';
    pytorch_image text := 'nvcr.io/nvidia/pytorch:25.02-py3';

BEGIN
BEGIN
        -- NeMo environment --

        -- Insert metadata for NeMo
        INSERT INTO assets_metadata ("id", "name", "description", "kind", "scope", "tenant_id", "created_by", "created_at", "updated_by", "updated_at") VALUES
            (nemo_uuid, nemo_name, nemo_description,  'environment', 'tenant', '2147483647',CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT)
            ON CONFLICT DO NOTHING;

        -- Insert environment spec for NeMo
        INSERT INTO assets_environment ("asset_id", "image", "command", "args", "image_pull_policy", "connections", "uid_gid_source", "working_dir", "routing_method") VALUES
             (nemo_uuid, nemo_image, common_command, common_path_base_args, common_image_pull_policy, common_connections, common_uid_gid_source, common_working_dir, common_path_routing_method)
            ON CONFLICT DO NOTHING;

        -- Insert supported workload for NeMo
        INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference") VALUES
            (nemo_uuid,     TRUE, TRUE, FALSE, FALSE)
            ON CONFLICT DO NOTHING;

        -- BioNeMo environment --

        -- Insert metadata for BioNeMo
        INSERT INTO assets_metadata ("id", "name", "description", "kind", "scope", "tenant_id", "created_by", "created_at", "updated_by", "updated_at") VALUES
            (bionemo_uuid, bionemo_name, bionemo_description,  'environment', 'tenant', '2147483647',CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT)
            ON CONFLICT DO NOTHING;

        -- Insert environment spec for BioNeMo
        INSERT INTO assets_environment ("asset_id", "image", "command", "args", "image_pull_policy", "connections", "uid_gid_source", "working_dir", "routing_method") VALUES
            (bionemo_uuid, bionemo_image, common_command, common_path_base_args, common_image_pull_policy, common_connections, common_uid_gid_source, common_working_dir, common_path_routing_method)
            ON CONFLICT DO NOTHING;

        -- Insert supported workload for BioNeMo
        INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference") VALUES
            (bionemo_uuid,     TRUE, TRUE, FALSE, FALSE)
            ON CONFLICT DO NOTHING;

        -- PyTorch environment --

        -- Insert metadata for PyTorch
        INSERT INTO assets_metadata ("id", "name", "description", "kind", "scope", "tenant_id", "created_by", "created_at", "updated_by", "updated_at") VALUES
            (pytorch_uuid, pytorch_name, pytorch_description,  'environment', 'tenant', '2147483647',CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT)
            ON CONFLICT DO NOTHING;

        -- Insert environment spec for PyTorch
        INSERT INTO assets_environment ("asset_id", "image", "command", "args", "image_pull_policy", "connections", "uid_gid_source", "working_dir", "routing_method") VALUES
            (pytorch_uuid, pytorch_image, common_command, common_path_base_args, common_image_pull_policy, common_connections, common_uid_gid_source, common_working_dir, common_path_routing_method)
            ON CONFLICT DO NOTHING;

        -- Insert supported workload for PyTorch
        INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference") VALUES
            (pytorch_uuid,     TRUE, TRUE, FALSE, FALSE)
            ON CONFLICT DO NOTHING;


EXCEPTION
        -- to ensure that the migration is idempotent
        WHEN unique_violation THEN
            -- do nothing, already exists
            RAISE NOTICE 'Assets already exist, skipping';
END;
END $$;