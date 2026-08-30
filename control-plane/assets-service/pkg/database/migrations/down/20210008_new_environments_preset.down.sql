
DO $$
DECLARE
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

BEGIN
    BEGIN
        -- Clear the routing method for the path base tensorboard and jupyter preset
        UPDATE assets_environment SET routing_method = '' WHERE asset_id::uuid = uuid_environment_tensorboard_path_based;
        UPDATE assets_environment SET routing_method = '' WHERE asset_id::uuid = uuid_environment_jupyter_path_based;
        UPDATE assets_environment SET routing_method = '' WHERE asset_id::uuid = uuid_environment_jupyter_tensorboard;

        -- Clear description
        UPDATE assets_metadata SET description = NULL WHERE id::uuid = uuid_environment_tensorboard_path_based;
        UPDATE assets_metadata SET description = NULL WHERE id::uuid = uuid_environment_jupyter_path_based;
        UPDATE assets_metadata SET description = NULL WHERE id::uuid = uuid_environment_jupyter_tensorboard;

        UPDATE assets_metadata SET description = NULL WHERE id::uuid = uuid_gpt2;
        UPDATE assets_metadata SET description = NULL WHERE id::uuid = uuid_llm_server;
        UPDATE assets_metadata SET description = NULL WHERE id::uuid = uuid_chatbot_ui;

        -- Delete the new preset from the support table
        DELETE FROM assets_supported_workload_types WHERE asset_id::uuid IN (
               uuid_environment_jupyter_host_based,
               uuid_environment_tensorboard_host_based,
               uuid_environment_rstudio,
               uuid_environment_vscdoe);

        -- Delete the new preset from the environment table
        DELETE FROM assets_environment WHERE asset_id::uuid IN (
               uuid_environment_jupyter_host_based,
               uuid_environment_tensorboard_host_based,
               uuid_environment_rstudio,
               uuid_environment_vscdoe);

        -- Delete the new preset from the metadata table
        DELETE FROM assets_metadata WHERE id::uuid IN (
               uuid_environment_jupyter_host_based,
               uuid_environment_tensorboard_host_based,
               uuid_environment_rstudio,
               uuid_environment_vscdoe);
END;
END $$;
