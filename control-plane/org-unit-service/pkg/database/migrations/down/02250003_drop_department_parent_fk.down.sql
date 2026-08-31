-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE department
    ADD CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES department(id);
