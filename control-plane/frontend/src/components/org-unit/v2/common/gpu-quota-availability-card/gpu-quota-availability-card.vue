<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card flat bordered class="gpu-quota-availability-card">
    <q-card-section class="q-px-sm q-py-xs">
      <div class="row items-center justify-between text-b-body1 text-b-sub">
        <div class="card-title">{{ title }}</div>
        <runai-tooltip v-if="tooltip" class="q-ml-xs" tooltip-position="right" :tooltip-text="tooltip" />
      </div>

      <div class="card-value text-b-h5">
        <span>{{ displayAvailable }}</span>

        <span v-if="hasTotal">/{{ total }}</span>

        <span class="text-b-body1 q-ml-xs">GPUs</span>
      </div>
    </q-card-section>
  </q-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export interface Props {
  title?: string;
  available?: number | null;
  total?: number | null;
  tooltip?: string;
  isCurrentEntityUnlimited?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  title: "",
  available: null,
  total: null,
  tooltip: "",
  isCurrentEntityUnlimited: false,
});

const isUnlimited = computed(() => props.total === null || props.total === undefined || props.total === -1);

const hasTotal = computed(() => !isUnlimited.value && props.total !== null && props.total !== undefined);

const displayAvailable = computed(() => {
  if (props.isCurrentEntityUnlimited) return 0;
  if (props.available && props.available < 0) return 0; // no free gpu quota
  return props.available ?? "Unlimited";
});
</script>

<style scoped lang="scss">
.gpu-quota-availability-card {
  min-width: 385px;
}
</style>
