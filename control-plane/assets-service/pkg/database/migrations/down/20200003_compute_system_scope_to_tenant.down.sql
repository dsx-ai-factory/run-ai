-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
DECLARE
    CREATED_AT timestamp := now();
    CREATED_BY text := 'Run:ai';

    uuid_compute_gpu_memory_16g uuid := '3ab6826a-2123-4b34-a686-64cdf47388ac';
    uuid_compute_gpu_memory_20g uuid := 'ecbb7a29-5e49-40f3-8145-15c2ffa4f710';
    uuid_compute_gpu_memory_40g uuid := 'f8f08816-53d9-47da-b3c7-2f003b175dc4';
    uuid_compute_gpu_memory_80g uuid := '07e99ad0-aa2f-4dd2-9385-e4a4b2ddc207';

    uuid_model_falcon_7b uuid := 'c6accd0c-3652-4975-ba11-7c9e3e05c7dd';
    uuid_model_falcon_40b uuid := 'b074bd87-6c31-41fa-89f2-16b470d46d29';
    uuid_model_llama_2_7b uuid := '45ab1c3e-df3b-4b8e-819e-b477fac1302f';
    uuid_model_llama_2_13b uuid := '5be388f5-2420-4aaa-b63d-73ee65b69211';
BEGIN
        -- reviving the system compute assets (as defined in the 20170003_model_presets migration script)
        INSERT INTO assets_metadata ("id", "kind", "scope", "name", "created_by", "created_at", "updated_by", "updated_at") VALUES
            (uuid_compute_gpu_memory_16g, 'compute', 'system', 'gpu-memory-16g', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT),
            (uuid_compute_gpu_memory_20g, 'compute', 'system', 'gpu-memory-20g', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT),
            (uuid_compute_gpu_memory_40g, 'compute', 'system', 'gpu-memory-40g', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT),
            (uuid_compute_gpu_memory_80g, 'compute', 'system', 'gpu-memory-80g', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT);
        INSERT INTO assets_compute ("asset_id", "cpu_core_request", "cpu_memory_request", "gpu_request_type", "gpu_memory_request", "gpu_portion_request", "gpu_devices_request") VALUES
            (uuid_compute_gpu_memory_16g, 1, '16G', 'memory', '16G', NULL, 1),
            (uuid_compute_gpu_memory_20g, 1, '20G', 'memory', '20G', NULL, 1),
            (uuid_compute_gpu_memory_40g, 1, '40G', 'memory', '40G', NULL, 1),
            (uuid_compute_gpu_memory_80g, 1, '80G', 'memory', '80G', NULL, 1);

        -- add them back to the models (as defined in the 20170003_model_presets migration script)
        UPDATE assets_model
        SET compute_asset_id = uuid_compute_gpu_memory_16g
        WHERE asset_id::uuid = uuid_model_falcon_7b;

        UPDATE assets_model
        SET compute_asset_id = uuid_compute_gpu_memory_16g
        WHERE asset_id::uuid = uuid_model_falcon_7b;

        UPDATE assets_model
        SET compute_asset_id = uuid_compute_gpu_memory_40g
        WHERE asset_id::uuid = uuid_model_llama_2_13b;
END $$;

-- last but not least, we need to set the asset_name_validation trigger as it was in 20190005_asset_name_validation
CREATE OR REPLACE FUNCTION asset_name_validation()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.name IN (select name FROM assets_metadata WHERE NEW.kind = assets_metadata.kind AND
        (assets_metadata.tenant_id is NULL or NEW.tenant_id = assets_metadata.tenant_id) AND NEW.scope != assets_metadata.scope)
    THEN
        RAISE EXCEPTION 'non-% % asset named % already exists', NEW.scope, NEW.kind, NEW.name;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;