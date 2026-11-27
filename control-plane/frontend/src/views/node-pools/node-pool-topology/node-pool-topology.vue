<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="node-pool-topology-view">
    <node-pool-wizard-load-state v-if="!pool" step="topology" :loading="loading" :error="loadError" @retry="refresh" />
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
        @open-identity="openStep('identity')"
        @back="openStep('scheduling')"
      />
    </template>
  </div>
</template>
<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useQuasar } from "quasar";
import { NodePoolTopologyStep } from "@/components/node-pools/node-pool-topology-step";
import { NodePoolWizardLoadState } from "@/components/node-pools/node-pool-wizard-load-state";
import { useNodePoolWizardData } from "@/composables/node-pool/use-node-pool-wizard-data.composable";
import { useNodePoolWizardNavigation } from "@/composables/node-pool/use-node-pool-wizard-navigation.composable";
import { useNodePoolTopologySupport } from "@/composables/node-pool/use-node-pool-topology-support.composable";
import { usePermissionStore } from "@/stores/permissions.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { alertUtil } from "@/utils/alert.util/alert.util";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

const route = useRoute();
const $q = useQuasar();
const permissions = usePermissionStore();
const poolId = computed(() => (typeof route.params.id === "string" ? route.params.id : ""));
const { pool, resources, loading, loadError, telemetryError, refresh } = useNodePoolWizardData(poolId);
const { saving, leaveApproved, openStep, finish } = useNodePoolWizardNavigation(pool);
const supports = useNodePoolTopologySupport(() => pool.value?.clusterId);
const canSave = computed(() => permissions.hasPermission(ResourceType.Nodepools, Action.Update));
const canReadTopologies = computed(() => permissions.hasPermission(ResourceType.NetworkTopologies, Action.Read));
const canCreateTopology = computed(() => permissions.hasPermission(ResourceType.NetworkTopologies, Action.Create));

async function onSaved(saved: Nodepool) {
  useNodePoolStore().removeNodePoolsByClusterId(saved.clusterId);
  $q.notify(alertUtil.getSuccess(`Node pool ${saved.name} saved`));
  await finish(saved);
}
</script>
