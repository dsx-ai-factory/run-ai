<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-environment-section column">
    <runai-details-panel-row v-for="(item, idx) in visibleItems" :key="`${item.label}-${idx}`" :label="item.label">
      <runai-details-panel-link v-if="item.link" :link="item.link" />
      <template v-else-if="item.isCode">
        <runai-inline-code-chip v-for="(line, lIdx) in valueAsLines(item.value)" :key="lIdx" :text="line" />
      </template>
      <template v-else>
        <div v-for="(line, lIdx) in valueAsLines(item.value)" :key="lIdx">{{ line }}</div>
      </template>
    </runai-details-panel-row>

    <runai-details-panel-row v-if="visibleEnvVars.length" label="Environment variables">
      <div class="labels column">
        <runai-inline-code-chip v-for="envVar in visibleEnvVars" :key="envVar" :text="envVar" />
      </div>
    </runai-details-panel-row>

    <runai-details-panel-row v-if="visibleLabels.length" label="Labels">
      <div class="labels column">
        <runai-inline-code-chip v-for="label in visibleLabels" :key="label" :text="label" />
      </div>
    </runai-details-panel-row>

    <runai-details-panel-row v-if="visibleAnnotations.length" label="Annotations">
      <div class="labels column">
        <runai-inline-code-chip v-for="annotation in visibleAnnotations" :key="annotation" :text="annotation" />
      </div>
    </runai-details-panel-row>

    <template v-if="visibleAppParams.length">
      <q-separator />
      <div class="app-params-title text-b-body1-md">Application parameters</div>
      <runai-details-panel-row v-for="(row, idx) in visibleAppParams" :key="`${row.label}-${idx}`" :label="row.label">
        {{ valueAsString(row.value) }}
      </runai-details-panel-row>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiDetailsPanelRow, RunaiDetailsPanelLink } from "@/components/common/runai-details-panel";
import { RunaiInlineCodeChip } from "@/components/common/runai-inline-code-chip";
import { filterItems, filterStrings } from "@/utils/details-panel.utils";
import type { IEnvironmentSectionData } from "@/models/workload-details.model";

defineOptions({ name: "workload-environment-section" });

export interface Props {
  data: IEnvironmentSectionData;
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), { searchTerm: "" });

const visibleItems = computed(() => filterItems(props.data.items, props.searchTerm));
const visibleEnvVars = computed(() => filterStrings(props.data.environmentVariables, props.searchTerm));
const visibleLabels = computed(() => filterStrings(props.data.labels, props.searchTerm));
const visibleAnnotations = computed(() => filterStrings(props.data.annotations, props.searchTerm));
const visibleAppParams = computed(() => filterItems(props.data.applicationParameters, props.searchTerm));

function valueAsLines(value: string | string[]): string[] {
  return Array.isArray(value) ? value : [value];
}

function valueAsString(value: string | string[]): string {
  return Array.isArray(value) ? value.join(", ") : value;
}
</script>

<style lang="scss" scoped>
.workload-environment-section {
  gap: var(--space-m);

  .labels {
    gap: var(--space-xs);
    align-items: flex-start;
  }

  .app-params-title {
    color: var(--b-text-main);
  }
}
</style>
