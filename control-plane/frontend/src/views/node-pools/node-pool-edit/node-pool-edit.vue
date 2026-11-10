<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <node-pool-edit-form-v2
    v-if="newFlowEnabled"
    :initial-cluster-id="initialClusterId"
    @created="onCreatedV2"
    @leave-approved="leaveApproved = true"
    @busy-change="savingV2 = $event"
  />
  <node-pool-edit-form
    v-else-if="nodePool"
    :loading="loading"
    :node-pool="nodePool"
    :is-new-node-pool="isNewNodePool"
    :is-node-pool-submitting="isNodePoolSubmitting"
    :node-pools="nodePools"
    @on-save="onSave"
    @on-cancel="onCancel"
    @update-cluster-id="updateClusterId"
  />
</template>
<script lang="ts">
import { computed, defineComponent, ref } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { useQuasar } from "quasar";
import { NodePoolEditFormV2, ENodePoolFormDestination } from "@/components/node-pools/node-pool-edit-form-v2";
import { ENABLE_NODE_POOL_CREATE_FLOW } from "@/constants/node-pool.constants";
import { errorMessages } from "@/common/error-message.constant";
import type { INodePoolFormDraft } from "@/components/node-pools/node-pool-edit-form-v2";
import NodePoolEditForm from "@/components/node-pools/node-pool-edit-form/node-pool-edit-form.vue";
import { useAppStore } from "@/stores/app.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { HttpErrorResponse } from "@/models/http-response.model";
import { useClusterStore } from "@/stores/cluster.store";
import { ETableFilters } from "@/models/table.model";
import { CLUSTER_COLUMN_FILTER_NAME, type IFilterBy } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import {
  type Nodepool,
  type NodepoolCreateFields,
  NodepoolSortFilterFields,
} from "@/swagger-models/cluster-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { deepCopy } from "@/utils/common.util";
import { DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY } from "@/models/node-pool.model";

export default defineComponent({
  name: "node-pool-edit",
  components: { NodePoolEditForm, NodePoolEditFormV2 },
  data() {
    return {
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      nodePoolStore: useNodePoolStore(),
      isNodePoolSubmitting: false as boolean,
      loading: false as boolean,
      nodePool: null as Nodepool | NodepoolCreateFields | null,
      clusterId: null as string | null,
      nodePools: [] as Nodepool[],
    };
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const $q = useQuasar();
    const savingV2 = ref(false);
    const leaveApproved = ref(false);
    const newFlowEnabled = computed(
      () => ENABLE_NODE_POOL_CREATE_FLOW && route.name === NODE_POOL_ROUTE_NAMES.NODE_POOL_NEW,
    );
    const initialClusterId = typeof route.query.clusterId === "string" ? route.query.clusterId : "";
    onBeforeRouteLeave(() => {
      if (!newFlowEnabled.value) return true;
      if (savingV2.value) {
        leaveApproved.value = false;
        $q.notify(alertUtil.getInfo(errorMessages.NODE_POOL_CREATION_IN_PROGRESS));
        return false;
      }
      return leaveApproved.value || requestToLeave();
    });
    async function onCreatedV2(pool: Nodepool, _draft: INodePoolFormDraft, destination: ENodePoolFormDestination) {
      useClusterStore().setSelectedClusterById(pool.clusterId);
      await router.push(
        destination === ENodePoolFormDestination.Close
          ? {
              name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
              query: { createdEntityId: pool.id, clusterId: pool.clusterId },
            }
          : { name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, params: { id: pool.id }, query: { stepName: "scheduling" } },
      );
    }
    return { newFlowEnabled, initialClusterId, onCreatedV2, savingV2, leaveApproved };
  },
  async created() {
    if (this.newFlowEnabled) {
      this.appStore.setPageLoading(false);
      return;
    }
    this.clusterId = (this.$route.query.clusterId as string) || null;
    await this.loadNodePool();
    if (this.clusterId) {
      await this.loadNodePools(this.clusterId);
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    isNewNodePool(): boolean {
      return this.$route.name === NODE_POOL_ROUTE_NAMES.NODE_POOL_NEW;
    },
    nodePoolId(): string {
      return this.$route.params.id as string;
    },
  },
  methods: {
    async loadNodePool(): Promise<void> {
      if (this.isNewNodePool) {
        this.loadEmptyNodePool();
      } else {
        await this.loadExistingNodePool();
      }
    },
    loadEmptyNodePool(): void {
      this.nodePool = nodePoolUtil.getEmptyNodePool(this.clusterId || "");
    },
    async loadExistingNodePool(): Promise<void> {
      try {
        this.nodePool = await clusterService.getNodepool(this.nodePoolId);
        this.clusterId = this.nodePool.clusterId;
        this.appStore.setPageTitleOverride(`Edit node pool: ${(this.nodePool as Nodepool).name}`);
      } catch (error: unknown) {
        this.handleLoadNodePoolError(error);
      }
    },
    handleLoadNodePoolError(error: unknown): void {
      if (error instanceof HttpErrorResponse) {
        console.error(error.serialize());
        this.$q.notify(alertUtil.getError(error.message));
      } else {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load node pool"));
      }
    },
    updateGpuNetworkAccelerationLabelKey(nodePool: Nodepool | NodepoolCreateFields): void {
      if (!nodePool.gpuNetworkAccelerationLabelKey) {
        nodePool.gpuNetworkAccelerationLabelKey = DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY;
      }
    },
    async onSave(nodePool: Nodepool | NodepoolCreateFields): Promise<void> {
      this.isNodePoolSubmitting = true;
      const nodePoolToSave = deepCopy(nodePool);
      this.updateGpuNetworkAccelerationLabelKey(nodePoolToSave);
      try {
        if (this.isNewNodePool) {
          await this.createNodePool(nodePoolToSave as NodepoolCreateFields);
        } else {
          await this.editNodePool(nodePoolToSave as Nodepool);
        }
        this.clusterStore.setSelectedClusterById(nodePoolToSave.clusterId);
      } catch (error: unknown) {
        this.handleError(error);
      } finally {
        this.isNodePoolSubmitting = false;
      }
    },
    async createNodePool(nodePool: NodepoolCreateFields): Promise<void> {
      const clusterId = nodePool.clusterId;
      const createdNodePool = await clusterService.createNodepool(nodePool);
      this.nodePoolStore.removeNodePoolsByClusterId(nodePool.clusterId);
      this.$router.push({
        name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
        query: { createdEntityId: createdNodePool.id, clusterId },
      });
      this.$q.notify(alertUtil.getSuccess(`Node pool ${nodePool.name} created`));
    },
    async editNodePool(nodePool: Nodepool): Promise<void> {
      const updateRequest = nodePoolUtil.getNodePoolUpdateRequest(nodePool);
      await clusterService.updateNodepool(this.nodePoolId, updateRequest);
      this.$q.notify(alertUtil.getSuccess(`Node pool ${nodePool.name} saved`));
      this.$router.push({ name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX, query: { createdEntityId: this.nodePoolId } });
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        await this.$router.push({ name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX, query: { clusterId: this.clusterId } });
      }
    },
    handleError(error: unknown) {
      if (error instanceof HttpErrorResponse) {
        console.error(error.serialize());
        this.$q.notify(alertUtil.getError(error.message));
      } else {
        this.$q.notify(alertUtil.getError(`Failed to ${this.isNewNodePool ? "create" : "save"} Node pool`));
      }
    },
    setClusterColumnFilter(clusterId: string): void {
      const filterBy: IFilterBy = this.getNodePoolsFilterBy();
      filterService.setColumnFilter(filterBy, clusterId, CLUSTER_COLUMN_FILTER_NAME, ETableFilters.NODE_POOL);
    },
    getNodePoolsFilterBy(): IFilterBy {
      return filterService.loadFilters(window.location, ETableFilters.NODE_POOL, {});
    },
    async updateClusterId(clusterId: string): Promise<void> {
      await this.loadNodePools(clusterId);
      if (this.isNewNodePool) {
        this.loadEmptyNodePool();
      }
    },
    async loadNodePools(clusterId: string): Promise<void> {
      try {
        this.loading = true;
        this.nodePools = await clusterService.getNodepools([
          filterUtil.getEqualsFilterString(NodepoolSortFilterFields.ClusterId, clusterId),
        ]);
      } catch (e) {
        console.error("Failed to load node pools", e);
      } finally {
        this.loading = false;
      }
    },
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});
</script>

<style scoped lang="scss"></style>
