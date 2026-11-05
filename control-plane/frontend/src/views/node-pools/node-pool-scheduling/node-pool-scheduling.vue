<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-form-wrapper v-if="!pool" :form-state="{}" full-width>
    <div v-if="loading" class="q-pa-lg" role="status">Loading node pool…</div>
    <q-banner v-else-if="loadError" class="bg-red-1 text-negative q-ma-lg" role="alert">
      {{ loadError }}
      <template #action
        ><runai-btn aid="retry-scheduling-pool" label="Retry" flat color="primary" @click="loadPool"
      /></template>
    </q-banner>
  </runai-form-wrapper>
  <node-pool-scheduling-step v-else :key="pool.id" :pool="pool" @leave-approved="leaveApproved = true" />
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute } from "vue-router";
import NodePoolSchedulingStep from "@/components/node-pools/node-pool-scheduling-step/node-pool-scheduling-step.vue";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { useAppStore } from "@/stores/app.store";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { extractErrorMessages } from "@/utils/error-alert.util";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

const route = useRoute();
const appStore = useAppStore();
const pool = ref<Nodepool>();
const loading = ref(true);
const loadError = ref("");
const leaveApproved = ref(false);
let requestVersion = 0;
async function loadPool() {
  const version = ++requestVersion;
  leaveApproved.value = false;
  const id = route.params.id;
  pool.value = undefined;
  loading.value = true;
  loadError.value = "";
  // This view owns loading/error states, including the header's leave-request listener.
  appStore.setPageLoading(false);
  appStore.setPageTitleOverride(null);
  try {
    if (typeof id !== "string" || !id) throw new Error("A node pool is required.");
    const result = await clusterService.getNodepool(id);
    if (version !== requestVersion) return;
    pool.value = result;
    appStore.setPageTitleOverride(`New node pool: ${result.name}`);
  } catch (error) {
    if (version === requestVersion) loadError.value = extractErrorMessages(error, "Failed to load node pool.").join(" ");
  } finally {
    if (version === requestVersion) {
      loading.value = false;
      appStore.setPageLoading(false);
    }
  }
}
function canLeave() {
  return leaveApproved.value || !pool.value || requestToLeave();
}
onBeforeRouteLeave(canLeave);
onBeforeRouteUpdate(canLeave);
watch(() => route.params.id, loadPool, { immediate: true });
onBeforeUnmount(() => {
  requestVersion++;
  appStore.setPageTitleOverride(null);
});
</script>
