<template>
  <runai-expansion-item
    class="settings-container"
    label="Branding"
    :model-value="isExpanded"
    @update:model-value="onExpansionChanged"
  >
    <template #header>
      <div class="header-container">
        <div>Branding</div>
        <div v-if="hasCurrentLogo && !isExpanded" class="header-logo-preview">
          <img :src="displayLogoDataUrl || ''" alt="logo preview" class="header-logo-image" />
        </div>
      </div>
    </template>
    <div class="q-pa-md">
      <div class="q-mb-sm">
        <div class="text-subtitle1">Logo</div>
        <div class="row items-center q-mt-xs">
          <div class="subtitle q-mr-sm">
            Upload the logo you want displayed in the top-right corner of the platform interface
          </div>
          <runai-tooltip
            aid="branding-logo-help"
            :tooltip-text="logoTooltipText"
            tooltip-position="right"
            width="420px"
          />
        </div>
      </div>
      <!-- Show + LOGO button when no logo exists -->
      <div v-if="!isBoxOpen && !hasCurrentLogo">
        <setting-readonly-input resource-type="BrandingSettings">
          <template v-slot:default="{ readonly }">
            <q-btn
              v-if="canEditBranding || readonly"
              aid="add-logo-btn"
              icon="ph-bold ph-plus"
              label="LOGO"
              class="btn-icon-sm"
              flat
              color="primary"
              @click="enterEditMode"
              :disable="readonly"
            />
          </template>
        </setting-readonly-input>
      </div>

      <!-- Show logo preview with edit/remove actions when logo exists -->
      <div v-if="!isBoxOpen && hasCurrentLogo" class="logo-read-box q-pa-md">
        <div class="preview-wrapper">
          <img :src="displayLogoDataUrl || ''" alt="logo preview" class="logo-preview" />
        </div>
        <div v-if="currentLogoFileName" class="logo-info">
          <div class="filename-display">
            <span class="filename-text">{{ currentLogoFileName }}</span>
          </div>
        </div>
        <setting-readonly-input resource-type="BrandingSettings">
          <template v-slot:default="{ readonly }">
            <div class="logo-actions">
              <q-btn
                v-if="canEditBranding || readonly"
                aid="branding-edit-logo-btn"
                round
                flat
                size="sm"
                icon="ph-bold ph-pencil-simple"
                color="black-54"
                class="collapsed-state-icon"
                :disable="readonly"
                @click="enterEditMode"
              >
                <q-tooltip v-if="!readonly">Edit logo</q-tooltip>
              </q-btn>
              <q-btn
                v-if="canDeleteBranding || readonly"
                aid="branding-remove-logo-btn"
                round
                flat
                size="sm"
                icon="ph-bold ph-x"
                color="black-54"
                class="collapsed-state-icon"
                :disable="readonly"
                @click="removeLogoQuick"
              >
                <q-tooltip v-if="!readonly">Remove logo</q-tooltip>
              </q-btn>
            </div>
          </template>
        </setting-readonly-input>
      </div>

      <setting-wrapper-box
        v-if="isBoxOpen"
        has-footer
        :show-primary="canEditBranding"
        :primary-disabled="false"
        :secondary-button-label="isReadOnly ? 'Close' : 'Cancel'"
        :validation-message="hasUnsavedChanges ? 'Unsaved changes' : ''"
        @primary-button-clicked="onSave"
        @secondary-button-clicked="onCancel"
        @edit-clicked="enterEditMode"
      >
        <div class="q-mb-sm">
          <div class="custom-logo-label">Custom logo</div>
        </div>
        <div class="row col-12">
          <setting-readonly-input class="col-12" resource-type="BrandingSettings">
            <template v-slot:default="{ readonly }">
              <!-- eslint-disable vuejs-accessibility/no-autofocus -- this uploader only renders once the user clicks Add/Edit logo (v-if="isBoxOpen"), never on a bare page load -->
              <q-file
                ref="logoUploader"
                v-model="selectedFile"
                :disable="isSaving || readonly"
                :loading="isSaving"
                :error="!!errorMessage"
                :error-message="errorMessage"
                no-error-icon
                :accept="acceptedTypes"
                label=""
                autofocus
                class="col-12"
                @update:model-value="onFileChangeEdit"
                :rules="[isValidFile]"
              >
                <template #prepend>
                  <div class="preview-wrapper" v-if="displayLogoDataUrl">
                    <img :src="displayLogoDataUrl || ''" alt="logo preview" class="logo-preview" />
                  </div>
                  <q-icon v-else name="ph-bold ph-cloud-arrow-up" size="20px" />
                </template>
                <template #append>
                  <q-icon
                    v-if="showRemoveX && !readonly"
                    name="ph-bold ph-x"
                    class="cursor-pointer"
                    size="xs"
                    @click.stop.prevent="onAppendXClick"
                  />
                </template>
              </q-file>
              <!-- eslint-enable vuejs-accessibility/no-autofocus -->
            </template>
          </setting-readonly-input>
        </div>
      </setting-wrapper-box>
    </div>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { brandingSettingsService } from "@/services/control-plane/branding-settings.service/branding-settings.service";
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { SettingWrapperBox } from "@/components/settings/setting-wrapper-box";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { LogoConstants, settingUpdateSuccessMessage } from "@/models/setting.model";
import { alertUtil } from "@/utils/alert.util";
import { imageUtil } from "@/utils/image.util";
import { ResourceType, Action } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "branding-section",
  components: { RunaiExpansionItem, RunaiTooltip, SettingWrapperBox, SettingReadonlyInput },
  data() {
    return {
      isSaving: false as boolean,
      errorMessage: "" as string,

      selectedFile: null as File | null,
      currentLogoDataUrl: null as string | null,
      previewDataUrl: null as string | null,
      authStore: useAuthStore(),
      permissionStore: usePermissionStore(),
      removePending: false as boolean,
      isBoxOpen: false as boolean,
      isExpanded: false as boolean,
      hidePreviewInEdit: false as boolean,
    };
  },
  created() {
    this.setCurrentLogoFromStore();
    this.loadCurrentLogo();
  },
  computed: {
    isReadOnly(): boolean {
      return this.permissionStore.isReadOnly(ResourceType.BrandingSettings);
    },
    canEditBranding(): boolean {
      return this.permissionStore.hasSomePermittedActions(ResourceType.BrandingSettings, [Action.Create, Action.Update]);
    },
    canDeleteBranding(): boolean {
      return this.permissionStore.hasPermission(ResourceType.BrandingSettings, Action.Delete);
    },
    hasUnsavedChanges(): boolean {
      return (!!this.selectedFile && !!this.previewDataUrl) || this.removePending;
    },
    isTransparentPng(): boolean {
      const currentBase64 = this.currentLogoDataUrl?.split(",")[1] || this.authStore.logo;
      return currentBase64 === LogoConstants.TRANSPARENT_PNG_BASE64;
    },
    hasCurrentLogo(): boolean {
      return !!this.currentLogoDataUrl && !this.isTransparentPng;
    },
    displayLogoDataUrl(): string | null {
      if (this.removePending) return null;
      if (this.isTransparentPng) return null;
      if (this.isRemovingInEditMode) return null;
      return this.previewDataUrl || this.currentLogoDataUrl;
    },
    showRemoveX(): boolean {
      return !!this.selectedFile || this.hasCurrentLogo || this.removePending;
    },
    currentLogoFileName(): string {
      if (this.selectedFile) {
        return this.selectedFile.name;
      }
      return "";
    },
    acceptedTypes(): string {
      return LogoConstants.ALLOWED_TYPES.join(",");
    },
    hasValidationError(): boolean {
      return !!this.errorMessage;
    },
    hasValidChangesToSave(): boolean {
      return (!!this.selectedFile && !!this.previewDataUrl) || this.removePending || this.isRemovingInEditMode;
    },
    canSave(): boolean {
      return this.hasValidChangesToSave && !this.hasValidationError;
    },
    logoTooltipText(): string {
      return `Upload an SVG or PNG image no larger than ${LogoConstants.MAX_SIZE_KB} KB. For best results, use a transparent background without padding.`;
    },
    isRemovingInEditMode(): boolean {
      return this.hidePreviewInEdit && this.isBoxOpen;
    },
  },
  methods: {
    setCurrentLogoFromStore(): void {
      if (this.authStore.logo && this.authStore.logo !== LogoConstants.TRANSPARENT_PNG_BASE64) {
        this.currentLogoDataUrl = imageUtil.createBase64DataUrl(this.authStore.logo) || null;
      }
    },
    async loadCurrentLogo(): Promise<void> {
      try {
        const res = await brandingSettingsService.getLogo();
        const logo = res?.logo;
        if (!logo) return;
        const logoBase64 = logo.split(",")[1] ?? logo;
        if (logoBase64 !== LogoConstants.TRANSPARENT_PNG_BASE64) {
          this.currentLogoDataUrl = logo;
        } else {
          this.currentLogoDataUrl = null;
          this.authStore.logo = "";
        }
      } catch {
        // ignore if no logo exists
      }
    },
    clearSelection(): void {
      this.selectedFile = null;
      this.previewDataUrl = null;
      this.errorMessage = "";
      if (this.isBoxOpen) {
        this.hidePreviewInEdit = true;
      }
    },
    isValidFile(file: File | null): boolean | string {
      if (!file) return `Upload a file up to ${LogoConstants.MAX_SIZE_KB} KB`;
      return true;
    },

    onFileChangeEdit(file: File | File[] | null): void {
      const picked = Array.isArray(file) ? file[0] : file;
      if (!picked) {
        this.clearSelection();
        return;
      }

      this.errorMessage = "";

      if (!LogoConstants.ALLOWED_TYPES.includes(picked.type)) {
        const supportedTypes = LogoConstants.ALLOWED_TYPES.map((type) =>
          type === "image/png" ? "PNG" : type === "image/svg+xml" ? "SVG" : type,
        ).join(" and ");
        this.errorMessage = `Only ${supportedTypes} images are supported`;
        this.clearSelection();
        return;
      }

      const maxSizeBytes = LogoConstants.MAX_SIZE_KB * 1024;
      if (picked.size > maxSizeBytes) {
        this.errorMessage = `Selected file exceeds ${LogoConstants.MAX_SIZE_KB} KB. Please choose a smaller image.`;
        this.clearSelection();
        return;
      }

      this.selectedFile = picked;
      const reader = new FileReader();
      reader.onload = () => {
        this.previewDataUrl = String(reader.result || "");
      };
      reader.readAsDataURL(picked);
      this.removePending = false;
      this.hidePreviewInEdit = false;
    },
    async onSave(): Promise<void> {
      if (!this.canSave || this.hasValidationError) {
        return;
      }

      try {
        this.isSaving = true;
        if (this.removePending || this.isRemovingInEditMode) {
          await brandingSettingsService.deleteLogo();
          this.currentLogoDataUrl = null;
          this.authStore.logo = "";
          this.removePending = false;
          this.hidePreviewInEdit = false;
          this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
          this.isBoxOpen = false;
        } else if (this.selectedFile && this.previewDataUrl) {
          const logoData = this.previewDataUrl;
          await brandingSettingsService.updateLogo(logoData);
          this.currentLogoDataUrl = logoData;
          this.authStore.logo = logoData;
          this.clearSelection();
          this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
          this.isBoxOpen = false;
        }
      } catch (err: unknown) {
        this.$q.notify(alertUtil.getError("Failed to save changes"));
        await this.loadCurrentLogo();
      } finally {
        this.isSaving = false;
      }
    },
    async removeLogoQuick(): Promise<void> {
      try {
        this.isSaving = true;
        await brandingSettingsService.deleteLogo();
        this.currentLogoDataUrl = null;
        this.authStore.logo = "";
        this.removePending = false;
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
      } catch (err: unknown) {
        this.$q.notify(alertUtil.getError("Failed to remove logo"));
      } finally {
        this.isSaving = false;
      }
    },
    onCancel(): void {
      this.clearSelection();
      this.removePending = false;
      this.hidePreviewInEdit = false;
      this.isBoxOpen = false;
    },
    enterEditMode(): void {
      this.isBoxOpen = true;
      this.hidePreviewInEdit = false;
    },
    onAppendXClick(): void {
      if (this.selectedFile) {
        this.clearSelection();
      } else if (this.hasCurrentLogo && !this.removePending && !this.isBoxOpen) {
        this.removePending = true;
      } else if (this.hasCurrentLogo && this.isBoxOpen) {
        this.hidePreviewInEdit = true;
      }
    },
    onExpansionChanged(isExpanded: boolean): void {
      this.isExpanded = isExpanded;
    },
  },
});
</script>

<style lang="scss" scoped>
.preview-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 38px;
  width: 60px;
  margin-right: 8px;
}

.preview-wrapper.always {
  background: white;
}

.logo-preview {
  max-height: 34px;
  max-width: 58px;
  object-fit: contain;
}

.logo-read-box {
  background: white;
  border: 1px solid $black-12;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.logo-info {
  flex: 1;
  display: flex;
  align-items: center;
}

.logo-actions {
  display: flex;
  gap: 8px;

  :deep(.collapsed-state-icon .q-icon) {
    font-size: 20px !important;
  }
}

.logo-file-field {
  background: white !important;
}

.logo-file-field :deep(.q-field__control) {
  min-height: 56px;
  background: white;
}

.logo-file-field :deep(.q-field__inner) {
  width: 100% !important;
  min-width: 700px;
}

.logo-file-field :deep(.q-field__control),
.logo-file-field :deep(.q-field__control-container),
.logo-file-field :deep(.q-field__native) {
  width: 100%;
  background: white;
}

.remove-logo-btn {
  color: $black-54 !important;

  .q-icon {
    color: $black-54;
  }
}

.filename-display {
  display: flex;
  align-items: center;

  .filename-text {
    color: $black-87;
    font-size: 14px;
    font-weight: 400;
  }
}

.header-container {
  display: flex;
  align-items: center;
  width: 100%;
  position: relative;
}

.header-logo-preview {
  position: absolute;
  left: 370px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  height: 20px;

  .header-logo-image {
    max-height: 18px;
    max-width: 30px;
    object-fit: contain;
  }
}

.logo-file-field :deep(.q-field__label) {
  color: $black-54;
}

.logo-file-field :deep(.q-field__native) {
  background: white;
}

.custom-logo-label {
  font-weight: 400;
  font-size: 12px;
  line-height: 100%;
  color: $black-54;
}
</style>
