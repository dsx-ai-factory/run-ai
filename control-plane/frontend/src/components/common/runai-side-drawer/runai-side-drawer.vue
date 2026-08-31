<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-dialog class="runai-side-drawer" :model-value="isOpen" seamless no-backdrop-dismiss position="right">
    <q-card :style="{ width: drawerWidth + 'px' }" class="drawer-wrapper" :class="{ 'disable-selection': resizing }">
      <div
        class="resizer"
        role="separator"
        aria-orientation="vertical"
        v-touch-pan.horizontal.prevent.mouse.preserveCursor="handleResize"
      ></div>
      <div class="drawer-content" :class="{ dark: darkMode && isEmpty }">
        <q-btn
          class="close-button runai-icon-btn--default text-icon-default"
          icon="ph-bold ph-x"
          flat
          round
          @click="close"
          aria-label="Hide details"
        >
          <q-tooltip anchor="top left" :offset="[50, -5]">Hide details</q-tooltip>
        </q-btn>
        <slot name="drawer-content" />
      </div>
    </q-card>
  </q-dialog>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { debounce } from "quasar";
import type { PropType } from "vue";
import type { TouchPanValue } from "quasar";

// Services
import { storageUtil } from "@/utils/storage.util";

// Components

// Models
import type { IDrawerSettings } from "@/models/global.model";

type TouchPanEvent = Parameters<NonNullable<TouchPanValue>>[0];

const RESIZE_RANGE = {
  DEFAULT: 654, // Fixed pixel value for default width
  MIN: 654, // Fixed pixel value for minimum width
  MAX: 0.8, // Percentage for maximum width
};
const GAP_WIDTH = 14;
const SAVE_DRAWER_SETTINGS_DEBOUNCE = 500;
const DRAWER_STORAGE_KEY = "drawerSettings";

export default defineComponent({
  emits: ["open", "close", "resize"],
  props: {
    isEmpty: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    emptyMessage: {
      type: String as PropType<string>,
      default: "",
    },
    isOpen: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    heightFromTop: {
      type: Number as PropType<number>,
      required: true,
    },
    darkMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      desiredWidth: 0 as number,
      widthPercentage: 0.5,
      isResizing: false as boolean,
      originalWidth: 0 as number,
      originalLeft: 0 as number,
    };
  },
  created() {
    this.saveDrawerSettings = debounce(this.saveDrawerSettings, SAVE_DRAWER_SETTINGS_DEBOUNCE);
  },
  mounted() {
    this.initDrawerSettings();
    this.setDrawerPadding();
  },
  computed: {
    drawerWidth(): number {
      return this.desiredWidth;
    },
    resizing(): boolean {
      return this.isResizing;
    },
  },
  methods: {
    open(): void {
      this.$emit("open");
      this.setDrawerPadding();
    },
    close(): void {
      this.$emit("close");
    },
    setDrawerPadding(): void {
      this.$nextTick(() => {
        const innerDialogElement: HTMLElement = document.getElementsByClassName("q-dialog__inner")[0] as HTMLElement;
        if (innerDialogElement) {
          innerDialogElement.style.paddingTop = `${this.heightFromTop}px`;
        }
      });
    },
    handleResize(ev: TouchPanEvent) {
      const left = ev.position?.left ?? 0;
      if (ev.isFirst) {
        this.originalWidth = this.desiredWidth;
        this.originalLeft = left;
      } else {
        const newDelta = left - this.originalLeft;
        // Assigning the Maximum/Minimum width if the widthPercentage is out of range.
        const newWidth = Math.max(
          RESIZE_RANGE.MIN,
          Math.min(window.innerWidth * RESIZE_RANGE.MAX, this.originalWidth - newDelta),
        );
        this.widthPercentage = this.desiredWidth / window.innerWidth;
        this.desiredWidth = newWidth;
        this.saveDrawerSettings();
      }
    },
    getDrawerSettings(): Record<string, IDrawerSettings> | null {
      const data: Record<string, IDrawerSettings> | null =
        storageUtil.get<Record<string, IDrawerSettings>>(DRAWER_STORAGE_KEY);
      return data ? data : null;
    },
    saveDrawerSettings(): void {
      let drawerSettings = this.getDrawerSettings();
      if (!drawerSettings) drawerSettings = {};
      drawerSettings[this.$route.name as string] = {
        isOpen: this.isOpen,
        widthPercentage: this.widthPercentage,
      };
      storageUtil.save<Record<string, IDrawerSettings>>(DRAWER_STORAGE_KEY, drawerSettings);
    },
    initDrawerSettings(): void {
      const drawerSettings: Record<string, IDrawerSettings> | null = this.getDrawerSettings();
      const route: string = this.$route.name as string;
      if (drawerSettings && drawerSettings[route]) {
        if (drawerSettings[route].isOpen) {
          this.$emit("open");
        }
        this.widthPercentage = drawerSettings[route].widthPercentage;
      }
      // Assigning the Maximum/Minimum width if the widthPercentage is out of range.
      this.desiredWidth = Math.max(
        RESIZE_RANGE.MIN,
        Math.min(
          window.innerWidth * RESIZE_RANGE.MAX,
          this.widthPercentage ? window.innerWidth * this.widthPercentage : RESIZE_RANGE.DEFAULT,
        ),
      );
      this.saveDrawerSettings();
    },
  },
  watch: {
    isOpen() {
      this.saveDrawerSettings();
      this.setDrawerPadding();
    },
    heightFromTop() {
      this.setDrawerPadding();
    },
    desiredWidth() {
      this.$emit("resize", this.desiredWidth - GAP_WIDTH + "px");
    },
  },
});
</script>

<style lang="scss">
.runai-side-drawer {
  ::-webkit-scrollbar {
    -webkit-appearance: none;
    width: 7px;
  }

  ::-webkit-scrollbar-thumb {
    border-radius: 2px;
    background-color: rgba(0, 0, 0, 0.5);
    box-shadow: 0 0 1px rgba(255, 255, 255, 0.5);
  }
  .q-dialog__inner {
    padding-bottom: 0 !important;
  }

  .q-dialog__inner--minimized > div {
    max-width: 100% !important;
    border: none;
    border-left: solid px $black-10;
    border-top: solid 1px $black-10;
  }

  .q-dialog__inner > div {
    box-shadow: unset;
  }
}
</style>
<style lang="scss" scoped>
$resizer-width: 14px;

.runai-side-drawer {
  display: flex;
  position: relative;

  .drawer-wrapper {
    width: 100%;
    height: 100%;
    background-color: transparent;
    position: relative;
    display: flex;
    border-radius: 0;
    overflow-x: clip;
    border-top: 0;
  }

  .resizer {
    cursor: ew-resize;
    height: 100%;
    width: $resizer-width;
  }

  .drawer-content {
    border-left: solid 1px $black-10;
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background-color: $white;

    &.dark {
      background-color: $navy;
      color: $white;
    }
  }

  .close-button {
    position: absolute;
    right: $resizer-width;
    top: 6px;
    z-index: 1000;
  }

  .drawer-header {
    height: 51px;
    padding-right: 50px;
    display: flex;
    border-bottom: solid 1px $black-10;
    align-items: center;
    padding-left: $resizer-width;
    font-size: 16px;
    font-weight: 500;
  }

  .drawer-body {
    overflow: hidden;
    height: 100%;
  }
}
</style>
