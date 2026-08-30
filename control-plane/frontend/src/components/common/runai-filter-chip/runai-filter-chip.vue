<template>
  <q-chip
    clickable
    size="md"
    :removable="removable"
    :aid="chipAid"
    class="runai-filter-chip"
    :icon-right="iconRight"
    role="button"
  >
    <div class="row items-center no-wrap">
      <q-icon style="margin: unset" tag="div" v-if="iconName" :name="iconName" size="16px" color="icon-default" />
      <span v-if="title" class="filter-chip-title"> {{ title }}:</span>
      <span :class="chipValueClass"> {{ value }}</span>
      <q-tooltip v-if="isTextOverflow(value)">{{ value }}</q-tooltip>
    </div>
    <q-menu
      class="page-filter-popup"
      transition-show="scale"
      transition-hide="scale"
      transition-duration="500"
      :model-value="display"
      @before-show="$emit('before-show')"
      @hide="$emit('close')"
    >
      <form @submit.prevent>
        <q-card-section class="row justify-between items-center q-py-sm">
          <div v-if="title || menuTitle" class="text-weight-medium">{{ title || menuTitle }}</div>
          <q-btn
            flat
            icon="ph-bold ph-x"
            class="q-pa-xs runai-icon-btn--default text-icon-default"
            round
            v-close-popup
            @click="$emit('close')"
            aid="close-btn"
            aria-label="Close filter"
          ></q-btn>
        </q-card-section>
        <q-card-section class="q-py-sm">
          <slot name="input-component"></slot>
        </q-card-section>
        <q-card-section class="row justify-end q-py-sm">
          <q-btn
            class="apply-chip-filter-btn"
            aid="apply-btn"
            type="button"
            flat
            color="primary"
            :icon-right="saveBtnIconRight"
            :label="saveBtnLabel"
            v-close-popup
            @click="$emit('save')"
          />
        </q-card-section>
      </form>
    </q-menu>
  </q-chip>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { stringUtil } from "@/utils/string.util";
import { MAX_STRING_LENGTH_BEFORE_TRUNCATION } from "@/models/global.model";

export default defineComponent({
  emits: ["save", "close", "before-show"],
  props: {
    title: {
      type: String as PropType<string>,
      required: false,
    },
    menuTitle: {
      type: String as PropType<string>,
      required: false,
    },
    value: {
      type: String as PropType<string>,
      required: true,
    },
    display: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    removable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    iconName: {
      type: String as PropType<string>,
      required: false,
    },
    chipValueTextWeight: {
      type: String as PropType<string>,
      default: "medium",
    },
    iconRight: {
      type: String as PropType<string>,
      required: false,
    },
    saveBtnLabel: {
      type: String as PropType<string>,
      default: "Apply",
    },
    saveBtnIconRight: {
      type: String as PropType<string>,
      required: false,
    },
  },
  computed: {
    chipAid(): string {
      return `${stringUtil.slug(this.title || this.menuTitle || "")}-filter-chip`;
    },
    chipValueClass(): string[] {
      return ["q-px-xs", "filter-chip-value", "ellipsis-chip", `text-weight-${this.chipValueTextWeight}`];
    },
  },
  methods: {
    isTextOverflow(text: string): boolean {
      return text.length > MAX_STRING_LENGTH_BEFORE_TRUNCATION;
    },
  },
});
</script>
<style lang="scss">
.runai-filter-chip {
  .q-chip__icon {
    font-size: 1.2em;
    margin-right: 0;
    color: var(--icon-default, #707070);
  }
}
.apply-chip-filter-btn {
  .q-icon {
    font-size: var(--icon-size-default) !important;
  }
  .on-right {
    margin-left: 5px;
  }
}
</style>
<style lang="scss" scoped>
.page-filter-popup {
  width: 250px;
}
.ellipsis-chip {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}
</style>
