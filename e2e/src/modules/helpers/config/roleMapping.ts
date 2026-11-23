/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const PROPERTY_TO_ROLE_MAPPING = {
    system_administrator: 'System administrator',
    department_admin: 'Department administrator',
    editor: 'Editor',
    viewer: 'Viewer',
    research_manager: 'Research manager',
    ml_engineer: 'ML engineer',
    data_volume_administrator: 'Data volume administrator',
    department_viewer: 'Department viewer',
    template_administrator: 'Template administrator',
    compute_resource_administrator: 'Compute resource administrator',
    data_source_administrator: 'Data source administrator',
    credentials_administrator: 'Credentials administrator',
    environment_administrator: 'Environment administrator',
    l1_researcher: 'L1 researcher',
    l2_researcher: 'L2 researcher',
    ai_practitioner: 'AI practitioner',
    application_administrator: 'Application administrator',
    project_administrator: 'Project administrator'
} as const;

export type UserRole = keyof typeof PROPERTY_TO_ROLE_MAPPING;
export type RoleDisplayName = (typeof PROPERTY_TO_ROLE_MAPPING)[UserRole];

export const ROLE_NAME_TO_PROPERTY = Object.fromEntries(
    Object.entries(PROPERTY_TO_ROLE_MAPPING).map(([property, roleName]) => [roleName, property])
) as Record<string, UserRole>;
