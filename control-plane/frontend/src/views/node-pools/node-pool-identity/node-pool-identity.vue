<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <node-pool-wizard-load-state v-if="!pool" step="identity" :loading="loading" :error="loadError" @retry="refresh" />
  <runai-wizard-layout
    v-else
    class="identity-wizard"
    :steps="steps"
    active-step="identity"
    responsive
    width="1000px"
    margin-top="26px"
    left-width="200px"
    right-width="200px"
    collapse-left-on-small-screen
    hide-right-on-small-screen
    :disable-navigation="saving"
    @step-click="onStepClick"
  >
    <runai-form-wrapper
      ref="form"
      :form-state="{ description }"
      :prevent-leave="saving"
      full-width
      @leave-page="leaveApproved = true"
    >
      <runai-expansion-wrapper no-first-top-padding>
        <runai-expansion-item
          v-for="section in sections"
          :key="section.title"
          :label="section.title"
          default-opened
          disable-closing
        >
          <runai-section-title v-if="section.instruction" :title="section.instruction" />
          <runai-tooltip-wrapper
            v-for="field in section.fields"
            :key="field.label"
            tag="div"
            class="locked-field"
            :class="field.size"
            :tooltip-text="NODE_POOL_IMMUTABLE_TOOLTIP"
          >
            <q-input
              :label="field.label"
              :model-value="field.value"
              disable
              dense
              :maxlength="field.label === 'Name' ? 40 : undefined"
              :counter="field.label === 'Name'"
            />
          </runai-tooltip-wrapper>
          <runai-description-input
            v-if="section.description"
            v-model="description"
            custom-aid="node-pool-description"
            class="description-field q-mt-lg"
            dense
            :debounce="0"
            :disable="!canSave"
            :readonly="saving"
            :rules="[nodePoolUtil.validateDescription]"
            no-error-icon
          />
        </runai-expansion-item>
      </runai-expansion-wrapper>
    </runai-form-wrapper>
    <template #content-footer>
      <div class="row justify-end q-gutter-md q-mt-lg">
        <runai-btn
          aid="identity-save-close"
          label="Save & close"
          flat
          color="primary"
          :disable="saving || !canSave"
          @click="save('close')"
        />
        <runai-btn
          aid="identity-save-continue"
          label="Save & continue"
          color="primary"
          :disable="saving || !canSave"
          @click="save('continue')"
        />
      </div>
    </template>
    <template #right-sidebar>
      <node-pool-wizard-summary :pool="pool" :resources="resources" :supports="supports" :cluster-name="clusterName" />
    </template>
  </runai-wizard-layout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useQuasar } from "quasar";
import { RunaiBtn } from "@/components/common/runai-btn";
import RunaiDescriptionInput from "@/components/common/runai-description-input/runai-description-input.vue";
import RunaiExpansionItem from "@/components/common/runai-expansion-item/runai-expansion-item.vue";
import RunaiExpansionWrapper from "@/components/common/runai-expansion-wrapper/runai-expansion-wrapper.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import RunaiSectionTitle from "@/components/common/runai-section-title/runai-section-title.vue";
import RunaiTooltipWrapper from "@/components/common/runai-tooltip-wrapper/runai-tooltip-wrapper.vue";
import RunaiWizardLayout from "@/components/common/runai-wizard-layout/runai-wizard-layout.vue";
import { NodePoolWizardLoadState } from "@/components/node-pools/node-pool-wizard-load-state";
import NodePoolWizardSummary from "@/components/node-pools/node-pool-wizard-summary/node-pool-wizard-summary.vue";
import { useNodePoolWizardData } from "@/composables/node-pool/use-node-pool-wizard-data.composable";
import { useNodePoolWizardNavigation } from "@/composables/node-pool/use-node-pool-wizard-navigation.composable";
import { useNodePoolTopologySupport } from "@/composables/node-pool/use-node-pool-topology-support.composable";
import { getNodePoolWizardSteps, NODE_POOL_IMMUTABLE_TOOLTIP } from "@/models/node-pool.model";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { useClusterStore } from "@/stores/cluster.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { alertUtil } from "@/utils/alert.util";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";

const route = useRoute();
const poolId = computed(() => String(route.params.id));
const { pool, resources, loading, loadError, refresh } = useNodePoolWizardData(poolId);
const { saving, leaveApproved, openStep, finish } = useNodePoolWizardNavigation(pool);
const supports = useNodePoolTopologySupport(() => pool.value?.clusterId);
const permissions = usePermissionStore();
const canSave = computed(() => permissions.hasPermission(ResourceType.Nodepools, Action.Update));
const clusterStore = useClusterStore();
const clusterName = computed(
  () => pool.value?.clusterName || clusterStore.clusterList.find((c) => c.uuid === pool.value?.clusterId)?.name,
);
const steps = getNodePoolWizardSteps("identity");
const sections = computed(() => [
  {
    title: "Cluster",
    instruction: "Select the cluster under which to create this node pool",
    fields: [{ label: "Cluster", value: clusterName.value || pool.value?.clusterId, size: "cluster-field" }],
  },
  {
    title: "Node pool name & description",
    description: true,
    fields: [{ label: "Name", value: pool.value?.name, size: "name-field" }],
  },
  {
    title: "Label",
    instruction: "Set the label for the nodes that belong to this node pool",
    fields: [
      { label: "Key", value: pool.value?.labelKey, size: "label-field" },
      { label: "Value", value: pool.value?.labelValue, size: "label-field" },
    ],
  },
]);
const description = ref("");
watch(
  pool,
  (value) => {
    description.value = value?.description ?? "";
  },
  { immediate: true, flush: "sync" },
);
const form = ref<InstanceType<typeof RunaiFormWrapper>>();
const $q = useQuasar();

function onStepClick({ name }: { name: string }) {
  if (name === "scheduling" || name === "topology") void openStep(name);
}

async function save(destination: "close" | "continue") {
  const current = pool.value;
  if (!current || !canSave.value || saving.value) return;
  saving.value = true;
  try {
    if (!(await form.value?.validate()) || !canSave.value) return;
    if (description.value !== (current.description ?? "")) {
      const saved = await clusterService.patchNodepool(current.id, { description: description.value });
      useNodePoolStore().removeNodePoolsByClusterId(saved.clusterId);
      $q.notify(alertUtil.getSuccess(`Node pool ${saved.name} saved`));
    }
    saving.value = false;
    leaveApproved.value = true;
    if (destination === "continue") await openStep("scheduling");
    else await finish(current);
  } catch {
    $q.notify(alertUtil.getError(`Node pool ${current.name} couldn't be saved`));
  } finally {
    saving.value = false;
  }
}
</script>

<style scoped lang="scss">
.identity-wizard {
  padding-inline: 20px;
}
.locked-field {
  cursor: not-allowed;
}
.cluster-field {
  max-width: 655px;
  margin-top: 26px;
}
.name-field {
  max-width: 510px;
}
.description-field {
  max-width: 702px;
}
.label-field {
  max-width: 380px;
  margin-top: 26px;
}
</style>
