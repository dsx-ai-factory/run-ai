<template>
  <runai-expansion-item class="settings-container" label="Workloads" :subheader="sectionSubHeader">
    <template v-if="canViewSettingsSection">
      <template v-for="setting in workloadSettingsValues" :key="setting.key">
        <template
          v-if="setting.key === SettingKeys.YamlSubmission && showYamlSubmissionSetting && yamlSubmissionSetting"
        >
          <yaml-submission-setting-section
            :yaml-submission-setting="yamlSubmissionSetting"
            class="q-mb-md"
            @update-value="onUpdateYamlSubmissionSetting"
          />
          <q-separator class="q-mb-md hide-if-last-child" />
        </template>
        <template
          v-else-if="
            setting.key === SettingKeys.ApplicationSubmission &&
            showApplicationSubmissionSetting &&
            applicationSubmissionSetting
          "
        >
          <application-submission-setting-section
            :application-submission-setting="applicationSubmissionSetting"
            class="q-mb-md"
            @update-value="onUpdateApplicationSubmissionSetting"
          />
          <q-separator class="q-mb-md hide-if-last-child" />
        </template>
        <template v-else-if="setting.key && shouldShowSetting(setting.key)">
          <toggle-setting class="q-mb-md" :setting="setting" @update-setting="onUpdateSetting" />
          <q-separator class="q-mb-md hide-if-last-child" />
        </template>
      </template>
    </template>
    <template v-if="showNgcPrivateRegistries">
      <toggle-setting
        v-if="ngcIntegrationSetting?.key"
        class="q-mb-md"
        :setting="ngcIntegrationSetting"
        resource-type="Registries"
        @update-setting="onUpdateSetting"
      />
      <q-separator class="q-mb-md hide-if-last-child" />
      <ngc-private-registry-section v-if="isNgcIntegrationEnabled" />
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// components
import ToggleSetting from "@/components/settings/toggle-setting/toggle-setting.vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { NgcPrivateRegistrySection } from "./ngc-private-registry-section";
import { YamlSubmissionSettingSection } from "./yaml-submission-setting";
import { ApplicationSubmissionSettingSection } from "./application-submission-setting";
// stores
import { useSettingStore } from "@/stores/setting.store";

// models
import type { IWorkloadsSectionSettings } from "@/models/setting.model";
import { SettingKeys, settingUpdateSuccessMessage } from "@/models/setting.model";
import { alertUtil } from "@/utils/alert.util";
import { settingsUtil } from "@/utils/settings.util";
import { stringUtil } from "@/utils/string.util";
import type { Setting } from "@/models/setting.model";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { useClusterStore } from "@/stores/cluster.store";
import {
  MIN_CLUSTER_VERSION_FOR_NGC_API_KEY,
  MIN_CLUSTER_VERSION_FOR_YAML_SUBMISSION,
  MIN_CLUSTER_VERSION_FOR_AI_APPLICATIONS,
} from "@/common/version.constant";
import type { YamlSubmissionSetting } from "@/swagger-models/workloads-manager-client";
import type { IApplicationSubmissionSetting } from "@/services/cluster/workload.service/workload.service";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { EFeatureToggle } from "@/services/infra/unleash.service/unleash.model";

export default defineComponent({
  name: "workloads-section",
  components: {
    ToggleSetting,
    RunaiExpansionItem,
    NgcPrivateRegistrySection,
    YamlSubmissionSettingSection,
    ApplicationSubmissionSettingSection,
  },
  props: {
    workloadSectionSettings: {
      type: Object as PropType<IWorkloadsSectionSettings>,
      required: true,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      permissionStore: usePermissionStore(),
      clusterStore: useClusterStore(),
      showPerformanceModal: false as boolean,
      SettingKeys,
    };
  },
  computed: {
    yamlSubmissionSetting(): YamlSubmissionSetting | null {
      return this.settingStore.yamlSubmissionSetting;
    },
    applicationSubmissionSetting(): IApplicationSubmissionSetting | null {
      return this.settingStore.applicationSubmissionSetting;
    },
    showApplicationSubmissionSetting(): boolean {
      return (
        unleashService.isFeatureEnabled(EFeatureToggle.ShowAIApplications) &&
        this.clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_AI_APPLICATIONS)
      );
    },
    canViewSettingsSection(): boolean {
      return this.permissionStore.hasSomePermittedActions(ResourceType.Settings, [
        Action.Read,
        Action.Create,
        Action.Update,
        Action.Delete,
      ]);
    },
    showNgcPrivateRegistries(): boolean {
      return (
        this.clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_NGC_API_KEY) &&
        this.permissionStore.hasPermission(ResourceType.Registries, Action.Read)
      );
    },
    ngcIntegrationSetting(): Setting | undefined {
      return this.workloadSectionSettings.ngcIntegration;
    },
    isNgcIntegrationEnabled(): boolean {
      return stringUtil.toBoolean(this.ngcIntegrationSetting?.value);
    },
    showYamlSubmissionSetting(): boolean {
      return this.clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_YAML_SUBMISSION);
    },
    sectionSubHeader(): string {
      if (!this.canViewSettingsSection) return "";
      const settings = { ...this.workloadSectionSettings };
      const isModelEnabled = Boolean(settings.models?.value);
      if (!isModelEnabled) {
        settings.allowGPT2Model = {};
      }
      if (!this.isHybridSubmissionEnabled) {
        settings.dataVolumes = {};
        settings.templatesV2 = {};
      }
      if (this.showYamlSubmissionSetting && this.yamlSubmissionSetting) {
        settings.yamlSubmission = { value: String(Boolean(this.yamlSubmissionSetting.enabled)) };
      }
      if (this.showApplicationSubmissionSetting && this.applicationSubmissionSetting) {
        settings.applicationSubmission = { value: String(Boolean(this.applicationSubmissionSetting.enabled)) };
      }
      return settingsUtil.getToggleSectionCollapsedTitle(settings);
    },
    workloadSettings(): Array<Setting> {
      return Object.values(this.workloadSectionSettings).filter((setting) => !!setting);
    },
    workloadSettingsValues(): Setting[] {
      return Object.values(this.workloadSettings).filter(
        (iWorkloadSetting: Setting) => iWorkloadSetting || iWorkloadSetting === "",
      );
    },
    isHybridSubmissionEnabled(): boolean {
      return stringUtil.toBoolean(this.workloadSectionSettings.hybridSubmission?.value);
    },
  },
  methods: {
    async onUpdateYamlSubmissionSetting(value: boolean): Promise<void> {
      if (!this.yamlSubmissionSetting) return;
      try {
        await this.settingStore.updateYamlSubmissionSetting(value);
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
      } catch (error) {
        this.$q.notify(settingsUtil.getSettingErrorMessage());
      }
    },
    async onUpdateApplicationSubmissionSetting(value: boolean): Promise<void> {
      if (!this.applicationSubmissionSetting) return;
      try {
        await this.settingStore.updateApplicationSubmissionSetting(value);
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
      } catch (error) {
        this.$q.notify(settingsUtil.getSettingErrorMessage());
      }
    },
    shouldShowSetting(key: string): boolean {
      if (key === SettingKeys.AllowGPT2Model) {
        return !!this.workloadSectionSettings.models?.value;
      }

      if (key === SettingKeys.EnableDataVolumes) {
        return this.isHybridSubmissionEnabled;
      }

      if (key === SettingKeys.EnableTemplatesV2) {
        return this.isHybridSubmissionEnabled;
      }

      if (key === SettingKeys.YamlSubmission) {
        return false;
      }

      if (key === SettingKeys.ApplicationSubmission) {
        return false;
      }

      if (key === SettingKeys.EnableNgcIntegration) {
        return false;
      }

      return true;
    },
    async saveSettings(key: SettingKeys, value: string | boolean | number): Promise<void> {
      try {
        await this.settingStore.updateSetting(key, value);
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
      } catch (error) {
        this.$q.notify(settingsUtil.getSettingErrorMessage());
      }
    },
    async onUpdateSetting(key: SettingKeys, value: boolean): Promise<void> {
      if (key === SettingKeys.EnableHybridSubmission && value === false) {
        try {
          await Promise.all([
            this.settingStore.refreshLocalSetting(SettingKeys.EnableTemplatesV2),
            this.settingStore.refreshLocalSetting(SettingKeys.EnableDataVolumes),
          ]);
        } catch (err) {
          console.error(err);
        }
      }
    },
  },
});
</script>
