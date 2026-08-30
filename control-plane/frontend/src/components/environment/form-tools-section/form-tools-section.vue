<template>
  <runai-expansion-item label="Tools" :section-invalid="!validSection" aid="form-tools-section">
    <template #subheader>
      {{ subheaderSummary }}
    </template>
    <form-tools-list :tools="tools" @update="$emit('update', $event)" :form-type="EToolBoxFormType.Environment" />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IUIConnection } from "@/models/environment.model";
// Services
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
// Cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { FormToolsList } from "@/components/environment/form-tools-section/form-tools-list";
import { EToolBoxFormType } from "@/components/environment/form-tools-section/tool-box";
export default defineComponent({
  components: {
    RunaiExpansionItem,
    FormToolsList,
  },
  emits: ["update"],
  props: {
    tools: {
      type: Array as PropType<Array<IUIConnection>>,
      required: true,
      default: () => [],
    },
  },
  data() {
    return {
      EToolBoxFormType,
    };
  },
  computed: {
    subheaderSummary(): string {
      const toolsLength: number = this.tools.length;
      return toolsLength ? `Tool(s) ${toolsLength}` : "None";
    },
    validSection(): boolean {
      return environmentService.validateUIConnections(this.tools);
    },
  },
});
</script>
