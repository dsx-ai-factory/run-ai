<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="oidc-metadata">
    <!-- OIDC FIELDS SECTION -->
    <div class="q-mb-xl">
      <oidc-fields
        :oidc-settings="oidcSettings"
        :is-edit-mode="false"
        :disabled="readOnly"
        @updated="handleOidcFieldsUpdate"
        ref="oidcFieldsRef"
      />
    </div>

    <!-- IDP SCOPE SECTION -->
    <runai-folding-sub-section label="User attributes" :default-opened="false">
      <div class="q-mb-xl">
        <idp-scope
          :scopes="sourceOidcSettings.scopes"
          idp-name="OIDC"
          :disabled="readOnly"
          @update-scope="handleScopeUpdate"
        />
      </div>

      <!-- IDP MAPPERS SECTION -->
      <div>
        <idp-mappers :idp-mappers="idpMappersItems" :disabled="readOnly" />
      </div>
    </runai-folding-sub-section>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { OidcFields } from "@/components/settings/sections/security/sso-settings/idps/oidc/oidc-fields";
import IdpScope from "@/components/settings/sections/security/sso-settings/idps/idp-scope/idp-scope.vue";
import IdpMappers from "@/components/settings/sections/security/sso-settings/idps/idp-mappers/idp-mappers.vue";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import type { OidcCreationData, Mappers } from "@/swagger-models/identity-manager-client";
import type { IIdpMapperItem } from "@/models/setting.model";

import { OPENID_SCOPE } from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";
import { settingsUtil } from "@/utils/settings.util";
import { deepCopy } from "@/utils/common.util";

// OIDC Metadata interface for onboarding
export interface OidcMetadata {
  discoveryUrl: string;
  clientId: string;
  clientSecret: string;
  scopes: string[];
  mappers: Mappers;
}

export default defineComponent({
  name: "oidc-metadata",
  components: { OidcFields, IdpScope, IdpMappers, RunaiFoldingSubSection },
  emits: ["update:metadata", "update:mappers"],
  props: {
    mappers: {
      type: Object as PropType<Mappers>,
      required: true,
    },
    oidcData: {
      type: Object as PropType<OidcCreationData>,
      default: () => ({
        discoverDocumentUrl: "",
        clientId: "",
        clientSecret: "",
        scopes: [OPENID_SCOPE],
        mandatoryClaim: null,
      }),
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      oidcConfigSettings: { ...this.oidcData } as OidcCreationData,
      sourceOidcSettings: { ...this.oidcData } as OidcCreationData,
      isValidFields: true,
      idpMappersItems: {} as Record<string, IIdpMapperItem>,
    };
  },
  created(): void {
    // Set source settings for comparison (like original oidc-settings)
    this.sourceOidcSettings = { ...this.oidcConfigSettings };
    this.idpMappersItems = deepCopy(this.getIdpMapper());
  },
  computed: {
    // Return current oidc settings (like original oidc-settings)
    oidcSettings(): OidcCreationData {
      return this.oidcConfigSettings;
    },
  },
  methods: {
    // Handle updates from oidc-fields component (exactly like oidc-settings updateOidcFields)
    handleOidcFieldsUpdate(newOidcFields: OidcCreationData): void {
      this.oidcConfigSettings = {
        ...this.oidcConfigSettings,
        ...newOidcFields,
      };
      this.emitCurrentMetadata();
    },

    // Handle scope updates (exactly like oidc-settings updateScope)
    handleScopeUpdate(scopes: string[]): void {
      this.oidcConfigSettings.scopes = scopes;
      this.handleOidcFieldsUpdate(this.oidcConfigSettings);
    },

    // Emit current metadata state
    emitCurrentMetadata(): void {
      const currentMetadata: OidcMetadata = {
        discoveryUrl: this.oidcConfigSettings.discoverDocumentUrl,
        clientId: this.oidcConfigSettings.clientId,
        clientSecret: this.oidcConfigSettings.clientSecret,
        scopes: this.oidcConfigSettings.scopes || [OPENID_SCOPE],
        mappers: settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems),
      };

      this.$emit("update:metadata", currentMetadata);
    },

    // Convert mappers from API format to component format (like settings components)
    getIdpMapper(): Record<string, IIdpMapperItem> {
      return this.mappers ? settingsUtil.getIdpMappersItems(this.mappers) : {};
    },
  },
  watch: {
    // Watch for changes in mappers and emit them back to parent
    idpMappersItems: {
      handler(newMappers: Record<string, IIdpMapperItem>) {
        const mappersForApi = settingsUtil.getMappersFromIdpMapperItems(newMappers);
        this.$emit("update:mappers", mappersForApi);
      },
      deep: true,
    },
  },
});
</script>
<style lang="scss">
.oidc-metadata {
  .client-inputs {
    flex-direction: columns;
    .q-field {
      width: 50%;
    }
  }
}
</style>
