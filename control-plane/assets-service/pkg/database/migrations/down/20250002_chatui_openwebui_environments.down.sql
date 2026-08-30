-- Rollback migration for Chat-UI and OpenWebUI environment presets

DO $$
DECLARE
    -- UUIDs for environment presets to delete
    uuid_environment_chatui_host uuid := '4ca2338b-1f43-4123-b349-57d6ad898542';
    uuid_environment_chatui_path uuid := '01a576e4-a02f-4866-ae90-912265ffee0a';
    uuid_environment_openwebui uuid := 'd0e72d44-c5c8-4a2b-b8b8-9d062de47b7c';

BEGIN
    -- Delete from assets_supported_workload_types
    DELETE FROM assets_supported_workload_types 
    WHERE asset_id IN (uuid_environment_chatui_host, uuid_environment_chatui_path, uuid_environment_openwebui);

    -- Delete from assets_environment
    DELETE FROM assets_environment 
    WHERE asset_id IN (uuid_environment_chatui_host, uuid_environment_chatui_path, uuid_environment_openwebui);

    -- Delete from assets_metadata
    DELETE FROM assets_metadata 
    WHERE id IN (uuid_environment_chatui_host, uuid_environment_chatui_path, uuid_environment_openwebui);

    RAISE NOTICE 'Chat-UI and OpenWebUI environment assets have been removed';
END $$;

