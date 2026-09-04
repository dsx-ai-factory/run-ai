-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE assets_environment
SET connections = '[{"name":"Chatbot UI","internalToolInfo":{"toolType":"chatbot-ui","containerPort":3000,"connectionType":"ExternalUrl","externalUrlInfo":{}}}]'
WHERE asset_id = '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
