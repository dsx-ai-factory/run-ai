-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop index audit_logs_tenant_entity_index;
drop index audit_logs_tenant_subject_index;
drop index audit_logs_tenant_time_index;

drop table if exists audit_logs;
