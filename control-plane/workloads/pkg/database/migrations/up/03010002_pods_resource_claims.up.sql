-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Per-pod DRA usage, reported as-is by cluster-sync. One jsonb array per pod;
-- each element mirrors a spec.resourceClaims[] entry enriched with the actual
-- bound ResourceClaim name once status publishes it. No index: matches the
-- existing jsonb-column convention on this table (tolerations, node_affinity).
-- The reverse-view lookup is added in a later task; an index can be introduced
-- there if profiling warrants.
ALTER TABLE pods ADD COLUMN IF NOT EXISTS resource_claims jsonb;

