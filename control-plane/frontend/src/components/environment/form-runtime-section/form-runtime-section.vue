<template>
  <runai-expansion-item label="Runtime settings" :section-invalid="!sectionValid" aid="form-runtime-section">
    <template #subheader>{{ summary }} </template>

    <form-runtime-settings
      :runtime-settings="runtimeSettings"
      :environment-variable-source-value-options="environmentVariablesSourceValueOptions || []"
      is-asset-creation-flow
      :support-config-map-as-env-var="supportConfigMapAsEnvVar"
      @update="$emit('update', $event)"
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Types
import type { IRuntimeSettings, TEnvVarNoneCustomSourceOption } from "@/models/environment.model";

// Services
import { environmentService } from "@/services/control-plane/environment.service/environment.service";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { FormRuntimeSettings } from "@/components/environment/form-runtime-section/form-runtime-settings";
export default defineComponent({
  components: {
    FormRuntimeSettings,
    RunaiExpansionItem,
  },
  emits: ["update"],
  props: {
    runtimeSettings: {
      type: Object as PropType<IRuntimeSettings>,
      required: false,
    },
    environmentVariablesSourceValueOptions: {
      type: Array as PropType<Array<TEnvVarNoneCustomSourceOption>>,
      required: false,
    },
    supportConfigMapAsEnvVar: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    summary(): string {
      const envVarsLength: number = this.envVarsLengthForSummary;
      const argumentsLength: number =
        !this.runtimeSettings?.args || this.runtimeSettings?.args?.length === 0
          ? 0
          : this.runtimeSettings?.args?.split(" ")?.length;
      const hasVariables: boolean = envVarsLength > 0;
      const hasArguments: boolean = argumentsLength > 0;
      const hasCommand = !!this.runtimeSettings?.command;
      const argumentsTitle = `${argumentsLength} Argument(s)`;
      const variablesTitle = `${envVarsLength} Environment Variable(s)`;

      const subheaderParts: Array<string> = [];
      hasCommand && subheaderParts.push("Command");
      hasArguments && subheaderParts.push(argumentsTitle);
      hasVariables && subheaderParts.push(variablesTitle);

      return subheaderParts.length ? subheaderParts.join(" / ") : "None";
    },
    envVarsLengthForSummary(): number {
      return this.runtimeSettings?.assetEnvironmentVariables?.length || 0;
    },
    sectionValid(): boolean {
      return environmentService.validateRuntimeSettings(
        this.runtimeSettings?.command,
        this.runtimeSettings?.args,
        this.runtimeSettings?.assetEnvironmentVariables,
        this.runtimeSettings?.workingDir,
      );
    },
  },
});
</script>
