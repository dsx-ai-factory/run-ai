<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="runai-name-section" :label="sectionLabel" :default-opened="defaultOpened" should-validate>
    <template #subheader>
      <span>{{ summary }}</span>
    </template>
    <section class="row">
      <slot name="name-input">
        <asset-name-input
          v-if="assetKind"
          class="col-8"
          :model-value="modelValue"
          @update:model-value="$emit('update:modelValue', $event)"
          :asset-kind="assetKind"
          :scope-model="scopeModel"
          :original-name="originalName"
          :disable="disableNameInput"
          :input-tooltip="nameInputTooltip"
          :auto-focus="autoFocus"
          :custom-existing-name-validation="customExistingNameValidation"
          label="Name"
          placeholder=""
        />
      </slot>
      <runai-description-input
        class="q-mt-lg col-12"
        v-if="supportDescription"
        :model-value="description"
        @update:model-value="updateDescription"
        :custom-aid="`${assetKind}-description-input`"
      />
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { AssetNameInput } from "@/components/common/asset-name-input";
import { RunaiDescriptionInput } from "@/components/common/runai-description-input";
// Models
import type { AssetKind } from "@/swagger-models/assets-service-client";
import type { IScopeModel, IValidationRule } from "@/models/global.model";

export default defineComponent({
  name: "runai-name-section",
  components: {
    RunaiExpansionItem,
    AssetNameInput,
    RunaiDescriptionInput,
  },
  emits: ["update:modelValue", "update-description"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    label: {
      type: String as PropType<string>,
      required: true,
    },
    assetKind: {
      type: String as PropType<AssetKind>,
      required: false,
    },
    supportDescription: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    description: {
      type: [String, null] as PropType<string | null>,
    },
    originalName: {
      type: String as PropType<string>,
      required: false,
    },
    disableNameInput: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    nameInputTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    customExistingNameValidation: {
      type: Function as PropType<IValidationRule>,
      required: false,
    },
  },
  data() {
    return {
      autoFocus: false as boolean,
    };
  },
  created() {
    this.autoFocus = !!this.$route.query.fromCopyId;
  },
  computed: {
    sectionLabel(): string {
      let label = `${this.label} name`;
      if (this.supportDescription) label += ` & description`;
      return label;
    },
    summary(): string {
      return this.modelValue || "None";
    },
  },
  methods: {
    updateDescription(val: string | number | null): void {
      this.$emit("update-description", val?.toString() || null);
    },
  },
});
</script>
