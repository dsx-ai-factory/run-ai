<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <node-pool-create-flow
    :initial-cluster-id="initialClusterId"
    @created="onCreated"
    @leave-approved="leaveApproved = true"
    @busy-change="saving = $event"
  />
</template>
<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import NodePoolCreateFlow from "@/components/node-pools/node-pool-create-flow/node-pool-create-flow.vue";
import { useNodePoolCreationStore } from "@/components/node-pools/node-pool-create-flow/node-pool-creation.store";
import type { NodePoolCreationDraft } from "@/components/node-pools/node-pool-create-flow/node-pool-create-flow.model";
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { useQuasar } from "quasar";
const $q = useQuasar();
const saving = ref(false);
const leaveApproved = ref(false);
onBeforeRouteLeave(() => {
  if (saving.value) {
    leaveApproved.value = false;
    $q.notify({ type: "info", message: "Wait for node pool creation to finish." });
    return false;
  }
  return leaveApproved.value || requestToLeave();
});
const route = useRoute();
const router = useRouter();
const store = useNodePoolCreationStore();
const clusterStore = useClusterStore();
const initialClusterId = typeof route.query.clusterId === "string" ? route.query.clusterId : "";
onMounted(() => useAppStore().setPageLoading(false));
async function onCreated(pool: Nodepool, draft: NodePoolCreationDraft, destination: "close" | "continue") {
  clusterStore.setSelectedClusterById(pool.clusterId);
  if (destination === "close") {
    store.handoff = null;
    await router.push({
      name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
      query: { createdEntityId: pool.id, clusterId: pool.clusterId },
    });
  } else {
    store.handoff = { pool, draft };
    // The new scheduling step will replace this handoff; keep the saved pool editable meanwhile.
    await router.push({
      name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT,
      params: { id: pool.id },
      query: { step: "scheduling" },
    });
  }
}
</script>
