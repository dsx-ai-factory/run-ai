<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="q-pt-sm q-mt-sm ai-application-workload-details">
    <div v-for="detail in visibleDetails" :key="detail.label">{{ detail.label }}: {{ detail.value }}</div>
    <a href="#" class="go-to-workload-link text-caption" @click.prevent.stop="$emit('go-to-workload')">
      Go to workload page
    </a>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Workload } from "@/swagger-models/workloads-client";
import { formatK8sResourceMemory, numberFormat } from "@/utils/format.util";

export interface Props {
  workload: Workload;
}

const props = defineProps<Props>();

defineEmits<{
  "go-to-workload": [];
}>();

export interface DetailItem {
  label: string;
  value: string;
}

const formatNumber = (val: number | undefined | null): string => (typeof val === "number" ? numberFormat(val) : "-");
const formatMemory = (val: string | undefined | null): string => (val ? formatK8sResourceMemory(val) : "-");

const visibleDetails = computed<DetailItem[]>(() => {
  const allocated = props.workload.allocatedResources;

  return [
    { label: "GPU memory allocation", value: formatMemory(allocated?.gpuMemory) },
    { label: "GPU compute allocation", value: formatNumber(allocated?.gpu) },
    { label: "CPU memory allocation", value: formatMemory(allocated?.cpuMemory) },
    { label: "CPU compute allocation", value: formatNumber(allocated?.cpu) },
  ];
});
</script>

<style lang="scss" scoped>
.ai-application-workload-details {
  border-top: 1px solid $black-12;
  font-size: 12px;
}

.go-to-workload-link {
  margin-top: 5px;
  display: block;
}
</style>
