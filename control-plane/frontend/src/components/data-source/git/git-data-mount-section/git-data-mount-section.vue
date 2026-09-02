<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="git-data-mount" default-opened label="Data snapshot" should-validate>
    <template #subheader>
      <section class="summary row">
        Origin:
        <div class="ellipsis-l">&lrm;{{ repositorySummary }}</div>
        <span> > </span> Target:
        <div class="ellipsis-l">&lrm;{{ containerPathSummary }}</div>
      </section>
    </template>
    <section>
      <div class="q-mb-md">
        <div class="q-py-md">Set the repository type</div>
        <runai-radio-options
          class="q-mb-md"
          :model-value="isPrivate"
          :options="[
            { label: 'Public', value: false },
            { label: 'Private', value: true },
          ]"
          @update:model-value="onSelectRepositoryType"
          horizontal
        />
        <div class="credentials-section row q-mb-md q-gutter-lg" v-if="isPrivate">
          <runai-select
            aid="credentials-select"
            class="col-5"
            :model-value="selectedCredentialId"
            @update:model-value="onSelectCredential"
            :loading="loadingCredentials"
            :options="credentialOptions"
            no-option-text="No credentials found"
            label="Credentials"
            emit-value
            :rules="[isCredSelected]"
            map-options
          />
        </div>
        <runai-section title="Set the data origin">
          <q-input
            aid="repository-input"
            :model-value="modelValue.repository"
            no-error-icon
            label="Repository URL"
            reactive-rules
            stack-label
            :placeholder="repositoryPlaceholder"
            input-class="placeholder-italic"
            :rules="repositoryRules"
            @update:model-value="$emit('update:model-value', { ...modelValue, repository: $event })"
          />
        </runai-section>
      </div>

      <div class="row q-mb-md q-gutter-lg justify-between">
        <q-input
          class="col-5"
          aid="git-revision-input"
          :model-value="modelValue.revision"
          no-error-icon
          :label="revisionLabel"
          stack-label
          placeholder="e.g., dae86e1950b1277e545cee1805517500..."
          input-class="placeholder-italic"
          hint="If left empty, will use the 'HEAD' (latest)"
          @update:model-value="$emit('update:model-value', { ...modelValue, revision: $event === '' ? null : $event })"
        />
      </div>

      <div>
        <div class="q-py-md">{{ containerPathTitle }}</div>
        <q-input
          aid="git-container-input"
          :model-value="modelValue.path"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/localdata"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, path: $event })"
        />
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiSection } from "@/components/common/runai-section";
// Types
import { errorMessages } from "@/common/error-message.constant";
import { AssetKind, type GitAssetSpec, type Password } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";
import type { ValidationRule } from "quasar";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSelect,
    RunaiSection,
    RunaiRadioOptions,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<GitAssetSpec>,
      required: true,
    },
    credentials: {
      type: Array as PropType<Password[]>,
      required: true,
    },
    loadingCredentials: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      errorMessages,
      validationRules,
      isPrivate: false,
    };
  },
  created() {
    this.initIsPrivate();
  },
  computed: {
    revisionLabel(): string {
      return "Revision (branch, tag or hash)";
    },
    selectedCredentialId(): string | null {
      if (!this.modelValue.credentialId && !this.modelValue.passwordAssetId) return "";
      return this.modelValue.credentialId || this.modelValue.passwordAssetId || null;
    },
    repositorySummary(): string {
      return this.modelValue.repository || "None";
    },
    containerPathSummary(): string {
      return this.modelValue.path || "None";
    },
    credentialOptions(): ISelectOption[] {
      return this.credentials.map((cred: Password) => {
        return { label: cred.meta.name, value: cred.meta.id };
      });
    },
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
    repositoryPlaceholder(): string {
      return this.isPrivate
        ? "e.g., git@github.com:kubernetes/kubernetes.git"
        : "e.g., https://github.com/kubernetes/kubernetes.git";
    },
    repositoryRules(): ValidationRule[] {
      return this.isPrivate
        ? [validationRules.validateUrlNotEmpty, validationRules.validatePrivateUrlFormat]
        : [validationRules.validateUrlNotEmpty, validationRules.validateUrlFormat];
    },
  },
  methods: {
    isCredSelected(val: string): boolean | string {
      return this.isCredValueValid(val) || errorMessages.SELECT_CREDENTIAL;
    },
    isCredValueValid(val: string | null | undefined) {
      return val === null || !!val?.length;
    },
    onSelectCredential(val: string): void {
      const selectedCred = this.credentials.find((c) => c.meta.id === val);

      const isPassword = selectedCred?.meta.kind === AssetKind.Password;
      const isSecret = selectedCred?.meta.kind === AssetKind.GenericSecret;

      if (!isSecret && !isPassword) {
        return;
      }

      const updatedSpec: GitAssetSpec = {
        ...this.modelValue,
        passwordAssetId: undefined,
        credentialId: val,
      };

      this.$emit("update:model-value", updatedSpec);
    },
    onSelectRepositoryType(val: boolean): void {
      this.isPrivate = val;

      const updatedSpec: GitAssetSpec = {
        ...this.modelValue,
        passwordAssetId: undefined,
        credentialId: undefined,
      };
      this.$emit("update:model-value", updatedSpec);
    },
    initIsPrivate(): void {
      // set it to private if there is a credential or password asset id
      this.isPrivate = !!this.modelValue.credentialId || !!this.modelValue.passwordAssetId;
    },
  },
});
</script>
<style>
.q-field--labeled .q-field__native {
  padding-bottom: 500;
}

.q-field__native.q-placeholder.placeholder-italic {
  padding-bottom: 0;
}
.summary {
  white-space: pre;
}
.ellipsis-l {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100px;
  direction: rtl;
  text-align: left;
}
</style>
