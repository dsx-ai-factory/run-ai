-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

---- Redo of 0010_compliance_to_json.down.sql in idempotent in case it didn't run
DELETE from policy_compliance WHERE aligned = false;

ALTER TABLE policy_compliance DROP COLUMN IF EXISTS reason;

ALTER TABLE policy_compliance ADD COLUMN IF NOT EXISTS reason TEXT;

