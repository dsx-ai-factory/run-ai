-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

INSERT INTO realm_settings (id, tenant_id, setting_key, setting_value, updated_at, updated_by_customer)
SELECT
    gen_random_uuid() AS id,
    tenant_id,
    'excludeGroupsFromToken' AS setting_key,  
    '{"enabled": true}'::jsonb AS setting_value,      
    updated_at, 
    true AS updated_by_customer    
FROM realms
WHERE exclude_groups_from_tokens = true
ON CONFLICT (tenant_id, setting_key) DO NOTHING;