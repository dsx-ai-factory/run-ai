-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE assets_environment
SET image = 'quay.io/jupyter/scipy-notebook'
WHERE asset_id IN ('102d4a7b-2345-47d4-bcc1-52347eef1b4e', '17af3f04-f632-4a43-9ceb-ae8c062eb250')
  AND image = 'jupyter/scipy-notebook';
