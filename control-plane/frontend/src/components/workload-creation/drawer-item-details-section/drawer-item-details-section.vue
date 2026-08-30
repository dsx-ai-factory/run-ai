<template>
  <overlay-drawer-section-wrapper @close="$emit('close')" :header-style="{ background: '#EEEEEE' }">
    <template #header-title>
      <span class="row items-center">
        <item-icon-with-counter :icons="displayedItem.icons || []" /><span>{{ displayedItem?.name }}</span>
      </span>
    </template>

    <template #main>
      <section class="item-details">
        <item-details :item="displayedItem" :entity="entity" />
      </section>
    </template>
  </overlay-drawer-section-wrapper>
</template>

<script lang="ts">
import type { IDrawerListItem, WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import { OverlayDrawerSectionWrapper } from "@/components/workload-creation/overlay-drawer-section-wrapper";
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { ItemIconWithCounter } from "../item-icon-with-counter";
import { ItemDetails } from "../item-details";
export default defineComponent({
  name: "drawer-item-details-section",
  components: {
    OverlayDrawerSectionWrapper,
    ItemIconWithCounter,
    ItemDetails,
  },
  emits: ["close"],
  props: {
    displayedItem: {
      type: Object as PropType<IDrawerListItem>,
      required: true,
    },
    entity: {
      type: [String, null] as PropType<WorkloadCreationDrawerEntities | null>,
      required: true,
    },
  },
});
</script>
