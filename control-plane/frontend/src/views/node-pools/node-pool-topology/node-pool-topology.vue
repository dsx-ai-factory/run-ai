<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="node-pool-topology-view">
    <runai-form-wrapper v-if="!pool" :form-state="{}" full-width>
      <div v-if="loading" class="q-pa-lg" role="status">Loading node pool…</div>
      <q-banner v-else-if="loadError" class="bg-red-1 text-negative q-ma-lg" role="alert"
        >{{ loadError
        }}<template #action
          ><runai-btn aid="retry-topology-pool" label="Retry" flat color="primary" @click="refresh" /></template
      ></q-banner>
    </runai-form-wrapper>
    <template v-else>
      <p v-if="telemetryError" class="q-px-md text-b-subtitle2" role="status">{{ telemetryError }}</p>
      <node-pool-topology-step
        :key="pool.id"
        :pool="pool"
        :resources="resources"
        :supports="supports"
        :can-save="canSave"
        :can-read-topologies="canReadTopologies"
        :can-create-topology="canCreateTopology"
        @saving="saving = $event"
        @leave-approved="leaveApproved = true"
        @saved="onSaved"
        @cancel="close"
        @back="back"
      />
    </template>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";
import { useQuasar } from "quasar";
import NodePoolTopologyStep from "@/components/node-pools/node-pool-topology-step/node-pool-topology-step.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { useNodePoolTopologyData } from "@/components/node-pools/node-pool-topology-step/use-node-pool-topology-data.composable";
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import {
  MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY,
  MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING,
} from "@/common/version.constant";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { alertUtil } from "@/utils/alert.util/alert.util";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
const route = useRoute();
const router = useRouter();
const $q = useQuasar();
const appStore = useAppStore();
const clusterStore = useClusterStore();
const permissions = usePermissionStore();
const poolId = computed(() => (typeof route.params.id === "string" ? route.params.id : ""));
const { pool, resources, loading, loadError, telemetryError, refresh } = useNodePoolTopologyData(poolId);
const saving = ref(false);
const leaveApproved = ref(false);
const canSave = computed(() => permissions.hasPermission(ResourceType.Nodepools, Action.Update));
const canReadTopologies = computed(() => permissions.hasPermission(ResourceType.NetworkTopologies, Action.Read));
const canCreateTopology = computed(() => permissions.hasPermission(ResourceType.NetworkTopologies, Action.Create));
const supports = computed(() => ({
  networkTopology: true,
  mnnvl: clusterStore.isClusterVersionSufficient(pool.value?.clusterId ?? null, "2.21"),
  numa: clusterStore.isClusterVersionSufficient(
    pool.value?.clusterId ?? null,
    MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING,
  ),
  gpuDevice: clusterStore.isClusterVersionSufficient(
    pool.value?.clusterId ?? null,
    MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY,
  ),
}));
watch(
  pool,
  (value) => {
    appStore.setPageLoading(false);
    appStore.setPageTitleOverride(value ? `Edit node pool: ${value.name}` : null);
  },
  { immediate: true },
);
watch(
  poolId,
  () => {
    leaveApproved.value = false;
  },
  { flush: "sync" },
);
function canLeave() {
  return !saving.value && (leaveApproved.value || !pool.value || requestToLeave());
}
onBeforeRouteLeave(canLeave);
onBeforeRouteUpdate(canLeave);
onBeforeUnmount(() => appStore.setPageTitleOverride(null));
async function close() {
  await router.push({ name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX });
}
async function back() {
  if (!pool.value || saving.value) return;
  await router.push({
    name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT,
    params: { id: pool.value.id },
    query: { stepName: "scheduling" },
  });
}
async function onSaved(saved: Nodepool) {
  leaveApproved.value = true;
  useNodePoolStore().removeNodePoolsByClusterId(saved.clusterId);
  $q.notify(alertUtil.getSuccess(`Node pool ${saved.name} saved`));
  await close();
}
</script>
