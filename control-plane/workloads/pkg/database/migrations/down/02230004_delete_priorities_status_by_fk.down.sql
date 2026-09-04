-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE priority_statuses DROP CONSTRAINT fk_priority;
ALTER TABLE priority_statuses ADD CONSTRAINT fk_priority FOREIGN KEY (priority_id) REFERENCES priorities(id);
