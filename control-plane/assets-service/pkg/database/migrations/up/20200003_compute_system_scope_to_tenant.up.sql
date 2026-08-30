-- This scripts migrates 4 compute assets under system scope to tenant scope
-- it looks for tenants that use the asset in either workloads or templates
-- and creates a new asset for each tenant with the same metadata and content
-- then updates the workloads_assets and template_content tables to use the new asset
-- and finally deletes the original assets from assets_metadata

-- but before doing all of that there are some perp work is required
--  1. need to fix the asset name validation function to allow the same asset name
--    when one of the tenant is null the the other isn't
--  2. need to remove those compute assets from the asset_model table

-- prep work 1: fix the asset name validation function
--   1. allow for the same asset name when one of the tenant is null the the other isn't
--   2. only if the assets is on the same scope (and not different scopes)
CREATE OR REPLACE FUNCTION asset_name_validation()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.name IN (select name FROM assets_metadata WHERE NEW.kind = assets_metadata.kind AND
        ((NEW.tenant_id is NULL and assets_metadata.tenant_id is NULL) or NEW.tenant_id = assets_metadata.tenant_id) AND NEW.scope = assets_metadata.scope)
    THEN
        RAISE EXCEPTION 'non-% % asset named % already exists', NEW.scope, NEW.kind, NEW.name;
    END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;


DO $$
DECLARE
   -- those the the target assets, they are hard coded in 20170003_model_presets.up.sql
   asset_ids UUID[] := ARRAY[
        '3ab6826a-2123-4b34-a686-64cdf47388ac'::uuid,
        'ecbb7a29-5e49-40f3-8145-15c2ffa4f710'::uuid,
        'f8f08816-53d9-47da-b3c7-2f003b175dc4'::uuid,
        '07e99ad0-aa2f-4dd2-9385-e4a4b2ddc207'::uuid
    ];
    input_asset_id UUID;
    tenant_id_var INT;
    new_asset_id UUID;
BEGIN
    -- prep work 2: remove the assets from the asset_model table
    UPDATE assets_model
    SET compute_asset_id = NULL
    WHERE compute_asset_id::uuid = ANY(asset_ids);

    -- Loop through each asset_id
    FOREACH input_asset_id IN ARRAY asset_ids LOOP
        -- Loop through all tenants that use the asset in either workloads or templates
        FOR tenant_id_var IN (
            SELECT DISTINCT tenant_id
            FROM (
                SELECT w.tenant_id
                FROM workloads w
                JOIN workloads_assets wa ON w.id = wa.workload_id
                WHERE wa.asset_id::UUID = input_asset_id
                UNION
                SELECT am.tenant_id
                FROM assets_metadata am
                JOIN assets_template_content tc ON am.id = tc.template_asset_id
                WHERE tc.contained_asset_id::UUID = input_asset_id
            ) tenants
        ) LOOP
            -- Generate a new asset ID for this tenant
            new_asset_id := gen_random_uuid();
            -- Copy the asset metadata for the new asset (emitting is_policy column as it isn't used and not defined in the gorm model)
            INSERT INTO assets_metadata (id, tenant_id, kind, scope, cluster_id, project_id, name, created_by, created_at, updated_by, updated_at, project_name, department_id, auto_delete, description)
            SELECT new_asset_id, tenant_id_var, kind, 'tenant', cluster_id, project_id, name, created_by, created_at, updated_by, updated_at, project_name, department_id, auto_delete, description
            FROM assets_metadata
            WHERE id::uuid = input_asset_id;
            -- Copy the asset compute content for the new asset (emitting node_type, gpu_request, gpu_compute_request and gpu_compute_limit columns as they aren't used and not defined in the gorm model)
            INSERT INTO assets_compute (asset_id, gpu_request_type, cpu_core_request, cpu_core_limit, cpu_memory_request, cpu_memory_limit, large_shm_request, extended_resources,  gpu_portion_request, gpu_memory_request, mig_profile, gpu_devices_request, gpu_memory_limit, gpu_portion_limit)
            SELECT new_asset_id, gpu_request_type, cpu_core_request, cpu_core_limit, cpu_memory_request, cpu_memory_limit, large_shm_request, extended_resources,  gpu_portion_request, gpu_memory_request, mig_profile, gpu_devices_request, gpu_memory_limit, gpu_portion_limit
            FROM assets_compute
            WHERE asset_id::uuid = input_asset_id;
            -- Update workloads_assets for this tenant
            UPDATE workloads_assets
            SET asset_id = new_asset_id
            WHERE asset_id::uuid = input_asset_id
            AND workload_id IN (
                SELECT id
                    FROM workloads
                    WHERE tenant_id = tenant_id_var
            );
            -- Update template_content for this tenant
            UPDATE assets_template_content
            SET contained_asset_id = new_asset_id
            WHERE contained_asset_id::uuid = input_asset_id
            AND template_asset_id IN (
                SELECT id
                FROM assets_metadata
                WHERE tenant_id = tenant_id_var
            );
        END LOOP;
    END LOOP;
    -- Delete original assets from assets_metadata, assets_compute and last_used_by_workload
    DELETE FROM assets_metadata
    WHERE id::uuid = ANY(asset_ids);
    DELETE FROM assets_compute
    WHERE asset_id::uuid = ANY(asset_ids);
    DELETE FROM last_used_by_workload
    WHERE asset_id::uuid = ANY(asset_ids);
END $$;