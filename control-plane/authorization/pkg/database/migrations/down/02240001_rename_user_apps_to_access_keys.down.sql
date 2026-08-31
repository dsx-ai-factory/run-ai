-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Rollback: rename access-keys resource type back to user-apps
UPDATE permissions 
SET resource_type = 'user-apps' 
WHERE resource_type = 'access-keys';

