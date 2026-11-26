<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="org-unit-wizard-summary-v3 column no-wrap">
    <template v-for="(section, index) in sections" :key="section.aid">
      <q-separator v-if="index > 0" class="summary-divider" />
      <section
        class="summary-section column no-wrap"
        :class="{ 'summary-section--count': section.isCount }"
        :aid="`wizard-summary-${section.aid}`"
      >
        <div class="summary-section-title row items-center no-wrap">
          <span class="text-b-overline-caps text-b-sub">{{ section.title }}</span>
          <runai-tooltip v-if="section.tooltip" :tooltip-text="section.tooltip" width="300px" />
        </div>
        <span class="summary-section-value" :class="section.isCount ? 'text-b-h6-md' : 'text-b-body1'">
          {{ section.value }}
        </span>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { useWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import type { SchedulingRules } from "@/swagger-models/org-unit-service-client";
import type { EQuotaEntity } from "@/models/resource.model";

interface ISummarySection {
  aid: string;
  title: string;
  value: string;
  tooltip?: string;
  isCount?: boolean;
}

export interface IProps {
  entityType: EQuotaEntity;
  accessRulesCount: number;
  schedulingRules?: SchedulingRules | null;
}

const props = defineProps<IProps>();

// Shown until the step that owns the section fills it.
const NOT_SET = "-";

const wizardPreview = useWizardPreview();

const schedulingRulesCount = computed(() => {
  const schedulingRules = wizardPreview.state.value.schedulingRules?.schedulingRules ?? props.schedulingRules;
  return Object.values(schedulingRules ?? {}).filter((rule) => rule !== undefined && rule !== null).length;
});

const sections = computed<ISummarySection[]>(() => [
  { aid: "cluster", title: "Cluster", value: NOT_SET },
  { aid: "administrators", title: "Administrator(s)", value: NOT_SET, isCount: true },
  { aid: "resources", title: "Resources", value: NOT_SET },
  { aid: "projects", title: "Projects", value: NOT_SET, isCount: true },
  {
    aid: "scheduling-rules",
    title: "Scheduling rules",
    tooltip: `Apply resource and runtime restrictions to all workloads associated with the ${props.entityType}`,
    value: String(schedulingRulesCount.value),
    isCount: true,
  },
  {
    aid: "access-rules",
    title: "Access rules",
    tooltip: `Authorize access to this ${props.entityType}`,
    value: String(props.accessRulesCount),
    isCount: true,
  },
]);
</script>

<style scoped lang="scss">
.org-unit-wizard-summary-v3 {
  gap: var(--space-s);
  padding: var(--space-xs) 0 var(--space-m);
  border-left: 1px solid var(--border-subtle);
  border-radius: inherit;
}

.summary-section {
  gap: var(--space-3xs);
  padding: 0 var(--space-m);

  &--count {
    gap: var(--space-2xs);
  }
}

.summary-section-title {
  height: 32px;
  gap: var(--space-xs);
}

.summary-section-value {
  word-break: break-word;
}

.summary-divider {
  background: var(--border-subtle);
}
</style>
