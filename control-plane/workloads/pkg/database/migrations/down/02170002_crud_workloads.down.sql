-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TRIGGER IF EXISTS crud_workloads_update_count_trigger ON crud_workloads;

DROP TABLE IF EXISTS crud_workloads;

DROP SEQUENCE IF EXISTS crud_workloads_sequence;
