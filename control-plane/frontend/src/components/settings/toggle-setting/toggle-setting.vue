<template>
  <div class="toggle-setting row items-center" :aid="`${setting.key}-setting`">
    <div class="col-4 items-center">
      <div class="row items-center">
        <div class="col-10">
          <span class="q-mr-md text-subtitle1 subtitle" v-dompurify-html="setting?.label"></span>
        </div>
        <div class="col-2">
          <setting-readonly-input :resource-type="resourceType">
            <template v-slot:default="{ readonly }">
              <q-toggle
                :model-value="isEnabled"
                @update:model-value="updateSettings(setting.key, $event)"
                :disable="disabled || readonly"
                :aid="toggleAid"
              >
                <q-tooltip v-if="tooltipMessage && !isReadOnly" anchor="bottom end">{{ tooltipMessage }}</q-tooltip>
              </q-toggle>
            </template>
          </setting-readonly-input>
        </div>
      </div>
    </div>
    <q-separator vertical class="q-mx-lg" />
    <div class="col-5">
      <div class="description" v-dompurify-html="setting?.description"></div>
      <notification-alert-bar v-if="warningMessage" class="flex-1 q-mt-md" :variant="ENotificationBarVariant.Warning">
        {{ warningMessage }}
      </notification-alert-bar>
    </div>
    <div class="col-2 toggle-item" v-if="badgeLabel !== null">
      <runai-stage-badge display-tooltip :badge="badgeLabel" use-default-cursor />
    </div>
    <slot></slot>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";

// components
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";

//stores
import { useSettingStore } from "@/stores/setting.store";

// models
import { AlphaStage, EBadgeLabel, settingUpdateSuccessMessage } from "@/models/setting.model";
import type { Setting } from "@/models/setting.model";
import type { SettingKeys } from "@/models/setting.model";

// utils
import { settingsUtil } from "@/utils/settings.util";
import { alertUtil } from "@/utils/alert.util";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import RunaiStageBadge from "@/components/common/runai-stage-badge/runai-stage-badge.vue";
import { stringUtil } from "@/utils/string.util";

export default defineComponent({
  name: "toggle-setting",
  components: { RunaiStageBadge, SettingReadonlyInput, NotificationAlertBar },
  emits: ["update-setting"],
  props: {
    setting: {
      type: Object as PropType<Setting>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    warningMessage: {
      type: String,
      default: "",
    },
    emitOnly: {
      type: Boolean,
      default: false,
    },
    tooltipMessage: {
      type: String,
      default: null,
    },
    resourceType: {
      type: String as PropType<keyof typeof ResourceType>,
      default: "Settings",
    },
  },
  data: () => ({
    settingStore: useSettingStore(),
    ENotificationBarVariant,
  }),
  computed: {
    isEnabled(): boolean {
      return stringUtil.toBoolean(this.setting?.value);
    },
    Action() {
      return Action;
    },
    ResourceType() {
      return ResourceType;
    },
    badgeLabel(): EBadgeLabel | null {
      switch (this.setting?.stage) {
        case AlphaStage:
          return EBadgeLabel.Experimental;
        case EBadgeLabel.Beta:
          return EBadgeLabel.Beta;
        case EBadgeLabel.Legacy:
          return EBadgeLabel.Legacy;
        default:
          return null;
      }
    },
    toggleAid(): string {
      return `toggle-setting-${this.setting?.key}`;
    },
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType[this.resourceType]);
    },
  },
  methods: {
    async updateSettings(key: string | undefined, value: string | number | boolean | null): Promise<void> {
      if (value === null || key === undefined) return;
      if (!this.emitOnly) {
        try {
          await this.settingStore.updateSetting(key as SettingKeys, value);
          this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
        } catch (err) {
          this.$q.notify(settingsUtil.getSettingErrorMessage());
        }
      }
      this.$emit("update-setting", key, value);
    },
  },
});
</script>
<style lang="scss" scoped>
.toggle-item {
  text-align: center;
}
.description {
  font-size: 14px;
  font-style: italic;
  font-weight: 400;
  line-height: 1.2;
  color: $black-54;
}
.subtitle {
  line-height: 1.5;
}
</style>
