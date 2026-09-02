<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="environment-variables-section">
    <div class="q-my-md">Set the environment variable(s)</div>
    <environment-variable-box
      class="q-mb-md"
      v-for="item in displayedItems"
      :key="item.id"
      :environment-variable="item.origin"
      @update-environment-variable="updateEnvironmentVariable(item.id, $event)"
      @remove-environment-variable="removeEnvironmentVariable(item.id)"
      :source-value-options="sourceValueOptions"
      :read-only="item.isLocked || readOnly"
      :disabled-tooptip="item.isLocked ? disabledByPolicyText : undefined"
      :edit-instruction-text="editInstructionText"
      :validate-empty-env-vars="validateEmptyEnvVars"
      :support-config-map="supportConfigMap"
    />
    <q-btn
      aid="add-environment-variable-button"
      flat
      class="btn-icon-sm"
      icon="ph-bold ph-plus"
      label="environment variable"
      color="primary"
      @click="addEnvironmentVariable"
      :disable="cannotAddDueToPolicy || readOnly"
    >
      <q-tooltip v-if="cannotAddDueToPolicy">{{ disabledByPolicyText }}</q-tooltip></q-btn
    >
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { EnvironmentVariableBox } from "./environment-variable-box";
// Models
import { type EnvironmentVariableOfAsset } from "@/swagger-models/assets-service-client";
import type { TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
// Utils
import type { EnvironmentVariablesDefaults, InstancesRules } from "@/swagger-models/policy-service-client";
import { useItemizedPolicy } from "@/composables/use-itemized-policy.composable";
import { ItemizedMapperEntities } from "@/composables/helpers/use-itemized-policy.mapper";

export default defineComponent({
  name: "environment-variables-section",
  components: {
    EnvironmentVariableBox,
  },
  emits: ["update-environment-variables", "remove-environment-variable"],
  props: {
    environmentVariables: {
      type: Array as PropType<Array<EnvironmentVariableOfAsset>>,
      required: false,
      default: () => [],
    },
    sourceValueOptions: {
      type: Array as PropType<Array<TEnvVarNoneCustomSourceOption>>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    validateEmptyEnvVars: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    editInstructionText: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<InstancesRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<EnvironmentVariablesDefaults | null>,
      required: false,
    },
    supportConfigMap: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  setup(props) {
    const { disabledByPolicyText, cannotAddDueToPolicy, addItem, removeItem, updateItem, displayedItems, fullList } =
      useItemizedPolicy(
        props.environmentVariables,
        ItemizedMapperEntities.EnvironmentVariables,
        props.policyRules?.instances,
        props.policyDefaults?.instances,
      );

    return {
      disabledByPolicyText,
      cannotAddDueToPolicy,
      addItem,
      removeItem,
      updateItem,
      displayedItems,
      fullList,
    };
  },
  methods: {
    addEnvironmentVariable() {
      const newEnvVar: EnvironmentVariableOfAsset = { name: "", value: "" };
      this.addItem(newEnvVar);
      this.$emit("update-environment-variables", this.fullList);
    },
    updateEnvironmentVariable(id: string, envVar: EnvironmentVariableOfAsset) {
      this.updateItem(id, envVar);
      this.$emit("update-environment-variables", this.fullList);
    },
    removeEnvironmentVariable(id: string) {
      this.removeItem(id);
      this.$emit("update-environment-variables", this.fullList);
    },
  },
});
</script>

<style scoped lang="scss"></style>
