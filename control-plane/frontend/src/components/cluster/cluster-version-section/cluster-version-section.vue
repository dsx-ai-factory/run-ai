<template>
  <runai-expansion-item
    class="runai-version"
    label="Run:ai version"
    :default-opened="true"
    :section-invalid="sectionInvalid"
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>
    <section class="row">
      <div class="col-4">
        <runai-select
          class="cluster-version-select"
          :model-value="selectedClusterVersionText"
          :options="allowedVersions"
          @update:model-value="$emit('update:model-value', $event)"
          placeholder="Select a version"
          :rules="[isSelected]"
        />
      </div>
    </section>
  </runai-expansion-item>
</template>
<script lang="ts">
// enums
import { errorMessages } from "@/common/error-message.constant";
// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSelect } from "@/components/common/runai-select";
// helpers
import { isEmpty, isNotEmpty } from "@/common/form.validators";
// types
import type { PropType } from "vue";
export default {
  components: {
    RunaiExpansionItem,
    RunaiSelect,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: String as PropType<string | undefined>,
      required: true,
    },
    allowedVersions: {
      type: Array as PropType<string[]>,
      required: true,
    },
  },
  computed: {
    selectedClusterVersionText(): string {
      return this.modelValue || "Select a version";
    },
    sectionInvalid(): boolean {
      return isEmpty(this.modelValue);
    },
    summary(): string {
      return this.modelValue || "None";
    },
  },
  methods: {
    isSelected(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.SELECT_VERSION;
    },
  },
};
</script>
<style lang="scss"></style>
