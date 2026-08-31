<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-radio-options
    class="q-mt-md"
    title="Set how volumes are used across projects within the selected scope"
    :model-value="modelValue"
    @update:model-value="$emit('update:model-value', $event)"
    :options="volumeSharingOptions"
  />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// Components
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
// Models
import type { ISelectOption } from "@/models/global.model";
import type { StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";

export default defineComponent({
  name: "volume-sharing-section",
  components: {
    RunaiRadioOptions,
  },
  emits: {
    "update:model-value": (_: boolean) => true,
  },
  props: {
    storageClasses: {
      type: Array as PropType<Array<StorageClassV2>>,
      required: true,
    },
    modelValue: {
      type: [Boolean, null] as PropType<boolean | null>,
      required: false,
      default: false,
    },
  },
  computed: {
    dataSharingModel(): boolean {
      return this.modelValue || false;
    },
    volumeSharingOptions(): Array<ISelectOption> {
      return [
        {
          label: "Separate volume for each project",
          value: false,
          toolTip:
            "A separate volume will be created for each project, and projects can only access their allocated storage space",
        },
        {
          label: "Shared volume between all projects ",
          value: true,
          disable: !this.storageClasses.some((sc) => sc.permissions?.allowDataSharing === true),
          disabledTooltip:
            "To enable this, a configured storage class that supports a shared volume is required. Contact your administrator",
          toolTip:
            "All projects will have access to the same underlying storage (best for collaboration or shared data usage)",
        },
      ];
    },
  },
});
</script>
