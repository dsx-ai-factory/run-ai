<template>
  <runai-tooltip-wrapper
    anchor="bottom middle"
    self="top middle"
    :display-tooltip="isDisabled"
    :tooltip-text="disabledTooltip"
  >
    <q-btn
      class="runai-dropdown"
      :class="customClasses"
      :label="label"
      :icon="icon"
      :flat="flat"
      :color="color"
      :disable="isDisabled"
      :size="size"
      :stack="stack"
      :round="round"
      :aria-label="ariaLabel || label || 'More actions'"
    >
      <div class="btn-container row items-center">
        <slot name="btn"></slot>
      </div>
      <q-menu :fit="fit" :auto-close="autoClose">
        <q-list role="menu">
          <q-item
            class="row items-center runai-dropdown-option"
            clickable
            role="menuitem"
            v-for="opt in options"
            :key="opt[rowKey]"
            @click="$emit('on-selected', opt)"
            :aid="opt.aid"
            :disable="opt.disable"
          >
            <span v-if="opt.svg" style="width: 25px" class="row items-center q-mr-sm">
              <runai-svg-icon :name="opt.svg" size="25px" />
            </span>

            <span v-if="opt.icon" style="width: 25px" class="row items-center q-mr-sm">
              <q-icon :name="opt.icon" />
            </span>

            <slot v-if="$slots.option" name="option" v-bind="opt"></slot>
            <div v-else>{{ opt.label }}</div>
            <runai-tooltip v-if="opt.tooltip" :tooltip-text="opt.tooltip" />
          </q-item>
        </q-list>
      </q-menu>
    </q-btn>
  </runai-tooltip-wrapper>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { RunaiTooltipWrapper } from "../runai-tooltip-wrapper";
import { RunaiSvgIcon } from "../runai-svg-icon";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export default defineComponent({
  components: { RunaiTooltip, RunaiTooltipWrapper, RunaiSvgIcon },
  emits: ["on-selected"],
  props: {
    options: {
      type: Array as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    rowKey: {
      type: String as PropType<string>,
      required: false,
      default: "id",
    },
    autoClose: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    isDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disabledTooltip: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    fit: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    flat: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    icon: {
      type: String as PropType<string>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    color: {
      type: String as PropType<string>,
      required: false,
    },
    size: {
      type: String as PropType<string>,
      required: false,
    },
    customClasses: {
      type: String as PropType<string>,
      required: false,
    },
    stack: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    round: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    ariaLabel: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-dropdown {
  .runai-dropdown-option {
    flex-direction: row;
    gap: 10px;
    justify-content: flex-start;
    align-items: center;
  }
}
.q-item {
  min-width: 173px;
}
</style>
