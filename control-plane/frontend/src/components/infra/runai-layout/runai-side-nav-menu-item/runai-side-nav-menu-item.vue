<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-item
    clickable
    v-ripple
    role="button"
    :aria-label="displayLabel"
    :aria-current="menuItem.isActive ? 'true' : undefined"
    @click="onClick"
    @keydown.enter.prevent="onClick"
    @keydown.space.prevent="onClick"
    :class="{ 'is-active-link': menuItem.isActive || isItemOrSubMenuHovered }"
    :aid="menuItem.aid"
    class="runai-side-nav-menu-item"
    @mouseenter="showHoverNavigation"
    @mouseleave="isItemHovered = false"
    @focusin="showHoverNavigation"
    @focusout="isItemHovered = false"
  >
    <q-icon :name="menuItem.icon" size="20px"> </q-icon>
    <span class="item-label">{{ displayLabel }}</span>
    <q-popup-proxy :offset="[-64, -64]" no-parent-event :ref="popupRef" :breakpoint="0" @hide="isMenuHovered = false">
      <runai-sub-menu
        ref="subMenuRef"
        @mouseenter="onSubMenuHovered"
        @mouseleave="isMenuHovered = false"
        @focusin="onSubMenuHovered"
        @focusout="isMenuHovered = false"
        :items="subMenuItems"
        :header-text="displayLabel"
      />
    </q-popup-proxy>
  </q-item>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiSubMenu } from "@/components/infra/runai-layout/runai-sub-menu";

// models
import type { IMenuItem, ISubjectRoutes } from "../routes-model/routes-model";
import type { QPopupProxy } from "quasar";

export default defineComponent({
  name: "runai-side-nav-menu-btn",
  components: {
    RunaiSubMenu,
  },
  emits: ["item-clicked", "item-hovered", "item-out-of-focus"],
  props: {
    menuItem: {
      type: Object as PropType<IMenuItem>,
      required: true,
    },
    subMenuItems: {
      type: Array as PropType<ISubjectRoutes[]>,
      required: true,
    },
  },
  data() {
    return {
      popupRef: `${this.menuItem.value}-popup` as string,
      isItemHovered: false as boolean,
      isMenuHovered: false as boolean,
      hideMenuTimeOut: null as null | ReturnType<typeof setTimeout>,
    };
  },
  computed: {
    isItemOrSubMenuHovered(): boolean {
      return this.isItemHovered || this.isMenuHovered;
    },
    displayLabel(): string {
      const label = this.menuItem.label;
      return label.includes("\n") ? label.split("\n")[0] : label;
    },
  },
  methods: {
    onClick(): void {
      this.$emit("item-clicked", this.menuItem);
    },
    showHoverNavigation(): void {
      if (!this.$refs?.[this.popupRef]) return;
      this.isItemHovered = true;
      this.$emit("item-hovered", this.$refs?.[this.popupRef]);
    },
    onSubMenuHovered(): void {
      if (!this.$refs.subMenuRef) return;
      this.isMenuHovered = true;
    },
  },
  watch: {
    isItemOrSubMenuHovered: {
      handler(newValue: boolean): void {
        if (newValue) {
          if (this.hideMenuTimeOut) {
            clearTimeout(this.hideMenuTimeOut);
            this.hideMenuTimeOut = null;
          }
        } else {
          this.hideMenuTimeOut = setTimeout(() => {
            (this.$refs?.[this.popupRef] as QPopupProxy)?.hide();
            this.hideMenuTimeOut = null;
          }, 600);
        }
      },
    },
  },
});
</script>
<style scoped lang="scss">
.runai-side-nav-menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-evenly !important;
  padding: 2px;
  width: 64px;
  height: 64px;
  border-radius: 5px;
  color: $white-80;

  &:deep(.q-focus-helper) {
    display: none;
  }

  &.is-active-link {
    background-color: $side-nav-selected-issue-background;
    color: $white;
  }

  .item-label {
    white-space: break-spaces;
    font-size: 10px;
    text-align: center;
    line-height: 11px;
  }
}
</style>
