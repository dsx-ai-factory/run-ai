<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="form-runtime-settings">
    <command-and-arguments
      class="q-mb-md"
      :command-and-args="commandAndArgs"
      @update="updateCommandAndArgsModel"
      :command-policy-rules="policy?.rules?.command"
      :args-policy-rules="policy?.rules?.args"
      :disable="disable"
    />

    <hybrid-environment-variables-section
      v-if="!!runtimeSettings.hybridEnvironmentVariables"
      class="q-mb-md"
      :key="'hybrid-ev' + renderKey"
      :environment-variables="runtimeSettings.hybridEnvironmentVariables"
      :source-value-options="environmentVariableSourceValueOptions"
      @update-environment-variables="updateHybridEnvironmentVariables"
      @credential-created="$emit('credential-created', $event)"
      :policy-rules="policy?.rules?.environmentVariables"
      :policy-defaults="policy?.defaults?.environmentVariables"
      :support-config-map="supportConfigMapAsEnvVar"
      :support-user-credentials="supportUserCredentials"
      :validate-empty-env-vars="validateEmptyEnvVars"
      :supporting-ngc-api-key="supportingNgcApiKey"
      :read-only="disable"
    />

    <environment-variables-section
      v-else
      class="q-mb-md"
      :key="'asset-ev' + renderKey"
      :environment-variables="runtimeSettings.assetEnvironmentVariables"
      :source-value-options="environmentVariableSourceValueOptions"
      @update-environment-variables="updateAssetEnvironmentVariables"
      :policy-rules="policy?.rules?.environmentVariables"
      :policy-defaults="policy?.defaults?.environmentVariables"
      :edit-instruction-text="isAssetCreationFlow"
      :validate-empty-env-vars="!isAssetCreationFlow || validateEmptyEnvVars"
      :support-config-map="supportConfigMapAsEnvVar"
      :read-only="disable"
    />

    <runai-section title="Set the container's working directory">
      <policy-string-field
        :model-value="runtimeSettings.workingDir"
        @update:model-value="updateWorkingDir"
        no-error-icon
        label="Directory"
        placeholder="Enter a path"
        :rules="[isDirectoryPath]"
        :policy-rules="policy?.rules?.workingDir"
        :disable="disable"
      />
    </runai-section>
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Types
import { errorMessages } from "@/common/error-message.constant";
import type { IUICommandAndArgs, IRuntimeSettings, TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
import type { EnvironmentVariableOfAsset } from "@/swagger-models/assets-service-client";
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";

// Services
import { isValidDirectoryPath } from "@/common/form.validators";

// Components
import CommandAndArguments from "@/components/environment/command-and-arguments/command-and-arguments.vue";
import { EnvironmentVariablesSection } from "@/components/environment/environment-variables-section";
import { HybridEnvironmentVariablesSection } from "@/components/environment/hybrid-environment-variables-section";
import { RunaiSection } from "@/components/common/runai-section";
import { PolicyStringField } from "@/components/common/policy-string-field";

export default defineComponent({
  name: "form-runtime-settings",
  components: {
    RunaiSection,
    CommandAndArguments,
    EnvironmentVariablesSection,
    HybridEnvironmentVariablesSection,
    PolicyStringField,
  },
  emits: ["update", "credential-created"],
  props: {
    runtimeSettings: {
      type: Object as PropType<IRuntimeSettings>,
      required: false,
      default: (): IRuntimeSettings => {
        return {
          command: null as string | null,
          args: null as string | null,
          workingDir: null as string | null,
        };
      },
    },
    environmentVariableSourceValueOptions: {
      type: Array as PropType<Array<TEnvVarNoneCustomSourceOption>>,
      required: true,
    },
    policy: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaType | null>,
      required: false,
    },
    renderKey: {
      type: String,
      required: false,
      default: "",
    },
    isAssetCreationFlow: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    supportConfigMapAsEnvVar: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportUserCredentials: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    validateEmptyEnvVars: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    supportingNgcApiKey: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    commandAndArgs(): IUICommandAndArgs {
      return {
        command: this.runtimeSettings.command || "",
        args: this.runtimeSettings.args || "",
      };
    },
  },
  methods: {
    updateWorkingDir(workingDir: string | null | number): void {
      this.$emit("update", { ...this.runtimeSettings, workingDir: workingDir === "" ? null : workingDir });
    },
    updateAssetEnvironmentVariables(environmentVariables: Array<EnvironmentVariableOfAsset>): void {
      this.$emit("update", { ...this.runtimeSettings, assetEnvironmentVariables: environmentVariables });
    },
    updateHybridEnvironmentVariables(environmentVariables: Array<EnvironmentVariableOfAsset>): void {
      this.$emit("update", { ...this.runtimeSettings, hybridEnvironmentVariables: environmentVariables });
    },
    updateCommandAndArgsModel(commandAndArgs: IUICommandAndArgs): void {
      this.$emit("update", {
        ...this.runtimeSettings,
        args: commandAndArgs.args === "" ? null : commandAndArgs.args,
        command: commandAndArgs.command === "" ? null : commandAndArgs.command,
      });
    },
    isDirectoryPath(val: string): boolean | string {
      if (val === null || val === "") return true;
      return isValidDirectoryPath(val) || errorMessages.INVALID_DIRECTORY_PATH;
    },
  },
});
</script>
