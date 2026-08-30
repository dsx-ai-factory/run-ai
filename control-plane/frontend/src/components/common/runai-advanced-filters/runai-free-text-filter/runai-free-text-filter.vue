<template>
  <runai-page-filter-wrapper display :title="filter.label" @save="saveFilter" @close="$emit('close')" max-width="250px">
    <template #body>
      <div class="row">
        <div class="col">
          <q-select
            aid="free-text-filter-select"
            outlined
            v-model="selectedOption"
            :options="options"
            :disable="isOperatorSelectDisabled"
            class="q-mb-lg"
          />
          <p class="text-black-54 filter-input-label">The following value:</p>
          <runai-autocomplete-filter-input
            v-model="term"
            aid="free-text-filter-input"
            :label="filter.label"
            :autocomplete-options-fn="filter.autocompleteOptionsFn"
          />
        </div>
      </div>
    </template>
  </runai-page-filter-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RunaiPageFilterWrapper } from "@/components/common/runai-advanced-filters/runai-page-filter-wrapper";
import { RunaiAutocompleteFilterInput } from "@/components/common/runai-autocomplete-filter-input";
//models
import {
  EFilterOperator,
  EFilterOperatorLabel,
  freeTextFilterSelectOptions,
  type IFreeTextFilterModel,
  type IFreeTextFilterSelectOption,
} from "@/models/filter.model";
import type { ISelectOption } from "@/models/global.model";

const DEFAULT_SELECTED_OPTION: IFreeTextFilterSelectOption = {
  label: EFilterOperatorLabel.Contains,
  value: EFilterOperator.Contains,
};
export default defineComponent({
  name: "runai-free-text-filter",
  components: { RunaiPageFilterWrapper, RunaiAutocompleteFilterInput },
  emits: ["save", "close"],
  props: {
    filter: {
      type: Object as PropType<IFreeTextFilterModel>,
      required: true,
    },
  },
  data() {
    return {
      term: "" as string,
      selectedOption: DEFAULT_SELECTED_OPTION as IFreeTextFilterSelectOption,
    };
  },
  created() {
    this.initFilter();
  },
  computed: {
    options(): IFreeTextFilterSelectOption[] {
      if (!this.filter.excludeOperators) {
        return freeTextFilterSelectOptions;
      }
      return freeTextFilterSelectOptions.filter((option: IFreeTextFilterSelectOption) => {
        return !this.filter?.excludeOperators?.includes(option.value);
      });
    },
    isOperatorSelectDisabled(): boolean {
      return this.options.length <= 1;
    },
  },
  methods: {
    initFilter(): void {
      this.term = this.filter.term;
      if (this.options.length === 1) {
        this.selectedOption = this.options[0];
        return;
      }
      const selectedOption: IFreeTextFilterSelectOption | undefined = this.options.find(
        (option: ISelectOption) => option.value === this.filter.value,
      );
      if (selectedOption) {
        this.selectedOption = selectedOption;
      } else if (this.filter.excludeOperators?.includes(DEFAULT_SELECTED_OPTION.value)) {
        this.selectedOption = this.options[0];
      }
    },
    saveFilter(): void {
      if (!this.term) {
        this.$emit("close");
        return;
      }
      const filter: IFreeTextFilterModel = {
        ...this.filter,
        term: this.term,
        value: this.selectedOption.value,
      };
      this.$emit("save", filter);
    },
  },
});
</script>

<style scoped lang="scss">
.filter-input-label {
  margin-bottom: 5px;
}
</style>
