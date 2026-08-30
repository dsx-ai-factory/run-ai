<template>
  <runai-expansion-item :default-opened="false" :disable-closing="false">
    <template #header>
      <div>
        {{ sectionOptions.mainTitle }}
        <q-toggle
          :disable="sectionDisabled"
          :model-value="isEnabled"
          @update:model-value="$emit('toggle-editing-changed', $event)"
          aid="edit-master-toggle"
        >
          <q-tooltip v-if="sectionDisabled && sectionOptions.disabledTooltip">
            {{ sectionOptions.disabledTooltip }}
          </q-tooltip>
        </q-toggle>
        <runai-tooltip
          v-if="sectionOptions.tooltip"
          tooltip-position="right"
          :tooltip-text="sectionOptions.tooltip"
          class="q-ml-md"
          width="460px"
        />
      </div>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import type { IWorkloadToggleEditingFormSectionOptions } from "./toggle-editing-form.model";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiTooltip,
  },
  emits: ["toggle-editing-changed"],
  props: {
    isEnabled: {
      type: Boolean as PropType<boolean>,
      required: true,
      default: false,
    },
    sectionOptions: {
      type: Object as PropType<IWorkloadToggleEditingFormSectionOptions>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
});
</script>
