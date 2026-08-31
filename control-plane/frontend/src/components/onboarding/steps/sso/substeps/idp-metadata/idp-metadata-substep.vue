<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="idp-metadata-substep">
    <sso-creation-error v-if="ssoData.ssoErrorType === SsoErrorType.IdpCreation" />
    <onboarding-content-wrapper v-else>
      <template #title>Set your identity provider metadata ({{ protocolNameToDisplay }})</template>
      <template #subtitle>
        <p>Add your identity provider (IdP) metadata to establish trust with NVIDIA Run:ai.</p>
        <p>
          For more information, see <a :href="ssoDocLink" target="_blank">Set up SSO with {{ protocolNameToDisplay }}</a>
        </p>
      </template>
      <template #content>
        <!-- SAML Metadata Component -->
        <saml-metadata
          v-if="ssoData.protocol === IdpTypeEnum.Saml"
          ref="samlMetadataRef"
          :mappers="ssoData.mappers"
          :saml-data="ssoData.samlData || defaultSamlData"
          :read-only="ssoData.hasIdpCreated"
          @update:metadata="handleMetadataChange"
          @update:mappers="handleMappersChange"
        />

        <!-- OIDC Metadata Component -->
        <oidc-metadata
          v-else-if="ssoData.protocol === IdpTypeEnum.Oidc"
          ref="oidcMetadataRef"
          :mappers="ssoData.mappers"
          :oidc-data="ssoData.oidcData || defaultOidcData"
          :read-only="ssoData.hasIdpCreated"
          @update:metadata="handleMetadataChange"
          @update:mappers="handleMappersChange"
        />

        <!-- Protocol not selected state -->
        <div v-else-if="!ssoData.protocol" class="no-protocol-selected">
          <q-icon name="info" size="24px" color="grey-6" />
          <p class="no-protocol-text">Please select a protocol first to configure metadata.</p>
        </div>
      </template>
    </onboarding-content-wrapper>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import SamlMetadata, { type SamlMetadata as SamlMetadataType } from "./saml-metadata.vue";
import OidcMetadata, { type OidcMetadata as OidcMetadataType } from "./oidc-metadata.vue";
import type { SsoData } from "@/components/onboarding/models/admin-onboarding.model";
import { SsoErrorType } from "@/components/onboarding/models/admin-onboarding.model";
import { ssoDataUpdateEmitValidator } from "@/components/onboarding/models/admin-onboarding.model";
import {
  IdpTypeEnum,
  type OidcCreationData,
  type SamlCreationData,
  SamlCreationDataMetadataXmlTypeEnum,
} from "@/swagger-models/identity-manager-client";
import { OPENID_SCOPE } from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";
import { onboardingUtil } from "@/utils/onboarding.util/onboarding.util";
import { SsoCreationError } from "@/components/onboarding/steps/sso/substeps/errors-state/sso-creation-error";
import type { Mappers } from "@/swagger-models/identity-manager-client";

export default defineComponent({
  name: "idp-metadata-substep",
  components: { OnboardingContentWrapper, SamlMetadata, OidcMetadata, SsoCreationError },
  emits: {
    "update:idp-metadata": (metadata: object) => typeof metadata === "object",
    "update:sso-data": ssoDataUpdateEmitValidator,
  },
  props: {
    idpMetadata: {
      type: [String, Object] as PropType<string | object>,
      default: () => ({}),
    },
    ssoData: {
      type: Object as PropType<SsoData>,
      default: () => ({}),
    },
  },
  data() {
    return {
      metadataValue: this.idpMetadata as string | object,
      IdpTypeEnum,
      SsoErrorType,
    };
  },
  computed: {
    protocolNameToDisplay(): string {
      return onboardingUtil.getProtocolDisplayName(this.ssoData.protocol);
    },
    ssoDocLink(): string {
      return onboardingUtil.getSsoDocumentationLink(this.ssoData.protocol);
    },
    defaultOidcData(): OidcCreationData {
      return {
        clientId: "",
        clientSecret: "",
        discoverDocumentUrl: "",
        scopes: [OPENID_SCOPE],
        mandatoryClaim: null,
      };
    },
    defaultSamlData(): SamlCreationData {
      return {
        entityId: "",
        metadataXmlUrl: "",
        metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.File,
        metadataXmlFile: "",
        fileName: "",
      };
    },
  },
  methods: {
    handleMetadataChange(value: object) {
      this.metadataValue = value;
      this.$emit("update:idp-metadata", value);

      // Also emit the structured data to the ssoData based on protocol
      if (this.ssoData.protocol === IdpTypeEnum.Oidc) {
        // Convert OIDC metadata to API format
        const oidcMetadata = value as OidcMetadataType;

        const oidcData: OidcCreationData = {
          discoverDocumentUrl: oidcMetadata.discoveryUrl,
          clientId: oidcMetadata.clientId,
          clientSecret: oidcMetadata.clientSecret,
          scopes: oidcMetadata.scopes,
          mandatoryClaim: null,
        };

        // Emit structured data to parent (but not mappers, those are handled separately)
        this.$emit("update:sso-data", "oidcData", oidcData);
      } else if (this.ssoData.protocol === IdpTypeEnum.Saml) {
        // For SAML, the value is now a structured SamlMetadata object (like OIDC)
        const samlMetadata = value as SamlMetadataType;

        // Emit structured data to parent (but not mappers, those are handled separately)
        this.$emit("update:sso-data", "samlData", samlMetadata.samlData);
      }
    },

    // Handle mappers changes from child components
    handleMappersChange(mappers: Mappers) {
      this.$emit("update:sso-data", "mappers", mappers);
    },
  },
  watch: {
    idpMetadata: {
      handler(newValue: string | object) {
        this.metadataValue = newValue;
      },
      immediate: true,
    },
  },
});
</script>
