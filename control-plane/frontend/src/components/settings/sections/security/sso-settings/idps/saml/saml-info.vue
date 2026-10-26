<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="saml-info">
    <div class="q-mb-lg">
      <p>Download the metadata file</p>
      <saml-metadata-download-button :idp-resources="idpResources" :is-new-idp="false" :is-file-source="isFileSource" />
      <p>Or copy the fields individually</p>
    </div>

    <setting-readable-field label="Redirect URL" v-if="idp?.redirectUri">
      <div class="row no-wrap items-center">
        <div class="col-12 q-pt-md q-pl-md">
          {{ idp.redirectUri }}
        </div>
      </div>
    </setting-readable-field>
    <setting-readable-field label="Entity ID">
      <div class="row no-wrap items-center">
        <div class="col-12 q-pt-md q-pl-md">
          {{ idp?.samlData?.entityId || defaultEntityId }}
        </div>
      </div>
    </setting-readable-field>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { Idp, IdpResources } from "@/swagger-models/identity-manager-client";
import { SamlCreationDataMetadataXmlTypeEnum } from "@/swagger-models/identity-manager-client";
import { settingsUtil } from "@/utils/settings.util";
import { useIdpsStore } from "@/stores/idps.store";
import SettingReadableField from "@/components/settings/setting-readable-field/setting-readable-field.vue";
import { SamlMetadataDownloadButton } from "./saml-metadata-download-button";

export default defineComponent({
  name: "saml-info",
  components: {
    SettingReadableField,
    SamlMetadataDownloadButton,
  },
  props: {
    idp: {
      type: Object as PropType<Idp>,
      required: true,
    },
  },

  data() {
    return {
      idpResources: {} as IdpResources,
      idpStore: useIdpsStore(),
    };
  },
  async created(): Promise<void> {
    // Load IDP resources if we have an idp
    if (this.idp?.alias) {
      try {
        this.idpResources = await this.idpStore.loadIdpResources();
      } catch (error) {
        console.error("Failed to load IDP resources:", error);
        // Continue without resources - download button will be disabled
        this.idpResources = {} as IdpResources;
      }
    }
  },
  computed: {
    hasIdpAttributes(): boolean {
      return (this.idp?.redirectUri || "") !== "";
    },
    defaultEntityId(): string {
      return settingsUtil.getDefaultEntityId(window.location.hostname);
    },
    isFileSource(): boolean {
      return this.idpResources?.resourceName !== SamlCreationDataMetadataXmlTypeEnum.Url;
    },
  },
});
</script>
