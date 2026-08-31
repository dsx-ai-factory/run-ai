<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="saml-metadata">
    <!-- SAML FIELDS SECTION -->
    <div class="q-mb-xl">
      <saml-fields
        :idp-resources="computedIdpResources"
        :is-new-idp="isNewIdp"
        :is-display-download-button="false"
        :disabled="readOnly"
        @updated="handleSamlFieldsUpdate"
        ref="samlFieldsRef"
      />
    </div>

    <!-- IDP MAPPERS SECTION -->
    <runai-folding-sub-section label="User attributes" :default-opened="false">
      <div>
        <idp-mappers :idp-mappers="idpMappersItems" :disabled="readOnly" />
      </div>
    </runai-folding-sub-section>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import SamlFields from "@/components/settings/sections/security/sso-settings/idps/saml/saml-fields/saml-fields.vue";
import IdpMappers from "@/components/settings/sections/security/sso-settings/idps/idp-mappers/idp-mappers.vue";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import type { SamlCreationData, Mappers, IdpResources } from "@/swagger-models/identity-manager-client";
import { SamlCreationDataMetadataXmlTypeEnum } from "@/swagger-models/identity-manager-client";
import type { IIdpMapperItem } from "@/models/setting.model";

import { settingsUtil } from "@/utils/settings.util";
import { useIdpsStore } from "@/stores/idps.store";
import { deepCopy } from "@/utils/common.util";

// SAML Metadata interface for onboarding (similar to OidcMetadata)
export interface SamlMetadata {
  samlData: SamlCreationData;
  mappers: Mappers;
}

export default defineComponent({
  name: "saml-metadata",
  components: { SamlFields, IdpMappers, RunaiFoldingSubSection },
  emits: ["update:metadata", "update:mappers"],
  props: {
    mappers: {
      type: Object as PropType<Mappers>,
      required: true,
    },
    samlData: {
      type: Object as PropType<SamlCreationData>,
      default: () => ({
        entityId: "",
        metadataXmlUrl: "",
        metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.File as SamlCreationDataMetadataXmlTypeEnum,
        metadataXmlFile: "",
        fileName: "",
      }),
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      samlFields: { ...this.samlData } as SamlCreationData,
      idpResources: {} as IdpResources, // Empty for new IDP in onboarding
      idpStore: useIdpsStore(),
      idpMappersItems: {} as Record<string, IIdpMapperItem>,
    };
  },
  created(): void {
    this.idpMappersItems = deepCopy(this.getIdpMapper());
  },
  mounted() {
    // Emit initial data to ensure parent has current state
    if (this.samlFields && Object.keys(this.samlFields).length > 0) {
      this.emitCurrentMetadata();
    }
  },
  computed: {
    // Transform samlData to the format expected by saml-fields component
    // For onboarding, we don't have server-side data yet, so return empty

    // Check if we have existing data to show
    isNewIdp(): boolean {
      // If we have existing data, treat it as existing IDP to load the data
      return !(this.samlFields.metadataXmlUrl || this.samlFields.metadataXmlFile || this.samlFields.fileName);
    },
    // Provide IDP resources when we have existing data
    computedIdpResources(): IdpResources {
      if (!this.isNewIdp && this.samlFields.metadataXmlUrl) {
        // For URL-based metadata
        return {
          resourceName: SamlCreationDataMetadataXmlTypeEnum.Url,
          resource: this.samlFields.metadataXmlUrl,
        } as IdpResources;
      } else if (!this.isNewIdp && this.samlFields.metadataXmlFile) {
        // For file-based metadata - DECODE Base64 back to raw XML
        // because saml-fields expects raw content and will re-encode it
        let rawXmlContent: string;
        try {
          // Decode Base64 to get raw XML content
          rawXmlContent = atob(this.samlFields.metadataXmlFile);
        } catch (error) {
          console.warn("Failed to decode Base64 metadata, using as-is:", error);
          rawXmlContent = this.samlFields.metadataXmlFile;
        }

        return {
          resourceName: this.samlFields.fileName || "metadata.xml",
          resource: rawXmlContent, // ← Pass RAW XML, not Base64
        } as IdpResources;
      }
      // Return empty for new IDP
      return {} as IdpResources;
    },
  },
  methods: {
    // Handle updates from saml-fields component (similar to saml-settings updateSamlFields)
    handleSamlFieldsUpdate(newSamlFields: SamlCreationData): void {
      if (JSON.stringify(this.samlFields) !== JSON.stringify(newSamlFields)) {
        this.samlFields = newSamlFields;
        this.emitCurrentMetadata();
      }
    },

    // Emit current metadata state
    emitCurrentMetadata(): void {
      const currentMetadata: SamlMetadata = {
        samlData: this.samlFields,
        mappers: settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems),
      };

      this.$emit("update:metadata", currentMetadata);
    },

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
    // Watch for changes in samlData prop and update local state
    samlData: {
      handler(newSamlData: SamlCreationData) {
        this.samlFields = { ...newSamlData };
      },
      deep: true,
      immediate: true,
    },
  },
});
</script>
