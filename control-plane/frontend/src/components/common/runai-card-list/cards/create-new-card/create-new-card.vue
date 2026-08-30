<template>
  <runai-select-card :value="isSelected" @clicked="$emit('card-clicked')" aid="create-new-card">
    <q-card class="create-new-card workload-card-content text-body1 text-weight-medium">
      <q-card-section class="create-new-card-content no-padding text-body1 text-weight-medium">
        Start from scratch
      </q-card-section>
    </q-card>
  </runai-select-card>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSelectCard } from "@/components/common/runai-select-card";
import type { ICreateNewCard } from "@/components/common/runai-card-list/runai-card-list.model";

export default defineComponent({
  components: {
    RunaiSelectCard,
  },
  emits: ["card-clicked"],
  props: {
    data: {
      type: Object as PropType<ICreateNewCard>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.create-new-card {
  @include runaiSelectCardSize;

  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
