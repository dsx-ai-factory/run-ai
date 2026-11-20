/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum EFeatureFlagsKeys {
    ENABLE_WORKSPACES = 'tenant.show_workspaces',
    LOGOUT_URI = 'tenant.logout_uri',
    INACTIVITY_LOGOUT_TIMEOUT = 'tenant.inactivity_minutes_till_logout',
    ENABLE_LIMIT_OQ_SUBSCRIPTION = 'tenant.limit_quota_over_subscription',
    ENABLE_NEW_POLICY = 'tenant.enable_policies',
    MODEL_CATALOG = 'tenant.enable_model_catalog',
    NIM = 'tenant.enable_nim',
    ALLOW_GPT2_MODEL = 'tenant.allow_gpt2_model',
    WORKLOAD_SUBMISSION_V2 = 'tenant.workload_submission_v2_enabled',
    CLI_V2 = 'tenant.is_new_cli_installer_enabled',
    ENABLE_NODE_POOLS = 'tenant.enable_node_pools',
    WORKLOAD_HYBRID_SUBMISSION = 'tenant.enable_hybrid_submission',
    ENABLE_DATA_VOLUMES = 'tenant.enable_data_volumes',
    ENABLE_TEMPLATES_V2 = 'tenant.enable_templates_v2',
    ENABLE_RESOURCE_OPTIMIZATION = 'tenant.gpu_over_provisioning',
    ENABLE_NGC_INTEGRATION = 'tenant.enable_ngc_integration'
}

export enum EFeatureFlagsNames {
    ENABLE_WORKSPACES = 'workspaces',
    ENABLE_SSO = 'sso',
    LOGOUT_URI = 'logout uri',
    INACTIVITY_LOGOUT_TIMEOUT = 'Inactivity Logout',
    ENABLE_LIMIT_OQ_SUBSCRIPTION = 'Limit quota over/under subscription',
    ENABLE_NEW_POLICY = 'new policy',
    ENABLE_JOBS_GRID = 'legacy jobs grid',
    MODEL_CATALOG = 'model catalog',
    NIM = 'NIM',
    ALLOW_GPT2_MODEL = 'gpt2 model',
    WORKLOAD_SUBMISSION_V2 = 'workload v2 submission',
    CLI_V2 = 'CLI V2',
    CSP_MANAGED = 'CSP Managed or DGXC Roles',
    ENABLE_NODE_POOLS = 'node pools',
    WORKLOAD_HYBRID_SUBMISSION = 'workload hybrid submission',
    ENABLE_DATA_VOLUMES = 'Enable data volumes',
    FLEXIBLE_WORKLOAD_TEMPLATE = 'flexible workload templates',
    RESOURCE_OPTIMIZATION = 'Resource optimization',
    ENABLE_NGC_INTEGRATION = 'NGC integration'
}
