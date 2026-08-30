<template>
  <setting-editable-field
    @edit="editClicked"
    @delete="deleteClicked"
    :delete-tool-tip="deleteToolTip"
    :edit-tool-tip="editToolTip"
    :label="readOnlyUrl"
    v-if="isReadMode"
    :delete-sub-title="deleteSubTitle"
    :resource-type="resourceType"
  />
  <setting-wrapper-box
    v-else
    :is-open="!isUrlEmpty"
    :opener-button-label="addButtonLabel"
    :validation-message="validationMessage"
    @primary-button-clicked="save"
    @secondary-button-clicked="cancel"
    :secondary-button-label="isReadOnly ? 'Close' : 'Cancel'"
    has-footer
    :show-primary="!isReadOnly"
    :resource-type="resourceType"
  >
    <div class="row col-12">
      <setting-readonly-input class="col-12" :resource-type="resourceType">
        <template v-slot:default="{ readonly }">
          <!-- eslint-disable vuejs-accessibility/no-autofocus -- this field only renders once the user leaves read mode (v-else on isReadMode, entered by clicking Edit), never on a bare page load -->
          <q-input
            class="col-12"
            :model-value="url"
            placeholder="Enter URL"
            debounce="300"
            @update:model-value="updateURL($event as string)"
            :rules="[isValidUrl]"
            no-error-icon
            stack-label
            :label="inputLabel"
            ref="urlInput"
            autofocus
            :disable="readonly"
          />
          <!-- eslint-enable vuejs-accessibility/no-autofocus -->
        </template>
      </setting-readonly-input>
    </div>
  </setting-wrapper-box>
</template>
<script lang="ts">
//models
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { SettingWrapperBox } from "@/components/settings/setting-wrapper-box";
import type { QInput } from "quasar";
import { SettingEditableField } from "@/components/settings/setting-editable-field";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType } from "@/swagger-models/authorization-client";
import { isValidUrl } from "@/common/form.validators";

export default defineComponent({
  name: "settings-url-field",
  components: { SettingReadonlyInput, SettingEditableField, SettingWrapperBox },
  emits: ["save"],
  props: {
    prefixInReadOnlyMode: {
      type: String as PropType<string>,
      default: "",
    },
    serverUrl: {
      type: String as PropType<string>,
      default: "",
    },
    deleteToolTip: {
      type: String as PropType<string>,
      default: "",
    },
    editToolTip: {
      type: String as PropType<string>,
      default: "",
    },
    addButtonLabel: {
      type: String as PropType<string>,
      default: "",
    },
    label: {
      type: String as PropType<string>,
      default: "",
    },
    inputLabel: {
      type: String as PropType<string>,
      default: "",
    },
    deleteSubTitle: {
      type: String as PropType<string>,
      default: "",
    },
    resourceType: {
      type: String as PropType<keyof typeof ResourceType>,
      default: ResourceType.Settings,
    },
  },
  data() {
    return {
      url: "" as string,
      isEditMode: false as boolean,
    };
  },
  computed: {
    validationMessage(): string {
      return this.serverUrl === this.url || (!this.serverUrl && !this.url) ? "" : "Unsaved changes";
    },
    isReadMode(): boolean {
      return !this.isEditMode && !this.isUrlEmpty;
    },
    isUrlEmpty(): boolean {
      return !this.serverUrl;
    },
    readOnlyUrl(): string {
      return this.prefixInReadOnlyMode + this.serverUrl;
    },
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType[this.resourceType]);
    },
  },
  methods: {
    isValidUrl(url: string | null): boolean | string {
      return (url && isValidUrl(String(url))) || "Enter a valid URL";
    },
    updateURL(url: string | null): void {
      this.url = String(url || "");
    },
    async save(): Promise<void> {
      const isValid = await (this.$refs.urlInput as QInput)?.validate();
      if (!isValid) return;
      this.$emit("save", this.url);
      this.isEditMode = false;
    },
    deleteClicked(): void {
      this.updateURL("");
      this.$emit("save", "");
    },
    editClicked(): void {
      this.isEditMode = true;
      this.url = this.serverUrl ?? "";
    },
    cancel(): void {
      this.isEditMode = false;
      this.url = this.serverUrl || "";
    },
  },
  watch: {
    serverUrl: {
      immediate: true,
      handler(newValue: string) {
        if (!this.isEditMode) {
          this.url = newValue ?? "";
        }
      },
    },
  },
});
</script>

<style scoped lang="scss">
.input-with-label-wrapper {
  width: 300px;
  .label {
    min-width: 40%;
  }
  .cost-input {
    width: 300px;
  }
}
</style>
