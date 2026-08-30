<template>
  <div class="system-notifications">
    <div class="row items-center q-mb-md">
      <span class="q-mr-md title">System notification</span>
    </div>
    <div class="row items-center q-mb-md">
      <span class="q-mr-md">{{ setting?.label }}</span>
      <runai-tooltip
        aid="tool-tip-system-notification"
        v-if="setting?.description"
        :tooltip-text="setting.description"
        width="400px"
        tooltip-position="right"
      />
    </div>
    <div v-if="showAddButton">
      <setting-readonly-input resource-type="Settings" :read-only="isReadOnly">
        <template v-slot:default="{ readonly }">
          <q-btn
            aid="add-system-notifications-btn"
            icon="ph-bold ph-plus"
            label="message"
            class="btn-icon-sm"
            flat
            color="primary"
            @click="openEditor"
            :disable="readonly"
          />
        </template>
      </setting-readonly-input>
    </div>
    <setting-wrapper-box
      v-if="editorOpen"
      primary-button-label="save & publish"
      :secondary-button-label="isReadOnly ? 'Close' : 'Cancel'"
      @primary-button-clicked="publishMessage"
      @secondary-button-clicked="discardMessage"
      has-footer
      is-wide-footer-buttons
      :validation-message="validationMessage"
      :show-primary="!isReadOnly"
    >
      <setting-readonly-input resource-type="Settings" :read-only="isReadOnly">
        <template v-slot:default="{ readonly }">
          <runai-text-editor v-model="message" class="col-12" :disable="readonly" />
        </template>
      </setting-readonly-input>
      <div class="row justify-between items-center col-12 q-mt-lg">
        <setting-readonly-input resource-type="Settings" :read-only="isReadOnly">
          <template v-slot:default="{ readonly }">
            <q-checkbox
              aid="enable-dismiss-checkbox"
              :label="checkboxLabel"
              v-model="enableDismiss"
              dense
              :disable="readonly"
            />
          </template>
        </setting-readonly-input>
      </div>
    </setting-wrapper-box>
    <div
      class="saved-message row justify-between no-wrap items-center q-pa-sm q-mt-md"
      v-if="!editorOpen && savedMessage"
    >
      <span class="ellipsis">{{ savedMessage }}</span>
      <div class="buttons row no-wrap q-gutter-sm">
        <q-btn aid="edit-message-btn" :icon="readEditIcon" class="q-pa-sm icon" flat size="sm" @click="openEditor">
          <q-tooltip>{{ readEditMessage }}</q-tooltip>
        </q-btn>
        <q-btn
          aid="delete-message-btn"
          icon="ph-fill ph-trash"
          class="q-pa-sm icon"
          flat
          size="sm"
          @click="deleteMessage"
          v-if="!isReadOnly"
        >
          <q-tooltip>Delete to unpublish the message</q-tooltip>
        </q-btn>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Models
import type { IAdminMessageSetting } from "@/models/setting.model";
import { STORAGE_KEYS } from "@/models/storage.model";
// Components
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiTextEditor } from "@/components/common/runai-text-editor";
// Stores
import { useSettingStore } from "@/stores/setting.store";
// Utils
import { alertUtil } from "@/utils/alert.util";
import { storageUtil } from "@/utils/storage.util";
import { HTMLUtil } from "@/utils/html.util/html.util";
import SettingWrapperBox from "@/components/settings/setting-wrapper-box/setting-wrapper-box.vue";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "system-notifications",
  components: {
    SettingReadonlyInput,
    SettingWrapperBox,
    RunaiTooltip,
    RunaiTextEditor,
  },
  props: {
    setting: {
      type: Object as PropType<IAdminMessageSetting | undefined>,
      required: false,
      default: undefined,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      message: "" as string,
      enableDismiss: false as boolean,
      editorOpen: false as boolean,
      savingMessage: false as boolean,
    };
  },
  computed: {
    checkboxLabel(): string {
      return `Display "Don't show this again" checkbox on message to users`;
    },
    savedMessage(): string {
      return this.setting?.value?.msg ? HTMLUtil.HTMLToPlainText(this.setting.value?.msg) : "";
    },
    showAddButton(): boolean {
      return !this.savedMessage && !this.editorOpen;
    },
    hasChanges(): boolean {
      return this.message !== this.setting?.value?.msg || this.enableDismiss !== this.setting.value?.enableDismiss;
    },
    validationMessage(): string {
      return this.hasChanges ? "Unsaved changes" : "";
    },
    isReadOnly(): boolean {
      return !usePermissionStore().hasPermission(ResourceType.Settings, Action.Update);
    },
    readEditIcon(): string {
      return this.isReadOnly ? "ph-fill ph-eye" : "ph-bold ph-pencil-simple";
    },
    readEditMessage(): string {
      return this.isReadOnly ? "View Message" : "Edit Message";
    },
  },
  methods: {
    syncLocalState(): void {
      const adminMessage = this.settingStore.adminMessage;
      this.message = adminMessage?.msg || "";
      this.enableDismiss = adminMessage?.enableDismiss || false;
    },
    openEditor(): void {
      this.syncLocalState();
      this.editorOpen = true;
    },
    discardMessage(): void {
      this.syncLocalState();
      this.editorOpen = false;
    },
    async publishMessage(): Promise<void> {
      try {
        this.savingMessage = true;
        if (this.hasChanges) {
          await this.settingStore.updateAdminMessage(this.message, this.enableDismiss);
          storageUtil.save(STORAGE_KEYS.ADMIN_MESSAGE_ID, this.settingStore.adminMessage?.id || undefined);
        }
        this.editorOpen = false;
        const successToast = this.message ? "Message published" : "Message unpublished";
        this.$q.notify(alertUtil.getSuccess(successToast));
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError("Failed to update message"));
        console.error(e);
      } finally {
        this.savingMessage = false;
      }
    },
    async deleteMessage(): Promise<void> {
      this.message = "";
      this.enableDismiss = false;
      await this.publishMessage();
    },
  },
  watch: {
    "settingStore.adminMessage": {
      handler() {
        if (!this.editorOpen || !this.message) {
          this.syncLocalState();
        }
      },
      deep: true,
      immediate: true,
    },
  },
});
</script>

<style scoped lang="scss">
.system-notifications {
  .icon {
    color: $black-54;

    :deep(.q-icon) {
      font-size: 20px;
    }
  }
  .title {
    font-size: 16px;
  }
  .saved-message {
    border: 1px solid $black-12;
    .q-icon {
      color: $black-54;
    }
  }
  .subheader {
    max-width: 350px;
  }
}
</style>
