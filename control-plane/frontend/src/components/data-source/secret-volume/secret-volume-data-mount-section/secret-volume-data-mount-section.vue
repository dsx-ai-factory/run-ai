<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="secret-volume-data-mount-section" default-opened label="Data mount" should-validate>
    <template #subheader>
      {{ summary }}
    </template>
    <div>
      <div class="q-py-md">Set the data origin</div>
      <div class="credentials-section q-mb-md row">
        <runai-select
          aid="s3-credentials-select"
          class="col-8"
          :model-value="modelValue.credentialAssetId"
          @update:model-value="$emit('update:model-value', { ...modelValue, credentialAssetId: $event })"
          :loading="loadingCredentials"
          :options="credentialOptions"
          map-options
          label="Credentials"
          emit-value
          no-option-text="No credentials found"
          :rules="[isCredSelected]"
        />
      </div>

      <div class="q-py-md">{{ containerPathTitle }}</div>
      <q-input
        aid="mount-path-input"
        :model-value="modelValue.mountPath"
        no-error-icon
        label="Container path"
        stack-label
        placeholder="e.g., /home/local/data"
        input-class="placeholder-italic"
        :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
        @update:model-value="$emit('update:model-value', { ...modelValue, mountPath: $event })"
      />
    </div>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSelect } from "@/components/common/runai-select";

// Models
import type { CredentialsListResponseEntry, SecretAssetSpec } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";

// Constants
import { errorMessages } from "@/common/error-message.constant";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  name: "secret-volume-data-mount-section",
  components: {
    RunaiExpansionItem,
    RunaiSelect,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<SecretAssetSpec>,
      required: true,
    },
    credentials: {
      type: Array as PropType<CredentialsListResponseEntry[]>,
      required: true,
    },
    loadingCredentials: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      validationRules,
    };
  },
  computed: {
    selectedCredentialsName(): string {
      const selectedCred = this.credentials.find((cred) => cred.meta.id === this.modelValue.credentialAssetId);
      return selectedCred?.meta.name || "";
    },
    credentialOptions(): ISelectOption[] {
      const mappedOptions: ISelectOption[] = this.credentials.map((cred: CredentialsListResponseEntry) => {
        return { label: cred.meta.name, value: cred.meta.id };
      });
      return mappedOptions;
    },
    summary(): string {
      return `Origin: ${this.selectedCredentialsName || "None"} > Target: ${this.modelValue.mountPath || "None"}`;
    },
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle(true);
    },
  },
  methods: {
    isCredSelected(val: string): boolean | string {
      return this.isCredValueValid(val) || errorMessages.SELECT_CREDENTIAL;
    },
    isCredValueValid(val: string | null | undefined) {
      return val === null || !!val?.length;
    },
  },
});
</script>
