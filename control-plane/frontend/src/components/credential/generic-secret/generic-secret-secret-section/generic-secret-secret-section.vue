<template>
  <runai-expansion-item class="generic-secret-secret-section" default-opened label="Secret" should-validate>
    <template #subheader>
      <section class="summary row">
        {{ summary }}
      </section>
    </template>

    <div class="q-mb-md">
      <div class="q-py-md">Set how the credential is created</div>
      <runai-radio-options
        :options="radioOptions"
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

    <generic-secret-key-value-section
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
import { GenericSecretKeyValueSection } from "./generic-secret-key-value-section";
// types
import type { ISelectOption } from "@/models/global.model";
import { errorMessages } from "@/common/error-message.constant";
import { type IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { GenericSecretCreationSpec, KeyValuePair } from "@/swagger-models/assets-service-client";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { validationRules } from "@/common/validation-rules";

export default {
  components: {
    RunaiExpansionItem,
    RunaiSelect,
    RunaiRadioOptions,
    GenericSecretKeyValueSection,
  },
  emits: ["update-secret-details"],
  props: {
    secretDetails: {
      type: Object as PropType<GenericSecretCreationSpec>,
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
  },
  data() {
    return {
      isExistingSecret: true as boolean,
      errorMessages,
      validationRules,
    };
  },
  computed: {
    radioOptions(): Array<ISelectOption> {
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
      const newSecretsAmount = this.secretDetails.keyValuePairs?.length;
      return newSecretsAmount ? `${newSecretsAmount} New secret(s)` : "None"; // TODO: Requires product texts
    },
    noSecretText(): string {
      return "No secrets found";
    },
    itemList(): IItemizedListItem[] {
      if (!this.secretDetails.keyValuePairs) return [];
      return this.secretDetails.keyValuePairs.map((s: KeyValuePair): IItemizedListItem => {
        return { name: s.key || "", value: s.value || "" };
      });
    },
  },
  methods: {
    updateFormState(isExisting: boolean): void {
      this.isExistingSecret = isExisting;
      const updatedModel = { ...this.secretDetails };
      if (isExisting) {
        updatedModel.keyValuePairs = null;
      } else {
        updatedModel.existingSecretName = null;
        updatedModel.keyValuePairs = [{ key: "", value: "" }];
      }
      this.$emit("update-secret-details", updatedModel);
    },
    updateKeyValuePairs(updatedKeyValuePairs: Array<IItemizedListItem>): void {
      if (!updatedKeyValuePairs) return;
      this.$emit("update-secret-details", {
        ...this.secretDetails,
        keyValuePairs: updatedKeyValuePairs.map((s) => {
          return { key: s.name, value: s.value };
        }),
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.generic-secret-secret-section {
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
