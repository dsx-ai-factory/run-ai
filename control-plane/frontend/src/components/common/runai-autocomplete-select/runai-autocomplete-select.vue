<template>
  <q-select
    aid="autocomplete-select"
    class="runai-auto-complete-select"
    :popup-content-class="popupContentClass"
    :model-value="modelValue"
    :options="options"
    :use-input="!modelValue?.length"
    :input-debounce="debounce"
    input-class="placeholder-italic"
    clearable
    clear-icon="ph-bold ph-x"
    @clear="$emit('clear')"
    @update:model-value="updateModelValue"
    @filter="filterFn"
    @input="filterFn"
    :rules="rules"
    no-error-icon
    hide-hint
    :stack-label="stackLabel"
    new-value-mode="add-unique"
    @new-value="handleNewValue"
    :ref="autocompleteSelectRef"
    :loading="loading"
    :square="hasOutline"
    :outlined="hasOutline"
    :label="label"
    :placeholder="placeholder"
    :error="!!errorMessage"
    :error-message="errorMessage"
    :disable="disable"
    :hide-dropdown-icon="disable"
  >
    <template v-slot:selected-item="scope">
      <span style="white-space: nowrap" class="ellipsis">{{ scope.opt }}</span>
    </template>
    <template v-slot:option="scope">
      <slot v-if="$slots.option" name="option" v-bind="scope" />
      <q-item v-else v-bind="scope.itemProps">
        <q-item-section>
          <q-item-label>{{ scope.opt.label || scope.opt }}</q-item-label>
          <q-item-label v-if="scope.opt.subLabel" caption class="ellipsis runai-autocomplete-select__sublabel">
            {{ scope.opt.subLabel }}
          </q-item-label>
        </q-item-section>
      </q-item>
    </template>
    <template v-slot:no-option>
      <q-item>
        <q-item-section class="text-grey" aid="autocomplete-no-result">
          <span v-if="isError">No options - contact your admin</span>
          <span v-else>{{ searchQuery?.length < minimumCharacters ? searchHint : noResultLabel }}</span>
        </q-item-section>
      </q-item>
    </template>
  </q-select>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { QInput, QSelect, ValidationRule } from "quasar";
import type { PropType } from "vue";

//models
import type { ISelectOption, SelectUpdateFilterFunction } from "@/models/global.model";
export default defineComponent({
  emits: ["clear", "update:model-value"],
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    placeholder: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    noResultLabel: {
      type: String as PropType<string>,
      required: true,
    },
    searchHint: {
      type: String as PropType<string>,
      required: true,
    },
    asyncOptionsPromise: {
      type: Function as PropType<(searchQuery: string) => Promise<ISelectOption[]>>,
      required: true,
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      required: false,
    },
    modelValue: {
      type: [String, null] as PropType<string | null>,
      required: false,
    },
    debounce: {
      type: Number as PropType<number>,
      required: false,
      default: 300,
    },
    minimumCharacters: {
      type: Number as PropType<number>,
      required: false,
      default: 0,
    },
    isError: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    canAddNewValues: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    newValuesLabelPostfix: {
      type: String as PropType<string>,
      default: "(New)",
    },
    /**
     * Show/ hide the add new value option as part of the list when the search query is empty
     * this is working only when @canAddNewValues set to true
     */
    showAddNewValuesOption: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    hasOutline: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    errorMessage: {
      type: String as PropType<string>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    popupContentClass: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      selectedItem: null as ISelectOption | null,
      options: [] as ISelectOption[],
      searchQuery: "" as string,
      loading: false,
      autocompleteSelectRef: "autocompleteSelect" as string,
    };
  },
  methods: {
    updateModelValue($event: string | ISelectOption) {
      if (this.canAddNewValues && $event && typeof $event !== "string") {
        // in case of new value added by user and not the options list
        $event = $event.value as string;
      }
      this.$emit("update:model-value", $event);
    },
    handleNewValue($event: string) {
      this.$emit("update:model-value", this.canAddNewValues ? $event : null);
      (this.$refs.autocompleteSelect as QSelect).hidePopup();
    },
    filterFn(searchQuery: string, update: SelectUpdateFilterFunction): void {
      this.searchQuery = searchQuery;
      if (update && typeof update === "function") {
        update(async () => {
          if (
            this.minimumCharacters != 0 &&
            (searchQuery === "" || searchQuery == null || searchQuery?.length < this.minimumCharacters)
          ) {
            this.options = [] as ISelectOption[];
          } else {
            this.loading = true;
            this.options = await this.asyncOptionsPromise(searchQuery);
            if (this.canAddNewValues) {
              const showAddNewValuesOption = this.showAddNewValuesOption ? true : !!searchQuery;
              if (showAddNewValuesOption) {
                const newSelection: ISelectOption = {
                  label: `${searchQuery} ${this.newValuesLabelPostfix}`,
                  value: searchQuery,
                };
                this.options.push(newSelection);
              }
            }
            this.loading = false;
          }
        });
      }
    },
    async validate(): Promise<void> {
      await (this.$refs[this.autocompleteSelectRef] as QInput)?.validate();
    },
  },
});
</script>

<style lang="scss">
.runai-auto-complete-select {
  .q-icon {
    font-size: 16px;
  }

  &__sublabel {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
</style>
