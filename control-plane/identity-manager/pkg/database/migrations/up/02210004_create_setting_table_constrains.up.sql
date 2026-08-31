-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE realm_settings DROP CONSTRAINT if exists unique_tenant_setting;
ALTER TABLE realm_settings ADD CONSTRAINT unique_tenant_setting UNIQUE (tenant_id, setting_key);