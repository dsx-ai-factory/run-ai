<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="saml-fields col-12">
    <notification-alert-bar
      v-if="signingDisabled"
      class="q-mb-md signing-disabled-warning"
      :variant="signingDisabledVariant"
      icon-override="ph-bold ph-warning-circle"
      show-h1-content
      :title="signingDisabledTitle"
      :message="signingDisabledMessage"
    />
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
              :rules="[isValidFile]"
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
              :rules="[isValidUrl]"
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
    <div class="row" v-if="samlSettings?.singleSignOnServiceUrl">
      <q-tooltip anchor="center middle">
        {{ disabledTooltip }}
      </q-tooltip>
      <q-input
        class="col-12 q-mb-lg"
        :model-value="samlSettings?.singleSignOnServiceUrl"
        no-error-icon
        stack-label
        label="Identity provider URL"
        :disable="true"
      />
    </div>
    <div class="row" v-if="samlSettings?.idpEntityId">
      <q-tooltip anchor="center middle">
        {{ disabledTooltip }}
      </q-tooltip>
      <q-input
        :model-value="samlSettings?.idpEntityId"
        class="col-12 q-mb-lg"
        no-error-icon
        stack-label
        label="Identity provider Entity ID (issuer)"
        :disable="true"
      />
    </div>
    <div class="row" v-if="isFileSource ? metadataFile : samlSettings?.signingCertificate">
      <q-tooltip anchor="center middle">
        {{ disabledTooltip }}
      </q-tooltip>
      <q-input
        :model-value="formattedExpirationDate"
        :rules="[iValidCertificateDate]"
        class="col-4 q-mb-lg"
        no-error-icon
        stack-label
        readonly
        :disable="isValidated"
        label="Certificate expiration date"
        ref="samlExpirationDateInput"
      />
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { isValidUrl } from "@/common/form.validators";
import type { QFile, QInput } from "quasar";
import type { IdpResources, SamlCreationData } from "@/swagger-models/identity-manager-client";
import { SamlCreationDataMetadataXmlTypeEnum } from "@/swagger-models/identity-manager-client";
import type { SamlConfigSettings } from "@/models/setting.model";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import { stringUtil } from "@/utils/string.util";
import { deepCopy, isEqual } from "@/utils/common.util";
import { HTMLUtil } from "@/utils/html.util/html.util";
import {
  DEFAULT_METADATA_FILE_NAME,
  DEFAULT_METADATA_FILE_TYPE,
  idpCommon,
  SAML_SIGNING_DISABLED_MESSAGE,
  SAML_SIGNING_DISABLED_TITLE,
} from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";
import { urlHintMessage } from "@/models/setting.model";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { RunaiBtn } from "@/components/common/runai-btn";
import { ENotificationBarVariant, NotificationAlertBar } from "@/components/common/notification-alert-bar";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "saml-fields",
  components: { RunaiBtn, SettingReadonlyInput, NotificationAlertBar },
  emits: ["updated"],
  props: {
    samlSettings: {
      type: Object as PropType<SamlConfigSettings>,
      required: false,
    },
    idpResources: {
      type: Object as PropType<IdpResources>,
      required: false,
    },
    isNewIdp: {
      type: Boolean as PropType<boolean>,
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
      expirationDate: null as Date | null,
      isValidated: false,
      samlSource: SamlCreationDataMetadataXmlTypeEnum.File as SamlCreationDataMetadataXmlTypeEnum,
      metadataFile: null as File | null,
      originalSamlFields: {} as SamlCreationData,
      fileSelectionToken: 0,
      signingDisabled: false,
    };
  },
  async created() {
    if (this.samlSettings?.signingCertificate) {
      await this.validateCertificate();
    }
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
    disabledTooltip(): string {
      return "This can't be modified. To make changes, modify the metadata file above";
    },
    formattedExpirationDate(): string {
      return this.expirationDate ? dateUtil.dateFormat(this.expirationDate, DATE_FORMATS.STANDARD_DATE) : "";
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
    isValidCertificate(): boolean {
      return idpCommon.iValidCertificateDate(this.expirationDate as Date);
    },
    securityResourceType(): keyof typeof ResourceType {
      return "SecuritySettings";
    },
    signingDisabledVariant(): ENotificationBarVariant {
      return ENotificationBarVariant.Error;
    },
    signingDisabledTitle(): string {
      return SAML_SIGNING_DISABLED_TITLE;
    },
    signingDisabledMessage(): string {
      return SAML_SIGNING_DISABLED_MESSAGE;
    },
  },
  methods: {
    async validate(): Promise<boolean> {
      if (this.samlSource === SamlCreationDataMetadataXmlTypeEnum.File) {
        (this.$refs.metadataFileInput as typeof QFile)?.validate();
        (this.$refs.samlExpirationDateInput as typeof QInput)?.validate();
        return !!this.metadataFile && this.isValidCertificate;
      } else {
        return (this.$refs.samlUrlInput as QInput).validate();
      }
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
          this.samlFields.metadataXmlUrl = this.idpResources.resource as string;
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
    updateMetadataXmlUrl(url: string | null | number): void {
      this.samlFields.metadataXmlUrl = url as string;
      this.samlFields.metadataXmlType = SamlCreationDataMetadataXmlTypeEnum.Url;
      this.onUpdated();
    },
    onUpdated(): void {
      this.$emit("updated", this.samlFields, this.isValidInput(), this.hasChanges());
    },
    isValidInput(): boolean {
      return this.isFileSource
        ? !!this.metadataFile && this.isValidCertificate
        : !!this.samlFields.metadataXmlUrl && isValidUrl(this.samlFields.metadataXmlUrl || "");
    },
    iValidCertificateDate(): boolean | string {
      if (this.isValidCertificate) return true;
      return this.isFileSource
        ? "Certificate expired. Replace the metadata file above."
        : "Certificate expired. Replace the metadata URL above.";
    },
    async validateCertificate(certificateOverride?: string | null, selectionToken?: number): Promise<boolean> {
      const token = selectionToken ?? this.fileSelectionToken;
      const certificate =
        certificateOverride !== undefined ? certificateOverride : this.samlSettings?.signingCertificate;
      this.isValidated = false;
      const expirationDate = certificate ? await idpCommon.getCertificateExpirationDate(certificate) : null;
      if (token !== this.fileSelectionToken) return false;
      this.expirationDate = expirationDate;
      (this.$refs.samlExpirationDateInput as typeof QInput)?.validate();
      if (this.expirationDate && idpCommon.iValidCertificateDate(this.expirationDate as Date)) {
        this.isValidated = true;
      }
      return true;
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
      this.expirationDate = null;
      this.isValidated = false;
      this.signingDisabled = false;
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
      const isCurrent = await this.validateCertificate(this.extractCertificateFromMetadataXml(fileContent), token);
      if (!isCurrent || token !== this.fileSelectionToken) return;
      this.signingDisabled = idpCommon.isSamlSigningDisabled(fileContent);
      this.samlFields.metadataXmlFile = stringUtil.stringToBase64(fileContent);
      this.samlFields.fileName = fileName;
      this.samlFields.metadataXmlType = SamlCreationDataMetadataXmlTypeEnum.File;
      this.samlFields.metadataXmlUrl = "";
      this.onUpdated();
    },
    extractCertificateFromMetadataXml(xmlContent: string): string | null {
      try {
        const doc = new DOMParser().parseFromString(xmlContent, "text/xml");
        const certEl = doc.getElementsByTagNameNS("*", "X509Certificate")[0];
        return certEl?.textContent?.replace(/\s/g, "") || null;
      } catch {
        return null;
      }
    },
    clearFileContent(): void {
      this.samlFields.metadataXmlFile = "";
      this.samlFields.fileName = "";
      this.expirationDate = null;
      this.isValidated = false;
      this.signingDisabled = false;
      (this.$refs.samlExpirationDateInput as typeof QInput)?.validate();
      this.onUpdated();
    },
    clearUrlContent(): void {
      this.samlFields.metadataXmlUrl = "";
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
