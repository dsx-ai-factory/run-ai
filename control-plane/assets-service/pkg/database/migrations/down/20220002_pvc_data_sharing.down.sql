-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TRIGGER IF EXISTS after_delete_pvc ON assets_metadata;

DROP FUNCTION IF EXISTS delete_pvc_datavolume_rows();

DROP TABLE IF EXISTS assets_pvc_datavolume;

ALTER TABLE assets_pvc DROP COLUMN IF EXISTS data_sharing;