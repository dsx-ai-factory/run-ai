<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-text-editor column">
    <q-editor
      :model-value="modelValue"
      @update:model-value="$emit('update:model-value', $event)"
      :toolbar="[
        [
          {
            icon: 'ph-light ph-text-aa',
            fixedIcon: true,
            options: ['size-3', 'size-4', 'size-5'],
            list: 'no-icons',
          },
        ],
        ['bold', 'italic', 'underline', 'link'],
      ]"
      :placeholder="placeholder"
      square
      :disable="disable"
    />
    <div v-if="hasFooter" class="editor-footer q-pa-sm">
      <slot name="footer"></slot>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "runai-text-editor",
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    placeholder: {
      type: String as PropType<string>,
      default: "Type your message here",
    },
    hideFooter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    hasFooter(): boolean {
      return !!this.$slots.footer;
    },
  },
});
</script>

<style lang="scss">
.runai-text-editor {
  .q-editor {
    word-break: break-word;
    .q-btn-item {
      padding: 8px;
      &:hover {
        text-decoration: none;
      }
      .q-icon {
        font-weight: 700;
        font-size: 20px;
      }
    }
  }
  .editor-footer {
    border: 1px solid $black-12;
    border-top: none;
  }
}
</style>
