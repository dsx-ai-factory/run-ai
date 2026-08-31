<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="icons-section row q-mr-md">
    <q-icon v-if="isFontIcon" :name="firstIcon" :key="icons[0]?.name" size="26px" color="grey-8" />
    <runai-svg-icon v-else :name="firstIcon" :key="icons[0]?.name" size="26" />
    <q-avatar v-if="showAdditionalsCounter" class="additionals-counter shadow-1" size="sm">
      +{{ additionalsCounter }}
    </q-avatar>
  </section>
</template>

<script lang="ts">
import type { IDrawerListItemIcon } from "@/models/workload-creation-drawer-model";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { defineComponent } from "vue";
import type { PropType } from "vue";

export default defineComponent({
  name: "item-icon-with-counter",
  components: {
    RunaiSvgIcon,
  },
  props: {
    icons: {
      type: Array as PropType<IDrawerListItemIcon[]>,
      required: true,
    },
  },
  computed: {
    firstIcon(): string {
      return this.icons[0]?.icon || "";
    },
    isFontIcon(): boolean {
      return this.firstIcon.startsWith("ph ") || this.firstIcon.startsWith("ph-") || this.firstIcon.startsWith("runai-");
    },
    additionalsCounter(): number {
      return (this.icons?.length || 0) - 1;
    },
    showAdditionalsCounter(): boolean {
      return this.additionalsCounter >= 1;
    },
  },
});
</script>
<style lang="scss" scoped>
.icons-section {
  justify-content: flex-start;
  position: relative;
  .additionals-counter {
    background-color: $body-background-color;
    position: absolute;
    top: 0;
    left: 0;
    transform: translate(50%, 50%);
  }
}
</style>
