DO $$
    DECLARE
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
        gpt2_env_id       uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';

        llm_server_image text := 'gcr.io/run-ai-demo/llm';
        chatbot_ui_image text := 'gcr.io/run-ai-demo/llm:app';
        gpt2_image       text := 'gcr.io/run-ai-demo/quickstart-inference:gpt2-cpu';
    BEGIN
        BEGIN
            UPDATE assets_environment
            SET image = llm_server_image
            WHERE asset_id = llm_server_env_id;

            UPDATE assets_environment
            SET image = chatbot_ui_image
            WHERE asset_id = chatbot_ui_env_id;

            UPDATE assets_environment
            SET image = gpt2_image
            WHERE asset_id = gpt2_env_id;
        END;
    END
$$;
