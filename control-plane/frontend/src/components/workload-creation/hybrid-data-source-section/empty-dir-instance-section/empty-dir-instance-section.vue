<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="empty-dir-instance-section">
    <div class="row">
      <volume-size-section
        class="col-6"
        :model-value="modelValue.sizeLimit || ''"
        @update:model-value="sizeChanged"
        :policy-rules="policyRules?.sizeLimit"
        :disable="disable"
        :disable-info="disableInfo"
      />
    </div>

    <div>
      <div class="q-py-md">{{ containerPathTitle }}</div>
      <div class="row">
        <policy-string-field
          :disable="disable"
          :tooltip="disableInfo"
          class="col-6"
          aid="mount-path-input"
          :model-value="modelValue.path"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/local/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="updatePath"
          :policy-rules="policyRules?.path"
        />
      </div>
    </div>

    <runai-radio-options
      class="q-mt-md q-mb-xs"
      title="Storage medium"
      :options="emptyDirVolumeOptions"
      :model-value="modelValue.medium || null"
      @update:model-value="updateEmptyDirVolume"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Models
import type { IUIEmptyDirInstance } from "@/models/workload.model";
import type { EmptyDirRules } from "@/swagger-models/policy-service-client";
import type { ISelectOption } from "@/models/global.model";
// Components
import { VolumeSizeSection } from "@/components/section/volume-section/volume-box/volume-size-section";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
// Utils
import { assetsUtil } from "@/utils/assets.util/assets.util";
// Constants
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  name: "empty-dir-instance-section",
  components: { VolumeSizeSection, PolicyStringField, RunaiRadioOptions },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIEmptyDirInstance>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<EmptyDirRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
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
    emptyDirVolumeOptions(): ISelectOption[] {
      return [
        {
          label: "Disk-backed storage",
          value: null,
          toolTip:
            "The data is stored on the node's filesystem, which persists across container restarts but not pod rescheduling, and is slower than memory",
        },
        {
          label: "Memory-backed storage",
          value: "Memory",
          toolTip:
            "The data is stored in RAM, which provides faster access but is not persistent, and will be lost if the pod restarts",
        },
      ];
    },
  },
  methods: {
    sizeChanged(size: string): void {
      this.$emit("update:model-value", { ...this.modelValue, sizeLimit: size });
    },
    updatePath(path: string): void {
      this.$emit("update:model-value", { ...this.modelValue, path });
    },
    updateEmptyDirVolume(medium: string | null): void {
      this.$emit("update:model-value", { ...this.modelValue, medium });
    },
  },
});
</script>
