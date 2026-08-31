<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-drawer-content dense-header>
    <template v-slot:header>
      <section class="row items-center q-gutter-x-md" v-if="dataSource">
        <runai-svg-icon :name="dataSourceIcon" size="34" :key="dataSource.meta.name"></runai-svg-icon>
        <div class="ellipsis">{{ dataSource.meta.name }}</div>
        <runai-status v-if="statusOptions" :options="statusOptions" />
      </section>
    </template>

    <template v-if="dataSource" v-slot:subheader>
      <runai-tab v-model="tab" :tabs="tabItems" />
    </template>

    <template v-slot:body>
      <runai-default-loader v-if="loading" />
      <runai-drawer-no-data
        v-else-if="!dataSource"
        :messages="[`Select a PVC to view it's details`]"
        icon-name="ph-bold ph-file"
        icon-class="text-icon-default"
      />

      <runai-tab-panels v-else-if="dataSource" v-model="tab">
        <runai-tab-panel class="q-pa-none" :name="EDataSourceDrawerTabs.details">
          <pvc-details-tab :pvc="dataSource" :storage-class-spec="storageClassSpec" />
        </runai-tab-panel>
        <runai-tab-panel class="q-pa-none" :name="EDataSourceDrawerTabs.eventHistory">
          <pvc-event-history-tab :pvc-id="dataSource.meta.id" />
        </runai-tab-panel>
      </runai-tab-panels>
    </template>
  </runai-drawer-content>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// COMPONENTS
import { RunaiDrawerContent } from "@/components/common/runai-drawer-content";
import { RunaiTab, RunaiTabPanels, RunaiTabPanel, type ITab } from "@/components/common/runai-tab";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiStatus } from "@/components/common/runai-status";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import { PvcDetailsTab } from "@/components/data-source/data-source-drawer-content/pvc-details-tab/";
import { PvcEventHistoryTab } from "@/components/data-source/data-source-drawer-content/pvc-event-history-tab/";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
// MODELS
import type { AssetKind, DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import type { IStatusColOptions } from "@/models/table.model";
// UTILS
import { assetsUtil } from "@/utils/assets.util/assets.util";
// CONSTANTS
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";
import { useClusterStore } from "@/stores/cluster.store";
import { MIN_CLUSTER_VERSION_FOR_PVC_EVENT_HISTORY } from "@/common/version.constant";
import type { StorageClassSpec } from "@/swagger-models/k8s-objects-tracker-client";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";

export enum EDataSourceDrawerTabs {
  details = "details",
  eventHistory = "event-history",
}

export default defineComponent({
  name: "data-source-drawer-content",
  components: {
    RunaiDrawerContent,
    RunaiSvgIcon,
    RunaiStatus,
    RunaiDrawerNoData,
    RunaiDefaultLoader,
    PvcDetailsTab,
    PvcEventHistoryTab,
    RunaiTab,
    RunaiTabPanels,
    RunaiTabPanel,
  },
  props: {
    dataSource: {
      type: Object as PropType<DatasourceListResponseEntry>,
      required: false,
    },
  },
  data() {
    return {
      tab: EDataSourceDrawerTabs.details,
      EDataSourceDrawerTabs,
      storageClassSpec: null as StorageClassSpec | null,
      loading: false,
    };
  },
  async created() {
    this.storageClassSpec = await this.loadStorageClassSpec();
  },
  computed: {
    tabItems(): ITab[] {
      return [
        { name: EDataSourceDrawerTabs.details, label: "details", aid: "data-source-drawer-tab-details" },
        {
          name: EDataSourceDrawerTabs.eventHistory,
          label: "event history",
          hidden: !this.supportsEventHistory,
          aid: "data-source-drawer-tab-event-history",
        },
      ];
    },
    statusOptions(): IStatusColOptions | null {
      if (!this.dataSource || !this.dataSource.status) return null;
      return assetsUtil.getStatusColOptions(this.dataSource.status, this.dataSource.meta.createdAt);
    },
    dataSourceIcon(): string {
      if (!this.dataSource) return DATA_SOURCE_ICONS_NAMES.default;
      const kind: AssetKind = this.dataSource.meta.kind;
      return (DATA_SOURCE_ICONS_NAMES[kind as keyof object] as string) || DATA_SOURCE_ICONS_NAMES.default;
    },
    supportsEventHistory(): boolean {
      return useClusterStore().isClusterVersionSufficient(
        this.dataSource?.meta.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_PVC_EVENT_HISTORY,
      );
    },
  },
  methods: {
    async loadStorageClassSpec(): Promise<StorageClassSpec | null> {
      if (!this.dataSource?.spec.pvc?.claimInfo?.storageClass) return null;
      this.loading = true;
      try {
        const storageClassSpec =
          (
            await dataSourceService.listStorageClassV2(
              this.dataSource?.meta.clusterId || "",
              this.dataSource?.spec.pvc?.claimInfo?.storageClass || undefined,
            )
          )?.[0]?.spec || null;
        return storageClassSpec;
      } catch (error: unknown) {
        console.error("failed to load storage class spec, error:", error);
      } finally {
        this.loading = false;
      }
      return null;
    },
  },
  watch: {
    dataSource: {
      async handler(oldVal, newVal) {
        if (oldVal?.spec.pvc?.claimInfo?.storageClass !== newVal?.spec.pvc?.claimInfo?.storageClass) {
          this.storageClassSpec = await this.loadStorageClassSpec();
        }
      },
      deep: true,
    },
  },
});
</script>
