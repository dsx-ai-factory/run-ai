<template>
  <q-btn-dropdown
    :style="customStyle"
    :size="size"
    :flat="flat"
    :dropdown-icon="dropdownIcon"
    :no-icon-animation="noIconAnimation"
    :color="color"
    :class="customClasses"
    :split="split"
    :label="split ? label : ''"
    @click.stop="split && $emit('on-select', options[0].value)"
    :loading="loading"
  >
    <q-list>
      <q-item
        clickable
        v-close-popup
        v-for="item in options"
        :key="item.label"
        :active="item.value === selectedValue"
        :focused="item.value === selectedValue"
        @click="$emit('on-select', item.value)"
        :aid="getAid(item.aid || item.label)"
      >
        <q-item-section>
          <div class="row items-center">
            <q-item-label>{{ item.label }}</q-item-label>
            <runai-tooltip
              class="q-ml-sm"
              v-if="item.toolTip"
              :tooltip-text="item.toolTip"
              :tooltip-position="tooltipPosition"
            />
          </div>
        </q-item-section>
      </q-item>
    </q-list>
  </q-btn-dropdown>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// models
import type { ISelectOption } from "@/models/global.model";

// utils
import { stringUtil } from "@/utils/string.util";

export default defineComponent({
  components: {
    RunaiTooltip,
  },
  emits: ["on-select"],
  props: {
    customStyle: {
      type: Object as PropType<Record<string, string>>,
      required: false,
      default: () => ({}),
    },
    dropdownIcon: {
      type: String as PropType<string>,
      required: false,
      default: "ph-bold ph-caret-down",
    },
    selectedValue: {
      type: String as PropType<string>,
      required: true,
    },
    options: {
      type: Array as PropType<ISelectOption[]>,
      required: true,
    },
    size: {
      type: String as PropType<string>,
      required: false,
      default: "xs",
    },
    flat: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    customClasses: {
      type: String as PropType<string>,
      required: false,
    },
    color: {
      type: String as PropType<string>,
      required: false,
      default: "black-54",
    },
    noIconAnimation: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    split: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    /**
     * This is required for the dropdown when split is true
     */
    label: {
      type: String as PropType<string>,
      required: false,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    tooltipPosition: {
      type: String as PropType<string>,
      required: false,
      default: "top",
    },
  },
  methods: {
    getAid(aid: string): string {
      return `runai-dropdown-btn-select-${stringUtil.slug(aid)}`;
    },
  },
});
</script>

<style lang="scss"></style>
