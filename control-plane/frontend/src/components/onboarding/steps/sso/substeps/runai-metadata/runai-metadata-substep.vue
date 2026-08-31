<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <onboarding-content-wrapper>
    <template #title>Add NVIDIA Run:AI metadata to your identity provider</template>
    <template #subtitle>
      <p>
        Use the following if required by the identity provider. For more information, see
        <a :href="ssoDocLink" target="_blank">Set up SSO with {{ protocolNameToDisplay }}</a>
      </p>
    </template>
    <template #content>
      <div class="q-mb-xl" v-if="ssoData?.idp">
        <oidc-info v-if="ssoData.protocol === IdpTypeEnum.Oidc" :idp="ssoData.idp" />
        <saml-info v-if="ssoData.protocol === IdpTypeEnum.Saml" :idp="ssoData.idp" />
      </div>
    </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import type { SsoData } from "@/components/onboarding/models/admin-onboarding.model";
import { onboardingUtil } from "@/utils/onboarding.util/onboarding.util";

// Import components
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { OidcInfo } from "@/components/settings/sections/security/sso-settings/idps/oidc";
import { SamlInfo } from "@/components/settings/sections/security/sso-settings/idps/saml";

export default defineComponent({
  name: "runai-metadata-substep",
  components: {
    OnboardingContentWrapper,
    OidcInfo,
    SamlInfo,
  },
  emits: {
    "update:runai-metadata": (metadata: string) => typeof metadata === "string",
  },
  props: {
    runaiMetadata: {
      type: String as PropType<string>,
      default: "",
    },
    ssoData: {
      type: Object as PropType<SsoData>,
      required: true,
    },
  },
  data() {
    return {
      metadataValue: this.runaiMetadata,
      IdpTypeEnum,
    };
  },
  computed: {
    protocolNameToDisplay(): string {
      return onboardingUtil.getProtocolDisplayName(this.ssoData.protocol);
    },
    ssoDocLink(): string {
      return onboardingUtil.getSsoDocumentationLink(this.ssoData.protocol);
    },
  },
  methods: {
    handleMetadataChange(value: string) {
      this.metadataValue = value;
      this.$emit("update:runai-metadata", value);
    },
  },
  watch: {
    runaiMetadata: {
      handler(newValue: string) {
        this.metadataValue = newValue;
      },
      immediate: true,
    },
  },
});
</script>
