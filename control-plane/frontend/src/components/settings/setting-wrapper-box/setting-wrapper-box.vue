<template>
  <div class="settings-box">
    <div v-if="!isBoxOpen && openerButtonLabel">
      <setting-readonly-input :resource-type="resourceType">
        <template v-slot:default="{ readonly }">
          <runai-btn
            icon="ph-bold ph-plus"
            :label="openerButtonLabel"
            flat
            color="primary"
            class="btn-icon-sm"
            @click="isBoxOpen = true"
            :aid="aidOpenerButton"
            :disable="readonly"
          />
        </template>
      </setting-readonly-input>
    </div>
    <div v-else>
      <div class="settings-box-header row" v-if="hasHeader">
        <div class="col-9 justify-start q-pt-xs">{{ settingBoxMessages?.headerTitle || "" }}</div>
        <div class="col-3 row no-wrap q-gutter-sm justify-end" v-if="hasHeaderButtons">
          <runai-btn
            :aid="`edit-${settingBoxMessages?.headerTitle}`"
            icon="ph-bold ph-pencil-simple"
            class="q-pa-sm icon"
            flat
            size="sm"
            round
            @click="editClicked"
            :tooltip="settingBoxMessages?.editToolTip || ''"
          />
          <runai-btn
            v-if="!readMode"
            :aid="`remove-${settingBoxMessages?.headerTitle}`"
            :disable="removeDisabled"
            icon="ph-bold ph-x"
            class="q-pa-sm icon"
            flat
            size="sm"
            rounded
            @click="isDeleteModalOpen = true"
            :tooltip="removeToolTip"
          />
        </div>
      </div>
      <div class="settings-box-body" :class="{ 'settings-box-body-no-footer': !hasFooter }">
        <slot></slot>
      </div>
      <div class="settings-box-footer row" v-if="hasFooter">
        <div
          class="justify-start validation-message"
          :class="{
            'col-10': !showPrimary,
            'col-9': !isWideFooterButtons && showPrimary,
            'col-8': isWideFooterButtons && showPrimary,
          }"
        >
          {{ validationMessage }}
        </div>
        <div
          class="justify-right no-wrap q-gutter-sm"
          :class="{
            'q-pl-lg': !showPrimary,
            'col-2': !showPrimary,
            'col-3': !isWideFooterButtons && showPrimary,
            'col-4': isWideFooterButtons && showPrimary,
          }"
        >
          <runai-btn
            :label="secondaryButtonLabel"
            flat
            color="primary"
            @click="SecondaryButtonClicked"
            aid="secondery-button"
          />
          <runai-btn
            v-if="showPrimary"
            :label="primaryButtonLabel"
            flat
            color="primary"
            @click="primaryButtonClicked"
            aid="primary-button"
            :disable="primaryDisabled"
          />
        </div>
      </div>
    </div>
  </div>
  <setting-delete-modal
    v-if="isDeleteModalOpen && settingBoxMessages"
    :delete-confirm-message="settingBoxMessages.confirmRemoveTitle"
    :delete-confirm-message-subtitle="settingBoxMessages.confirmRemoveSubTitle"
    @cancel="isDeleteModalOpen = false"
    @delete="confirmRemoveClicked"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import SettingDeleteModal from "@/components/settings/setting-delete-modal/setting-delete-modal.vue";
import type { IWrapperBoxSettingMessages } from "@/models/setting.model";
import { replaceAll } from "@/utils/string.util/string.util";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "setting-wrapper-box",
  components: { SettingReadonlyInput, SettingDeleteModal, RunaiBtn },
  emits: ["primary-button-clicked", "secondary-button-clicked", "edit-clicked", "removed-clicked"],
  props: {
    settingBoxMessages: {
      type: Object as PropType<IWrapperBoxSettingMessages>,
      required: false,
    },
    openerButtonLabel: {
      type: String as PropType<string>,
    },
    validationMessage: {
      type: String as PropType<string>,
    },
    primaryButtonLabel: {
      type: String as PropType<string>,
      default: "Save",
    },
    secondaryButtonLabel: {
      type: String as PropType<string>,
      default: "Cancel",
    },
    isOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasFooter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasHeader: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasHeaderButtons: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    removeDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isWideFooterButtons: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showPrimary: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    primaryDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    readMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    resourceType: {
      type: String as PropType<keyof typeof ResourceType>,
      default: ResourceType.Settings,
    },
  },
  data() {
    return {
      isDeleteModalOpen: false as boolean,
      isBoxOpen: this.isOpen as boolean,
    };
  },
  created() {
    this.isBoxOpen = this.isOpen;
  },
  computed: {
    removeToolTip(): string {
      if (this.removeDisabled) {
        return this.settingBoxMessages?.removeDisable || "";
      }
      return this.settingBoxMessages?.removeToolTip || "";
    },
    aidOpenerButton(): string {
      const aidLabel = replaceAll(replaceAll(this.openerButtonLabel, " ", "-"), "+", "").toLowerCase();
      return `opener-button${aidLabel}`;
    },
  },
  methods: {
    primaryButtonClicked(): void {
      this.$emit("primary-button-clicked");
    },
    SecondaryButtonClicked(): void {
      this.isBoxOpen = false;
      this.$emit("secondary-button-clicked");
    },
    editClicked(): void {
      this.$emit("edit-clicked");
    },
    confirmRemoveClicked(): void {
      this.isDeleteModalOpen = false;
      this.isBoxOpen = false;
      this.$emit("removed-clicked");
    },
  },
});
</script>
<style lang="scss" scoped>
.settings-box {
  .settings-box-body,
  .settings-box-footer,
  .settings-box-header {
    border: 1px solid $black-12;
  }
  .settings-box-body {
    border-bottom: none;
    padding: 20px 20px 20px 20px;
  }
  .settings-box-footer {
    padding: 10px 0 10px 10px;
  }
  .settings-box-header {
    font-size: 16px;
    padding: 15px 0 15px 18px;
    border-bottom: none;
  }
  .settings-box-body-no-footer {
    border-bottom: 1px solid $black-12;
  }
  .validation-message {
    padding: 10px 0 0 10px;
    font-size: 12px;
    color: $negative;
  }
  .icon {
    color: $black-54;

    :deep(.q-icon) {
      font-size: 20px;
    }
  }
}
</style>
