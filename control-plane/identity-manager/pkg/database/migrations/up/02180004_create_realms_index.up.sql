-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create unique index if  not exists realms_realm_idx on realms (realm);

create unique index if  not exists realms_tenant_id_idx on realms (tenant_id);