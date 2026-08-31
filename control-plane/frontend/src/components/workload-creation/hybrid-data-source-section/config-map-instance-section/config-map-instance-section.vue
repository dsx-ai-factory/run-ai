<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="config-map-instance-section">
    <div class="q-mb-md">
      <div class="q-py-md">Set the data origin</div>
      <div class="row">
        <runai-tooltip-wrapper
          :display-tooltip="allowOnlyContainerPath"
          tooltip-text="This was defined when the setup was created and can't be modified"
          anchor="bottom middle"
          self="bottom middle"
          class="col-6"
        >
          <runai-select
            :disable="allowOnlyContainerPath || disable"
            class="configmap-name-select"
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
        </runai-tooltip-wrapper>
      </div>
    </div>

    <div>
      <div class="q-py-md">{{ containerPathTitle }}</div>
      <div class="row">
        <policy-string-field
          v-if="supportsSubPath"
          :disable="disable"
          class="col-6"
          aid="mount-subpath-input"
          :model-value="modelValue.subPath"
          no-error-icon
          label="Sub path"
          stack-label
          placeholder="e.g., local/data"
          input-class="placeholder-italic"
          @update:model-value="$emit('update:model-value', { ...modelValue, subPath: $event })"
          :rules="[validSubPath]"
        />
        <q-tooltip
          class="explanatory-tooltip"
          target=".config-map-mount-subpath-input"
          v-if="disableInfo"
          max-width="350px"
          >{{ disableInfo }}</q-tooltip
        >
      </div>
      <div class="row">
        <policy-string-field
          :disable="disable"
          aid="mount-path-input"
          class="config-map-mount-path-input col-6"
          :model-value="modelValue.mountPath"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/local/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, mountPath: $event })"
        />
        <q-tooltip
          class="explanatory-tooltip"
          target=".config-map-mount-path-input"
          v-if="disableInfo"
          max-width="350px"
          >{{ disableInfo }}</q-tooltip
        >
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { PolicyStringField } from "@/components/common/policy-string-field";
// Models
import type { IUIConfigMapInstance } from "@/models/workload.model";
// Constants
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";
import { isValidSubPath } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "config-map-instance-section",
  components: {
    RunaiSelect,
    RunaiTooltipWrapper,
    PolicyStringField,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIConfigMapInstance>,
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
    allowOnlyContainerPath: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
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
