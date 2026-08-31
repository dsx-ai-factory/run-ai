-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS pvc_claim_name_locks(
    id UUID NOT NULL PRIMARY KEY,
    tenant_id INT NOT NULL,
    claim_name TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at timestamp NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS pvc_claim_name_unique_lock ON pvc_claim_name_locks (tenant_id, claim_name);
