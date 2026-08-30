<template>
  <runai-base-popover @hide="onHide" :no-padding="true" size="width-sm" ref="popover" max-height="50vh">
    <template #header> <span class=""> Access </span> </template>
    <template #body>
      <runai-list-modal-search v-model="searchTerm" :placeholder="placeholderForSearch" />
      <q-list separator>
        <q-item>
          <q-item-section class="text-b-body1-md">{{ listTitle }}</q-item-section>
        </q-item>
        <template v-if="displayList.length > 0">
          <q-item clickable v-for="item in displayList" :key="item">
            <q-item-section class="text-b-body1">{{ item }}</q-item-section>
          </q-item>
        </template>
        <q-item v-else>
          <q-item-section class="text-center text-b-subtitle2">{{ noResultsMessage }}</q-item-section>
        </q-item>
      </q-list>
    </template>
    <template #footer>
      <section>
        <q-btn aid="close-tool-access-list-modal-btn" label="Close" color="primary" @click="onClose" />
      </section>
    </template>
  </runai-base-popover>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiListModalSearch } from "@/components/common/runai-list-modal-search";
import { RunaiBasePopover } from "@/components/common/runai-base-popover";

export type RunaiBasePopoverType = InstanceType<typeof RunaiBasePopover>;

export default defineComponent({
  name: "tool-access-list-popover",
  components: { RunaiListModalSearch, RunaiBasePopover },
  props: {
    searchPlaceholder: {
      type: String as PropType<string>,
      required: true,
    },
    listTitle: {
      type: String as PropType<string>,
      required: true,
    },
    list: {
      type: Array as PropType<string[]>,
      required: true,
    },
  },
  data() {
    return {
      searchTerm: "" as string,
    };
  },
  computed: {
    displayList(): string[] {
      return this.list.filter((item) => item.toLowerCase().includes(this.searchTerm.toLowerCase()));
    },
    placeholderForSearch(): string {
      return `Search ${this.searchPlaceholder}`;
    },
    noResultsMessage(): string {
      return `No ${this.searchPlaceholder} match your filters`;
    },
  },
  methods: {
    onHide(): void {
      this.searchTerm = "";
    },
    onClose(): void {
      this.searchTerm = "";
      const popover = this.$refs.popover as RunaiBasePopoverType;
      popover.hide();
    },
  },
});
</script>
