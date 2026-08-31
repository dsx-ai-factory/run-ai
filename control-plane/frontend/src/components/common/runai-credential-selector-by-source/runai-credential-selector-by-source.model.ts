/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TCredentialKinds } from "@/models/credential.model";
import {
  AssetKind,
  type GenericSecretCreationRequest,
  type NgcApiKeyCreationRequest,
  type DockerRegistryCreationRequest,
  type KeyValuePair,
  type CredentialsListResponseEntry,
} from "@/swagger-models/assets-service-client";
import type { credentialService } from "@/services/control-plane/credential.service/credential.service";

export type TCredentialNameResolver = (credential: CredentialsListResponseEntry) => string;

export function createResourceNameResolver(
  ngcResourcePreference: typeof AssetKind.GenericSecret | typeof AssetKind.DockerRegistry = AssetKind.DockerRegistry,
): TCredentialNameResolver {
  return (credential: CredentialsListResponseEntry): string => {
    if (credential.meta.private) return credential.meta.name;

    const resourceNameKey =
      credential.meta.kind === AssetKind.NgcApiKey
        ? ngcResourcePreference.toLocaleLowerCase()
        : credential.meta.kind.toLocaleLowerCase();

    return (
      credential.clusterInfo?.resources?.find((resource) => resource.name?.includes(resourceNameKey))?.name ||
      credential.meta.name
    );
  };
}

export enum ERunaiCredentialSource {
  SharedSecret = "shared-secret",
  MyCredentials = "my-credentials",
}

export enum ERunaiCredentialType {
  NgcApiKey = "ngc-api-key",
  GenericSecret = "generic-secret",
  DockerRegistry = "docker-registry",
}

export enum ECredentialSelectorEvents {
  UpdateModelValue = "update:model-value",
}

export interface ICredentialSelectorModel {
  source: ERunaiCredentialSource | undefined;
  type?: ERunaiCredentialType | undefined;
  credentialName: string;
  id?: string;
}

export interface ICredentialOption {
  name: string;
  source: ERunaiCredentialSource;
  type: ERunaiCredentialType;
  id: string;
  disable?: boolean;
  disabledTooltip?: string;
}

export interface ICredentialTypeConfig {
  label: string;
  assetKind: TCredentialKinds;
  createFn: (
    model: GenericSecretCreationRequest | NgcApiKeyCreationRequest | DockerRegistryCreationRequest,
    service: typeof credentialService,
  ) => Promise<void>;
  supportTooltip?: string;
}

export interface ISourceSelectConfig {
  hidden?: boolean;
  disable?: boolean;
  showPersonalCredentialsOption?: boolean;
}

export interface IDisabledTypeConfig {
  tooltip?: string;
}

export interface ITypeSelectConfig {
  hidden?: boolean;
  disable?: boolean;
  allowedTypes?: ERunaiCredentialType[];
  disabledTypes?: Partial<Record<ERunaiCredentialType, IDisabledTypeConfig>>;
}

export interface ICredentialSelectConfig {
  required?: boolean;
  showCreateNewBtn?: boolean;
  supportMultipleGenericSecretItems?: boolean;
  defaultGenericSecretItems?: KeyValuePair[];
}

/** Optional grid classes per field. Omitted keys fall back to `DEFAULT_CREDENTIAL_SELECTOR_GRID`. */
export interface IGridConfig {
  sourceSelect?: string;
  typeSelect?: string;
  credentialListSelect?: string;
}

export const DEFAULT_CREDENTIAL_SELECTOR_GRID = {
  sourceSelect: "col-3 q-mr-lg",
  typeSelect: "col-3 q-mr-lg",
  credentialListSelect: "col",
} as const;

export type TResolvedCredentialSelectorGrid = {
  sourceSelect: string;
  typeSelect: string;
  credentialListSelect: string;
};

export interface ICredentialChangedEvent {
  credentialName: string | null;
  credentialSource: ERunaiCredentialSource;
}
