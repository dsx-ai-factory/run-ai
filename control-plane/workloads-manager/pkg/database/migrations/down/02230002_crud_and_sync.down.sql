-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


DROP TRIGGER IF EXISTS workloads_queue_update_count_trigger ON workloads_queue;

DROP TABLE IF EXISTS workloads_queue;

DROP SEQUENCE IF EXISTS workloads_queue_sequence;

