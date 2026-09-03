<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card flat square class="wrapper" :class="{ 'q-mx-md': !noGutters }">
    <template v-if="title">
      <div class="row items-center q-pa-md">
        <q-icon v-if="phIconName" class="q-mr-sm" :name="phIconName" size="22px" color="black-54" />
        <runai-svg-icon v-else-if="iconName" class="q-mr-sm" :name="iconName" size="22" />
        <div class="q-pt-xs italic text-black-54">{{ title }}</div>
      </div>
      <q-separator />
    </template>

    <div class="row items-center">
      <div class="col-11">
        <q-card-section> <slot /> </q-card-section>
      </div>
      <div class="col-1">
        <runai-tooltip
          v-if="resetTooltipText"
          aid="reset-button"
          icon="ph-bold ph-arrow-counter-clockwise"
          @click="$emit('reset')"
          :tooltip-text="resetTooltipText"
        />
        <q-btn
          v-else-if="!hideCloseButton"
          @click="$emit('close')"
          size="12px"
          round
          flat
          color="black-54"
          aid="close-button"
          icon="ph-bold ph-x"
          aria-label="Remove"
        ></q-btn>
      </div>
    </div>
  </q-card>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export default defineComponent({
  components: { RunaiTooltip, RunaiSvgIcon },
  emits: ["close", "reset"],
  props: {
    hideCloseButton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    noGutters: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    title: {
      type: String as PropType<string>,
      required: false,
    },
    iconName: {
      type: String as PropType<string>,
      required: false,
    },
    phIconName: {
      type: String as PropType<string>,
      required: false,
    },
    resetTooltipText: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>
<style scoped lang="scss">
.wrapper {
  background-color: $body-background-color;
  margin-bottom: 10px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
</style>
