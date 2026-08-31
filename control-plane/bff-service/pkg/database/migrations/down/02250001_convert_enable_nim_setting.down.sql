-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Revert JSON format {enabled: true/false} back to boolean

-- Update the value column to convert JSON object back to boolean
UPDATE settings
SET value = to_jsonb((value->>'enabled')::boolean)
WHERE key = 'enableNIM'
  AND value IS NOT NULL
  AND value ? 'enabled';
