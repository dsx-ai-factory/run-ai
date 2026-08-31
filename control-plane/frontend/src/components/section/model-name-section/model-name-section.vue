<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    :default-opened="defaultOpen"
    label="Model name"
    :disable-closing="disableClosing"
    :section-invalid="sectionInvalid"
    class="model-name-section"
    aid="model-name-section"
  >
    <template #subheader>
      <span aid="model-name-section-summary">{{ summary }}</span>
    </template>
    <section class="main-section">
      <div class="row items-center q-mb-md">
        <span class="q-mr-md">Set the model name as displayed in Hugging Face</span>
        <runai-tooltip tooltip-text="Copy the full model name from Hugging Face and paste it here" width="370" />
      </div>
      <div class="row">
        <q-input
          aid="model-name-input"
          :model-value="modelName.name"
          placeholder="e.g., meta-llama/Meta-Llama-3-8B"
          input-class="placeholder-italic"
          label="Model name"
          stack-label
          @update:model-value="nameChanged"
          :rules="[isNameEmpty, isNameFormatValid]"
          no-error-icon
          reactive-rules
          class="col-6"
        ></q-input>
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// models
import type { IUIWorkloadModelNameSectionModel } from "./model-name-section.model";
import type { EWorkloadFormType } from "@/models/workload.model";

// others
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty, isEmpty, huggingFaceModelNameValid } from "@/common/form.validators";

import { is } from "quasar";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiTooltip,
  },
  emits: ["model-name-changed", "is-section-invalid"],
  props: {
    entityType: {
      type: String as PropType<EWorkloadFormType>,
      required: true,
    },
    modelName: {
      type: Object as PropType<IUIWorkloadModelNameSectionModel>,
      required: true,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    defaultOpen: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  computed: {
    sectionInvalid(): boolean {
      return isEmpty(this.modelName.name) || !huggingFaceModelNameValid(this.modelName.name);
    },
    summary(): string {
      return this.modelName.name || "None";
    },
  },
  methods: {
    isNameEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    isNameFormatValid(val: string): boolean | string {
      return huggingFaceModelNameValid(val) || errorMessages.ENTER_A_VALID_NAME;
    },
    nameChanged(value: string | number | null): void {
      let name = "";
      if (is.number(value)) {
        name = String(value);
      } else {
        name = value || "";
      }
      this.$emit("model-name-changed", { ...this.modelName, name });
    },
  },
  watch: {
    sectionInvalid: {
      handler(newVal: boolean): void {
        this.$emit("is-section-invalid", newVal);
      },
      immediate: true,
    },
  },
});
</script>
