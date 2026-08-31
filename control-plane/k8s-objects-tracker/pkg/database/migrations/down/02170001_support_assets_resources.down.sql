-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE pvcs    DROP COLUMN IF EXISTS origin;
ALTER TABLE secrets DROP COLUMN IF EXISTS origin;

