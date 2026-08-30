<template>
  <q-tr class="text-left" :props="props">
    <q-th class="q-table--col-auto-width" v-if="isMultipleSelection">
      <q-checkbox
        size="md"
        :model-value="isSelected"
        @click="$emit('checkbox-clicked', $event)"
        aria-label="Select all rows"
      />
    </q-th>
    <q-th v-else-if="selection !== 'none'" class="q-table--col-auto-width checkbox-col"></q-th>

    <q-th v-for="col in props.cols" :key="col.name" :props="props">
      <component
        v-if="col.customHeader"
        :is="col.customHeader"
        :col="col"
        :filter-by="filterBy"
        @custom-header-event="
          col.customHeaderEventName &&
          $emit('custom-header-event', { emitName: col.customHeaderEventName, value: $event })
        "
      />
      <span v-else>{{ col.label }}</span>
    </q-th>
  </q-tr>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RadioButtonsTableHeader } from "@/components/common/runai-table/runai-custom-header-columns";
import type { IFilterBy } from "@/models/filter.model";
//model

export default defineComponent({
  name: "runai-table-header",
  components: {
    RadioButtonsTableHeader,
  },
  emits: ["checkbox-clicked", "custom-header-event"],
  props: {
    props: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      type: Object as PropType<any>,
      required: false,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
    selection: {
      type: String as PropType<"none" | "single" | "multiple">,
      default: "single",
    },
  },
  computed: {
    isMultipleSelection(): boolean {
      return this.selection === "multiple";
    },
  },
});
</script>

<style scoped lang="scss">
.checkbox-col {
  min-height: 55px;
}
</style>
