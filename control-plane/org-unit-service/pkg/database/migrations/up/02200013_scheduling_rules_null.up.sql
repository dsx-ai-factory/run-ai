-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE org_unit_data
SET
    interactive_job_time_limit = CASE
        WHEN interactive_job_time_limit = 0 THEN NULL
        ELSE interactive_job_time_limit
        END,
    interactive_job_time_limit_effective = CASE
        WHEN interactive_job_time_limit_effective = 0 THEN NULL
        ELSE interactive_job_time_limit_effective
        END,
    interactive_job_max_idle_duration = CASE
        WHEN interactive_job_max_idle_duration = 0 THEN NULL
        ELSE interactive_job_max_idle_duration
        END,
    interactive_job_max_idle_duration_effective = CASE
        WHEN interactive_job_max_idle_duration_effective = 0 THEN NULL
        ELSE interactive_job_max_idle_duration_effective
        END,
    interactive_job_preempt_idle_duration = CASE
        WHEN interactive_job_preempt_idle_duration = 0 THEN NULL
        ELSE interactive_job_preempt_idle_duration
        END,
    interactive_job_preempt_idle_duration_effective = CASE
        WHEN interactive_job_preempt_idle_duration_effective = 0 THEN NULL
        ELSE interactive_job_preempt_idle_duration_effective
        END,
    training_job_time_limit = CASE
        WHEN training_job_time_limit = 0 THEN NULL
        ELSE training_job_time_limit
        END,
    training_job_time_limit_effective = CASE
        WHEN training_job_time_limit_effective = 0 THEN NULL
        ELSE training_job_time_limit_effective
        END,
    training_job_max_idle_duration = CASE
        WHEN training_job_max_idle_duration = 0 THEN NULL
        ELSE training_job_max_idle_duration
        END,
    training_job_max_idle_duration_effective = CASE
        WHEN training_job_max_idle_duration_effective = 0 THEN NULL
        ELSE training_job_max_idle_duration_effective
        END
WHERE deleted_at IS NULL
  AND (
    interactive_job_time_limit = 0 OR
    interactive_job_time_limit_effective = 0 OR
    interactive_job_max_idle_duration = 0 OR
    interactive_job_max_idle_duration_effective = 0 OR
    interactive_job_preempt_idle_duration = 0 OR
    interactive_job_preempt_idle_duration_effective = 0 OR
    training_job_time_limit = 0 OR
    training_job_time_limit_effective = 0 OR
    training_job_max_idle_duration = 0 OR
    training_job_max_idle_duration_effective = 0
    );
