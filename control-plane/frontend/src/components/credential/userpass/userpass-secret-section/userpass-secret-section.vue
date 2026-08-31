<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="userpass-secret-section" default-opened label="Secret" should-validate>
    <template #subheader>
      <section class="summary row">
        {{ summary }}
      </section>
    </template>
    <div class="q-mb-md">
      <div class="q-pb-sm">Set how the credential is created</div>
      <runai-radio-options
        :options="userpassOptions"
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
          :model-value="secretDetails.user"
          no-error-icon
          label="Username"
          :rules="[validationRules.validateUsernameNotEmpty]"
          aid="userpass-username-input"
          @update:model-value="$emit('update-secret-details', { ...secretDetails, user: $event })"
        />

        <q-input
          class="col-5"
          :type="isPassHidden ? 'password' : 'text'"
          :model-value="secretDetails.password"
          no-error-icon
          label="Password"
          :rules="[validationRules.validatePasswordNotEmpty]"
          aid="userpass-password-input"
          @update:model-value="$emit('update-secret-details', { ...secretDetails, password: $event })"
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
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";

// types
import type { PasswordCreationSpec } from "@/swagger-models/assets-service-client";
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
      type: Object as PropType<PasswordCreationSpec>,
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
    userpassOptions(): Array<ISelectOption> {
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
.userpass-secret-section {
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
