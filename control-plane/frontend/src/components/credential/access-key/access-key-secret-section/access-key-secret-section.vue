<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="access-key-secret-section" default-opened label="Secret" should-validate>
    <template #subheader>
      <section class="summary row">
        {{ summary }}
      </section>
    </template>
    <div class="q-mb-md">
      <div class="q-pb-sm">Set how the credential is created</div>
      <runai-radio-options
        :options="accessKeyOptions"
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

    <section class="new-secret-details" v-else>
      <div class="row q-mb-md q-gutter-lg">
        <q-input
          class="col-5"
          :model-value="secretDetails.accessKeyId"
          no-error-icon
          label="Access key"
          :rules="[validationRules.validateKeyNotEmpty]"
          aid="access-key-id-input"
          @update:model-value="$emit('update-secret-details', { ...secretDetails, accessKeyId: $event })"
        />

        <q-input
          class="col-5"
          :type="isPassHidden ? 'password' : 'text'"
          :model-value="secretDetails.secretAccessKey"
          no-error-icon
          label="Access secret"
          :rules="[validationRules.validateSecretNotEmpty]"
          aid="access-key-secret-input"
          @update:model-value="$emit('update-secret-details', { ...secretDetails, secretAccessKey: $event })"
        >
          <template v-slot:append>
            <q-icon
              size="20px"
              :name="isPassHidden ? 'ph-fill ph-eye-slash' : 'ph-fill ph-eye'"
              class="cursor-pointer text-black-54"
              @click="isPassHidden = !isPassHidden"
            />
          </template>
        </q-input>
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import type { PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiSelect } from "@/components/common/runai-select";

// types
import type { AccessKeyCreationSpec } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import { validationRules } from "@/common/validation-rules";

export default {
  components: {
    RunaiExpansionItem,
    RunaiSelect,
    RunaiRadioOptions,
  },
  emits: ["update-secret-details"],
  props: {
    secretDetails: {
      type: Object as PropType<AccessKeyCreationSpec>,
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
      isPassHidden: true as boolean,
      validationRules,
    };
  },
  computed: {
    accessKeyOptions(): Array<ISelectOption> {
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
      return this.secretDetails.accessKeyId || "None";
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
.access-key-secret-section {
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
