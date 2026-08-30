<template>
  <q-card flat bordered square class="q-mb-md">
    <q-card-section class="card-header q-mb-md">
      <span class="text-subtitle2">{{ header }}</span>
      <q-btn
        v-if="!hideCloseButton"
        @click="$emit('close')"
        size="12px"
        aid="remove-rule-button"
        round
        flat
        color="black-54"
        icon="ph-bold ph-x"
        aria-label="Remove rule"
      ></q-btn>
    </q-card-section>
    <q-card-section class="card-subheader no-padding">
      <div class="text-body2 q-mb-md q-px-md">{{ subheader }}</div>
      <slot name="body"></slot>
    </q-card-section>
    <q-card-actions class="q-mx-sm q-pt-none">
      <slot name="actions"></slot>
    </q-card-actions>
  </q-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

export default defineComponent({
  emits: ["close"],
  props: {
    header: {
      type: String as PropType<string>,
      default: "",
    },
    subheader: {
      type: String as PropType<string>,
      default: "",
    },
    hideCloseButton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>
<style scoped lang="scss">
.card-header {
  height: 50px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid $grey-4;
}
</style>
