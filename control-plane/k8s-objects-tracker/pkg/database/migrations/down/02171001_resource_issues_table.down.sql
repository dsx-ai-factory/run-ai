-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TRIGGER IF EXISTS delete_pvc_issues_trigger ON pvcs;

DROP TRIGGER IF EXISTS delete_secret_issues_trigger ON secrets;

DROP TRIGGER IF EXISTS delete_configmap_issues_trigger ON configmaps;

DROP FUNCTION IF EXISTS delete_pvc_issues;

DROP FUNCTION IF EXISTS delete_secret_issues;

DROP FUNCTION IF EXISTS delete_configmap_issues;

DROP TABLE IF EXISTS resource_issues;
