<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <p v-if="isAccessKey">Enter the access key name</p>
  <p v-else>Enter the service account name</p>
  <runai-name-validation
    :model-value="modelValue"
    :rules="[isEmpty, isValidFormat, isNameNotTaken]"
    debounce="300"
    auto-focus
    @update:model-value="$emit('update:model-value', $event)"
  ></runai-name-validation>
</template>
<script lang="ts">
import { errorMessages } from "@/common/error-message.constant";
import { isFirstLetterIsLowerCase, isLowerCaseAndNumbersAndUnderscore, isNotEmpty } from "@/common/form.validators";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import type { PropType } from "vue";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import { accessKeysService } from "@/services/control-plane/access-keys.service/access-keys.service";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";

export default {
  components: { RunaiNameValidation },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    isAccessKey: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      apps: [] as ServiceAccount[],
    };
  },
  methods: {
    async isNameNotTaken(name: string): Promise<boolean | string> {
      if (this.apps.length === 0) {
        try {
          if (this.isAccessKey) {
            this.apps = (await accessKeysService.getAccessKeys()) as ServiceAccount[];
          } else {
            this.apps = await serviceAccountsService.list({});
          }
        } catch (error) {
          console.warn(`Failed to fetch applications list, cannot verify if app ${name} exist`, error);
          return true;
        }
      }
      return !this.apps.find((app) => app.name === name) || errorMessages.NAME_ALREADY_EXIST;
    },
    isEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    isValidFormat(val: string): boolean | string {
      const isAppValid = isLowerCaseAndNumbersAndUnderscore(val) && isFirstLetterIsLowerCase(val);
      return isAppValid || errorMessages.VALID_APP_FORMAT;
    },
  },
};
</script>
<style lang=""></style>
