<template>
  <runai-section title="Select a source to load the model from" class="model-source-selector">
    <div class="row">
      <runai-select
        class="model-source-selector__select col-6"
        aid="model-source-selector-select"
        label="Source"
        stack-label
        :options="sourceOptions"
        emit-value
        :model-value="modelValue"
        @update:model-value="$emit('update:model-value', $event)"
        :rules="rules"
        placeholder="Select a repository"
      />
    </div>
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
import { errorMessages } from "@/common/error-message.constant";
import { validationRules } from "@/common/validation-rules";
import type { ISelectOption } from "@/models/global.model";
import { EModelSource } from "./model-source-selector.model";
import { allWorkloadsIconsMap } from "@/common/icons.constant";

const SOURCE_OPTIONS: ISelectOption[] = [
  { label: "NVIDIA NGC catalog", value: EModelSource.Nim, icon: allWorkloadsIconsMap.Nvidia },
  { label: "Hugging Face", value: EModelSource.HuggingFace, icon: allWorkloadsIconsMap.HuggingFace },
];

export default defineComponent({
  name: "ModelSourceSelector",
  components: {
    RunaiSection,
    RunaiSelect,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: String as PropType<EModelSource | "">,
      default: () => EModelSource.Nim,
    },
    required: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  computed: {
    sourceOptions(): ISelectOption[] {
      return SOURCE_OPTIONS;
    },
    rules() {
      if (!this.required) return [];
      return [(value: string) => validationRules.validateValueSelected(value) || errorMessages.SELECT_A_REPOSITORY];
    },
  },
});
</script>
