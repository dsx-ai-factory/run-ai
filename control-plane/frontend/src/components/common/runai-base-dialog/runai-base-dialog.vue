<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-dialog
    class="runai-base-dialog"
    :model-value="modelValue"
    @hide="$emit('close')"
    @escape-key="$emit('close')"
    :persistent="isModalPersistent"
    :position="position"
    @show="$emit('dialog-loaded')"
    @keyup.enter="handleEnterKey"
  >
    <q-card class="dialog-content-container" :class="[customClass, size]">
      <q-card-section class="runai-base-dialog-header row justify-between items-center no-wrap">
        <h6 class="ellipsis">
          <b><slot name="header"></slot></b>
        </h6>
        <q-btn
          v-if="!hideX"
          class="runai-base-dialog-close-btn runai-icon-btn--large text-icon-default"
          icon="ph-bold ph-x"
          flat
          round
          @click="$emit('close')"
          aria-label="Close dialog"
        />
      </q-card-section>

      <q-separator />
      <q-card-section v-if="$slots.subheader">
        <slot name="subheader"></slot>
      </q-card-section>
      <q-separator v-if="$slots.subheader && !hideSubheaderSeparator" />

      <q-card-section
        class="runai-base-dialog-body"
        :class="[noPadding ? 'no-padding' : '', scroll ? 'scroll' : '', dialogBodySize ? dialogBodySize : '']"
        :style="bodyStyles"
      >
        <q-linear-progress v-if="loading" indeterminate />
        <slot name="body" v-else></slot>
      </q-card-section>

      <q-separator />

      <q-card-actions class="runai-base-dialog-footer row">
        <div class="footer-left">
          <slot name="footer-left"></slot>
        </div>
        <div class="footer">
          <slot name="footer"></slot>
        </div>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export interface IStyles {
  minHeight?: string;
  maxHeight?: string;
}

export default defineComponent({
  emits: ["close", "dialog-loaded", "enter-pressed"],
  props: {
    modelValue: {
      type: Boolean,
      required: true,
    },
    customClass: {
      type: String,
      required: false,
    },
    size: {
      type: String as PropType<"width-xs" | "width-sm" | "width-md" | "width-lg" | "width-xl" | "dynamic">,
      required: false,
      default: "",
    },
    noPadding: {
      type: Boolean,
      required: false,
    },
    position: {
      type: String as PropType<"standard" | "top" | "right" | "bottom" | "left" | undefined>,
      required: false,
    },
    scroll: {
      type: Boolean,
      required: false,
      default: true,
    },
    bodyMinHeight: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    bodyMaxHeight: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    isModalPersistent: {
      type: Boolean,
      default: false,
    },
    hideX: {
      type: Boolean,
      default: false,
    },
    hideSubheaderSeparator: {
      type: Boolean,
      default: false,
    },
    dialogBodySize: {
      type: String as PropType<"medium" | "large" | "extra-large">,
      default: "medium",
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    onEnterCallback: {
      type: Function as PropType<() => void>,
      required: false,
    },
  },
  computed: {
    bodyStyles(): IStyles {
      const styles: IStyles = {};
      if (this.bodyMinHeight) {
        styles.minHeight = this.bodyMinHeight;
      }

      if (this.bodyMaxHeight) {
        styles.maxHeight = this.bodyMaxHeight;
      }
      return styles;
    },
  },
  methods: {
    handleEnterKey() {
      if (this.onEnterCallback) {
        this.onEnterCallback();
      }
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-base-dialog {
  .dialog-content-container {
    &.width-xs {
      width: 300px;
    }

    &.width-sm {
      width: 540px;
    }

    &.width-md {
      width: 700px;
      max-width: 80vw;
    }

    &.width-lg {
      width: 1000px;
      max-width: 80vw;
    }
    &.width-xl {
      min-width: 80vw;
    }
    &.dynamic {
      min-width: 300px;
      max-width: 80vw;
    }
  }

  .runai-base-dialog-body {
    .extra-large.scroll {
      max-height: 80vh;
    }
    &.large.scroll {
      max-height: 70vh;
    }
    &.medium.scroll {
      max-height: 70vh;
    }
  }

  .runai-base-dialog-footer {
    .footer-left {
      flex: 1;
      &:empty {
        display: none;
        flex: unset;
      }
    }
    .footer {
      display: flex;
      justify-content: flex-end;
      flex: 1;
    }
  }
}
</style>

<style lang="scss">
.runai-base-dialog {
  .runai-base-dialog-footer .q-btn + .q-btn {
    margin-left: 8px;
    padding-inline: 16px;
  }
}
</style>
