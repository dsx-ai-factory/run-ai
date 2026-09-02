-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Convert boolean enable_nim setting to JSON format {enabled: true/false}

-- Update the value column to convert boolean to JSON object
-- Only update if the value hasn't already been converted (idempotent)
UPDATE settings
SET value = jsonb_build_object('enabled', (value::text)::boolean)
WHERE key = 'enableNIM'
  AND value IS NOT NULL
  AND NOT (value ? 'enabled');
