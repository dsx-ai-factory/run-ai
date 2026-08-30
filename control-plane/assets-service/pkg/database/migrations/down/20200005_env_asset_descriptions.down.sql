DO $$
    DECLARE
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        llm_server_env_vars jsonb := '[{"name": "RUNAI_MODEL", "value": "MODEL_NAME_OR_PATH_HERE"}, {"name": "RUNAI_MODEL_NAME", "value": "MODEL_NAME_HERE"}]';

        chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
        chatbot_ui_env_vars jsonb := '[{"name": "RUNAI_MODEL_NAME", "value": "MODEL_NAME_HERE"}, {"name": "RUNAI_MODEL_BASE_URL", "value": "http://WORKLOAD.runai-PROJECT.svc"}, {"name": "RUNAI_MODEL_MAX_LENGTH", "value": "4096"}, {"name": "RUNAI_MODEL_TOKEN_LIMIT", "value": "2048"}, {"name": "PATH_PREFIX", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}"}]';

        gpt2_env_id uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';
        gpt2_env_vars jsonb := '[{"name": "PATH_PREFIX", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}"}]';

        jupyter_tensorboard_env_id uuid := '6d9a2313-f447-40af-bd61-56bafa249e52';
        jupyter_tensorboard_env_vars jsonb := '[{"name": "LOG_DIR", "value": "./"}]';
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
