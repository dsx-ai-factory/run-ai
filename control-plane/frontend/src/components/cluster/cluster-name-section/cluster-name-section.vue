<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="cluster-name"
    label="Cluster name"
    :default-opened="!isClusterEditMode"
    :disable-opening="isClusterEditMode"
    :hide-expend-icon="isClusterEditMode"
    :section-invalid="sectionInvalid"
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>
    <section class="row">
      <div class="col-8">
        <runai-name-validation
          :model-value="modelValue"
          @update:model-value="$emit('update:modelValue', $event)"
          :rules="[isEmpty, isValidFormat, isNameExists]"
          @invalid="nameAlreadyExist = $event"
          debounce="300"
        />
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";

import { errorMessages } from "@/common/error-message.constant";
import { isValidRunaiEntityNameFormat, isEmpty, isNotEmpty } from "@/common/form.validators";
import { useClusterStore } from "@/stores/cluster.store";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiNameValidation,
  },
  emits: ["update:modelValue", "is-section-invalid"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    isClusterEditMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      nameAlreadyExist: false as boolean,
    };
  },
  computed: {
    sectionInvalid(): boolean {
      if (this.isClusterEditMode) return isEmpty(this.modelValue);
      return isEmpty(this.modelValue) || !isValidRunaiEntityNameFormat(this.modelValue) || this.nameAlreadyExist;
    },
    summary(): string {
      return this.modelValue || "None";
    },
  },
  methods: {
    isEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    isValidFormat(val: string): boolean | string {
      if (this.isClusterEditMode) return true;
      return isValidRunaiEntityNameFormat(val) || errorMessages.VALID_FORMAT;
    },
    isNameExists(val: string): boolean | string {
      if (this.isClusterEditMode) return true;
      return !useClusterStore().isClusterNameExists(val) || errorMessages.NAME_ALREADY_EXIST;
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
