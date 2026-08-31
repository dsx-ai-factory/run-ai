<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="config-map-data-mount" default-opened label="Data mount" should-validate>
    <template #subheader>
      <section class="summary row">
        Origin:
        <div class="ellipsis-l">&lrm;{{ configMapNameSummary }}</div>
        <span> > </span> Target:
        <div class="ellipsis-l">&lrm;{{ containerPathSummary }}</div>
      </section>
    </template>
    <section>
      <runai-section title="Set the data origin">
        <div class="row q-mb-lg">
          <runai-select
            class="configmap-name-select col-7"
            standard
            :model-value="modelValue.configMap"
            :options="configMapNames"
            label="ConfigMap name"
            stack-label
            placeholder="Select..."
            :rules="[validationRules.validateNameSelected]"
            @update:model-value="$emit('update:model-value', { ...modelValue, configMap: $event })"
            :loading="loadingConfigMapNames"
            no-option-text="No names found"
          />
        </div>
      </runai-section>

      <runai-section :title="containerPathTitle">
        <div class="row q-mb-lg" v-if="supportsSubPath">
          <q-input
            class="col-7"
            aid="mount-subpath-input"
            :model-value="modelValue.subPath"
            no-error-icon
            label="Sub path"
            stack-label
            placeholder="e.g., local/data"
            input-class="placeholder-italic"
            :rules="[validSubPath]"
            @update:model-value="$emit('update:model-value', { ...modelValue, subPath: $event })"
          />
        </div>
        <div class="row">
          <q-input
            class="col-7"
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
      </runai-section>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { isValidSubPath } from "@/common/form.validators";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";

// Types
import { RunaiSelect } from "@/components/common/runai-select";
import type { ConfigMapAssetSpec } from "@/swagger-models/assets-service-client";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";
import { RunaiSection } from "@/components/common/runai-section";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSelect,
    RunaiSection,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<ConfigMapAssetSpec>,
      required: true,
    },
    configMapNames: {
      type: Array as PropType<string[]>,
      required: true,
    },
    loadingConfigMapNames: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    supportsSubPath: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      validationRules,
    };
  },
  computed: {
    configMapNameSummary(): string {
      return `${this.modelValue.configMap ? this.modelValue.configMap : "None"}`;
    },
    containerPathSummary(): string {
      return `${this.modelValue.mountPath ? this.modelValue.mountPath : "None"}`;
    },
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
  },
  methods: {
    validSubPath(subPath: string | undefined): boolean | string {
      if (!subPath) return true;
      return isValidSubPath(subPath) || errorMessages.INVALID_SUB_PATH;
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
