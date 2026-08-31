<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="saml-metadata-download-button">
    <runai-btn
      label="metadata XML"
      class="metadata-xml-download-btn q-mt-md q-mb-lg"
      @click="downloadFile"
      icon="ph-fill ph-download"
      :disable="!canDownload"
      aid="download-metadata"
      :tooltip="downloadTooltip"
      color="grey-4"
      text-color="grey-9"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IdpResources } from "@/swagger-models/identity-manager-client";
import { HTMLUtil } from "@/utils/html.util/html.util";
import {
  DEFAULT_METADATA_FILE_NAME,
  DEFAULT_METADATA_FILE_TYPE,
} from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  name: "saml-metadata-download-button",
  components: { RunaiBtn },
  props: {
    idpResources: {
      type: Object as PropType<IdpResources>,
      required: false,
    },
    isNewIdp: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isFileSource: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  computed: {
    canDownload(): boolean {
      return !this.isNewIdp && !!this.idpResources?.resource;
    },
    downloadTooltip(): string {
      if (this.isNewIdp) {
        return "To download, first save the settings";
      }
      if (!this.idpResources?.resource) {
        return "No metadata available to download";
      }
      return "Download the SAML metadata XML file";
    },
  },
  methods: {
    downloadFile(): void {
      if (!this.idpResources?.resource) return;

      const fileContent: string = this.idpResources.resource as string;
      const fileName: string =
        (((this.idpResources.resourceName as string) && this.isFileSource) as string) || DEFAULT_METADATA_FILE_NAME;

      HTMLUtil.downloadFileFromString(fileContent, fileName, DEFAULT_METADATA_FILE_TYPE);
    },
  },
});
</script>

<style lang="scss">
.metadata-xml-download-btn {
  .q-icon {
    font-size: 20px;
  }
}
</style>
