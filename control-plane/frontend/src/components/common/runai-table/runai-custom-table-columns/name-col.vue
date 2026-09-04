<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td class="name-col">
    <div class="col-container row items-center">
      <span class="q-mr-sm" :class="{ 'icon-container': !brandingLogo }" v-if="icon">
        <img height="34" v-if="brandingLogo" :src="brandingLogo" alt="logo" />
        <runai-svg-icon v-else :key="icon" :name="icon" size="34"></runai-svg-icon>
      </span>
      {{ name }}
    </div>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";

// cmps
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

// services
import { getFieldContent } from "@/utils/table-format.util";

// stores
import { useAppStore } from "@/stores/app.store";

// constants
import { NVIDIA_WORKLOAD_ICON } from "@/common/icons.constant";

export default defineComponent({
  components: {
    RunaiSvgIcon,
  },
  props: {
    row: {
      type: Object as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
    icon: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
    };
  },
  computed: {
    name(): string {
      return getFieldContent(this.column, this.row);
    },
    brandingLogo(): string | undefined {
      return this.icon === NVIDIA_WORKLOAD_ICON && this.appStore.isWhiteLabel
        ? this.appStore.theme?.logo.main.dark.src
        : undefined;
    },
  },
});
</script>

<style lang="scss" scoped>
.name-col {
  .col-container {
    flex-wrap: nowrap;
  }

  .icon-container {
    width: 34px;
    display: flex;
  }
}
</style>
