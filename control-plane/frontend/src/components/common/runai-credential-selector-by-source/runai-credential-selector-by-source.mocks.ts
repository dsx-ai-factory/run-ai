/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  AssetKind,
  AssetSyncStatusEnum,
  type CredentialsListResponseEntry,
  type GenericSecret,
  type GenericSecretCreationRequest,
  type KeyValuePair,
  type NgcApiKey,
  type NgcApiKeyCreationRequest,
  Scope,
} from "@/swagger-models/assets-service-client";

// Shared (non-private) credentials - mixed types
export const mockSharedCredentials: CredentialsListResponseEntry[] = [
  {
    meta: {
      id: "shared-gen1",
      name: "shared-generic-secret-1",
      kind: AssetKind.GenericSecret,
      scope: Scope.Project,
      createdBy: "admin",
      createdAt: "2023-01-01T00:00:00Z",
      updatedBy: "admin",
      updatedAt: "2023-01-01T00:00:00Z",
      private: false,
    },
    spec: {
      genericSecret: { keyId: 0 },
    },
  },
  {
    meta: {
      id: "shared-gen2",
      name: "shared-generic-secret-2",
      kind: AssetKind.GenericSecret,
      scope: Scope.Project,
      createdBy: "admin",
      createdAt: "2023-01-01T00:00:00Z",
      updatedBy: "admin",
      updatedAt: "2023-01-01T00:00:00Z",
      private: false,
    },
    spec: {
      genericSecret: { keyId: 0 },
    },
  },
  {
    meta: {
      id: "shared-ngc1",
      name: "shared-ngc-key-1",
      kind: AssetKind.NgcApiKey,
      scope: Scope.Project,
      createdBy: "admin",
      createdAt: "2023-01-01T00:00:00Z",
      updatedBy: "admin",
      updatedAt: "2023-01-01T00:00:00Z",
      private: false,
    },
    spec: {
      ngcApiKey: { keyId: 0 },
    },
  },
  {
    meta: {
      id: "shared-docker1",
      name: "shared-docker-registry-1",
      kind: AssetKind.DockerRegistry,
      scope: Scope.Project,
      createdBy: "admin",
      createdAt: "2023-01-01T00:00:00Z",
      updatedBy: "admin",
      updatedAt: "2023-01-01T00:00:00Z",
      private: false,
    },
    spec: {
      dockerRegistry: { keyId: 0 },
    },
  },
];

// Private credentials
export const mockPrivateCredentials: CredentialsListResponseEntry[] = [
  {
    meta: {
      id: "private-gen1",
      name: "my-generic-secret-1",
      kind: AssetKind.GenericSecret,
      scope: Scope.Project,
      createdBy: "user1",
      createdAt: "2023-01-01T00:00:00Z",
      updatedBy: "user1",
      updatedAt: "2023-01-01T00:00:00Z",
      private: true,
    },
    spec: {
      genericSecret: { keyId: 0 },
    },
  },
  {
    meta: {
      id: "private-ngc1",
      name: "my-ngc-key-1",
      kind: AssetKind.NgcApiKey,
      scope: Scope.Project,
      createdBy: "user1",
      createdAt: "2023-01-01T00:00:00Z",
      updatedBy: "user1",
      updatedAt: "2023-01-01T00:00:00Z",
      private: true,
    },
    spec: {
      ngcApiKey: { keyId: 0 },
    },
  },
];

// All credentials combined (what listCredentials returns with includePrivate: true)
export const mockAllCredentials: CredentialsListResponseEntry[] = [...mockSharedCredentials, ...mockPrivateCredentials];

// NGC API key with cluster resources (both generic secret and docker registry are created per NGC API key)
export const mockNgcApiKeyWithClusterResources: CredentialsListResponseEntry = {
  meta: {
    id: "ngc-with-resources",
    name: "ngc-api-key-with-resources",
    kind: AssetKind.NgcApiKey,
    scope: Scope.Project,
    createdBy: "admin",
    createdAt: "2023-01-01T00:00:00Z",
    updatedBy: "admin",
    updatedAt: "2023-01-01T00:00:00Z",
    private: false,
  },
  spec: {
    ngcApiKey: { keyId: 0 },
  },
  clusterInfo: {
    resources: [{ name: "ngc-api-key-genericsecret-abc123" }, { name: "ngc-api-key-dockerregistry-xyz789" }],
  },
  status: {
    status: AssetSyncStatusEnum.NoIssuesFound,
  },
};

// Private NGC API key with cluster resources
export const mockPrivateNgcApiKeyWithClusterResources: CredentialsListResponseEntry = {
  meta: {
    id: "private-ngc-with-resources",
    name: "my-ngc-api-key-with-resources",
    kind: AssetKind.NgcApiKey,
    scope: Scope.Project,
    createdBy: "user1",
    createdAt: "2023-01-01T00:00:00Z",
    updatedBy: "user1",
    updatedAt: "2023-01-01T00:00:00Z",
    private: true,
  },
  spec: {
    ngcApiKey: { keyId: 0 },
  },
  clusterInfo: {
    resources: [{ name: "my-ngc-genericsecret-abc123" }, { name: "my-ngc-dockerregistry-xyz789" }],
  },
  status: {
    status: AssetSyncStatusEnum.NoIssuesFound,
  },
};

export const mockDefaultGenericSecretItems: KeyValuePair[] = [
  { key: "API_KEY", value: "" },
  { key: "SECRET_TOKEN", value: "" },
];

export const mockCredentialModel: GenericSecretCreationRequest = {
  meta: {
    name: "test-credential",
    scope: Scope.Project,
  },
  spec: {
    keyValuePairs: [{ key: "API_KEY", value: "test-value" }],
    existingSecretName: null,
  },
};

export const mockGenericSecret: GenericSecret = {
  meta: {
    ...mockCredentialModel.meta,
    id: "",
    kind: AssetKind.GenericSecret,
    createdBy: "",
    createdAt: "",
    updatedBy: "",
    updatedAt: "",
  },
  spec: {
    ...mockCredentialModel.spec,
    keyId: 0,
  },
};

export const mockNgcApiKeyCreationRequest: NgcApiKeyCreationRequest = {
  meta: {
    name: "test-ngc-credential",
    scope: Scope.Project,
  },
  spec: {
    ngc_api_key: "nvapi-test-key",
  },
};

export const mockNgcApiKey: NgcApiKey = {
  meta: {
    ...mockNgcApiKeyCreationRequest.meta,
    id: "",
    kind: AssetKind.NgcApiKey,
    createdBy: "",
    createdAt: "",
    updatedBy: "",
    updatedAt: "",
  },
  spec: {
    ...mockNgcApiKeyCreationRequest.spec,
    keyId: 0,
  },
};
