<template>
  <section class="topology-section">
    <div class="row items-center">
      <span aid="topology-section-title">
        Set topology (pod affinity) to let the workload be scheduled on nodes with a matching topology
      </span>
      <runai-tooltip
        aid="topology-title-tooltip"
        class="q-ml-sm"
        tooltip-text="Topology (pod affinity) lets the workload be scheduled on nodes within the same region, zone, placement group, or any other topology you define."
        tooltip-position="right"
      />
    </div>

    <div v-if="podAffinity" class="bg-grey-3 q-pa-md q-mt-md row items-center justify-between">
      <div class="col row items-center">
        <policy-string-field
          aid="topology-key-input"
          class="col-6"
          label="Key"
          stack-label
          placeholder="Enter a key"
          :disable="readOnly"
          :rules="[isEmptyKey, isValidKey]"
          :model-value="podAffinity.key"
          @update:model-value="updateTopologyKey"
          no-error-icon
          :policy-rules="policyRules?.key || undefined"
        />

        <runai-select
          aid="topology-operator-select"
          class="col-2 q-ml-md"
          label="Operator"
          :disable="operatorSelectDisabled"
          emit-value
          :options="podAffinityOperatorOptions"
          :model-value="podAffinity.type"
          @update:model-value="updateTopologyType"
          :rules="[() => true]"
          unclearable
          :policy-rules="typePolicy"
        />
      </div>

      <q-btn
        aid="remove-topology-btn"
        icon="ph-bold ph-x"
        flat
        round
        color="black-54"
        @click="removeTopology"
        :disable="readOnly"
      >
        <q-tooltip v-if="!readOnly">Remove</q-tooltip>
      </q-btn>
    </div>

    <q-btn
      v-else
      aid="add-topology-btn"
      class="q-mt-sm btn-icon-sm"
      color="primary"
      flat
      icon="ph-bold ph-plus"
      label="topology"
      @click="addTopology"
      :disable="readOnly"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Models
import { type PodAffinity, PodAffinityType } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
// Components
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiSelect } from "@/components/common/runai-select";
import { isEmpty, isValidTopologyKey } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import type { PodAffinityRules } from "@/swagger-models/policy-service-client";
import { PolicyStringField } from "@/components/common/policy-string-field";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";

export default defineComponent({
  name: "topology-section",
  components: {
    RunaiTooltip,
    RunaiSelect,
    PolicyStringField,
  },
  emits: ["update-topology"],
  props: {
    podAffinity: {
      type: [Object, null] as PropType<PodAffinity | null>,
      required: false,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PodAffinityRules | null>,
      required: false,
    },
  },
  data() {
    return {
      operatorOptions: [
        {
          value: PodAffinityType.Required,
          label: PodAffinityType.Required,
          toolTip: "If the scheduler can't schedule all pods within the same topology, the workload will be pending",
        },
        {
          value: PodAffinityType.Preferred,
          label: PodAffinityType.Preferred,
          toolTip:
            "The scheduler will try to schedule all pods within the specified topology but may schedule some pods on nodes that aren't part of the topology",
        },
      ] as ISelectOption[],
    };
  },
  computed: {
    podAffinityOperatorOptions(): ISelectOption[] {
      return this.operatorOptions;
      // return policyUtil.getOptionsWithRules(this.operatorOptions, this.policyRules?.type?.options, true);
    },
    operatorSelectDisabled(): boolean {
      return this.readOnly || this.policyRules?.type?.canEdit === false;
    },
    typePolicy(): IGenericSelectPolicyRules | undefined {
      return this.policyRules?.type ? { ...this.policyRules?.type, isClosedList: true } : undefined;
    },
  },
  methods: {
    addTopology(): void {
      const type = this.policyRules?.type?.options?.length ? null : PodAffinityType.Required;
      const podAffinity: PodAffinity = {
        type,
        key: "",
      };
      this.$emit("update-topology", podAffinity);
    },
    removeTopology(): void {
      this.$emit("update-topology", null);
    },
    updateTopologyKey(key: string | number | null): void {
      this.$emit("update-topology", { ...this.podAffinity, key });
    },
    updateTopologyType(type: PodAffinityType): void {
      this.$emit("update-topology", { ...this.podAffinity, type });
    },
    // RULES
    isEmptyKey(val: string): boolean | string {
      return !isEmpty(val) || errorMessages.KEY_NOT_EMPTY;
    },
    isValidKey(val: string): boolean | string {
      return isValidTopologyKey(val) || errorMessages.VALID_K8S_KEY;
    },
  },
});
</script>
