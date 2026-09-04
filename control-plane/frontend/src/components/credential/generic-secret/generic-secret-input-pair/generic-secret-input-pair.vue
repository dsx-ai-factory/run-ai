<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="row q-mb-md q-gutter-lg">
    <q-input
      class="col-5"
      :model-value="secretPairDetails.identifierInput.value"
      no-error-icon
      :label="secretPairDetails.identifierInput.label"
      :rules="secretPairDetails.identifierInput.rules"
      aid="generic-pair-identifier-input"
      @update:model-value="updateIdentifierInputValue"
    />

    <q-input
      class="col-5"
      :type="isPassHidden ? 'password' : 'text'"
      :model-value="secretPairDetails.secretInput.value"
      no-error-icon
      :label="secretPairDetails.secretInput.label"
      :rules="secretPairDetails.secretInput.rules"
      aid="generic-pair-secret-input"
      @update:model-value="updateSecretInputValue"
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
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// types
import { errorMessages } from "@/common/error-message.constant";
import { validationRules } from "@/common/validation-rules";
import type { ValidationRule } from "quasar";

export type GenericSecretInputPairProps = {
  identifierInput: {
    value: string;
    label: string;
    rules: Array<ValidationRule>;
  };
  secretInput: {
    value: string;
    label: string;
    rules: Array<ValidationRule>;
  };
};

export default defineComponent({
  name: "generic-secret-input-pair",
  emits: ["update"],
  props: {
    secretPairDetails: {
      type: Object as PropType<GenericSecretInputPairProps>,
      required: true,
    },
  },
  data() {
    return {
      errorMessages,
      validationRules,

      isPassHidden: true as boolean,
    };
  },
  methods: {
    updateIdentifierInputValue(value: string | null | number): void {
      this.$emit("update", { identifier: value || "", secret: this.secretPairDetails.secretInput.value });
    },
    updateSecretInputValue(value: string | null | number): void {
      this.$emit("update", { identifier: this.secretPairDetails.identifierInput.value, secret: value || "" });
    },
  },
});
</script>
