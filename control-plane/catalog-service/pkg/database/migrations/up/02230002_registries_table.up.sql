-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS registries (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id TEXT,
    cluster_id UUID,
    department_id TEXT,
    project_id TEXT,
    scope_type TEXT NOT NULL,
    scope_id TEXT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by TEXT
);

-- Put scope type in front of the key, to allow searching by scope type system with no tenant id
CREATE UNIQUE INDEX IF NOT EXISTS registries_scope_url_idx ON registries (scope_type, tenant_id, scope_id, url);