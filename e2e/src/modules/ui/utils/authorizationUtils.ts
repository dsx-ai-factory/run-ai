/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from './sidenavUtils';

export const SYSTEM_ADMIN_SIDEBAR_LINKS = [
    'Dashboards',
    'clusters',
    'nodes',
    'projects',
    'workloads',
    'environments',
    'data-sources',
    'compute-resource',
    'templates',
    'credentials',
    'events'
];

export const MENU_SUBJECT_LIST = [
    MENU_SUBJECT_NAMES.Analytics,
    MENU_SUBJECT_NAMES.WorkloadsManager,
    MENU_SUBJECT_NAMES.Resources,
    MENU_SUBJECT_NAMES.Organization,
    MENU_SUBJECT_NAMES.Access,
    MENU_SUBJECT_NAMES.Policies,
    MENU_SUBJECT_NAMES.Admin
];

export const MENU_SUBJECT_TO_ITEMS = {
    [MENU_SUBJECT_NAMES.Analytics]: [
        MENU_SUB_ITEMS.Dashboard,
        MENU_SUB_ITEMS.Overview,
        MENU_SUB_ITEMS.Resources
    ],
    [MENU_SUBJECT_NAMES.WorkloadsManager]: [
        MENU_SUB_ITEMS.Workloads,
        MENU_SUB_ITEMS.Environments,
        MENU_SUB_ITEMS.DataSources,
        MENU_SUB_ITEMS.ComputeResource,
        MENU_SUB_ITEMS.Credentials
    ],
    [MENU_SUBJECT_NAMES.Resources]: [MENU_SUB_ITEMS.Clusters, MENU_SUB_ITEMS.NodePools, MENU_SUB_ITEMS.Nodes],
    [MENU_SUBJECT_NAMES.Organization]: [MENU_SUB_ITEMS.Departments, MENU_SUB_ITEMS.Projects],
    [MENU_SUBJECT_NAMES.Access]: [
        MENU_SUB_ITEMS.Users,
        MENU_SUB_ITEMS.ServiceAccounts,
        MENU_SUB_ITEMS.Roles,
        MENU_SUB_ITEMS.AccessRules
    ],
    [MENU_SUBJECT_NAMES.Policies]: [MENU_SUB_ITEMS.Policies],
    [MENU_SUBJECT_NAMES.Admin]: [MENU_SUB_ITEMS.GeneralSettings, MENU_SUB_ITEMS.Audit]
};

export const L1_RESEARCHER_SIDEBAR_LINKS = [
    'Dashboards',
    'projects',
    'workloads',
    'environments',
    'data-sources',
    'compute-resource',
    'templates',
    'credentials'
];

export const L1_RESEARCHER_MENU_SUBJECT_NAMES = [
    MENU_SUBJECT_NAMES.Analytics,
    MENU_SUBJECT_NAMES.WorkloadsManager,
    MENU_SUBJECT_NAMES.Organization,
    MENU_SUBJECT_NAMES.Policies
];

export const L1_RESEARCHER_MENU_SUBJECT_NAMES_TO_ITEMS = {
    [MENU_SUBJECT_NAMES.Analytics]: [MENU_SUB_ITEMS.Dashboard, MENU_SUB_ITEMS.Overview],
    [MENU_SUBJECT_NAMES.WorkloadsManager]: [
        MENU_SUB_ITEMS.Workloads,
        MENU_SUB_ITEMS.Environments,
        MENU_SUB_ITEMS.DataSources,
        MENU_SUB_ITEMS.ComputeResource,
        MENU_SUB_ITEMS.Credentials
    ],
    [MENU_SUBJECT_NAMES.Organization]: [MENU_SUB_ITEMS.Projects],
    [MENU_SUBJECT_NAMES.Policies]: [MENU_SUB_ITEMS.Policies]
};

export const L1_RESEARCHER_MENU_SUBJECT_NAMES_TO_NON_VISIBLE_ITEMS = {
    [MENU_SUBJECT_NAMES.Analytics]: [MENU_SUB_ITEMS.Resources],
    [MENU_SUBJECT_NAMES.Resources]: [MENU_SUB_ITEMS.Clusters, MENU_SUB_ITEMS.Nodes],
    [MENU_SUBJECT_NAMES.Organization]: [MENU_SUB_ITEMS.Departments],
    [MENU_SUBJECT_NAMES.Resources]: [MENU_SUB_ITEMS.NodePools]
};
