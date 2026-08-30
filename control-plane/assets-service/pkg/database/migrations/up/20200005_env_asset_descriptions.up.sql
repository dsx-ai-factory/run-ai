DO $$
    DECLARE
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        llm_server_env_vars jsonb := '[{"name": "RUNAI_MODEL", "description": "If the model is loaded from Hugging Face, specify the Hugging Face model name (e.g. meta-llama/Llama-3.1-8B-Instruct).\n\nIf the model is loaded from a local storage, specify the path to the model files."}, {"name": "RUNAI_MODEL_NAME", "description": "If the model is loaded from local storage, specify the model name to be used as the served model (this name must be included within requests). Otherwise, remove this environment variable entirely."}]';

        chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
        chatbot_ui_env_vars jsonb := '[{"name": "RUNAI_MODEL_NAME", "description": "The name of the model to interact with, as configured on the server (e.g. meta-llama/Llama-3.1-8B-Instruct)."}, {"name": "RUNAI_MODEL_BASE_URL", "description": "The base URL of the model to interact with, available in the Connection modal on the Workloads grid. For example, http://WORKLOAD.runai-PROJECT.svc. The model should be OpenAI API compatible."}, {"name": "RUNAI_MODEL_TOKEN_LIMIT", "value": "8192", "description": "The model’s context window size, which is the maximum amount of text (request + response) the model can process in a single interaction."}, {"name": "RUNAI_MODEL_MAX_LENGTH", "value": "16384", "description": "The maximum request length, in characters."}, {"name": "PATH_PREFIX", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}", "description": "If the cluster uses path-based URLs (e.g., http://my.domain.com/proj/workload), leave the value as is.\n\nIf the cluster uses host-based URLs (e.g., http://proj-workload.my.domain.com), remove this environment variable entirely."}]';

        gpt2_env_id uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';
        gpt2_env_vars jsonb := '[{"name": "PATH_PREFIX", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}", "description": "If the cluster uses path-based URLs (e.g., http://my.domain.com/proj/workload), leave the value as is.\n\nIf the cluster uses host-based URLs (e.g., http://proj-workload.my.domain.com), remove this environment variable entirely."}]';

        jupyter_tensorboard_env_id uuid := '6d9a2313-f447-40af-bd61-56bafa249e52';
        jupyter_tensorboard_env_vars jsonb := '[{"name": "LOG_DIR", "value": "./", "description": "The directory where TensorBoard searches for log files generated during the model training. These log files typically include event data such as metrics (e.g., loss, accuracy.)"}]';
    BEGIN
        BEGIN
            UPDATE assets_environment
            SET environment_variables = llm_server_env_vars
            WHERE asset_id = llm_server_env_id;

            UPDATE assets_environment
            SET environment_variables = chatbot_ui_env_vars
            WHERE asset_id = chatbot_ui_env_id;

            UPDATE assets_environment
            SET environment_variables = gpt2_env_vars
            WHERE asset_id = gpt2_env_id;

            UPDATE assets_environment
            SET environment_variables = jupyter_tensorboard_env_vars
            WHERE asset_id = jupyter_tensorboard_env_id;
        END;
   END
$$;
