<template>
  <!-- Object-based search mode (label/value options, emits value) -->
  <q-select
    v-if="hasSearchOptions"
    :aid="aid"
    outlined
    :dense="dense"
    :model-value="modelValue"
    use-input
    hide-selected
    fill-input
    input-debounce="600"
    :options="mergedSearchOptions"
    :loading="searchObjectLoading"
    :label="label"
    option-label="label"
    option-value="value"
    emit-value
    map-options
    @filter="onSearchObjectFilter"
    @update:model-value="emit('update:model-value', $event)"
  >
    <template #no-option>
      <q-item>
        <q-item-section class="text-grey">
          {{ searchObjectLoading ? "Searching..." : "No results found" }}
        </q-item-section>
      </q-item>
    </template>
    <template #option="scope">
      <q-item v-bind="scope.itemProps">
        <q-item-section>
          <q-item-label>{{ scope.opt.label }}</q-item-label>
        </q-item-section>
      </q-item>
    </template>
  </q-select>
  <!-- String-based autocomplete mode (backward compatible) -->
  <q-select
    v-else-if="hasAutocomplete"
    :aid="aid"
    outlined
    :dense="dense"
    :model-value="modelValue"
    use-input
    hide-selected
    fill-input
    input-debounce="600"
    :options="autocompleteOptions"
    :loading="autocompleteLoading"
    :label="label"
    new-value-mode="add-unique"
    @filter="onAutocompleteFilter"
    @input-value="emit('update:model-value', $event)"
  >
    <template #no-option>
      <q-item>
        <q-item-section class="text-grey">
          {{ autocompleteLoading ? "Searching..." : "No results found" }}
        </q-item-section>
      </q-item>
    </template>
    <template #option="scope">
      <q-item v-bind="scope.itemProps">
        <q-item-section>
          <q-item-label>{{ scope.opt }}</q-item-label>
        </q-item-section>
      </q-item>
    </template>
  </q-select>
  <!-- Plain input fallback -->
  <!-- eslint-disable vuejs-accessibility/no-autofocus -- the only consumer passing autofocus=true (runai-page-filter-popup) only mounts this component inside a q-menu that opens on click, never on a bare page load -->
  <q-input
    v-else
    :aid="aid"
    outlined
    :dense="dense"
    :model-value="modelValue"
    :label="label"
    :autofocus="autofocus"
    @update:model-value="emit('update:model-value', String($event ?? ''))"
  />
  <!-- eslint-enable vuejs-accessibility/no-autofocus -->
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

const props = defineProps<{
  modelValue?: string | null;
  autocompleteOptionsFn?: (searchQuery: string) => Promise<string[]>;
  searchOptionsFn?: (searchQuery: string) => Promise<Array<{ label: string; value: string }>>;
  displayValue?: string;
  label?: string;
  aid?: string;
  autofocus?: boolean;
  dense?: boolean;
}>();

const emit = defineEmits<{
  "update:model-value": [value: string];
}>();

// String-based autocomplete state
const autocompleteOptions = ref<string[]>([]);
const autocompleteLoading = ref(false);

// Object-based search state
const searchObjectOptions = ref<Array<{ label: string; value: string }>>([]);
const searchObjectLoading = ref(false);

const hasSearchOptions = computed(() => !!props.searchOptionsFn);
const hasAutocomplete = computed(() => !!props.autocompleteOptionsFn);

// Ensure the current selection is always available for map-options to resolve,
// even before the dropdown is opened and options are fetched.
const mergedSearchOptions = computed(() => {
  const options = [...searchObjectOptions.value];
  if (props.modelValue && props.displayValue) {
    const exists = options.some((o) => o.value === props.modelValue);
    if (!exists) {
      options.unshift({ label: props.displayValue, value: props.modelValue });
    }
  }
  return options;
});

function onAutocompleteFilter(searchQuery: string, update: (callback: () => void) => void): void {
  if (!props.autocompleteOptionsFn) {
    update(() => {
      autocompleteOptions.value = [];
    });
    return;
  }

  const fetchFn = props.autocompleteOptionsFn;
  update(async () => {
    autocompleteLoading.value = true;
    try {
      autocompleteOptions.value = await fetchFn(searchQuery || "");
    } catch {
      autocompleteOptions.value = [];
    } finally {
      autocompleteLoading.value = false;
    }
  });
}

function onSearchObjectFilter(searchQuery: string, update: (callback: () => void) => void): void {
  if (!props.searchOptionsFn) {
    update(() => {
      searchObjectOptions.value = [];
    });
    return;
  }

  const fetchFn = props.searchOptionsFn;
  searchObjectLoading.value = true;
  const applyResults = (results: Array<{ label: string; value: string }>) => {
    update(() => {
      searchObjectOptions.value = results;
      searchObjectLoading.value = false;
    });
  };

  Promise.resolve()
    .then(() => fetchFn(searchQuery || ""))
    .then(applyResults)
    .catch(() => {
      applyResults([]);
    });
}
</script>
