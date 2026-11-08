<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="saml-fields col-12">
    <div class="row">
      <div class="q-mb-md">
        <div class="col-12 q-mb-md">Upload your identity provider metadata file</div>
        <div class="col-12 text-body1 source-radio">
          <setting-readonly-input :resource-type="securityResourceType">
            <template v-slot:default="{ readonly }">
              <q-radio
                v-model="samlSource"
                :val="samlSourceType.File"
                label="From computer"
                @update:model-value="clearUrlContent()"
                :disable="readonly || disabled"
              />
            </template>
          </setting-readonly-input>
        </div>
        <div class="col-12 text-body1 source-radio">
          <setting-readonly-input :resource-type="securityResourceType">
            <template v-slot:default="{ readonly }">
              <q-radio
                v-model="samlSource"
                :val="samlSourceType.Url"
                label="From URL"
                @update:model-value="clearFile()"
                :disable="readonly || disabled"
              />
            </template>
          </setting-readonly-input>
        </div>
      </div>
      <div class="row col-12">
        <setting-readonly-input class="col-12" :resource-type="securityResourceType">
          <template v-slot:default="{ readonly }">
            <q-file
              v-if="isFileSource"
              v-model="metadataFile"
              label="Metadata XML file"
              accept=".xml"
              stack-label
              ref="metadataFileInput"
              class="col-12"
              @update:model-value="getFileContent"
              :rules="[isValidFile, metadataStatus]"
              no-error-icon
              :disable="readonly || disabled"
            >
              <template v-slot:prepend>
                <q-icon name="ph-bold ph-cloud-arrow-up" @click.stop.prevent size="20px" />
              </template>
              <template v-slot:append>
                <q-icon name="ph-bold ph-x" class="cursor-pointer" size="16px" @click="clearFile()" />
              </template>
            </q-file>
          </template>
        </setting-readonly-input>
        <setting-readonly-input class="col-12" :resource-type="securityResourceType">
          <template v-slot:default="{ readonly }">
            <q-input
              v-if="!isFileSource"
              class="col-12"
              :model-value="samlFields.metadataXmlUrl"
              placeholder="Enter URL"
              debounce="300"
              :rules="[isValidUrl, metadataStatus]"
              no-error-icon
              @update:model-value="updateMetadataXmlUrl"
              stack-label
              label="Metadata URL"
              ref="samlUrlInput"
              :hint="urlHint"
              :disable="readonly || disabled"
            />
          </template>
        </setting-readonly-input>
      </div>
      <div v-if="isDisplayDownloadButton">
        <runai-btn
          label="Download idp metadata XML file"
          class="download-btn q-mt-md q-mb-lg"
          @click="downloadFile"
          icon="ph-fill ph-download"
          :disable="isNewIdp || idpResources?.resource == undefined"
          aid="download-metadata"
          color="grey-4"
          disabled-tooltip="To download, first save the settings"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { isValidUrl } from "@/common/form.validators";
import type { QFile, QInput } from "quasar";
import type { IdpResources, SamlCreationData, SamlData } from "@/swagger-models/identity-manager-client";
import {
  IdpCreationRequestTypeEnum,
  SamlCreationDataMetadataXmlTypeEnum,
} from "@/swagger-models/identity-manager-client";
import { idpService } from "@/services/control-plane/idp.service/idp.service";
import type { SamlCertificateDetails } from "@/models/setting.model";
import { stringUtil } from "@/utils/string.util";
import { deepCopy, isEqual } from "@/utils/common.util";
import { HTMLUtil } from "@/utils/html.util/html.util";
import {
  DEFAULT_METADATA_FILE_NAME,
  DEFAULT_METADATA_FILE_TYPE,
} from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";
import { urlHintMessage } from "@/models/setting.model";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { RunaiBtn } from "@/components/common/runai-btn";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "saml-fields",
  components: { RunaiBtn, SettingReadonlyInput },
  emits: ["updated", "certificate-details"],
  props: {
    idpResources: {
      type: Object as PropType<IdpResources>,
      required: false,
    },
    isNewIdp: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    idpAlias: {
      type: String as PropType<string>,
      required: false,
    },
    isDisplayDownloadButton: {
      type: Boolean as PropType<boolean>,
      default: true,
      required: false,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
  },
  data() {
    return {
      samlFields: {} as SamlCreationData,
      samlSource: SamlCreationDataMetadataXmlTypeEnum.File as SamlCreationDataMetadataXmlTypeEnum,
      metadataFile: null as File | null,
      originalSamlFields: {} as SamlCreationData,
      fileSelectionToken: 0,
      metadataAccepted: false,
      validationError: "",
      returnedSaml: null as SamlData | null,
    };
  },
  async created() {
    if (!this.isNewIdp && this.idpResources?.resource) {
      await this.setSamlResources();
    }
    this.originalSamlFields = deepCopy(this.samlFields);
    this.onUpdated();
  },
  computed: {
    urlHint(): string {
      return !this.isNewIdp ? urlHintMessage : "";
    },
    samlSourceType() {
      return SamlCreationDataMetadataXmlTypeEnum;
    },
    certificateDetails(): SamlCertificateDetails {
      return {
        showExpirationDate: !!this.validationError,
        expirationDate: null,
        expirationError: this.validationError,
        idpEntityId: this.returnedSaml?.entityId || "",
        singleSignOnServiceUrl: this.returnedSaml?.singleSignOnServiceUrl || "",
      };
    },
    isFileSource(): boolean {
      return this.samlSource === SamlCreationDataMetadataXmlTypeEnum.File;
    },
    getSourceType(): SamlCreationDataMetadataXmlTypeEnum {
      return this.idpResources?.resourceName === SamlCreationDataMetadataXmlTypeEnum.Url
        ? SamlCreationDataMetadataXmlTypeEnum.Url
        : SamlCreationDataMetadataXmlTypeEnum.File;
    },
    fileName(): string {
      return this.samlFields.fileName || "";
    },
    securityResourceType(): keyof typeof ResourceType {
      return "SecuritySettings";
    },
  },
  methods: {
    async validate(): Promise<boolean> {
      if (this.samlSource === SamlCreationDataMetadataXmlTypeEnum.File) {
        await (this.$refs.metadataFileInput as typeof QFile)?.validate();
        return this.isValidInput();
      }
      const urlValid = await (this.$refs.samlUrlInput as QInput).validate();
      return urlValid !== false && this.isValidInput();
    },
    isEmpty(): boolean {
      switch (this.samlSource) {
        case SamlCreationDataMetadataXmlTypeEnum.File:
          return !this.samlFields.fileName;
        case SamlCreationDataMetadataXmlTypeEnum.Url:
          return !this.samlFields.metadataXmlUrl;
      }
    },
    hasChanges(): boolean {
      return !this.isEmpty() && !isEqual(this.samlFields, this.originalSamlFields);
    },
    async setSamlResources(): Promise<void> {
      if (this.idpResources) {
        this.samlSource = this.getSourceType;
        if (this.samlSource === SamlCreationDataMetadataXmlTypeEnum.File) {
          this.metadataFile = new File([this.idpResources.resource as string], this.idpResources.resourceName as string);
          await this.setFileContent(this.idpResources.resourceName as string, this.idpResources.resource as string);
        } else {
          // For URL sources the backend may return the downloaded metadata XML as the resource, not the URL itself.
          const resource = this.idpResources.resource as string;
          if (isValidUrl(resource)) this.samlFields.metadataXmlUrl = resource;
          this.samlFields.metadataXmlType = SamlCreationDataMetadataXmlTypeEnum.Url;
        }
      }
    },
    isValidUrl(url: string): boolean | string {
      const canLeaveBlank = !this.isNewIdp && this.getSourceType === SamlCreationDataMetadataXmlTypeEnum.Url;
      return isValidUrl(url) || (canLeaveBlank && !url) || "Enter a valid URL";
    },
    isValidFile(file: string): boolean | string {
      return !!file || "Upload a file";
    },
    metadataStatus(): boolean | string {
      if (this.validationError) return this.validationError;
      if (this.requiresDryRun() && !this.metadataAccepted) return "Checking metadata";
      return true;
    },
    requiresDryRun(): boolean {
      if (this.isFileSource) return !!this.samlFields.metadataXmlFile;
      const url = this.samlFields.metadataXmlUrl || "";
      return !!url && url !== this.originalSamlFields.metadataXmlUrl && this.isValidUrl(url) === true;
    },
    async updateMetadataXmlUrl(url: string | null | number): Promise<void> {
      const token = ++this.fileSelectionToken;
      this.samlFields.metadataXmlUrl = url as string;
      this.samlFields.metadataXmlType = SamlCreationDataMetadataXmlTypeEnum.Url;
      this.metadataAccepted = false;
      this.validationError = "";
      this.returnedSaml = null;
      if (!this.samlFields.metadataXmlUrl || this.isValidUrl(this.samlFields.metadataXmlUrl) !== true) {
        this.onUpdated();
        return;
      }
      if (this.samlFields.metadataXmlUrl === this.originalSamlFields.metadataXmlUrl) {
        this.onUpdated();
        return;
      }
      await this.validateMetadata(token);
    },
    onUpdated(): void {
      this.$emit("updated", this.samlFields, this.isValidInput(), this.hasChanges());
    },
    isValidInput(): boolean {
      if (this.validationError) return false;
      if (this.isFileSource) {
        return !!this.metadataFile && !!this.samlFields.metadataXmlFile && this.metadataAccepted;
      }
      const url = this.samlFields.metadataXmlUrl || "";
      if (this.isValidUrl(url) !== true) return false;
      if (!url || url === this.originalSamlFields.metadataXmlUrl) return true;
      return this.metadataAccepted;
    },
    clearFile(): void {
      this.fileSelectionToken++;
      this.metadataFile = null;
      this.samlFields.metadataXmlUrl = "";
      this.clearFileContent();
      (this.$refs.metadataFileInput as typeof QFile)?.validate();
      this.onUpdated();
    },
    getFileContent(file: File): void {
      const selectionToken = ++this.fileSelectionToken;
      this.metadataFile = file;
      this.samlFields.metadataXmlFile = "";
      this.samlFields.fileName = "";
      this.metadataAccepted = false;
      this.validationError = "";
      this.returnedSaml = null;
      (this.$refs.metadataFileInput as typeof QFile)?.validate();
      this.onUpdated();
      const reader = new FileReader();
      reader.onload = (e: ProgressEvent<FileReader>) => {
        if (selectionToken !== this.fileSelectionToken) return;
        const fileContent = e.target?.result as string;
        void this.setFileContent(file.name, fileContent, selectionToken);
      };
      reader.readAsText(file);
    },
    async setFileContent(fileName: string, fileContent: string, selectionToken?: number): Promise<void> {
      const token = selectionToken ?? this.fileSelectionToken;
      if (token !== this.fileSelectionToken) return;
      this.samlFields.metadataXmlFile = stringUtil.stringToBase64(fileContent);
      this.samlFields.fileName = fileName;
      this.samlFields.metadataXmlType = SamlCreationDataMetadataXmlTypeEnum.File;
      this.samlFields.metadataXmlUrl = "";
      this.metadataAccepted = false;
      this.validationError = "";
      await this.validateMetadata(token);
    },
    async validateMetadata(token: number): Promise<void> {
      const request = {
        type: IdpCreationRequestTypeEnum.Saml,
        samlData: { ...this.samlFields },
        dryRun: true,
      };
      try {
        const response = this.idpAlias
          ? await idpService.dryRunIdp(request, this.idpAlias)
          : await idpService.dryRunIdp(request);
        if (token !== this.fileSelectionToken) return;
        this.applyReturnedSaml(response.samlData);
        this.metadataAccepted = true;
        this.validationError = "";
      } catch (err: unknown) {
        if (token !== this.fileSelectionToken) return;
        this.applyReturnedSaml(this.samlDataFromError(err));
        this.metadataAccepted = false;
        this.validationError = this.validationMessage(err);
      }
      this.onUpdated();
      await this.$nextTick();
      if (this.isFileSource) {
        (this.$refs.metadataFileInput as typeof QFile)?.validate();
      } else {
        (this.$refs.samlUrlInput as QInput)?.validate();
      }
    },
    applyReturnedSaml(samlData?: SamlData): void {
      this.returnedSaml = samlData ?? null;
      if (samlData?.entityId) this.samlFields.entityId = samlData.entityId;
    },
    samlDataFromError(err: unknown): SamlData | undefined {
      if (typeof err !== "object" || err === null || !("samlData" in err)) return undefined;
      return (err as { samlData?: SamlData }).samlData;
    },
    validationMessage(err: unknown): string {
      if (
        typeof err === "object" &&
        err !== null &&
        "message" in err &&
        typeof err.message === "string" &&
        err.message
      ) {
        return err.message;
      }
      return "SAML metadata is invalid";
    },
    clearFileContent(): void {
      this.samlFields.metadataXmlFile = "";
      this.samlFields.fileName = "";
      this.metadataAccepted = false;
      this.validationError = "";
      this.returnedSaml = null;
      this.onUpdated();
    },
    clearUrlContent(): void {
      this.fileSelectionToken++;
      this.samlFields.metadataXmlUrl = "";
      this.metadataAccepted = false;
      this.validationError = "";
      this.returnedSaml = null;
      this.onUpdated();
    },
    downloadFile(): void {
      if (!this.idpResources?.resource) return;
      const fileContent: string = this.idpResources.resource as string;
      const fileName: string =
        (((this.idpResources.resourceName as string) && this.isFileSource) as string) || DEFAULT_METADATA_FILE_NAME;
      HTMLUtil.downloadFileFromString(fileContent, fileName, DEFAULT_METADATA_FILE_TYPE);
    },
  },
  watch: {
    certificateDetails: {
      handler(details: SamlCertificateDetails): void {
        this.$emit("certificate-details", details);
      },
      immediate: true,
    },
    samlSource(): void {
      this.$nextTick(() => {
        if (this.isFileSource) {
          (this.$refs.metadataFileInput as QFile)?.focus();
        } else {
          (this.$refs.samlUrlInput as QInput)?.focus();
        }
      });
    },
  },
});
</script>
<style lang="scss" scoped>
.saml-fields {
  .source-radio {
    margin-left: -10px;
  }
}
</style>
<style lang="scss">
.saml-fields {
  .download-btn {
    color: $black-54 !important;
    .q-icon {
      color: $black-54;
    }
  }
}
</style>
