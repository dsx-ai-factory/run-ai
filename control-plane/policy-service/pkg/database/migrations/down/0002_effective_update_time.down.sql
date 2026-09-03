-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE effective_policies DROP COLUMN IF EXISTS updated_at;

DROP TRIGGER IF EXISTS set_timestamp ON effective_policies;
