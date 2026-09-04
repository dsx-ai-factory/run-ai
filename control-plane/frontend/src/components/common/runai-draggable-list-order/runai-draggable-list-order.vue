<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section
    class="runai-draggable-list-order"
    :class="{ 'no-border': !showBorder }"
    :style="{ '--item-size': `${itemSizeMap[itemSize]}px` }"
  >
    <section class="list-container">
      <ul class="list-numbers-col">
        <li class="title" v-if="showHeader">Order of priority</li>

        <li
          v-for="item in listOfNumbers"
          class="item"
          :class="{ 'item-number': itemSize === 'sm' }"
          :key="item"
          :style="currentHeight"
          :disabled="disabled || null"
        >
          {{ item }}
        </li>
      </ul>

      <ul class="main-list-col" ref="main-list-col" :class="{ 'list-disabled': disabled }">
        <li class="title" v-if="showHeader">
          <span>{{ mainListTitle }}</span>
          <runai-tooltip
            v-if="!disabled"
            icon="ph-bold ph-arrow-counter-clockwise"
            @click="onReset"
            :tooltip-text="resetTooltipText"
          />
        </li>
        <vue-draggable
          class="dragArea list-group w-full"
          ghost-class="ghost"
          :model-value="list"
          :disabled="disabled"
          @update:model-value="onChange"
        >
          <transition-group type="transition" name="flip-list">
            <li class="item" :disabled="disabled || null" v-for="item in list" :key="item[rowKey]">
              <q-tooltip v-if="disabled">{{ disabledTooltipText }}</q-tooltip>
              <q-icon
                v-if="!disabled"
                name="ph-bold ph-dots-six-vertical"
                class="draggable-icon"
                size="24px"
                style="color: rgba(0, 0, 0, 0.12)"
              ></q-icon>

              <div class="item-content">
                <slot name="list-item" v-bind="{ ...item, disabled }"></slot>
              </div>
            </li>
          </transition-group>
        </vue-draggable>
      </ul>
    </section>
    <div class="footer" v-if="$slots.footer">
      <slot name="footer"></slot>
    </div>
  </section>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";

// cmps
import { VueDraggable } from "vue-draggable-plus";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// constants
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "runai-draggable-list-order",
  components: {
    VueDraggable,
    RunaiTooltip,
  },
  emits: ["change", "reset"],
  props: {
    list: {
      type: Array as PropType<Array<any>>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    rowKey: {
      type: String as PropType<string>,
      required: false,
      default: "id",
    },
    mainListTitle: {
      type: String as PropType<string>,
      required: false,
    },
    resetTooltipText: {
      type: String as PropType<string>,
      required: false,
      default: "Reset to default",
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disabledTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    showHeader: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    showBorder: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    itemSize: {
      type: String as PropType<"sm" | "md">,
      required: false,
      default: "sm",
    },
  },
  data() {
    return {
      defaultHeight: 50 as number,
      itemHeight: null as number | null,
      itemSizeMap: {
        sm: 50,
        md: 80,
      },
    };
  },
  mounted() {
    const mainListCol = this.$refs["main-list-col"] as HTMLElement;
    this.itemHeight = mainListCol.querySelector(".item")?.clientHeight || null;
  },
  computed: {
    listOfNumbers(): Array<number> {
      return Array(this.list.length)
        .fill(0)
        .map((_, i) => i + 1);
    },
    currentHeight(): string {
      return `height: ${this.itemHeight ? this.itemHeight : this.defaultHeight}px`;
    },
    disabledTooltipText(): string {
      return this.disabledTooltip || errorMessages.CANT_BE_MODIFIED_POLICY;
    },
  },
  methods: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onChange(updatedList: Array<any>): void {
      this.$emit("change", updatedList);
    },
    onReset(): void {
      this.$emit("reset");
    },
  },
});
</script>

<style lang="scss">
.ghost {
  background: transparent !important;
  border: $black-54 1px dashed;

  > div,
  > svg {
    opacity: 0;
  }
}

.flip-list-move {
  transition: transform 0.5s;
}
</style>

<style lang="scss" scoped>
.runai-draggable-list-order {
  --item-size: 50px; /* Default fallback */
  border: 1px solid $black-12;
  padding: 5px 15px 15px 15px;

  &.no-border {
    border: none;
    padding: 0;
  }

  .list-container {
    display: flex;

    &:not(:last-child) {
      margin-bottom: 10px;
    }
  }

  .main-list-col {
    flex: 1;

    .title {
      padding: 0 10px 0 54px;
      justify-content: space-between;
    }

    &.list-disabled {
      .title {
        padding-left: 20px;
      }
    }
  }

  .item {
    width: 100%;
    min-height: var(--item-size);
    background-color: $body-background-color;
    border-radius: 4px;
    cursor: move;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 0 20px;

    &:hover {
      background-color: $grey-4;
    }
    &.item-number {
      width: var(--item-size);
    }

    &:not(:last-child) {
      margin-bottom: 10px;
    }
  }

  .list-numbers-col {
    padding-right: 10px;
  }

  .title {
    height: 42px;
    font-size: 12px;
    line-height: 14px;
    margin-bottom: 5px;
    display: flex;
    align-items: center;
  }

  .item-content {
    flex: 1;
    padding-left: 10px;

    &:first-child {
      padding-left: 0px;
    }
  }
}
</style>
