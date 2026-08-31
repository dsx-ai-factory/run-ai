<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card flat bordered class="gpu-quota-card bg-grey-1">
    <q-card-section>
      <div class="row items-center gap-15">
        <div class="org-unit-name text-b-h6-md col-3">
          {{ orgUnitName }}
        </div>

        <gpu-quota-input
          v-model="localGpuQuota"
          :state="state"
          :message="message"
          :entity-type="entityType"
          class="col q-mr-lg"
        />

        <org-unit-quota-distribution-chart
          class="col"
          :entity-name="orgUnitName"
          :entity-type="entityType"
          :assigned-quota="localGpuQuota === -1 ? 0 : (localGpuQuota ?? 0)"
          :parent-name="parentName"
          :node-pool-name="nodePoolName"
          :parent-available-quota="totalQuota ?? 0"
          :other-entities-assigned-quota="otherEntitiesAssignedQuota"
          :is-unlimited="isUnlimited"
          :is-assigned-quota-unlimited="isCurrentEntityUnlimited"
          :is-other-entities-quota-unlimited="isOtherEntitiesQuotaUnlimited"
        />
      </div>
    </q-card-section>
  </q-card>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { GpuQuotaInput } from "@/components/org-unit/v2/common/gpu-quota-input";
import { OrgUnitQuotaDistributionChart } from "@/components/org-unit/v2/common/org-unit-quota-distribution-chart";
import { EQuotaInputState } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";

export interface Props {
  orgUnitName: string;
  gpuQuota: number | null;
  totalQuota: number | null;
  unassigned: number;
  state?: EQuotaInputState;
  message?: string;
  parentName?: string;
  nodePoolName?: string;
  otherEntitiesAssignedQuota?: number;
  isUnlimited?: boolean;
  isCurrentEntityUnlimited?: boolean; // True when this entity's quota is -1 (unlimited)
  isOtherEntitiesQuotaUnlimited?: boolean; // True when other entities include unlimited quota
  entityType?: EQuotaEntity;
}

const props = withDefaults(defineProps<Props>(), {
  gpuQuota: 0,
  totalQuota: null,
  unassigned: 0,
  state: EQuotaInputState.Default,
  message: "",
  otherEntitiesAssignedQuota: 0,
  isUnlimited: false,
  isCurrentEntityUnlimited: false,
  isOtherEntitiesQuotaUnlimited: false,
  entityType: EQuotaEntity.project,
});

const emit = defineEmits<{
  (e: "update:gpu-quota", value: number): void;
}>();

const localGpuQuota = ref<number>(props.gpuQuota ?? 0);

watch(
  () => props.gpuQuota,
  (newValue) => {
    localGpuQuota.value = newValue ?? 0;
  },
);

watch(localGpuQuota, (newValue) => {
  emit("update:gpu-quota", newValue);
});
</script>

<style scoped lang="scss">
.org-unit-name {
  word-break: break-word;
}
</style>
