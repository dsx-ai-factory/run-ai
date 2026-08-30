<template>
  <section class="runai-list-modal-search row">
    <q-input
      class="col-12"
      aid="runai-list-modal-search"
      :model-value="modelValue"
      @update:model-value="$emit('update:model-value', $event)"
      :placeholder="placeholder"
      :aria-label="placeholder || 'Search'"
      :debounce="debounce"
      :input-style="{ paddingInline: '16px' }"
    >
      <template v-slot:append>
        <q-icon
          v-if="modelValue"
          aid="runai-list-modal-search-clear"
          name="ph-bold ph-x"
          size="20px"
          class="cursor-pointer"
          role="button"
          tabindex="0"
          aria-label="Clear search"
          @click="clearSearch"
          @keydown.enter.prevent="clearSearch"
          @keydown.space.prevent="clearSearch"
        />
        <q-icon name="ph-bold ph-magnifying-glass" size="20px" class="q-mr-md" />
      </template>
    </q-input>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [String, undefined] as PropType<string | undefined>,
      required: true,
    },
    placeholder: {
      type: String as PropType<string>,
      required: false,
    },
    debounce: {
      type: Number as PropType<number>,
      required: false,
    },
  },
  methods: {
    clearSearch(): void {
      this.$emit("update:model-value", "");
    },
  },
});
</script>

<style lang="scss" scoped>
.cursor-pointer {
  margin-right: 10px;
}
</style>
