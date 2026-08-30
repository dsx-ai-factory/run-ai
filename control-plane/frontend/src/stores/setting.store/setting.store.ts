import { defineStore } from "pinia";
import { settingService } from "@/services/control-plane/setting.service/setting.service";
import { notificationsSettingsService } from "@/services/control-plane/notifications-settings.service/notifications-settings.service";
import { workloadsSettingsService } from "@/services/control-plane/workloads-settings.service/workloads-settings.service";
import { workloadsServiceSettingsService } from "@/services/control-plane/workloads-service-settings.service/workloads-service-settings.service";
import { catalogServiceSettingsService } from "@/services/control-plane/catalog-service-settings.service/catalog-service-settings.service";
import { datavolumesServiceSettingsService } from "@/services/control-plane/datavolumes-service-settings.service/datavolumes-service-settings.service";
import { clusterServiceSettingsService } from "@/services/control-plane/cluster-service-settings.service/cluster-service-settings.service";
import { orgUnitSettingsService } from "@/services/control-plane/org-unit-settings.service/org-unit-settings.service";
import { nodePoolSettingsService } from "@/services/control-plane/node-pool-settings.service/node-pool-settings.service";
import { SettingKeys } from "@/models/setting.model";
import type {
  TSettingValue,
  IAdminMessage,
  IEnableNIMSetting,
  IDeletedWorkloadsViewSetting,
  IAllowGPT2ModelSetting,
  IEnableModelCatalogSetting,
  IEnablePoliciesSetting,
  IEnableHybridSubmissionSetting,
  IEnableTemplatesV2Setting,
  IGpuOverProvisioningSetting,
  IEnableNgcIntegrationSetting,
  IEnableDataVolumesSetting,
  IEnableAdvancedGpuMetricsSetting,
  ILimitQuotaOverSubscriptionSetting,
  IProjectManualOverQuotaSetting,
  ICPUResourcesQuotaSetting,
} from "@/models/setting.model";
import type { Setting } from "@/models/setting.model";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import type { YamlSubmissionSetting } from "@/swagger-models/workloads-manager-client";
import type { IApplicationSubmissionSetting } from "@/services/cluster/workload.service/workload.service";
import { appearanceSettingsService } from "@/services/control-plane/appearance-settings.service/appearance-settings.service";
import type { GetBannerConfigValue } from "@/swagger-models/bff-service-client";
import type { IFlagSetting } from "@/models/setting.model";

interface SettingStoreState {
  settings: Map<string, Setting>;
  yamlSubmissionSetting: YamlSubmissionSetting | null;
  applicationSubmissionSetting: IApplicationSubmissionSetting | null;
  adminMessage: IAdminMessage | undefined;
  systemMessage: IAdminMessage | undefined;
  enableNIM: IEnableNIMSetting;
  deletedWorkloadsView: IDeletedWorkloadsViewSetting;
  allowGPT2Model: IAllowGPT2ModelSetting;
  enableModelCatalog: IEnableModelCatalogSetting;
  bannerConfig: GetBannerConfigValue | undefined;
  enablePolicies: IEnablePoliciesSetting;
  enableHybridSubmission: IEnableHybridSubmissionSetting;
  enableTemplatesV2: IEnableTemplatesV2Setting;
  gpuOverProvisioning: IGpuOverProvisioningSetting;
  enableNgcIntegration: IEnableNgcIntegrationSetting;
  enableDataVolumes: IEnableDataVolumesSetting;
  enableAdvancedGpuMetrics: IEnableAdvancedGpuMetricsSetting;
  limitQuotaOverSubscription: ILimitQuotaOverSubscriptionSetting;
  projectManualOverQuota: IProjectManualOverQuotaSetting;
  cpuResourcesQuota: ICPUResourcesQuotaSetting;
}

const bffSettingsState: Record<string, keyof SettingStoreState> = {
  [SettingKeys.EnableNIM]: "enableNIM",
  [SettingKeys.DeletedWorkloadsView]: "deletedWorkloadsView",
  [SettingKeys.AllowGPT2Model]: "allowGPT2Model",
  [SettingKeys.EnableModelCatalog]: "enableModelCatalog",
  [SettingKeys.EnablePolicies]: "enablePolicies",
  [SettingKeys.EnableHybridSubmission]: "enableHybridSubmission",
  [SettingKeys.EnableTemplatesV2]: "enableTemplatesV2",
  [SettingKeys.GpuOverProvisioning]: "gpuOverProvisioning",
  [SettingKeys.EnableNgcIntegration]: "enableNgcIntegration",
  [SettingKeys.EnableDataVolumes]: "enableDataVolumes",
  [SettingKeys.AdvancedGPUMetrics]: "enableAdvancedGpuMetrics",
  [SettingKeys.LimitQuotaOverSubscription]: "limitQuotaOverSubscription",
  [SettingKeys.ProjectManualOverQuota]: "projectManualOverQuota",
  [SettingKeys.CPUResourcesQuota]: "cpuResourcesQuota",
};

type BffSettingUpdater = (value: boolean) => Promise<void>;

const bffSettingsUpdaters: Record<string, (store: ReturnType<typeof useSettingStore>) => BffSettingUpdater> = {
  [SettingKeys.EnableNIM]: (store) => (value) => store.updateEnableNIM(value),
  [SettingKeys.AllowGPT2Model]: (store) => (value) => store.updateAllowGPT2Model(value),
  [SettingKeys.EnableModelCatalog]: (store) => (value) => store.updateEnableModelCatalog(value),
  [SettingKeys.EnablePolicies]: (store) => (value) => store.updateEnablePolicies(value),
  [SettingKeys.EnableHybridSubmission]: (store) => (value) => store.updateEnableHybridSubmission(value),
  [SettingKeys.EnableTemplatesV2]: (store) => (value) => store.updateEnableTemplatesV2(value),
  [SettingKeys.GpuOverProvisioning]: (store) => (value) => store.updateGpuOverProvisioning(value),
  [SettingKeys.EnableNgcIntegration]: (store) => (value) => store.updateEnableNgcIntegration(value),
  [SettingKeys.EnableDataVolumes]: (store) => (value) => store.updateEnableDataVolumes(value),
  [SettingKeys.AdvancedGPUMetrics]: (store) => (value) => store.updateEnableAdvancedGpuMetrics(value),
  [SettingKeys.LimitQuotaOverSubscription]: (store) => (value) => store.updateLimitQuotaOverSubscription(value),
  [SettingKeys.ProjectManualOverQuota]: (store) => (value) => store.updateProjectManualOverQuota(value),
  [SettingKeys.CPUResourcesQuota]: (store) => (value) => store.updateCpuResourcesQuota(value),
};

type MigratedSettingRefresher = (store: ReturnType<typeof useSettingStore>) => Promise<void>;

const migratedSettingsRefreshers: Record<string, MigratedSettingRefresher> = {
  [SettingKeys.EnablePolicies]: async (store) => {
    store.enablePolicies = await workloadsServiceSettingsService.getEnablePolicies();
  },
  [SettingKeys.EnableHybridSubmission]: async (store) => {
    store.enableHybridSubmission = await workloadsServiceSettingsService.getEnableHybridSubmission();
  },
  [SettingKeys.EnableTemplatesV2]: async (store) => {
    store.enableTemplatesV2 = await workloadsServiceSettingsService.getEnableTemplatesV2();
  },
  [SettingKeys.GpuOverProvisioning]: async (store) => {
    store.gpuOverProvisioning = await workloadsServiceSettingsService.getGpuOverProvisioning();
  },
  [SettingKeys.EnableNgcIntegration]: async (store) => {
    store.enableNgcIntegration = await catalogServiceSettingsService.getEnableNgcIntegration();
  },
  [SettingKeys.EnableDataVolumes]: async (store) => {
    store.enableDataVolumes = await datavolumesServiceSettingsService.getEnableDataVolumes();
  },
  [SettingKeys.AdvancedGPUMetrics]: async (store) => {
    store.enableAdvancedGpuMetrics = await clusterServiceSettingsService.getEnableAdvancedGpuMetrics();
  },
  [SettingKeys.LimitQuotaOverSubscription]: async (store) => {
    store.limitQuotaOverSubscription = await orgUnitSettingsService.getLimitQuotaOverSubscription();
  },
  [SettingKeys.ProjectManualOverQuota]: async (store) => {
    store.projectManualOverQuota = await nodePoolSettingsService.getProjectManualOverQuota();
  },
  [SettingKeys.CPUResourcesQuota]: async (store) => {
    store.cpuResourcesQuota = await nodePoolSettingsService.getCpuResourcesQuota();
  },
};

export const useSettingStore = defineStore("Setting", {
  state: (): SettingStoreState => ({
    settings: new Map(),
    yamlSubmissionSetting: null,
    applicationSubmissionSetting: null,
    adminMessage: undefined,
    systemMessage: undefined,
    enableNIM: { value: true },
    deletedWorkloadsView: { value: true },
    allowGPT2Model: { value: false },
    enableModelCatalog: { value: true },
    bannerConfig: undefined,
    enablePolicies: { value: false },
    enableHybridSubmission: { value: false },
    enableTemplatesV2: { value: false },
    gpuOverProvisioning: { value: false },
    enableNgcIntegration: { value: false },
    enableDataVolumes: { value: false },
    enableAdvancedGpuMetrics: { value: false },
    limitQuotaOverSubscription: { value: false },
    projectManualOverQuota: { value: false },
    cpuResourcesQuota: { value: false },
  }),
  getters: {
    settingList(): Array<Setting> {
      return Array.from(this.settings.values());
    },
    isCPUResourcesQuotaEnabled(): boolean {
      return this.cpuResourcesQuota?.value || false;
    },
    isLimitQuotaOverSubscriptionEnabled(): boolean {
      return this.limitQuotaOverSubscription?.value || false;
    },
    isOverQuotaPriorityEnabled(): boolean {
      return this.projectManualOverQuota?.value || false;
    },
    isAdvancedGpuMetricsEnabled(): boolean {
      return this.enableAdvancedGpuMetrics?.value || false;
    },
    isPolicyManagerEnabled(): boolean {
      return this.enablePolicies?.value || false;
    },
    isGPUOverProvisioningEnabled(): boolean {
      return this.gpuOverProvisioning?.value || false;
    },
    isAllowDeployingGPT2ModelEnabled(): boolean {
      return this.allowGPT2Model?.value || false;
    },

    isHybridSubmissionEnabled(): boolean {
      return this.enableHybridSubmission?.value || false;
    },
    isDataVolumeEnabled(): boolean {
      return this.enableDataVolumes?.value || false;
    },
    isNgcIntegrationEnabled(): boolean {
      if (!unleashService.showNgcIntegration()) return false;
      return this.enableNgcIntegration?.value || false;
    },
    isYamlSubmissionEnabled(): boolean {
      return this.yamlSubmissionSetting?.enabled || false;
    },
    isApplicationSubmissionEnabled(): boolean | null {
      return this.applicationSubmissionSetting?.enabled ?? null;
    },
  },
  actions: {
    async loadSystemMessage(): Promise<IAdminMessage | undefined> {
      try {
        this.systemMessage = await notificationsSettingsService.getSystemMessage();
      } catch (error) {
        console.warn("Failed to load system message:", error);
      }
      return this.systemMessage;
    },
    async loadAdminMessage(): Promise<IAdminMessage | undefined> {
      try {
        this.adminMessage = await notificationsSettingsService.getAdminMessage();
      } catch (error) {
        console.warn("Failed to load admin message:", error);
      }
      return this.adminMessage;
    },
    async updateAdminMessage(msg: string, enableDismiss: boolean): Promise<void> {
      await notificationsSettingsService.updateAdminMessage(msg, enableDismiss);
      await this.loadAdminMessage();
    },
    async updateEnableNIM(enabled: boolean): Promise<void> {
      const original = this.enableNIM;
      this.enableNIM = { ...this.enableNIM, value: enabled };
      try {
        await workloadsSettingsService.updateEnableNIM(enabled);
      } catch (err) {
        this.enableNIM = original;
        throw err;
      }
    },
    async updateAllowGPT2Model(enabled: boolean): Promise<void> {
      const original = this.allowGPT2Model;
      this.allowGPT2Model = { ...this.allowGPT2Model, value: enabled };
      try {
        await workloadsSettingsService.updateAllowGPT2Model(enabled);
      } catch (err) {
        this.allowGPT2Model = original;
        throw err;
      }
    },
    async updateEnableModelCatalog(enabled: boolean): Promise<void> {
      const original = this.enableModelCatalog;
      this.enableModelCatalog = { ...this.enableModelCatalog, value: enabled };
      try {
        await workloadsSettingsService.updateEnableModelCatalog(enabled);
      } catch (err) {
        this.enableModelCatalog = original;
        throw err;
      }
    },
    async loadWorkloadsSettings(): Promise<void> {
      try {
        const settings = await workloadsSettingsService.getWorkloadsSettings();
        this.enableNIM = settings.enableNIM;
        this.deletedWorkloadsView = settings.deletedWorkloadsView;
        this.allowGPT2Model = settings.allowGPT2Model;
        this.enableModelCatalog = settings.enableModelCatalog;
      } catch (error) {
        console.warn("Failed to load workloads settings:", error);
      }
    },
    async loadWorkloadsServiceSettings(): Promise<void> {
      try {
        const settings = await workloadsServiceSettingsService.getAll();
        this.enablePolicies = settings.enablePolicies;
        this.enableHybridSubmission = settings.enableHybridSubmission;
        this.enableTemplatesV2 = settings.enableTemplatesV2;
        this.gpuOverProvisioning = settings.gpuOverProvisioning;
      } catch (error) {
        console.warn("Failed to load workloads service settings:", error);
      }
    },
    async loadCatalogServiceSettings(): Promise<void> {
      try {
        const settings = await catalogServiceSettingsService.getAll();
        this.enableNgcIntegration = settings.enableNgcIntegration;
      } catch (error) {
        console.warn("Failed to load catalog service settings:", error);
      }
    },
    async loadDatavolumesServiceSettings(): Promise<void> {
      try {
        const settings = await datavolumesServiceSettingsService.getAll();
        this.enableDataVolumes = settings.enableDataVolumes;
      } catch (error) {
        console.warn("Failed to load datavolumes service settings:", error);
      }
    },
    async loadClusterServiceSettings(): Promise<void> {
      try {
        const settings = await clusterServiceSettingsService.getAll();
        this.enableAdvancedGpuMetrics = settings.enableAdvancedGpuMetrics;
      } catch (error) {
        console.warn("Failed to load cluster service settings:", error);
      }
    },
    async loadOrgUnitServiceSettings(): Promise<void> {
      try {
        const settings = await orgUnitSettingsService.getAll();
        this.limitQuotaOverSubscription = settings.limitQuotaOverSubscription;
      } catch (error) {
        console.warn("Failed to load org unit service settings:", error);
      }
    },
    async loadNodePoolServiceSettings(): Promise<void> {
      try {
        const settings = await nodePoolSettingsService.getAll();
        this.projectManualOverQuota = settings.projectManualOverQuota;
        this.cpuResourcesQuota = settings.cpuResourcesQuota;
      } catch (error) {
        console.warn("Failed to load node pool service settings:", error);
      }
    },
    async updateEnablePolicies(enabled: boolean): Promise<void> {
      const original = this.enablePolicies;
      this.enablePolicies = { ...this.enablePolicies, value: enabled };
      try {
        await workloadsServiceSettingsService.updateEnablePolicies(enabled);
      } catch (err) {
        this.enablePolicies = original;
        throw err;
      }
    },
    async updateEnableHybridSubmission(enabled: boolean): Promise<void> {
      const original = this.enableHybridSubmission;
      this.enableHybridSubmission = { ...this.enableHybridSubmission, value: enabled };
      try {
        await workloadsServiceSettingsService.updateEnableHybridSubmission(enabled);
      } catch (err) {
        this.enableHybridSubmission = original;
        throw err;
      }
    },
    async updateEnableTemplatesV2(enabled: boolean): Promise<void> {
      const original = this.enableTemplatesV2;
      this.enableTemplatesV2 = { ...this.enableTemplatesV2, value: enabled };
      try {
        await workloadsServiceSettingsService.updateEnableTemplatesV2(enabled);
      } catch (err) {
        this.enableTemplatesV2 = original;
        throw err;
      }
    },
    async updateGpuOverProvisioning(enabled: boolean): Promise<void> {
      const original = this.gpuOverProvisioning;
      this.gpuOverProvisioning = { ...this.gpuOverProvisioning, value: enabled };
      try {
        await workloadsServiceSettingsService.updateGpuOverProvisioning(enabled);
      } catch (err) {
        this.gpuOverProvisioning = original;
        throw err;
      }
    },
    async updateEnableNgcIntegration(enabled: boolean): Promise<void> {
      const original = this.enableNgcIntegration;
      this.enableNgcIntegration = { ...this.enableNgcIntegration, value: enabled };
      try {
        await catalogServiceSettingsService.updateEnableNgcIntegration(enabled);
      } catch (err) {
        this.enableNgcIntegration = original;
        throw err;
      }
    },
    async updateEnableDataVolumes(enabled: boolean): Promise<void> {
      const original = this.enableDataVolumes;
      this.enableDataVolumes = { ...this.enableDataVolumes, value: enabled };
      try {
        await datavolumesServiceSettingsService.updateEnableDataVolumes(enabled);
      } catch (err) {
        this.enableDataVolumes = original;
        throw err;
      }
    },
    async updateEnableAdvancedGpuMetrics(enabled: boolean): Promise<void> {
      const original = this.enableAdvancedGpuMetrics;
      this.enableAdvancedGpuMetrics = { ...this.enableAdvancedGpuMetrics, value: enabled };
      try {
        await clusterServiceSettingsService.updateEnableAdvancedGpuMetrics(enabled);
      } catch (err) {
        this.enableAdvancedGpuMetrics = original;
        throw err;
      }
    },
    async updateLimitQuotaOverSubscription(enabled: boolean): Promise<void> {
      const original = this.limitQuotaOverSubscription;
      this.limitQuotaOverSubscription = { ...this.limitQuotaOverSubscription, value: enabled };
      try {
        await orgUnitSettingsService.updateLimitQuotaOverSubscription(enabled);
      } catch (err) {
        this.limitQuotaOverSubscription = original;
        throw err;
      }
    },
    async updateProjectManualOverQuota(enabled: boolean): Promise<void> {
      const original = this.projectManualOverQuota;
      this.projectManualOverQuota = { ...this.projectManualOverQuota, value: enabled };
      try {
        await nodePoolSettingsService.updateProjectManualOverQuota(enabled);
      } catch (err) {
        this.projectManualOverQuota = original;
        throw err;
      }
    },
    async updateCpuResourcesQuota(enabled: boolean): Promise<void> {
      const original = this.cpuResourcesQuota;
      this.cpuResourcesQuota = { ...this.cpuResourcesQuota, value: enabled };
      try {
        await nodePoolSettingsService.updateCpuResourcesQuota(enabled);
      } catch (err) {
        this.cpuResourcesQuota = original;
        throw err;
      }
    },
    async loadBannerConfig(): Promise<void> {
      try {
        const response = await appearanceSettingsService.getBannerConfig();
        this.bannerConfig = response;
      } catch (error) {
        console.warn("Failed to load banner config:", error);
      }
    },
    async loadSetting(): Promise<Map<string, Setting>> {
      const [settings, yamlSubmissionSetting, applicationSubmissionSetting] = await Promise.all([
        settingService.get(),
        workloadService.getYamlSubmissionSetting(),
        workloadService.getApplicationSubmissionSetting(),
        this.loadAdminMessage(),
        this.loadSystemMessage(),
        this.loadWorkloadsSettings(),
        this.loadBannerConfig(),
        this.loadWorkloadsServiceSettings(),
        this.loadCatalogServiceSettings(),
        this.loadDatavolumesServiceSettings(),
        this.loadClusterServiceSettings(),
        this.loadOrgUnitServiceSettings(),
        this.loadNodePoolServiceSettings(),
      ]);
      this.settings = settings;
      this.yamlSubmissionSetting = yamlSubmissionSetting;
      this.applicationSubmissionSetting = applicationSubmissionSetting;
      return this.settings;
    },
    async loadYamlSubmissionSetting(): Promise<YamlSubmissionSetting> {
      return workloadService.getYamlSubmissionSetting();
    },
    async loadApplicationSubmissionSetting(): Promise<IApplicationSubmissionSetting> {
      const applicationSubmissionSetting = await workloadService.getApplicationSubmissionSetting();
      this.applicationSubmissionSetting = applicationSubmissionSetting;
      return applicationSubmissionSetting;
    },
    async updateSetting(key: SettingKeys, value: TSettingValue): Promise<Map<string, Setting>> {
      const bffUpdaterFactory = bffSettingsUpdaters[key];
      if (bffUpdaterFactory) {
        const boolValue = typeof value === "boolean" ? value : value === "true";
        await bffUpdaterFactory(this)(boolValue);
        return this.settings;
      }

      const originalValue = this.settings.get(key);
      this.settings.set(key, { ...originalValue, value: value as unknown as string });
      try {
        await settingService.update(key, value);
      } catch (err) {
        if (originalValue !== undefined) {
          this.settings.set(key, originalValue);
        } else {
          this.settings.delete(key);
        }
        throw err;
      }
      return this.settings;
    },
    async updateYamlSubmissionSetting(value: boolean): Promise<YamlSubmissionSetting | void> {
      if (!this.yamlSubmissionSetting) return;
      this.yamlSubmissionSetting.enabled = value;
      try {
        await workloadService.updateYamlSubmissionSetting(value);
      } catch (err) {
        this.yamlSubmissionSetting.enabled = !value;
        throw err;
      }
      return this.yamlSubmissionSetting;
    },
    async updateApplicationSubmissionSetting(value: boolean): Promise<IApplicationSubmissionSetting | void> {
      if (!this.applicationSubmissionSetting) return;
      const prev = this.applicationSubmissionSetting.enabled;
      this.applicationSubmissionSetting.enabled = value;
      try {
        await workloadService.updateApplicationSubmissionSetting(value);
      } catch (err) {
        this.applicationSubmissionSetting.enabled = prev;
        throw err;
      }
      return this.applicationSubmissionSetting;
    },
    isFeatureEnabled(featureFlagName: SettingKeys): boolean {
      const setting = this.settings.get(featureFlagName);
      if (setting !== undefined) {
        return Boolean(setting.value);
      }

      const cacheProp = bffSettingsState[featureFlagName];
      if (cacheProp) {
        const cached = this[cacheProp];
        if (cached && typeof cached === "object" && "value" in cached) {
          return Boolean((cached as { value: boolean }).value);
        }
      }

      return false;
    },
    getBffSettingAsOldFormat(key: string): Setting | undefined {
      const cacheProp = bffSettingsState[key];
      if (!cacheProp) {
        return undefined;
      }

      const cached = this[cacheProp] as IFlagSetting | undefined;
      if (!cached) {
        return undefined;
      }

      const setting: Setting = {
        key,
        value: cached.value ? "true" : "false",
        label: cached.displayName,
        description: cached.description,
      };
      if ("stage" in cached && cached.stage !== undefined) {
        setting.stage = cached.stage;
      }
      return setting;
    },
    async refreshLocalSetting(key: SettingKeys): Promise<void> {
      const refresher = migratedSettingsRefreshers[key];
      if (refresher) {
        await refresher(this);
        return;
      }
      const value = await settingService.getByKey(key);
      this.settings.set(key, { ...this.settings.get(key), value: value as unknown as string });
    },
  },
});
