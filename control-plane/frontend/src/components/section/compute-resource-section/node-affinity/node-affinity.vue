<template>
  <div class="node-affinity-section column q-mt-lg">
    <div class="q-mb-md">Set node affinity to schedule the workload on a specific node type</div>
    <div class="row">
      <runai-select
        v-if="inputType === 'select'"
        aid="node-affinity-select"
        class="col-5 node-affinity-select"
        :model-value="nodeType"
        @update:model-value="selectNodeAffinity"
        no-error-icon
        emit-value
        :options="selectedNodesAffinityOptions"
        label="Node type (Affinity)"
        :rules="[isSelected]"
        :disable="disabled"
        :policy-rules="policyRules"
      />

      <policy-string-field
        v-if="inputType === 'input'"
        aid="node-affinity-input"
        :model-value="nodeType"
        @update:model-value="updateNodeAffinity"
        class="col-5"
        :debounce="300"
        label="Node type (Affinity)"
        :disable="disabled"
        :policy-rules="policyRules"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiSelect } from "@/components/common/runai-select";

// constant
import { errorMessages } from "@/common/error-message.constant";
import type { StringRules } from "@/swagger-models/policy-service-client";

export default defineComponent({
  name: "node-affinity",
  components: { PolicyStringField, RunaiSelect },
  emits: ["node-type-changed"],
  props: {
    nodesAffinity: {
      type: Array as PropType<string[]>,
      required: false,
    },
    inputType: {
      type: String as PropType<"select" | "input">,
      required: true,
    },
    nodeType: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    policyRules: {
      type: Object as PropType<StringRules>,
      required: false,
    },
  },
  computed: {
    selectedNodesAffinityOptions(): string[] {
      return this.nodesAffinity || [];
    },
  },
  methods: {
    selectNodeAffinity(selectedNodeAffinity: string | null): void {
      this.updateNodeAffinity(selectedNodeAffinity || null);
    },
    updateNodeAffinity(nodeAffinity: string | number | null): void {
      this.$emit("node-type-changed", nodeAffinity);
    },
    isSelected(selectedOption: string | null): boolean | string {
      return !!selectedOption || errorMessages.SELECT_NODE_TYPE;
    },
  },
});
</script>
<style lang="scss" scoped></style>
