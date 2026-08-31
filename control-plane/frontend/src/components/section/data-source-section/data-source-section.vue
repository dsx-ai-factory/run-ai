<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="data-source-section"
    :label="sectionLabel"
    :default-opened="defaultOpened"
    :section-invalid="sectionInvalid"
  >
    <template #subheader>
      <span>{{ dataSourceSummary }}</span>
    </template>

    <data-source-gallery
      :entity-type="entityType"
      :loading="loading"
      :data-sources="dataSources"
      :selected-data-sources="selectedDataSources"
      :cards-main-title="cardsTitle"
      :section-disabled="sectionDisabled"
      :cluster-id="clusterId"
      :created-data-source-id="createdDataSourceId"
      :allowed-created-types="allowedCreatedTypes"
      :policy-rules="policyRules"
      @datasource-changed="$emit('datasource-changed', $event)"
      @create-new="$emit('create-new', $event)"
      @is-section-invalid="sectionInvalid = $event"
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import {
  AssetKind,
  type AssetIdAndKind,
  type DatasourceListResponseEntry,
  type PolicyInfo,
} from "@/swagger-models/assets-service-client";
import type { IWorkloadDataSourceSectionOptions } from "./data-source-section.model";

// components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { DataSourceGallery } from "./data-source-gallery";
import { EWorkloadFormType } from "@/models/workload.model";

// stores
import { useClusterStore } from "@/stores/cluster.store";

// constants
import { MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH } from "@/common/version.constant";
import { useSettingStore } from "@/stores/setting.store";
import { dataSourceUtil } from "@/utils/data-source.util";
import type { TDataSourceKinds } from "@/models/data-source.model";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    DataSourceGallery,
  },
  emits: ["datasource-changed", "create-new"],
  props: {
    entityType: {
      type: String as PropType<EWorkloadFormType>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    dataSources: {
      type: Array as PropType<Array<DatasourceListResponseEntry>>,
      required: true,
    },
    selectedDataSources: {
      type: Array as PropType<Array<AssetIdAndKind>>,
      required: true,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    sectionOptions: {
      type: Object as PropType<IWorkloadDataSourceSectionOptions>,
      required: false,
      default: () => ({}),
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    createdDataSourceId: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PolicyInfo | null>,
      required: false,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      sectionInvalid: false,
    };
  },
  computed: {
    sectionLabel(): string {
      return this.settingStore.isDataVolumeEnabled ? "Data & storage" : "Data sources";
    },
    cardsTitle(): string {
      return (
        this.sectionOptions.cardsTitle ||
        `Select the ${
          this.settingStore.isDataVolumeEnabled ? "data and storage" : "data sources"
        } your workload needs to access`
      );
    },
    isClusterSupportedNfsAndHostPath(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH);
    },
    allowedCreatedTypes(): Set<TDataSourceKinds> {
      const allowedDsTypes = new Set<TDataSourceKinds>([
        AssetKind.Pvc,
        AssetKind.Nfs,
        AssetKind.S3,
        AssetKind.Git,
        AssetKind.HostPath,
        AssetKind.SecretVolume,
        AssetKind.ConfigMap,
      ]);

      if (this.entityType === EWorkloadFormType.Inference) {
        allowedDsTypes.delete(AssetKind.S3);
        allowedDsTypes.delete(AssetKind.SecretVolume);

        if (!this.isClusterSupportedNfsAndHostPath) {
          allowedDsTypes.delete(AssetKind.Nfs);
          allowedDsTypes.delete(AssetKind.HostPath);
        }
      }

      return dataSourceUtil.filterDataSourcesByPolicyRules(allowedDsTypes, this.policyRules?.rules?.assets?.datasources);
    },
    lockedCards() {
      return this.dataSources.filter((dataSource: DatasourceListResponseEntry) => dataSource.compliance?.imposed);
    },
    dataSourceSummary(): string {
      if (!this.selectedDataSourceCards?.length && !this.lockedCards.length) return "None";

      const lockedIds = this.lockedCards.map((ds: DatasourceListResponseEntry) => ds.meta.id);
      const allRelevantCards = [...this.selectedDataSourceCards, ...lockedIds];

      const firstSelected = this.dataSourceMap[allRelevantCards[0]];
      if (!firstSelected) return "None";

      let dataSourceName: string = firstSelected.meta.name;

      if (allRelevantCards.length > 1) {
        dataSourceName += ` +${allRelevantCards.length - 1}`;
      }
      return dataSourceName;
    },
    dataSourceMap(): Record<string, DatasourceListResponseEntry> {
      return this.dataSources.reduce(
        (acc: Record<string, DatasourceListResponseEntry>, dataSource: DatasourceListResponseEntry) => {
          acc[dataSource.meta.id] = dataSource;
          return acc;
        },
        {},
      );
    },
    selectedDataSourceCards(): Array<string> {
      return this.selectedDataSources.map((dataSource: AssetIdAndKind) => dataSource.id);
    },
  },
});
</script>

<style lang="scss" scoped></style>
