/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import { ScopeType } from "@/swagger-models/authorization-client";
import type { IScopeModel } from "@/models/global.model";
import type { IOrgTreeNode } from "@/models/org-tree.model";
import { TenantStatusEnum, TenantContractTypeEnum } from "@/swagger-models/tenants-manager-client";

export const MOCK_IDS = {
  TENANT_ID: "1001",
  CLUSTER_ID: "a0d63713-4a9f-49b5-b36f-42af6f7f7bb6",
  DEPARTMENT_ID: "4500000",
  PROJECT_ONE_ID: "4500001",
  PROJECT_TWO_ID: "4500002",
} as const;

export const MOCK_NAMES = {
  TENANT_NAME: "runai",
  CLUSTER_NAME: "test-cluster",
  DEPARTMENT_NAME: "default",
  PROJECT_ONE_NAME: "project-one",
  PROJECT_TWO_NAME: "project-two",
} as const;

export const MOCK_PATHS = {
  TENANT: MOCK_NAMES.TENANT_NAME,
  CLUSTER: `${MOCK_NAMES.TENANT_NAME}/${MOCK_NAMES.CLUSTER_NAME}`,
  DEPARTMENT: `${MOCK_NAMES.TENANT_NAME}/${MOCK_NAMES.CLUSTER_NAME}/${MOCK_NAMES.DEPARTMENT_NAME}`,
  PROJECT_ONE: `${MOCK_NAMES.TENANT_NAME}/${MOCK_NAMES.CLUSTER_NAME}/${MOCK_NAMES.DEPARTMENT_NAME}/${MOCK_NAMES.PROJECT_ONE_NAME}`,
  PROJECT_TWO: `${MOCK_NAMES.TENANT_NAME}/${MOCK_NAMES.CLUSTER_NAME}/${MOCK_NAMES.DEPARTMENT_NAME}/${MOCK_NAMES.PROJECT_TWO_NAME}`,
} as const;

export const mockTreeData: IOrgTreeNode[] = [
  {
    id: MOCK_IDS.TENANT_ID,
    path: MOCK_PATHS.TENANT,
    type: ScopeType.Tenant,
    children: [
      {
        id: MOCK_IDS.CLUSTER_ID,
        path: MOCK_PATHS.CLUSTER,
        type: ScopeType.Cluster,
        children: [
          {
            id: MOCK_IDS.DEPARTMENT_ID,
            path: MOCK_PATHS.DEPARTMENT,
            type: ScopeType.Department,
            children: [
              {
                id: MOCK_IDS.PROJECT_TWO_ID,
                path: MOCK_PATHS.PROJECT_TWO,
                type: ScopeType.Project,
                children: [],
                icon: "ph-bold ph-suitcase",
                label: MOCK_NAMES.PROJECT_TWO_NAME,
              },
              {
                id: MOCK_IDS.PROJECT_ONE_ID,
                path: MOCK_PATHS.PROJECT_ONE,
                type: ScopeType.Project,
                children: [],
                icon: "ph-bold ph-suitcase",
                label: MOCK_NAMES.PROJECT_ONE_NAME,
              },
            ],
            icon: "ph-bold ph-garage",
            label: MOCK_NAMES.DEPARTMENT_NAME,
          },
        ],
        icon: "ph-bold ph-graph",
        label: MOCK_NAMES.CLUSTER_NAME,
      },
    ],
    icon: "ph-bold ph-tree-structure",
    label: MOCK_NAMES.TENANT_NAME,
  },
];

export const mockOrgUnits: OrgUnit[] = [
  {
    id: MOCK_IDS.TENANT_ID,
    type: "tenant",
    name: MOCK_NAMES.TENANT_NAME,
  },
  {
    id: MOCK_IDS.CLUSTER_ID,
    type: "cluster",
    name: MOCK_NAMES.CLUSTER_NAME,
    parentId: MOCK_IDS.TENANT_ID,
    parentType: "tenant",
  },
  {
    id: MOCK_IDS.DEPARTMENT_ID,
    type: "department",
    name: MOCK_NAMES.DEPARTMENT_NAME,
    parentId: MOCK_IDS.CLUSTER_ID,
    parentType: "cluster",
  },
  {
    id: MOCK_IDS.PROJECT_TWO_ID,
    type: "project",
    name: MOCK_NAMES.PROJECT_TWO_NAME,
    parentId: MOCK_IDS.DEPARTMENT_ID,
    parentType: "department",
  },
  {
    id: MOCK_IDS.PROJECT_ONE_ID,
    type: "project",
    name: MOCK_NAMES.PROJECT_ONE_NAME,
    parentId: MOCK_IDS.DEPARTMENT_ID,
    parentType: "department",
  },
];

export const mockTenant = {
  id: parseInt(MOCK_IDS.TENANT_ID),
  mimirId: "3af1e53b-cb30-41c1-930c-be214d1782a7",
  uuid: "3af1e53b-cb30-41c1-930c-be214d1782a7",
  name: MOCK_NAMES.TENANT_NAME,
  displayName: MOCK_NAMES.TENANT_NAME,
  status: TenantStatusEnum.Ready,
  createdAt: "2025-05-28T10:36:31.912731Z",
  updatedAt: "2025-05-28T10:36:35.510796Z",
  contractType: TenantContractTypeEnum.Normal,
};

export const mockClusters = [
  {
    uuid: MOCK_IDS.CLUSTER_ID,
    name: MOCK_NAMES.CLUSTER_NAME,
  },
];

export const mockScopeModels = {
  empty: {
    scope: null,
    projectId: null,
    departmentId: null,
    clusterId: null,
  } as IScopeModel,

  project: {
    scope: ScopeType.Project,
    projectId: parseInt(MOCK_IDS.PROJECT_ONE_ID),
    departmentId: MOCK_IDS.DEPARTMENT_ID,
    clusterId: MOCK_IDS.CLUSTER_ID,
  } as IScopeModel,

  department: {
    scope: ScopeType.Department,
    departmentId: MOCK_IDS.DEPARTMENT_ID,
    clusterId: MOCK_IDS.CLUSTER_ID,
  } as IScopeModel,

  cluster: {
    scope: ScopeType.Cluster,
    departmentId: null,
    clusterId: MOCK_IDS.CLUSTER_ID,
  } as IScopeModel,
};

export const mockAllowedScopes = {
  full: {
    tenant: MOCK_NAMES.TENANT_NAME,
    cluster: { [MOCK_IDS.CLUSTER_ID]: MOCK_NAMES.CLUSTER_NAME },
    department: { [MOCK_IDS.DEPARTMENT_ID]: MOCK_NAMES.DEPARTMENT_NAME },
    project: {
      [MOCK_IDS.PROJECT_ONE_ID]: MOCK_NAMES.PROJECT_ONE_NAME,
      [MOCK_IDS.PROJECT_TWO_ID]: MOCK_NAMES.PROJECT_TWO_NAME,
    },
  },

  limited: {
    tenant: MOCK_NAMES.TENANT_NAME,
    cluster: { [MOCK_IDS.CLUSTER_ID]: MOCK_NAMES.CLUSTER_NAME },
    department: { [MOCK_IDS.DEPARTMENT_ID]: MOCK_NAMES.DEPARTMENT_NAME },
    project: { [MOCK_IDS.PROJECT_ONE_ID]: MOCK_NAMES.PROJECT_ONE_NAME },
  },
};

export const mockExpandedPermittedScopes = {
  tenants: [MOCK_IDS.TENANT_ID],
  clusters: [MOCK_IDS.CLUSTER_ID],
  departments: [MOCK_IDS.DEPARTMENT_ID],
  projects: [MOCK_IDS.PROJECT_ONE_ID, MOCK_IDS.PROJECT_TWO_ID],
};

export const mockScopesSupportedVersions = {
  tenant: "1.0.0",
  cluster: "1.1.0",
  department: "1.2.0",
  project: "1.3.0",
};

export const mockForcedDisabledScopes = {
  [ScopeType.Tenant]: true,
  [ScopeType.Cluster]: false,
  [ScopeType.Department]: false,
  [ScopeType.Project]: false,
};

export const mockNodePathMappings = {
  [MOCK_IDS.TENANT_ID]: MOCK_PATHS.TENANT,
  [MOCK_IDS.CLUSTER_ID]: MOCK_PATHS.CLUSTER,
  [MOCK_IDS.DEPARTMENT_ID]: MOCK_PATHS.DEPARTMENT,
  [MOCK_IDS.PROJECT_ONE_ID]: MOCK_PATHS.PROJECT_ONE,
  [MOCK_IDS.PROJECT_TWO_ID]: MOCK_PATHS.PROJECT_TWO,
};

export const mockDefaultFactoryProps = {
  readOnly: false,
  kindOfAsset: "environment",
  scopeModel: mockScopeModels.empty,
  allowedScopes: mockAllowedScopes.full,
  selectedClusterId: MOCK_IDS.CLUSTER_ID,
};

export const mockInformationText = "<strong>Important:</strong> This is test information";

export const runaiScopeSectionMock = {
  MOCK_IDS,
  MOCK_NAMES,
  MOCK_PATHS,
  mockTreeData,
  mockOrgUnits,
  mockTenant,
  mockClusters,
  mockScopeModels,
  mockAllowedScopes,
  mockExpandedPermittedScopes,
  mockScopesSupportedVersions,
  mockForcedDisabledScopes,
  mockNodePathMappings,
  mockDefaultFactoryProps,
  mockInformationText,
};
