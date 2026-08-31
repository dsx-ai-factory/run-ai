-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS ep_by_tenant_and_id;

ALTER TABLE effective_policies 
  DROP COLUMN IF EXISTS tenant_id;

