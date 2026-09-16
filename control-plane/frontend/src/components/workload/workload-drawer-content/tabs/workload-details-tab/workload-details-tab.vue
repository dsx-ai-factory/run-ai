<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-details" :key="workload.id">
    <details-panel-view
      :loading="loading"
      :sections="accordionSections"
      v-model:selected-spec="selectedSpec"
      :show-spec-select="showSpecSelect"
      :spec-options="specOptions"
      :empty-messages="['There are no details for this workload yet.']"
    />
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//models
import type {
  Workload,
  WorkloadDetailed,
  WorkloadNodeAffinity,
  DistributedWorkload,
} from "@/swagger-models/workloads-client";
import type { ISelectOption } from "@/models/global.model";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";
import { EWorkloadType, EDistributedReplicaType } from "@/models/workload.model";
//services
import { workloadService, type CrudWorkload } from "@/services/cluster/workload.service/workload.service";
//stores
import { useClusterStore } from "@/stores/cluster.store";
//cmps
import DetailsPanelView from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/details-panel-view.vue";
//utils
import { workloadDetailsUtil, type WorkloadSpec } from "@/utils/workload-details.util";
import { workloadV2ManifestUtil } from "@/utils/workload-v2-manifest.util";
import { alertUtil } from "@/utils/alert.util";
import { SourceApi } from "@/swagger-models/workloads-client";
import { NotFoundError, type HttpErrorResponse } from "@/models/http-response.model";
import { topologyLevelsFetcherService } from "@/services/topology-levels-fetcher.service/topology-levels-fetcher.service";

export default defineComponent({
  name: "workload-details-tab",
  components: {
    DetailsPanelView,
  },
  props: {
    workload: {
      type: Object as PropType<Workload>,
      required: true,
    },
    isDeletedViewType: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      workloadDetailed: null as WorkloadDetailed | null,
      workloadCrud: null as CrudWorkload | null,
      nodeAffinities: [] as WorkloadNodeAffinity[],
      topologyLevels: [] as string[],
      fetchTopologyLevelsGuarded: topologyLevelsFetcherService.createTopologyLevelsFetcher(),
      loading: false,
      selectedSpec: EDistributedReplicaType.Master,
      clusterStore: useClusterStore(),
    };
  },
  async created() {
    await this.fetchWorkloadDetails();
  },
  computed: {
    isNativeWorkload(): boolean {
      const nativeTypes = [
        EWorkloadType.Workspace,
        EWorkloadType.Training,
        EWorkloadType.Inference,
        EWorkloadType.Distributed,
        EWorkloadType.DistributedInference,
      ];
      return nativeTypes.includes(this.workload.type as EWorkloadType);
    },

    isWorkloadsV2Workload(): boolean {
      return this.workload.sourceApi === SourceApi.WorkloadsV2;
    },

    // Soft-deleted workloads remain in v1; v2 queue rows are hard-deleted after agent confirm.
    useV2Api(): boolean {
      return this.isWorkloadsV2Workload && !this.isDeletedViewType;
    },

    isDistributed(): boolean {
      return this.workload.type === EWorkloadType.Distributed;
    },

    isDistributedInference(): boolean {
      return this.workload.type === EWorkloadType.DistributedInference;
    },

    hasReplicaSelector(): boolean {
      return this.isDistributed || this.isDistributedInference;
    },

    showSpecSelect(): boolean {
      // Hide the selector when there's only one replica role (e.g., a TFJob with workers and
      // no master, or a single-pod DistributedInference with only a leader). A 1-item dropdown
      // adds no value — the section data is the same regardless of the selection.
      return this.hasReplicaSelector && this.specOptions.length > 1;
    },

    specOptions(): ISelectOption[] {
      if (!this.workloadCrud) return [];

      // "Worker" is pluralized as "Workers" in the UI because a workload can have many worker pods,
      // unlike the single master/leader. The underlying value stays as the enum string.
      const workerOption: ISelectOption = { label: "Workers", value: EDistributedReplicaType.Worker };

      if (this.isDistributed) {
        const distributedWorkload = this.workloadCrud as DistributedWorkload;
        const options: ISelectOption[] = [];
        if (distributedWorkload.masterSpec) {
          options.push({ label: "Master", value: EDistributedReplicaType.Master });
        }
        if (distributedWorkload.spec) options.push(workerOption);
        return options;
      }

      if (this.isDistributedInference) {
        const spec = (this.workloadCrud as { spec?: { leader?: unknown; worker?: unknown } }).spec;
        const options: ISelectOption[] = [];
        if (spec?.leader) options.push({ label: "Leader", value: EDistributedReplicaType.Leader });
        if (spec?.worker) options.push(workerOption);
        return options;
      }

      return [];
    },

    currentSpec(): WorkloadSpec | undefined {
      const crudWorkload = this.workloadCrud ?? undefined;
      const selectedReplicaType = this.hasReplicaSelector ? this.selectedSpec : undefined;

      return workloadDetailsUtil.getSpecSection(crudWorkload, selectedReplicaType);
    },

    currentNodeAffinity(): WorkloadNodeAffinity | undefined {
      if (!this.nodeAffinities.length) return undefined;
      if (this.nodeAffinities.length === 1) return this.nodeAffinities[0];
      return (
        this.nodeAffinities.find((entry) => entry.specName?.toLowerCase() === this.selectedSpec?.toLowerCase()) ??
        this.nodeAffinities[0]
      );
    },

    // The accordion-section model consumed by the shared details-panel-view. Built here (rather
    // than in a separate "view" component) since this is the only consumer of the data this tab
    // already fetched — no reuse to justify the extra layer.
    accordionSections(): IDetailsPanelSectionDef[] {
      if (!this.workloadDetailed) return [];
      return workloadDetailsUtil.buildSearchableSections({
        workloadDetailed: this.workloadDetailed,
        workloadCrud: this.workloadCrud ?? undefined,
        currentSpec: this.currentSpec,
        clusterName: this.getClusterName(),
        topologyLevels: this.topologyLevels,
        nodeAffinity: this.currentNodeAffinity?.nodeAffinity,
      });
    },
  },
  methods: {
    async fetchWorkloadDetails() {
      this.loading = true;
      this.nodeAffinities = [];
      this.workloadCrud = null;
      this.workloadDetailed = null;

      // Supported (external) workloads have no CRUD/asset representation — the CRUD endpoint
      // throws for non-native types. Fetch it only for native workloads and render everything
      // else from WorkloadDetailed alone.
      // Deleted view: always use v1 get-by-id (soft-deleted rows); v2 get-by-id returns 404.
      const detailedPromise = this.useV2Api ? Promise.resolve(false) : this.fetchWorkloadDetailed();
      const crudPromise =
        !this.isWorkloadsV2Workload && this.isNativeWorkload ? this.fetchWorkloadCrud() : Promise.resolve(false);
      const workloadV2Promise = this.useV2Api ? this.fetchWorkloadV2Details() : Promise.resolve(false);

      const [hasDetailed, hasCrud, hasWorkloadV2] = await Promise.all([detailedPromise, crudPromise, workloadV2Promise]);

      // Show error only if everything we tried to load failed.
      if (!hasDetailed && !hasCrud && !hasWorkloadV2) {
        this.$q.notify(alertUtil.getError("Failed to load workload details"));
      }

      // Snap selectedSpec to the first available replica option — the data-level default
      // (Master) is invalid for DistributedInference, which only has Leader/Worker.
      const options = this.specOptions;
      if (options.length > 0 && !options.some((o) => o.value === this.selectedSpec)) {
        this.selectedSpec = options[0].value as EDistributedReplicaType;
      }

      const workloadDetailed = this.workloadDetailed as WorkloadDetailed | null;
      this.nodeAffinities = workloadDetailed?.nodeAffinities ?? [];
      await this.fetchTopologyLevels();

      this.loading = false;
    },

    async fetchWorkloadDetailed(): Promise<boolean> {
      try {
        this.workloadDetailed = await workloadService.getWorkloadById(this.workload.id);
        return true;
      } catch (error) {
        return false;
      }
    },

    async fetchWorkloadV2Details(): Promise<boolean> {
      try {
        const response = await workloadService.getWorkloadV2ById(this.workload.id);
        this.workloadCrud =
          (workloadV2ManifestUtil.getCrudWorkloadFromManifest(
            response.workload.rawManifest,
            this.workload.type,
          ) as CrudWorkload) ?? null;
        this.workloadDetailed ??= this.workload as WorkloadDetailed;
        return true;
      } catch (error) {
        if ((error as HttpErrorResponse)?.statusCode === NotFoundError.statusCode) {
          return this.fetchWorkloadDetailed();
        }
        return false;
      }
    },

    async fetchWorkloadCrud(): Promise<boolean> {
      try {
        this.workloadCrud = await workloadService.getWorkloadCrudById(this.workload.id, this.workload.type);
        return true;
      } catch (error) {
        return false;
      }
    },

    getClusterName(): string {
      if (!this.workloadDetailed?.clusterId) return "";
      const cluster = this.clusterStore.getClusterById(this.workloadDetailed.clusterId);
      return cluster?.name || this.workloadDetailed.clusterId;
    },

    async fetchTopologyLevels(): Promise<void> {
      this.topologyLevels = [];
      const levels = await this.fetchTopologyLevelsGuarded(
        this.workloadDetailed?.topology?.name,
        this.workloadDetailed?.clusterId,
      );
      if (levels !== undefined) this.topologyLevels = levels;
    },
  },
  watch: {
    "workload.id": {
      handler() {
        this.fetchWorkloadDetails();
      },
    },
  },
});
</script>

<style scoped lang="scss">
.workload-details {
  height: 100%;
}

:deep(.item-details-item) {
  margin-bottom: 24px;
}
</style>
