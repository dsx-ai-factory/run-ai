<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-select-card
    :value="isSelected"
    @clicked="$emit('card-clicked')"
    :aid="getAid(data.name)"
    @mouseenter="scheduleTooltip"
    @mouseleave="hideTooltip"
    @focusin="scheduleTooltip"
    @focusout="hideTooltip"
  >
    <div class="project-card workload-card-content">
      <section class="top-content">
        <div
          class="col text-subtitle1 text-weight-medium q-mb-xm ellipsis"
          :class="isSelected ? 'text-primary' : 'text-b-main'"
        >
          {{ data.name }}
        </div>
        <div class="col text-caption text-b-label">Resources</div>
      </section>

      <section class="bottom-content text-caption text-b-sub">
        <runai-resources-chart
          :show-tool-tip="showToolTip"
          :over-quota-value="overQuotaValue"
          :quota-value="quotaValue"
          :allocated-value="allocatedValue"
        />
      </section>
    </div>
  </runai-select-card>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiResourcesChart } from "@/components/common/runai-resources-chart";
import { RunaiSelectCard } from "@/components/common/runai-select-card";
// models

// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";
import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";

export default defineComponent({
  components: {
    RunaiResourcesChart,
    RunaiSelectCard,
  },
  emits: ["card-clicked"],
  props: {
    data: {
      type: Object as PropType<ProjectMetadata>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      showTip: false as boolean,
      tooltipTimer: undefined as ReturnType<typeof setTimeout> | undefined,
    };
  },
  computed: {
    quotaValue(): number {
      return this.data.deserved.gpu || 0;
    },
    allocatedValue(): number {
      return this.data.allocated.gpu || 0;
    },
    overQuotaValue(): number {
      return this.allocatedValue > this.quotaValue ? this.allocatedValue - this.quotaValue : 0;
    },
    showToolTip(): boolean {
      return this.showTip;
    },
  },
  methods: {
    scheduleTooltip(): void {
      this.clearTooltipTimer();
      this.tooltipTimer = setTimeout(() => {
        this.showTip = true;
        this.tooltipTimer = undefined;
      }, 500);
    },
    hideTooltip(): void {
      this.clearTooltipTimer();
      this.showTip = false;
    },
    clearTooltipTimer(): void {
      if (this.tooltipTimer) {
        clearTimeout(this.tooltipTimer);
        this.tooltipTimer = undefined;
      }
    },
    getAid(name: string): string {
      return `${toLowerCase(spaceToDash(name))}-card`;
    },
  },
  beforeUnmount() {
    this.clearTooltipTimer();
  },
});
</script>
