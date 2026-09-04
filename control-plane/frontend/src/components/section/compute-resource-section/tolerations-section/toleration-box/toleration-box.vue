<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="toleration-box bg-grey-3 q-my-md q-pa-md row justify-between">
    <div class="row items-center">
      <policy-string-field
        aid="toleration-key-input"
        label="Key"
        stack-label
        placeholder="Enter a key"
        :style="{ minWidth: '190px' }"
        :model-value="toleration.key"
        @update:model-value="updateKey"
        no-error-icon
        :rules="[isEmptyKey, isValidKey]"
        :disable="readOnly"
        :policy-rules="policyRules?.key || undefined"
      />

      <policy-string-field
        v-if="showValueInput"
        aid="toleration-value-input"
        stack-label
        placeholder="Enter a value"
        class="q-ml-md"
        label="Value"
        :style="{ minWidth: '190px' }"
        :model-value="toleration.value"
        @update:model-value="updateValue"
        no-error-icon
        :rules="[isEmptyValue, isValidValue]"
        :disable="readOnly"
        :policy-rules="policyRules?.value || undefined"
      />
    </div>

    <div class="row justify-evenly items-center">
      <runai-select
        aid="toleration-operator-select"
        unclearable
        label="Operator"
        :style="{ minWidth: '100px' }"
        emit-value
        :options="tolerationOptions"
        :model-value="toleration.operator"
        @update:model-value="updateOperator"
        :rules="[() => true]"
        :disable="readOnly"
        :policy-rules="operatorPolicy"
      />

      <runai-select
        aid="toleration-effect-select"
        class="q-mx-md"
        label="Effect"
        emit-value
        unclearable
        :style="{ minWidth: '145px' }"
        :model-value="toleration.effect"
        @update:model-value="updateEffect"
        :options="effectOptions"
        :rules="[() => true]"
        :disable="readOnly"
        :policy-rules="effectPolicy"
      />

      <q-btn
        aid="remove-toleration-btn"
        icon="ph-bold ph-x"
        flat
        round
        color="black-54"
        @click="removeToleration"
        :disable="readOnly"
      >
        <q-tooltip v-if="disabledTooptip">{{ disabledTooptip }}</q-tooltip>
        <q-tooltip v-else-if="!readOnly">Remove</q-tooltip>
      </q-btn>
    </div>

    <q-tooltip target=".toleration-box" v-if="readOnlyInfo">{{ readOnlyInfo }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// Models
import { TolerationOperator, type TolerationRules, type Toleration } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyStringField } from "@/components/common/policy-string-field";
// Utils
import { isEmpty, isValidRunaiEntityNameFormat } from "@/common/form.validators";
// Constants
import { errorMessages } from "@/common/error-message.constant";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import { isValidK8sKey } from "@/utils/k8s-validation.util/k8s-validation.util";

export default defineComponent({
  name: "toleration-box",
  components: { RunaiSelect, PolicyStringField },
  emits: ["remove-toleration", "update-toleration"],
  props: {
    toleration: {
      type: Object as PropType<Toleration>,
      required: true,
    },
    readOnly: {
      type: Boolean,
      default: false,
    },
    disabledTooptip: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<TolerationRules | null>,
      required: false,
    },
    allowAnyEffect: {
      type: Boolean,
      default: false,
    },
    readOnlyInfo: {
      type: String,
      required: false,
    },
  },
  computed: {
    tolerationOptions(): string[] {
      return [TolerationOperator.Exists, TolerationOperator.Equal];
    },
    effectOptions(): ISelectOption[] {
      const options = [
        {
          label: "NoExecute",
          value: "NoExecute",
          toolTip: "Pods that do not tolerate the taint are evicted immediately",
        },
        {
          label: "NoSchedule",
          value: "NoSchedule",
          toolTip:
            "No new pods will be scheduled on the tainted node unless they have a matching toleration. Pods currently running on the node are not evicted.",
        },
        {
          label: "PreferNoSchedule",
          value: "PreferNoSchedule",
          toolTip:
            "The control plane will try to avoid placing a pod that does not tolerate the taint on the node, but it is not guaranteed",
        },
      ];

      if (this.allowAnyEffect) {
        options.unshift({ label: "Any", value: "Any", toolTip: "All effects match" });
      }

      return options;
    },
    showValueInput(): boolean {
      return !!(this.toleration.operator === TolerationOperator.Equal);
    },
    effectPolicy(): IGenericSelectPolicyRules | undefined {
      return this.policyRules?.effect ? { ...this.policyRules?.effect, isClosedList: true } : undefined;
    },
    operatorPolicy(): IGenericSelectPolicyRules | undefined {
      return this.policyRules?.operator ? { ...this.policyRules?.operator, isClosedList: true } : undefined;
    },
  },
  methods: {
    removeToleration(): void {
      this.$emit("remove-toleration");
    },
    updateKey(key: string | number | null): void {
      this.$emit("update-toleration", { ...this.toleration, key });
    },
    updateValue(value: string | number | null): void {
      this.$emit("update-toleration", { ...this.toleration, value });
    },
    updateOperator(operator: string | number | null): void {
      if (operator === TolerationOperator.Exists) {
        const toleration = { ...this.toleration, value: null, operator };
        this.$emit("update-toleration", toleration);
        return;
      }
      this.$emit("update-toleration", { ...this.toleration, operator });
    },
    updateEffect(effect: string | number | null): void {
      this.$emit("update-toleration", { ...this.toleration, effect });
    },
    // Rules
    isValidKey(val: string): boolean | string {
      return isValidK8sKey(val) || errorMessages.VALID_K8S_KEY_SHORT;
    },
    isEmptyKey(val: string): boolean | string {
      return !isEmpty(val) || errorMessages.KEY_NOT_EMPTY;
    },
    isValidValue(val: string): boolean | string {
      return isValidRunaiEntityNameFormat(val) || errorMessages.VALID_FORMAT_VALUE;
    },
    isEmptyValue(val: string): boolean | string {
      return !isEmpty(val) || errorMessages.VALUE_NOT_EMPTY;
    },
  },
});
</script>
