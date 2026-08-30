
DO $$
DECLARE
    CREATED_BY text := 'Run:ai';
    CREATED_AT timestamp := now();

    -- uuids for the existed environments preset
    uuid_environment_tensorboard_path_based uuid := '01ef611d-ce9b-4bff-a29f-6ba4bb4224dc';
    uuid_environment_jupyter_path_based uuid := '102d4a7b-2345-47d4-bcc1-52347eef1b4e';
    uuid_environment_jupyter_tensorboard uuid := '6d9a2313-f447-40af-bd61-56bafa249e52';

    -- uuids for existing system assets
    uuid_gpt2 uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';
    uuid_llm_server uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
    uuid_chatbot_ui uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';

    -- uuids for the new environments preset
    uuid_environment_jupyter_host_based uuid := '17af3f04-f632-4a43-9ceb-ae8c062eb250';
    uuid_environment_tensorboard_host_based uuid := '490ed99b-7706-4ad3-be19-a383dbf7bf27';
    uuid_environment_rstudio uuid := '97db5c9e-24d4-4029-b980-c8c4bca0226f';
    uuid_environment_vscdoe uuid := '16acccda-d80f-402a-aa21-31fdddc67bc4';

    -- description text for environments preset
    description_tensorboard text := 'A visualization toolkit for TensorFlow that helps users monitor and analyze ML models, displaying various metrics and model architecture';
    description_rstudio text := 'An integrated development environment (IDE) for R, commonly used for statistical computing and data analysis';
    description_vscode text := 'A fast, lightweight code editor with powerful features like intelligent code completion, debugging, Git integration, and extensions, ideal for web development, data science, and more';
    description_jupyter text := 'An interactive development environment for Jupyter notebooks, code, and data visualization';
    description_jupyter_tensorboard text := 'An integrated combination of the interactive Jupyter development environment and TesnorFlow''s visualization toolkit for monitoring and analyzing ML models';

    description_chatbot_ui text := 'A user interface for interacting with chat-based AI models, often used for testing and deploying chatbot applications';
    description_llm_server text := 'A vLLM-based server that hosts and serves large language models for inference, enabling API-based access to AI models';
    description_gpt2 text := 'A package of an inference server, gpt2 model and chat UI, often used for quick demos';

BEGIN
    BEGIN
        -- Insert metadata for new environments preset
        INSERT INTO assets_metadata ("id", "kind", "scope", "tenant_id", "name", "created_by", "created_at", "updated_by", "updated_at", "description") VALUES
            (uuid_environment_jupyter_host_based,     'environment', 'tenant', '2147483647', 'jupyter-scipy',           CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, description_jupyter),
            (uuid_environment_tensorboard_host_based, 'environment', 'tenant', '2147483647', 'tensorboard-tensorflow',  CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, description_tensorboard),
            (uuid_environment_rstudio,                'environment', 'tenant', '2147483647', 'rstudio',                 CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, description_rstudio),
            (uuid_environment_vscdoe,                 'environment', 'tenant', '2147483647', 'vscode',                  CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, description_vscode);

        -- Insert tensorboard and jupyter "host base routing" environments preset spec
        INSERT INTO assets_environment ("asset_id", "command", "args", "image", "image_pull_policy", "connections", "uid_gid_source", "routing_method") VALUES
            (uuid_environment_tensorboard_host_based, 'tensorboard', '--logdir ./ --host 0.0.0.0', 'tensorflow/tensorflow:2.20.0', 'IfNotPresent', '[{"name": "TensorBoard", "internalToolInfo": {"toolType": "tensorboard", "containerPort": 6006, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]',  'fromTheImage', 'host-base'),
            (uuid_environment_jupyter_host_based, 'start-notebook.sh', '--NotebookApp.token=''''', 'jupyter/scipy-notebook', 'IfNotPresent', '[{"name": "Jupyter", "internalToolInfo": {"toolType": "jupyter-notebook", "containerPort": 8888, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]', 'fromTheImage', 'host-base');

        -- Insert the new environment for rstudio preset
        INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "connections", "uid_gid_source", "routing_method") VALUES
            (uuid_environment_rstudio, 'rocker/rstudio', 'IfNotPresent', '[{"name": "RStudio", "internalToolInfo": {"toolType": "rstudio", "containerPort": 8787, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]', 'fromTheImage', 'host-base');

        -- Insert the new environment for vscode preset
        INSERT INTO assets_environment ("asset_id", "image", "args" ,"image_pull_policy", "connections", "uid_gid_source", "routing_method") VALUES
            (uuid_environment_vscdoe, 'ghcr.io/coder/code-server', '--auth none','IfNotPresent', '[{"name": "VSCode", "internalToolInfo": {"toolType": "visual-studio-code", "containerPort": 8080, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]', 'fromTheImage','host-base');

        -- set new  preset supported workloads type
        INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference") VALUES
            (uuid_environment_jupyter_host_based,     TRUE, TRUE, FALSE, FALSE),
            (uuid_environment_tensorboard_host_based, TRUE, TRUE, FALSE, FALSE),
            (uuid_environment_rstudio,                TRUE, TRUE, FALSE, FALSE),
            (uuid_environment_vscdoe,                 TRUE, TRUE, FALSE, FALSE);

        -- Update the current tensorboard and jupyter preset as path-based routing
        -- note that this doesn't update the already tenant copied, as it could have been edited by the customer to be host-based
        UPDATE assets_environment SET routing_method = 'path-base' WHERE asset_id::uuid = uuid_environment_tensorboard_path_based;
        UPDATE assets_environment SET routing_method = 'path-base' WHERE asset_id::uuid = uuid_environment_jupyter_path_based;
        UPDATE assets_environment SET routing_method = 'path-base' WHERE asset_id::uuid = uuid_environment_jupyter_tensorboard;

        -- Update the current tensorboard and jupyter preset description
        UPDATE assets_metadata SET description = description_tensorboard         WHERE id::uuid = uuid_environment_tensorboard_path_based;
        UPDATE assets_metadata SET description = description_jupyter             WHERE id::uuid = uuid_environment_jupyter_path_based;
        UPDATE assets_metadata SET description = description_jupyter_tensorboard WHERE id::uuid = uuid_environment_jupyter_tensorboard;

        -- Update the current system assets description
        UPDATE assets_metadata SET description = description_gpt2       WHERE id::uuid = uuid_gpt2;
        UPDATE assets_metadata SET description = description_llm_server WHERE id::uuid = uuid_llm_server;
        UPDATE assets_metadata SET description = description_chatbot_ui WHERE id::uuid = uuid_chatbot_ui;

EXCEPTION
        -- to ensure that the migration is idempotent
        WHEN unique_violation THEN
            -- do nothing, already exists
            RAISE NOTICE 'Assets already exist, skipping';
END;
END $$;