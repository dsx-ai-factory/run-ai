<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-select
    aid="auto-gen-select"
    :options="options"
    :model-value="currValue"
    label=""
    @update:model-value="optionSelected"
    :disable="disable"
  />
</template>

<script lang="ts">
import type { ISelectOption } from "@/models/global.model";
import { defineComponent, type PropType } from "vue";
import { RunaiSelect } from "@/components/common/runai-select";

export default defineComponent({
  components: {
    RunaiSelect,
  },
  props: {
    customLabel: {
      type: String as PropType<string>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    initValue: {
      type: Boolean as PropType<boolean>,
      required: true,
      default: false,
    },
  },
  data() {
    return {
      currValue: { label: "Auto generate", value: true } as ISelectOption,
      options: [
        {
          label: "Auto generate",
          value: false,
        },
        {
          label: this.customLabel,
          value: true,
        },
      ] as Array<ISelectOption>,
    };
  },
  created() {
    this.init();
  },
  methods: {
    init(): void {
      const currValue: ISelectOption | undefined = this.options.find(
        (opt: ISelectOption): boolean => opt.value === this.initValue,
      );
      if (!currValue) return;
      this.currValue = currValue;
    },
    optionSelected(selectedOption: ISelectOption | null): void {
      if (!selectedOption) return;
      this.currValue = selectedOption;
      this.$emit("value-changed", selectedOption.value);
    },
  },
  watch: {
    initValue: {
      immediate: true,
      handler(): void {
        this.init();
      },
    },
  },
});
</script>

<style lang="scss" scoped></style>
