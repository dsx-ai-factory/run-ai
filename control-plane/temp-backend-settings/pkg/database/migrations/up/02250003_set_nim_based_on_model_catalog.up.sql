-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

INSERT INTO tenant_settings (tenant_id, key, value, type)
SELECT tenant_id, 'tenant.enable_nim', 'false', 'Boolean'
FROM tenant_settings
WHERE key = 'tenant.enable_model_catalog'
  AND value = 'false'
  AND tenant_id NOT IN (
    SELECT tenant_id
    FROM tenant_settings
    WHERE key = 'tenant.enable_nim'
  );
