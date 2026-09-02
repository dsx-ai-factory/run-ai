<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-expansion-item
    class="runai-sub-expansion-text-item"
    expand-icon-class="runai-expansion-icon"
    expand-icon="ph-bold ph-caret-down"
    expand-separator
    expand-icon-toggle
    v-model="isOpen"
  >
    <template #header>
      <div
        class="runai-sub-expansion-header"
        role="button"
        tabindex="0"
        :aria-expanded="isOpen"
        @click="isOpen = !isOpen"
        @keydown.enter.prevent="isOpen = !isOpen"
        @keydown.space.prevent="isOpen = !isOpen"
        :aid="aid"
      >
        <slot name="header">
          {{ label }}
        </slot>
      </div>
    </template>
    <slot name="body" />
  </q-expansion-item>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";

export default defineComponent({
  props: {
    label: {
      type: String as PropType<string>,
      required: false,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      isOpen: !!this.defaultOpened,
    };
  },
});
</script>

<style lang="scss">
.runai-sub-expansion-text-item {
  width: 100%;

  hr {
    display: none;
  }
  .runai-expansion-icon {
    color: var(--icon-default, #707070);
    .q-icon {
      font-size: var(--icon-size-default) !important;
    }
  }
  .q-item[role="listitem"] {
    padding: 0;
  }
  .q-item.q-hoverable {
    margin-top: 20px;
    padding: 10px 0;

    &:hover {
      .q-focus-helper {
        opacity: 0;
      }
    }
  }

  .runai-sub-expansion-header {
    font-weight: 500;
    color: $primary;
    display: flex;
    align-items: center;
    padding-right: 20px;
    cursor: pointer;
  }
}
</style>
