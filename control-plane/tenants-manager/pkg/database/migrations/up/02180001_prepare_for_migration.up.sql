-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

Delete from tenants where status= 'Failed' and name = 'runai';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS "smg" BOOLEAN NOT NULL DEFAULT(TRUE);
