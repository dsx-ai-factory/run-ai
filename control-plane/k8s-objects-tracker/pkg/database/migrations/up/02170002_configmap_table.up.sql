-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


CREATE TABLE IF NOT EXISTS configmaps (
                                    id UUID NOT NULL PRIMARY KEY,
                                    tenant_id INT NOT NULL,
                                    cluster_id UUID NOT NULL,
                                    scope_type TEXT NOT NULL,
                                    scope_id TEXT NOT NULL,
                                    name TEXT NOT NULL,
                                    UNIQUE (tenant_id, cluster_id, scope_type, scope_id, name)
);

CREATE INDEX IF NOT EXISTS configmaps_idx ON configmaps(tenant_id, cluster_id, scope_type, scope_id, name);
