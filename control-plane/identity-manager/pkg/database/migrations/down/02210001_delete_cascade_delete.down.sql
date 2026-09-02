-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE reconciler_versions DROP CONSTRAINT if exists fk_tenant_id;

ALTER TABLE idps DROP CONSTRAINT if exists fk_tenant_id;
