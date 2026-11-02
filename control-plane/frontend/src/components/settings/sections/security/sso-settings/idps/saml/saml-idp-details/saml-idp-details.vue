<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="saml-idp-details col-12" aid="saml-idp-details">
    <div v-if="showExpirationDate" class="row">
      <q-tooltip v-if="tooltip" anchor="center middle">{{ tooltip }}</q-tooltip>
      <q-input
        class="col-12 q-mb-lg"
        :model-value="formattedExpirationDate"
        :error="!!expirationError"
        :error-message="expirationError"
        :rules="[certificateRule]"
        no-error-icon
        stack-label
        readonly
        label="Certificate expiration date"
      />
    </div>
    <div v-if="singleSignOnServiceUrl" class="row">
      <q-tooltip v-if="tooltip" anchor="center middle">{{ tooltip }}</q-tooltip>
      <q-input
        class="col-12 q-mb-lg"
        :model-value="singleSignOnServiceUrl"
        no-error-icon
        stack-label
        label="Identity provider URL"
        disable
      />
    </div>
    <div v-if="idpEntityId" class="row">
      <q-tooltip v-if="tooltip" anchor="center middle">{{ tooltip }}</q-tooltip>
      <q-input
        class="col-12 q-mb-lg"
        :model-value="idpEntityId"
        no-error-icon
        stack-label
        label="Identity provider Entity ID (issuer)"
        disable
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

defineOptions({ name: "SamlIdpDetails" });

const props = defineProps<{
  singleSignOnServiceUrl?: string;
  idpEntityId?: string;
  expirationDate?: Date | null;
  showExpirationDate?: boolean;
  expirationError?: string;
  tooltip?: string;
}>();

// The field is readonly (not disabled) because Quasar skips the rules of disabled fields: with an expiration error
// the rule makes the q-form that contains this field fail its validation.
const certificateRule = (): boolean | string => !props.expirationError || props.expirationError;

const formattedExpirationDate = computed((): string =>
  props.expirationDate ? dateUtil.dateFormat(props.expirationDate, DATE_FORMATS.STANDARD_DATE) : "",
);
</script>

<style lang="scss" scoped>
// A readonly field draws a dashed underline; keep it solid like the disabled fields next to it.
.saml-idp-details :deep(.q-field--standard.q-field--readonly .q-field__control:before) {
  border-bottom-style: solid;
}
</style>
