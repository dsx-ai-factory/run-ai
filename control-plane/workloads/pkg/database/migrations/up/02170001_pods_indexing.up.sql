-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE INDEX IF NOT EXISTS idx_pods_node ON pods (node_name);
CREATE INDEX IF NOT EXISTS idx_pods_nodepool ON pods (current_nodepool);
CREATE INDEX IF NOT EXISTS idx_pods_created_at ON pods (created_at);
CREATE INDEX IF NOT EXISTS idx_pods_completed_at ON pods (completed_at) where completed_at is not null;
CREATE INDEX IF NOT EXISTS idx_pods_deleted_at ON pods (deleted_at) where deleted_at is not null;
