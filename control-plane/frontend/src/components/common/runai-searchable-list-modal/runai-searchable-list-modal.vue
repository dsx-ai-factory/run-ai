<template>
  <section class="node-modal">
    <runai-base-dialog :model-value="true" @close="$emit('close')" size="width-sm" no-padding>
      <template #header> {{ title }} </template>

      <template #body>
        <div class="q-pb-sm">
          <runai-list-modal-search v-model="searchTerm" :placeholder="placeholder" />
          <q-separator />
        </div>
        <span class="list-title" v-if="listTitle">{{ listTitle }}</span>
        <div>
          <q-list>
            <q-item v-for="item in filteredList" :key="item" clickable v-ripple>
              <q-item-section>{{ item }}</q-item-section>
            </q-item>
          </q-list>
        </div>
      </template>

      <template #footer>
        <q-btn label="Close" color="primary" @click="$emit('close')" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";

export default defineComponent({
  components: {
    RunaiBaseDialog,
    RunaiListModalSearch,
  },
  emits: ["close"],
  props: {
    title: {
      type: String as PropType<string>,
      required: true,
    },
    placeholder: {
      type: String as PropType<string>,
      default: "Search",
    },
    itemList: {
      type: Array as PropType<Array<string>>, // Adjust the type based on your data structure
      required: true,
    },
    listTitle: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      searchTerm: "",
    };
  },
  computed: {
    filteredList(): Array<string> {
      return this.itemList.filter((item) => item.includes(this.searchTerm));
    },
  },
});
</script>
<style lang="scss" scoped>
.list-title {
  font-weight: 700;
  margin-left: 15px;
}
</style>
