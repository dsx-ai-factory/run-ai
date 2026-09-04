<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="host-path-data-mount" default-opened label="Data mount" should-validate>
    <template #subheader>
      <section class="summary row">
        Origin:
        <div class="ellipsis-l">&lrm;{{ hostPathSummary }}</div>
        <span> > </span> Target:
        <div class="ellipsis-l">&lrm;{{ containerPathSummary }}</div>
      </section>
    </template>
    <section>
      <div class="q-mb-md">
        <div class="q-py-md">Set the data origin</div>
        <q-input
          aid="host-path-input"
          :model-value="modelValue.path"
          no-error-icon
          label="Host path"
          stack-label
          placeholder="e.g., /usr/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, path: $event })"
        />
      </div>

      <div>
        <div class="q-py-md">{{ containerPathTitle }}</div>
        <q-input
          aid="mount-path-input"
          :model-value="modelValue.mountPath"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/local/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, mountPath: $event })"
        />
      </div>
    </section>
    <section>
      <runai-sub-expansion-item label="Restrictions">
        <section class="row items-center q-mb-md">
          <q-toggle
            aid="read-only-toggle"
            :model-value="modelValue.readOnly"
            label="Prevent data modification"
            left-label
            class="text-no-wrap"
            @update:model-value="$emit('update:model-value', { ...modelValue, readOnly: $event })"
          />
          <runai-tooltip
            class="q-ml-md"
            tooltip-position="right"
            tooltip-text="When enabled, the data will be mounted with read-only permissions."
          />
        </section>
      </runai-sub-expansion-item>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSubExpansionItem } from "@/components/common/runai-sub-expansion-item";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// Types
import type { HostPathSpec } from "@/swagger-models/assets-service-client";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSubExpansionItem,
    RunaiTooltip,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<HostPathSpec>,
      required: true,
    },
  },
  data() {
    return {
      validationRules,
    };
  },
  computed: {
    hostPathSummary(): string {
      return `${this.modelValue.path ? this.modelValue.path : "None"}`;
    },
    containerPathSummary(): string {
      return `${this.modelValue.mountPath ? this.modelValue.mountPath : "None"}`;
    },
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
  },
});
</script>
<style lang="scss" scoped>
.summary {
  white-space: pre;
}
.ellipsis-l {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100px;
  direction: rtl;
  text-align: left;
}
</style>
