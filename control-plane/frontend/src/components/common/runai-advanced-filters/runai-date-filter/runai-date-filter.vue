<template>
  <runai-page-filter-wrapper
    :is-apply-button-disabled="!isValidDate"
    display
    :title="filter.label"
    @save="saveFilter"
    @close="$emit('close')"
    width="310px"
  >
    <template #body>
      <div class="row">
        <div class="col">
          <q-select aid="date-filter-select" outlined v-model="selectedOption" :options="options" />
        </div>
      </div>
      <q-date
        flat
        minimal
        v-model="selectedDate"
        mask="DD/MM/YYYY"
        class="filter-date-picker q-pr-sm"
        style="max-height: 240px"
      />
    </template>
  </runai-page-filter-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiPageFilterWrapper } from "@/components/common/runai-advanced-filters/runai-page-filter-wrapper";
import { type IDateFilterModel, EDateFilterOperator } from "@/models/filter.model";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

export default defineComponent({
  name: "runai-date-filter",
  components: { RunaiPageFilterWrapper },
  emits: ["save", "close"],
  props: {
    filter: {
      type: Object as PropType<IDateFilterModel>,
      required: true,
    },
  },
  data() {
    return {
      options: [EDateFilterOperator.Before, EDateFilterOperator.On, EDateFilterOperator.After] as EDateFilterOperator[],
      selectedOption: EDateFilterOperator.Before as EDateFilterOperator,
      selectedDate: "" as string | null,
    };
  },
  created() {
    this.initFilter();
  },
  computed: {
    isValidDate(): boolean {
      return !!this.selectedDate;
    },
  },
  methods: {
    initFilter(): void {
      this.selectedOption = this.filter.value;
      this.selectedDate = this.filter.date || dateUtil.dateFormat(new Date(), DATE_FORMATS.STANDARD_DATE);
    },
    saveFilter(): void {
      if (this.selectedDate === null) return;

      const filter: IDateFilterModel = {
        ...this.filter,
        value: this.selectedOption,
        date: this.selectedDate,
      };
      this.$emit("save", filter);
    },
  },
});
</script>

<style scoped lang="scss">
.date-filter-select {
  width: 170px;
}
</style>
