-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE assets_environment
SET args = '--NotebookApp.token='''''
WHERE asset_id = '17af3f04-f632-4a43-9ceb-ae8c062eb250'
  AND args = '--NotebookApp.token=''';
