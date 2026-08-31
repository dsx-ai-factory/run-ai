-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Fix preemptibility for all rows based on priority value
UPDATE external_workloads_type_config AS ewtc
SET preemptibility = CASE
    WHEN ewtc.priority_id IS NULL THEN NULL
    WHEN (
        SELECT p.value
        FROM priorities p
        WHERE p.id = ewtc.priority_id
    ) >= 100 THEN 'non-preemptible'
    ELSE 'preemptible'
END;