/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { clusterApiServiceOld } from "@/services/cluster/cluster-api.service/cluster-api-old.service";
import { filterService } from "@/services/filter.service/filter.service";
import { TRAINING_DISTRIBUTED_MIN_VERSION } from "@/models/cluster.model";
import { useAuthStore } from "@/stores/auth.store";
import { isNewerVersion } from "@/utils/version.util";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

import { MIN_VERSION_BACKOFF_LIMIT } from "@/common/version.constant";
import type { ClustersMinimalData, DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { storageUtil } from "@/utils/storage.util";
import { CLUSTER_ID_FILTER } from "@/common/storage.constant";
import { clusterApi } from "@/services/infra/client-apis/cluster-api/cluster-api";

export const useClusterStore = defineStore("Cluster", {
  state: () => ({
    clusters: [] as Array<ClustersMinimalData>,
    currentCluster: {} as DisplayedCluster,
    lastCreatedClusterUuid: null as null | string,
    firstConnectedCluster: undefined as DisplayedCluster | undefined,
    selectedCluster: null as ClustersMinimalData | null,
    isClustersLoaded: false,
  }),
  getters: {
    selectedClusterName(): string {
      return (
        this.clusterList.find((cluster) => cluster.uuid === this.selectedCluster?.uuid)?.name || this.defaultClusterName
      );
    },
    selectedClusterId(): string {
      return (
        this.clusterList.find((cluster) => cluster.uuid === this.selectedCluster?.uuid)?.uuid || this.defaultClusterId
      );
    },
    defaultClusterId(): string {
      return this.clusterList[0]?.uuid || "";
    },
    defaultClusterName(): string {
      return this.clusterList[0]?.name || "";
    },
    defaultCluster(): ClustersMinimalData | null {
      return this.clusterList[0] || null;
    },
    clusterList(): Array<ClustersMinimalData> {
      return this.clusters;
    },
    clustersAmount(): number {
      return this.clusters.length;
    },
    isMultiCluster(): boolean {
      return this.clustersAmount > 1;
    },
    isSingleCluster(): boolean {
      return this.clustersAmount === 1;
    },
    clusterListLastCreatedExcluded(): Array<ClustersMinimalData> {
      if (!this.lastCreatedClusterUuid) return this.clusters;
      return this.clusters.filter((cluster) => cluster.uuid !== this.lastCreatedClusterUuid);
    },
    lastCreatedCluster(): ClustersMinimalData | null {
      if (!this.lastCreatedClusterUuid) return null;
      return this.clusters.find((cluster) => cluster.uuid === this.lastCreatedClusterUuid) || null;
    },
    // Uses dependenciesStatus path (minimal API)
    isConfigDomainEnabled:
      (state) =>
      (clusterId: string): boolean => {
        const cluster = state.clusters.find((cluster) => cluster.uuid === clusterId);
        if (!cluster?.dependenciesStatus?.optional?.knative) return false;
        const kourierComponents = cluster.dependenciesStatus.optional.knative.components?.kourier?.components;
        return kourierComponents?.["external-config-domain"]?.available || false;
      },
    // Uses dependenciesStatus path (minimal API)
    availableDependenciesByClusterUuid:
      (state) =>
      (clusterUuid: string): Array<string> => {
        const relevantCluster = state.clusters.find((c) => c.uuid === clusterUuid);
        if (!relevantCluster) return [];
        const requiredDependenciesAvailable: Array<string> = Object.keys(
          relevantCluster.dependenciesStatus?.required || {},
        ).filter((key: string) => relevantCluster.dependenciesStatus?.required?.[key]?.available);

        const optionalDependenciesAvailable: Array<string> = Object.keys(
          relevantCluster.dependenciesStatus?.optional || {},
        ).filter((key: string) => relevantCluster.dependenciesStatus?.optional?.[key]?.available);

        return [...requiredDependenciesAvailable, ...optionalDependenciesAvailable];
      },
    isVersionSupportDistributedTraining:
      (state) =>
      (clusterUuid: string): boolean => {
        const clusterVersion = state.clusters.find((cluster) => clusterUuid === cluster.uuid)?.version || "";
        return isNewerVersion(clusterVersion, TRAINING_DISTRIBUTED_MIN_VERSION);
      },
    isClusterVersionSupportBackoffLimit:
      (state) =>
      (clusterUuid: string): boolean => {
        const clusterVersion = state.clusters.find((cluster) => clusterUuid === cluster.uuid)?.version || "";
        return isNewerVersion(clusterVersion, MIN_VERSION_BACKOFF_LIMIT);
      },
    canCreateCluster(): boolean {
      const permissionStore = usePermissionStore();
      const authStore = useAuthStore();
      return permissionStore.hasPermission(ResourceType.Cluster, Action.Read) || authStore.isAdmin;
    },
    clusterByUuid:
      (state) =>
      (clusterUuid: string): ClustersMinimalData | undefined => {
        return state.clusters.find((cluster) => clusterUuid === cluster.uuid);
      },
    clusterNameByUuid:
      (state) =>
      (clusterUuid: string): string | undefined => {
        return state.clusters.find((cluster) => clusterUuid === cluster.uuid)?.name;
      },
  },
  actions: {
    // Loads minimal cluster data (default)
    async loadClusters(): Promise<Array<ClustersMinimalData>> {
      this.clusters = await clusterService.listMinimal();
      this.isClustersLoaded = true;
      return this.clusters;
    },
    async create(name: string): Promise<DisplayedCluster> {
      const newCluster: DisplayedCluster = await clusterService.create(name);
      await useAuthStore().loadUserOrgUnits();
      this.lastCreatedClusterUuid = newCluster?.uuid || null;
      if (newCluster) {
        this.clusters.push(newCluster);
      }
      return newCluster;
    },
    async deleteCluster(uuid: string): Promise<void> {
      await clusterService.remove(uuid);
      this.clusters = this.clusters.filter((cluster) => cluster.uuid !== uuid);
      await useAuthStore().loadUserOrgUnits();
    },
    async updateCluster(uuid: string, updatedCluster: DisplayedCluster): Promise<void> {
      await clusterService.update(uuid, updatedCluster);
    },
    isClusterNameExists(name: string): boolean {
      return !!this.clusters.find((cluster) => cluster.name === name);
    },
    getClusterById(clusterId: string): ClustersMinimalData | null {
      return this.clusters.find((cluster) => cluster.uuid === clusterId) || null;
    },
    getClusterVersion(clusterId: string): string | null {
      const cluster = this.getClusterById(clusterId);
      return cluster?.version || null;
    },
    isClusterVersionSufficient(clusterUuid: string | null, checkForVersion: string): boolean {
      if (!clusterUuid) return false;
      const clusterVersion = this.clusterList.find((c) => c.uuid === clusterUuid)?.version;
      if (!clusterVersion) return false;

      return isNewerVersion(clusterVersion, checkForVersion);
    },
    isAtleastOneClusterSufficientForVersion(checkForVersion: string): boolean {
      return this.clusterList.some((c) => isNewerVersion(c.version || "", checkForVersion));
    },
    isAllClustersSupportSpecificVersion(version: string): boolean {
      if (!version) return false;
      return this.clusterList.every((c) => this.isClusterVersionSufficient(c.uuid, version));
    },
    /**
     * This method is intended for local use only and should not be exposed publicly.
     * @param {ClustersMinimalData} selectedCluster - The cluster to be selected.
     */
    _setSelectedCluster(selectedCluster: ClustersMinimalData): void {
      this.selectedCluster = selectedCluster;

      if (!this.selectedCluster) {
        clusterApiServiceOld.setBaseURL("");
        clusterApi.setBaseUrl("");
        return;
      }

      try {
        const domain = this.selectedCluster.domain || "";
        clusterApiServiceOld.setBaseURL(domain);
        clusterApi.setBaseUrl(domain);
      } catch (e) {
        console.error("Error while setting selected cluster", e);
      }
    },
    setSelectedClusterById(clusterId: string): void {
      const cluster = this.clusterList.find((c) => c.uuid === clusterId) || this.defaultCluster;
      if (!cluster) return;
      this._setSelectedCluster(cluster);
      storageUtil.save(CLUSTER_ID_FILTER, cluster.uuid);
      filterService.setClusterUuidForMultiClusterTableFilters(cluster.uuid);
    },
  },
});
