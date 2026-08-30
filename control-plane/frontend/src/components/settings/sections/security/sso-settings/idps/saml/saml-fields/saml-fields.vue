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
    <div class="row" v-if="samlSettings?.signingCertificate">
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
} from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";
import { urlHintMessage } from "@/models/setting.model";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { RunaiBtn } from "@/components/common/runai-btn";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "saml-fields",
  components: { RunaiBtn, SettingReadonlyInput },
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
    };
  },
  async created() {
    if (this.samlSettings?.signingCertificate) {
      await this.validateCertificate();
    }
    if (!this.isNewIdp && this.idpResources?.resource) {
      this.setSamlResources();
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
  },
  methods: {
    async validate(): Promise<boolean> {
      if (this.samlSource === SamlCreationDataMetadataXmlTypeEnum.File) {
        return !!this.metadataFile;
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
    setSamlResources() {
      if (this.idpResources) {
        this.samlSource = this.getSourceType;
        if (this.samlSource === SamlCreationDataMetadataXmlTypeEnum.File) {
          this.setFileContent(this.idpResources.resourceName as string, this.idpResources.resource as string);
          this.metadataFile = new File([this.idpResources.resource as string], this.idpResources.resourceName as string);
        }
      }
    },
    isValidUrl(url: string): boolean | string {
      return isValidUrl(url) || (!this.isNewIdp && !url) || "Enter a valid URL";
    },
    isValidFile(file: string): boolean | string {
      return file || !this.isNewIdp || "Upload a file";
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
        ? !!this.metadataFile
        : !!this.samlFields.metadataXmlUrl && isValidUrl(this.samlFields.metadataXmlUrl || "");
    },
    iValidCertificateDate(): boolean | string {
      return this.isValidCertificate || "Certificate expired. To replace it, enter another metadata URL above.";
    },
    async validateCertificate(): Promise<void> {
      this.expirationDate = await idpCommon.getCertificateExpirationDate(
        this.samlSettings?.signingCertificate as string,
      );
      (this.$refs.samlExpirationDateInput as typeof QInput)?.validate();
      if (this.expirationDate && idpCommon.iValidCertificateDate(this.expirationDate as Date)) {
        this.isValidated = true;
      }
    },
    clearFile(): void {
      this.metadataFile = null;
      this.samlFields.metadataXmlUrl = "";
      this.clearFileContent();
      this.onUpdated();
    },
    getFileContent(file: File): void {
      this.metadataFile = file;
      const reader = new FileReader();
      reader.onload = (e: ProgressEvent<FileReader>) => {
        const fileContent = e.target?.result as string;
        this.setFileContent(file.name, fileContent);
      };
      reader.readAsText(file);
    },
    setFileContent(fileName: string, fileContent: string): void {
      this.samlFields.metadataXmlFile = stringUtil.stringToBase64(fileContent);
      this.samlFields.fileName = fileName;
      this.samlFields.metadataXmlType = SamlCreationDataMetadataXmlTypeEnum.File;
      this.samlFields.metadataXmlUrl = "";
      this.onUpdated();
    },
    clearFileContent(): void {
      this.samlFields.metadataXmlFile = "";
      this.samlFields.fileName = "";
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
