<template>
  <runai-expansion-item class="docker-registry-secret-section" default-opened label="Secret" should-validate>
    <template #subheader>
      <section class="summary row">
        {{ summary }}
      </section>
    </template>
    <div class="q-mb-md">
      <div class="q-pb-sm">Set how the credential is created</div>
      <runai-radio-options
        :options="secretOptions"
        :model-value="isExistingSecret"
        @update:model-value="updateFormState"
        horizontal
      />
    </div>

    <div class="row q-mb-md" v-if="isExistingSecret">
      <runai-select
        class="secret-select col-8"
        standard
        :model-value="secretDetails.existingSecretName || null"
        :options="existingSecretNames || []"
        @update:model-value="$emit('update-secret-details', { ...secretDetails, existingSecretName: $event })"
        :rules="[validationRules.validateValueSelected]"
        aid="existing-secret-select"
        :loading="loading"
        :no-option-text="noSecretText"
        placeholder="Select a secret..."
      />
    </div>

    <docker-registry-new-secret-section
      v-else
      :secret-details="secretDetails"
      @update-secret-details="$emit('update-secret-details', $event)"
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import type { PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";

// types
import type { DockerRegistryCreationSpec } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import { validationRules } from "@/common/validation-rules";
import { DockerRegistryNewSecretSection } from "./docker-registry-new-secret-section";

export default {
  name: "docker-registry-secret-section",
  components: {
    RunaiExpansionItem,
    RunaiSelect,
    RunaiRadioOptions,
    DockerRegistryNewSecretSection,
  },
  emits: ["update-secret-details"],
  props: {
    secretDetails: {
      type: Object as PropType<DockerRegistryCreationSpec>,
      required: true,
    },
    existingSecretNames: {
      type: [Array, null] as PropType<Array<string> | null>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    projectId: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: true,
    },
  },
  data() {
    return {
      isExistingSecret: true as boolean,
      validationRules,
    };
  },
  computed: {
    secretOptions(): Array<ISelectOption> {
      return [
        {
          label: "Existing secret",
          value: true,
        },
        {
          label: "New secret",
          value: false,
        },
      ];
    },
    summary(): string {
      if (this.isExistingSecret) return this.secretDetails.existingSecretName || "None";
      return this.secretDetails.user || "None";
    },
    noSecretText(): string {
      return this.projectId ? "No secrets found in the selected project" : "No secrets found";
    },
  },
  methods: {
    updateFormState(isExisting: boolean): void {
      this.isExistingSecret = isExisting;
      if (!isExisting) this.$emit("update-secret-details", { ...this.secretDetails, existingSecretName: null });
    },
  },
};
</script>

<style lang="scss" scoped>
.docker-registry-secret-section {
  .type-options {
    margin-left: -18px;
    gap: 40px;
  }
  .cursor-pointer {
    margin-right: 10px;
    margin-top: 10px;
  }
}
</style>
