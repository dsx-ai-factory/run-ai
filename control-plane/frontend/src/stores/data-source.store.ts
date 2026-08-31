/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

// models
import {
  type DatasourceListResponseEntry,
  type GitAsset,
  type NFSAsset,
  type HostPathAsset,
  type S3Asset,
  type PVCAsset,
  type NFSUpdateRequest,
  type GitUpdateRequest,
  type PVCUpdateRequest,
  type HostPathUpdateRequest,
  type S3UpdateRequest,
  type ConfigMapUpdateRequest,
  type ConfigMapAsset,
  type HttpResponse,
  type SecretAsset,
  type SecretAssetUpdateRequest,
  AssetKind,
} from "@/swagger-models/assets-service-client";
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IAssetsFilter } from "@/models/filter.model";
import type { FallbackPolicy } from "@/swagger-models/policy-service-client";

// services
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { arrayToObjectMap } from "@/utils/common.util";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import type { TDataSourceKinds } from "@/models/data-source.model";
// Stores
import { useSettingStore } from "./setting.store";

export const useDataSourceStore = defineStore("DataSource", {
  state: () => ({
    dataSources: [] as Array<DatasourceListResponseEntry>,
    selectedHostPath: null as HostPathAsset | null,
    selectedS3: null as S3Asset | null,
    selectedPVC: null as PVCAsset | null,
    selectedGit: null as GitAsset | null,
    selectedNFS: null as NFSAsset | null,
    selectedConfigMap: null as ConfigMapAsset | null,
    selectedSecretVolume: null as SecretAsset | null,
    dataSourcesActionPermissions: {} as Record<ResourceType, Record<Action, PermittedScopes>>,
    fallbackPolicy: null as FallbackPolicy | null,
    fallbackPolicyLoadingError: false as boolean,
  }),
  getters: {
    dataSourceList(): Array<DatasourceListResponseEntry> {
      return this.dataSources;
    },
    dataSourcesMapByIds(): Record<string, DatasourceListResponseEntry> {
      return arrayToObjectMap<DatasourceListResponseEntry>(
        this.dataSources,
        (dataSource: DatasourceListResponseEntry) => dataSource.meta.id,
      );
    },
  },
  actions: {
    async loadFallbackPolicy(): Promise<void> {
      try {
        this.fallbackPolicy = await policyService.getTenantFallbackPolicy();
        this.fallbackPolicyLoadingError = false;
      } catch (e: unknown) {
        this.fallbackPolicyLoadingError = true;
        this.fallbackPolicy = null;
        console.error("failed to load tenant fallback policy, error:", e);
      }
    },
    async getAvailableDataSources(): Promise<Set<TDataSourceKinds>> {
      // lazy loading, if it does not exists in the store load it
      if (!this.fallbackPolicy) await this.loadFallbackPolicy();

      const dsAllowedTypes = new Set([
        AssetKind.Nfs,
        AssetKind.Pvc,
        AssetKind.S3,
        AssetKind.Git,
        AssetKind.HostPath,
        AssetKind.ConfigMap,
        AssetKind.SecretVolume,
        AssetKind.DataVolume,
      ]);

      if (this.fallbackPolicy?.effective?.rules?.storage) {
        const { pvc, nfs, s3, git, hostPath, secretVolume, configMapVolume, dataVolume } =
          this.fallbackPolicy.effective.rules.storage;
        [
          { kind: AssetKind.Pvc, rule: pvc },
          { kind: AssetKind.Nfs, rule: nfs },
          { kind: AssetKind.S3, rule: s3 },
          { kind: AssetKind.Git, rule: git },
          { kind: AssetKind.HostPath, rule: hostPath },
          { kind: AssetKind.SecretVolume, rule: secretVolume },
          { kind: AssetKind.ConfigMap, rule: configMapVolume },
          { kind: AssetKind.DataVolume, rule: dataVolume },
        ].forEach(({ kind, rule }) => {
          if (rule?.instances?.canAdd === false) {
            dsAllowedTypes.delete(kind);
          }
        });
      }

      if (!useSettingStore().isDataVolumeEnabled) {
        dsAllowedTypes.delete(AssetKind.DataVolume);
      }

      return dsAllowedTypes;
    },
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.dataSourcesActionPermissions).length) return;
      const dsResourceTypes: Array<ResourceType> = [
        ResourceType.NfsAssets,
        ResourceType.GitAssets,
        ResourceType.PvcAssets,
        ResourceType.S3Assets,
        ResourceType.HostPathAssets,
        ResourceType.CmVolumeAssets,
        ResourceType.SecretVolumeAssets,
        ResourceType.Datavolumes,
      ];
      return await Promise.all(
        dsResourceTypes.map((dsResourceType: ResourceType) =>
          permissionService.getPermittedScopesForAction(dsResourceType),
        ),
      ).then((actions: Array<Record<Action, PermittedScopes>>) => {
        actions.forEach((permissions: Record<Action, PermittedScopes>, index: number) => {
          this.dataSourcesActionPermissions[dsResourceTypes[index]] = permissions;
        });
      });
    },
    async loadDataSources(filterBy: IAssetsFilter = {}): Promise<Array<DatasourceListResponseEntry>> {
      this.dataSources = await dataSourceService.list(filterBy);
      return this.dataSources;
    },
    async deleteDataSource(dataSource: DatasourceListResponseEntry): Promise<HttpResponse> {
      return await dataSourceService.remove(dataSource);
    },
    async loadHostPathById(hostPathId: string): Promise<HostPathAsset> {
      this.selectedHostPath = await dataSourceService.getHostPathById(hostPathId);
      return this.selectedHostPath;
    },
    async loadS3ById(s3Id: string): Promise<S3Asset> {
      this.selectedS3 = await dataSourceService.getS3ById(s3Id);
      return this.selectedS3;
    },
    async loadNFSById(nfsId: string): Promise<NFSAsset> {
      this.selectedNFS = await dataSourceService.getNFSById(nfsId);
      return this.selectedNFS;
    },
    async loadGitById(gitId: string): Promise<GitAsset> {
      this.selectedGit = await dataSourceService.getGitById(gitId);
      return this.selectedGit;
    },
    async loadPVCById(pvcId: string): Promise<PVCAsset> {
      this.selectedPVC = await dataSourceService.getPvcById(pvcId);
      return this.selectedPVC;
    },
    async loadConfigMapById(cmId: string): Promise<ConfigMapAsset> {
      this.selectedConfigMap = await dataSourceService.getConfigMapById(cmId);
      return this.selectedConfigMap;
    },
    async loadSecretVolumeById(secretVolumeId: string): Promise<ConfigMapAsset> {
      this.selectedSecretVolume = await dataSourceService.getSecretVolumeById(secretVolumeId);
      return this.selectedSecretVolume;
    },
    async updateNfs(nfsId: string, dataSource: NFSUpdateRequest): Promise<void> {
      await dataSourceService.updateNFS(nfsId, dataSource);
    },
    async updateHostPath(hostPathId: string, dataSource: HostPathUpdateRequest): Promise<void> {
      await dataSourceService.updateHostPath(hostPathId, dataSource);
    },
    async updatePVC(pvcId: string, dataSource: PVCUpdateRequest): Promise<void> {
      await dataSourceService.updatePVC(pvcId, dataSource);
    },
    async updateGit(gitId: string, dataSource: GitUpdateRequest): Promise<void> {
      await dataSourceService.updateGit(gitId, dataSource);
    },
    async updateS3(s3Id: string, dataSource: S3UpdateRequest): Promise<void> {
      await dataSourceService.updateS3(s3Id, dataSource);
    },
    async updateConfigMap(cmId: string, dataSource: ConfigMapUpdateRequest): Promise<void> {
      await dataSourceService.updateConfigMap(cmId, dataSource);
    },
    async updateSecretVolume(secretVolumeId: string, dataSource: SecretAssetUpdateRequest): Promise<void> {
      await dataSourceService.updateSecretVolume(secretVolumeId, dataSource);
    },
    clearSelectedS3(): void {
      this.selectedS3 = null;
    },
    clearSelectedHostPath(): void {
      this.selectedHostPath = null;
    },
    clearSelectedPVC(): void {
      this.selectedPVC = null;
    },
    clearSelectedGit(): void {
      this.selectedGit = null;
    },
    clearSelectedNFS(): void {
      this.selectedNFS = null;
    },
    clearSelectedConfigMap(): void {
      this.selectedConfigMap = null;
    },
    clearSelectedSecretVolume(): void {
      this.selectedSecretVolume = null;
    },
  },
});
