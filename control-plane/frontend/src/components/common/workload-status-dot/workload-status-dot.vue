<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-status-dot" :class="statusClass">
    <span class="dot" />
    <span class="label">{{ displayLabel }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { StatusColor } from "@/models/table.model";

export interface Props {
  status: string;
  color?: StatusColor;
  label?: string;
}

const props = withDefaults(defineProps<Props>(), {
  color: undefined,
  label: undefined,
});

// Auto-detect color based on workload phase (matches WorkloadPhaseMap)
const autoColor = computed<StatusColor>(() => {
  if (props.color) return props.color;

  const status = props.status.toLowerCase();

  // Success states (green)
  if (["running", "succeeded"].includes(status)) {
    return "success";
  }

  // Warning states (yellow)
  if (status === "pending") {
    return "warning";
  }

  // Error states (red)
  if (status === "failed") {
    return "negative";
  }

  // All other states use neutral/gray (stopped, degraded, completed, creating, initializing, etc.)
  return "neutral";
});

const statusClass = computed(() => `status-${autoColor.value}`);

const displayLabel = computed(() => props.label || props.status);
</script>

<style lang="scss" scoped>
.workload-status-dot {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;

  .dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  &.status-success {
    color: $success;
    .dot {
      background-color: $success;
    }
  }

  &.status-warning {
    color: $warning;
    .dot {
      background-color: $warning;
    }
  }

  &.status-negative {
    color: $negative;
    .dot {
      background-color: $negative;
    }
  }

  &.status-info {
    color: $info;
    .dot {
      background-color: $info;
    }
  }

  &.status-neutral {
    color: $black-54;
    .dot {
      background-color: $black-40;
    }
  }
}
</style>
