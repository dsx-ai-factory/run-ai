<template>
  <q-menu
    class="page-filter-popup"
    transition-show="scale"
    transition-hide="scale"
    transition-duration="500"
    :model-value="display"
    @hide="$emit('close')"
    :max-width="maxWidth"
    :style="menuStyle"
  >
    <q-card-section class="row justify-between items-center q-py-sm">
      <div class="text-weight-medium">{{ title }}</div>
      <q-btn
        aid="close-filter-btn"
        flat
        icon="ph-bold ph-x"
        class="q-pa-xs runai-icon-btn--default text-icon-default"
        round
        size="sm"
        aria-label="Close filter"
        v-close-popup
        @click="$emit('close')"
      ></q-btn>
    </q-card-section>
    <q-card-section class="q-py-sm">
      <slot name="body"></slot>
    </q-card-section>
    <q-card-section class="row justify-end q-py-sm">
      <q-btn
        aid="submit-filter-btn"
        type="submit"
        flat
        color="primary"
        label="Apply"
        v-close-popup
        @click="$emit('save')"
        :disable="isApplyButtonDisabled"
      />
    </q-card-section>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
export default defineComponent({
  name: "runai-page-filter-wrapper",
  emits: ["save", "close"],
  props: {
    title: {
      type: String as PropType<string>,
      required: true,
    },
    display: {
      type: Boolean as PropType<boolean>,
    },
    maxWidth: {
      type: String as PropType<string>,
      required: false,
    },
    width: {
      type: String as PropType<string>,
      required: false,
    },
    isApplyButtonDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  computed: {
    menuStyle(): { width: string } | NonNullable<unknown> {
      return this.width ? { width: this.width } : {};
    },
  },
});
</script>

<style scoped lang="scss"></style>
