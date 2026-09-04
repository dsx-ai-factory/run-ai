/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

// models
import type {
  AccessKeyUpdateRequest,
  CredentialsListResponseEntry,
  DockerRegistryUpdateRequest,
  GenericSecretUpdateRequest,
  NgcApiKeyUpdateRequest,
  PasswordUpdateRequest,
} from "@/swagger-models/assets-service-client";
import type { Action } from "@/swagger-models/authorization-client";
import { ResourceType, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IAssetsFilter } from "@/models/filter.model";

// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";

// constants
import type { TCredentialKinds } from "@/models/credential.model";
import { permissionService } from "@/services/authorization/permission.service/permission.service";

export const useCredentialStore = defineStore("Credential", {
  state: () => ({
    credentials: [] as Array<CredentialsListResponseEntry>,
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,
  }),
  getters: {
    credentialsList(): Array<CredentialsListResponseEntry> {
      return this.credentials;
    },
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Credentials);
    },
    async loadCredentials(filterBy: IAssetsFilter = {}): Promise<Array<CredentialsListResponseEntry>> {
      this.credentials = await credentialService.listCredentials(filterBy);
      return this.credentials;
    },
    async remove(credentialId: string, kind: TCredentialKinds): Promise<void> {
      await credentialService.remove(credentialId, kind);
      await this.loadCredentials({ statusInfo: true, usageInfo: true });
    },
    async updateUserpass(userpassId: string, userpass: PasswordUpdateRequest): Promise<void> {
      await credentialService.updateUserpass(userpassId, userpass);
    },
    async updateAccessKey(accessKeyId: string, accessKey: AccessKeyUpdateRequest): Promise<void> {
      await credentialService.updateAccessKey(accessKeyId, accessKey);
    },
    async updateDockerRegistry(dockerRegistryId: string, dockerRegistry: DockerRegistryUpdateRequest): Promise<void> {
      await credentialService.updateDockerRegistry(dockerRegistryId, dockerRegistry);
    },
    async updateGenericSecret(genericSecretId: string, genericSecret: GenericSecretUpdateRequest): Promise<void> {
      await credentialService.updateGenericSecret(genericSecretId, genericSecret);
    },
    async updateNgcApiKey(ngcApiKeyId: string, ngcApiKey: NgcApiKeyUpdateRequest): Promise<void> {
      await credentialService.updateNgcApiKey(ngcApiKeyId, ngcApiKey);
    },
  },
});
