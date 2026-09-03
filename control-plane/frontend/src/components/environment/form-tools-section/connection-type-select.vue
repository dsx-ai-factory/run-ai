<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="tool-type-select">
    <runai-select
      :model-value="currConnectionType"
      @update:model-value="updateConnectionType"
      :options="options"
      label="Connection type"
      :disable="disable"
      aid="connection-type-select"
    />
    <q-tooltip v-if="disable" max-width="350px">
      {{ toolTipText }}
    </q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IConnectionTypeSelectOption } from "@/models/environment.model";
import { RunaiSelect } from "@/components/common/runai-select";

export default defineComponent({
  components: {
    RunaiSelect,
  },
  props: {
    connectionType: {
      type: String as PropType<string>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    toolTipText: {
      type: String as PropType<string>,
      required: true,
    },
    options: {
      type: Array as PropType<Array<IConnectionTypeSelectOption>>,
      required: true,
    },
  },
  data() {
    return {
      currConnectionType: {} as IConnectionTypeSelectOption,
    };
  },
  created() {
    this.initCurrConnectionType();
  },
  methods: {
    initCurrConnectionType(): void {
      const currOption: IConnectionTypeSelectOption | undefined = this.options.find(
        (option: IConnectionTypeSelectOption): boolean => option.value === this.connectionType,
      );
      if (!currOption) return;
      this.currConnectionType = currOption;
    },
    updateConnectionType(selectedConnection: IConnectionTypeSelectOption | null): void {
      if (!selectedConnection) return;
      this.currConnectionType = selectedConnection;
      this.$emit("value-changed", selectedConnection);
    },
  },
  watch: {
    connectionType: {
      immediate: true,
      handler(): void {
        this.initCurrConnectionType();
      },
    },
  },
});
</script>

<style lang="scss" scoped>
.tool-type-select {
  label.q-field {
    width: 190px;
  }
}
</style>
