-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


DO $$
DECLARE
    nemo_uuid := '844edfab-940c-47a6-81aa-60d8a87dd72f';
    bionemo_uuid := '1312b1f8-223c-42f9-8076-64cac7103f66';
    pytorch_uuid := '6618d68d-3aa1-43aa-ac3a-0a3f60ee5364';

BEGIN
BEGIN

    -- Delete the new preset from the support table
    DELETE FROM assets_supported_workload_types WHERE asset_id::uuid IN (
               nemo_uuid,
               bionemo_uuid,
               pytorch_uuid);

    -- Delete the new preset from the environment table
    DELETE FROM assets_environment WHERE asset_id::uuid IN (
               nemo_uuid,
               bionemo_uuid,
               uuid_environment_rstudio);

    -- Delete the new preset from the metadata table
    DELETE FROM assets_metadata WHERE id::uuid IN (
               nemo_uuid,
               bionemo_uuid,
               pytorch_uuid);
END;
END $$;
