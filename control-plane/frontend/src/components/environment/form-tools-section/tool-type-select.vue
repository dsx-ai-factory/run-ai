<template>
  <section class="tool-type-select">
    <runai-select
      :model-value="currToolType"
      @update:model-value="updateToolType"
      :options="options"
      label="Tool"
      stack-label
      placeholder="Select tool"
      no-error-icon
      :rules="[isValidTool]"
      :disable="disable"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSelect } from "@/components/common/runai-select";
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  components: {
    RunaiSelect,
  },
  props: {
    toolType: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    options: {
      type: Array as PropType<Array<ISelectOption>>,
      required: true,
    },
    notRequired: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      currToolType: null as ISelectOption | null,
    };
  },
  created() {
    this.initCurrToolType();
  },
  methods: {
    initCurrToolType(): void {
      const currOption: ISelectOption | undefined = this.options.find((option) => option.value === this.toolType);
      if (currOption) this.currToolType = currOption;
    },
    isValidTool(val: ISelectOption): string | boolean {
      if (this.notRequired) return true;
      return !!val?.value || "select tool";
    },
    updateToolType(selectedToolType: ISelectOption | null): void {
      if (!selectedToolType) return;
      this.currToolType = selectedToolType;
      this.$emit("value-changed", selectedToolType);
    },
  },
  watch: {
    toolType: {
      immediate: true,
      handler(): void {
        this.initCurrToolType();
      },
    },
  },
});
</script>

<style lang="scss">
.tool-type-select {
  label.q-field {
    width: 190px;
  }
  .select-placeholder {
    opacity: 0.7;
  }
}
</style>
