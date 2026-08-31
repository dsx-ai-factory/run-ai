-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE pvcs    ADD COLUMN IF NOT EXISTS origin TEXT DEFAULT 'existing';
ALTER TABLE secrets ADD COLUMN IF NOT EXISTS origin TEXT DEFAULT 'existing';

