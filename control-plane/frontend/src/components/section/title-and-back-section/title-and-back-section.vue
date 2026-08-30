<template>
  <runai-expansion-item
    :label="label"
    :default-opened="false"
    :hide-expend-icon="false"
    :disable-opening="true"
    :section-invalid="false"
    :aid="aid"
  >
    <template #subheader>
      <div class="row items-center justify-between q-pr-sm">
        <span>{{ summary }}</span>
        <runai-btn
          v-if="backTooltip"
          icon="ph-bold ph-pencil-simple"
          flat
          round
          size="sm"
          class="runai-icon-btn--default"
          @click="onBack"
          aid="back-btn"
        >
          <q-tooltip anchor="center right" self="center left">{{ backTooltip }}</q-tooltip>
        </runai-btn>
      </div>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiBtn,
  },
  emits: ["on-back"],
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
    summary: {
      type: String as PropType<string>,
      required: true,
    },
    backTooltip: {
      type: String as PropType<string>,
      required: false,
    },
  },
  computed: {
    aid(): string {
      return `title-${this.label.replace(" ", "-").toLowerCase()}-and-back-section`;
    },
  },
  methods: {
    onBack(): void {
      this.$emit("on-back");
    },
  },
});
</script>

<style scoped lang="scss"></style>
