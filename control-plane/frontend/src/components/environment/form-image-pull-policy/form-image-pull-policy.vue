<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="form-image-pull-policy">
    <runai-section title="Set the condition for pulling the image">
      <runai-radio-options
        :options="imageRadioOptions"
        :model-value="imagePullPolicy || null"
        @update:model-value="$emit('update', $event)"
        :policy-rules="radioOptionsRules"
        :disable="disable"
      />
    </runai-section>

    <runai-section
      v-if="supportImagePullSecrets && shouldShowImagePullSecrets"
      class="q-mt-lg"
      title="Set the credentials for pulling the image"
    >
      <runai-credential-selector-by-source
        class="col-8"
        :model-value="credentialSelectorModel"
        :cluster-id="clusterId"
        @update:model-value="onCredentialSelectorChanged"
        :source-select-config="{ disable: disable || disabledByPolicy, showPersonalCredentialsOption: true }"
        :type-select-config="{
          hidden: false,
          allowedTypes: [ERunaiCredentialType.DockerRegistry, ERunaiCredentialType.NgcApiKey],
          disabledTypes: !supportNgcApiKey ? { [ERunaiCredentialType.NgcApiKey]: {} } : undefined,
        }"
        :credential-select-config="{ required: false, showCreateNewBtn: !disable && !disabledByPolicy }"
        :disable-all="disable || disabledByPolicy"
        :name-resolver="credentialNameResolver"
      />
    </runai-section>
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Cmp
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
// Types
import type { ISelectOption } from "@/models/global.model";
import { IMAGE_PULL_POLICY_LABELS } from "@/models/environment.model";
import type { ImagePullSecret } from "@/swagger-models/assets-service-client";
import { ImagePullPolicy } from "@/swagger-models/assets-service-client";
import type { ImagePullPolicyRules, ImagePullSecretsRules } from "@/swagger-models/policy-service-client";
import type { IGenericRadioOptionsPolicyRules } from "@/models/policy.model";
import {
  ERunaiCredentialSource,
  ERunaiCredentialType,
  type ICredentialSelectorModel,
  createResourceNameResolver,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

const credentialNameResolver = createResourceNameResolver();

export default defineComponent({
  name: "form-image-pool-policy",
  components: {
    RunaiRadioOptions,
    RunaiSection,
    RunaiCredentialSelectorBySource,
  },
  emits: ["update", "update-image-pull-secrets", "credential-created"],
  props: {
    imagePullPolicy: {
      type: [String, null] as PropType<ImagePullPolicy | null>,
      required: false,
    },
    supportImagePullSecrets: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    imagePullSecrets: {
      type: [Array, null] as PropType<ImagePullSecret[] | null>,
      required: false,
      default: () => [],
    },
    imagePullSecretsRules: {
      type: [Object, null] as PropType<ImagePullSecretsRules | null>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<ImagePullPolicyRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    supportNgcApiKey: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      imageRadioOptions: [
        {
          label: `${IMAGE_PULL_POLICY_LABELS.IfNotPresent} (recommended)`,
          value: ImagePullPolicy.IfNotPresent,
        },
        { label: IMAGE_PULL_POLICY_LABELS.Always, value: ImagePullPolicy.Always },
        {
          label: `${IMAGE_PULL_POLICY_LABELS.Never} (if the image is already present on the host, it will be used)`,
          value: ImagePullPolicy.Never,
        },
      ] as ISelectOption[],
      ERunaiCredentialType,
      credentialNameResolver,
      credentialSelectorType: this.supportNgcApiKey
        ? ERunaiCredentialType.NgcApiKey
        : ERunaiCredentialType.DockerRegistry,
      credentialSelectorSource: ERunaiCredentialSource.MyCredentials,
    };
  },
  computed: {
    radioOptionsRules(): IGenericRadioOptionsPolicyRules {
      return this.policyRules as IGenericRadioOptionsPolicyRules;
    },
    shouldShowImagePullSecrets(): boolean {
      return this.imagePullPolicy !== ImagePullPolicy.Never;
    },

    credentialSelectorModel(): ICredentialSelectorModel {
      const imagePullSecret = this.imagePullSecrets?.[0];
      const imagePullSecretName = imagePullSecret?.name || "";

      let secretSource = this.credentialSelectorSource;
      if (imagePullSecret) {
        secretSource = imagePullSecret.userCredential
          ? ERunaiCredentialSource.MyCredentials
          : ERunaiCredentialSource.SharedSecret;
      }

      return {
        source: secretSource,
        type: this.credentialSelectorType,
        credentialName: imagePullSecretName,
      };
    },
    disabledByPolicy(): boolean {
      return (
        this.imagePullSecretsRules?.attributes?.name?.canEdit === false ||
        this.imagePullSecretsRules?.instances?.canAdd === false
      );
    },
  },
  methods: {
    onCredentialSelectorChanged(value: ICredentialSelectorModel): void {
      this.credentialSelectorType = value.type || this.credentialSelectorType;
      this.credentialSelectorSource = value.source || ERunaiCredentialSource.MyCredentials;
      const imagePullSecrets: ImagePullSecret[] = value.credentialName
        ? [
            {
              name: value.credentialName,
              userCredential: value.source === ERunaiCredentialSource.MyCredentials,
            },
          ]
        : [];
      this.$emit("update-image-pull-secrets", imagePullSecrets);

      // If a credential was just created, emit the credential-created event
      if (value.credentialName && !this.imagePullSecrets?.[0]?.name) {
        this.$emit("credential-created");
      }
    },
  },
});
</script>
