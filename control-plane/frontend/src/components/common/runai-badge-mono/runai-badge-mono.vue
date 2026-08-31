<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div :aid="aid" class="runai-badge-mono">
    <span v-if="content" :aid="`${aid}-content`" class="runai-badge-mono-content text-b-mono">
      {{ content }}
    </span>

    <div v-if="showBadgeGroup" :aid="`${aid}-status-group`" class="runai-badge-mono-status-group">
      <span
        v-if="outlinedLabel"
        :aid="`${aid}-outlined-label`"
        class="runai-badge-mono-status-badge runai-badge-mono-status-badge--outline text-b-body1"
      >
        {{ outlinedLabel }}
      </span>

      <span
        v-if="filledLabel"
        :aid="`${aid}-filled-label`"
        class="runai-badge-mono-status-badge runai-badge-mono-status-badge--filled text-b-body1"
      >
        {{ filledLabel }}
      </span>

      <q-icon
        v-if="statusIconName"
        :aid="`${aid}-status-icon`"
        :name="statusIconName"
        :class="statusIconClass"
        size="14px"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

defineOptions({
  name: "runai-badge-mono",
});

const props = defineProps({
  aid: {
    type: String,
    default: "runai-badge-mono",
  },
  content: {
    type: String,
    default: "",
  },
  showBadgeGroup: {
    type: Boolean,
    default: false,
  },
  outlinedLabel: {
    type: String,
    default: "Required",
  },
  filledLabel: {
    type: String,
    default: "Actual",
  },
  status: {
    type: String,
    default: "success",
    validator(value: string) {
      return ["success", "warning", "none"].includes(value);
    },
  },
});

const statusIconName = computed(() => {
  if (props.status === "success") {
    return "ph-bold ph-check-circle";
  }

  if (props.status === "warning") {
    return "ph-bold ph-warning-circle";
  }

  return "";
});

const statusIconClass = computed(() => {
  if (props.status === "warning") {
    return "runai-badge-mono-status-icon--warning";
  }

  return "runai-badge-mono-status-icon--success";
});
</script>

<style scoped lang="scss">
.runai-badge-mono {
  display: inline-flex;
  align-items: center;
  gap: var(--space-s);
  max-width: 100%;
}

.runai-badge-mono-content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-3xs) var(--space-xs);
  background-color: var(--bg-secondary-default);
  border-radius: var(--space-2xs);
  white-space: nowrap;
}

.runai-badge-mono-status-group {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3xs);
}

.runai-badge-mono-status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 var(--space-xs);
  border-radius: var(--space-3xs);
  text-align: center;
  white-space: nowrap;
}

.runai-badge-mono-status-badge--outline {
  color: var(--b-text-sub);
  border: 1px solid var(--border-strong);
}

.runai-badge-mono-status-badge--filled {
  color: var(--text-inverse);
  background-color: var(--feedback-default-bg);
}

.runai-badge-mono-status-icon--success {
  color: var(--text-success);
}

.runai-badge-mono-status-icon--warning {
  color: var(--icon-warning);
}
</style>
