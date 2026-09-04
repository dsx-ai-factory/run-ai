-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Update permissions table: rename user-apps resource type to access-keys
UPDATE permissions 
SET resource_type = 'access-keys' 
WHERE resource_type = 'user-apps';

