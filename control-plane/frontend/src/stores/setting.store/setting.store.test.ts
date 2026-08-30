import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useSettingStore } from "./setting.store";
import { settingService } from "@/services/control-plane/setting.service/setting.service";
import { workloadsSettingsService } from "@/services/control-plane/workloads-settings.service/workloads-settings.service";
import { workloadsServiceSettingsService } from "@/services/control-plane/workloads-service-settings.service/workloads-service-settings.service";
import { catalogServiceSettingsService } from "@/services/control-plane/catalog-service-settings.service/catalog-service-settings.service";
import { datavolumesServiceSettingsService } from "@/services/control-plane/datavolumes-service-settings.service/datavolumes-service-settings.service";
import { notificationsSettingsService } from "@/services/control-plane/notifications-settings.service/notifications-settings.service";
import { SettingKeys } from "@/models/setting.model";
import type { Setting } from "@/models/setting.model";
import type { IEnableNIMSetting } from "@/models/setting.model";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

type TestSetting = Partial<Setting>;

vi.mock("@/services/control-plane/setting.service/setting.service", () => ({
  settingService: {
    get: vi.fn(),
    update: vi.fn(),
    getByKey: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/workloads-settings.service/workloads-settings.service", () => ({
  workloadsSettingsService: {
    getWorkloadsSettings: vi.fn(),
    updateEnableNIM: vi.fn(),
    updateAllowGPT2Model: vi.fn(),
    updateEnableModelCatalog: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/notifications-settings.service/notifications-settings.service", () => ({
  notificationsSettingsService: {
    getSystemMessage: vi.fn(),
    getAdminMessage: vi.fn(),
    updateAdminMessage: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/workloads-service-settings.service/workloads-service-settings.service", () => ({
  workloadsServiceSettingsService: {
    getAll: vi.fn(),
    getEnablePolicies: vi.fn(),
    updateEnablePolicies: vi.fn(),
    getEnableHybridSubmission: vi.fn(),
    updateEnableHybridSubmission: vi.fn(),
    getEnableTemplatesV2: vi.fn(),
    updateEnableTemplatesV2: vi.fn(),
    getGpuOverProvisioning: vi.fn(),
    updateGpuOverProvisioning: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/catalog-service-settings.service/catalog-service-settings.service", () => ({
  catalogServiceSettingsService: {
    getAll: vi.fn(),
    getEnableNgcIntegration: vi.fn(),
    updateEnableNgcIntegration: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/datavolumes-service-settings.service/datavolumes-service-settings.service", () => ({
  datavolumesServiceSettingsService: {
    getAll: vi.fn(),
    getEnableDataVolumes: vi.fn(),
    updateEnableDataVolumes: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/org-unit-settings.service/org-unit-settings.service", () => ({
  orgUnitSettingsService: {
    getAll: vi.fn(),
    getLimitQuotaOverSubscription: vi.fn(),
    updateLimitQuotaOverSubscription: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/node-pool-settings.service/node-pool-settings.service", () => ({
  nodePoolSettingsService: {
    getAll: vi.fn(),
    getProjectManualOverQuota: vi.fn(),
    getCpuResourcesQuota: vi.fn(),
    updateProjectManualOverQuota: vi.fn(),
    updateCpuResourcesQuota: vi.fn(),
  },
}));

vi.mock("@/services/infra/unleash.service/unleash.service", () => ({
  unleashService: {
    showHybridSubmission: vi.fn(() => true),
    showNgcIntegration: vi.fn(() => true),
  },
}));

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getYamlSubmissionSetting: vi.fn(),
    getApplicationSubmissionSetting: vi.fn(),
    updateApplicationSubmissionSetting: vi.fn(),
  },
}));

describe("setting.store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  describe("isFeatureEnabled", () => {
    describe("regular settings from Map", () => {
      it("should return true when setting value is boolean true", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: true,
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(true);
      });

      it("should return false when setting value is boolean false", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: false,
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(false);
      });

      it("should return true when setting value is number 1", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: 1,
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(true);
      });

      it("should return false when setting value is number 0", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: 0,
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(false);
      });

      it("should return false when setting value is empty string", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: "",
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(false);
      });

      it("should return false when setting value is null", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: null,
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(false);
      });

      it("should return false when setting value is undefined", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: undefined,
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(false);
      });

      it("should return false when setting does not exist in Map", () => {
        const store = useSettingStore();

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(false);
      });
    });

    describe("BFF settings (EnableNIM)", () => {
      it("should return true when enableNIM.value is true", () => {
        const store = useSettingStore();
        store.enableNIM = { value: true };

        expect(store.isFeatureEnabled(SettingKeys.EnableNIM)).toBe(true);
      });

      it("should return false when enableNIM.value is false", () => {
        const store = useSettingStore();
        store.enableNIM = { value: false };

        expect(store.isFeatureEnabled(SettingKeys.EnableNIM)).toBe(false);
      });

      it("should return false when enableNIM.value is not present", () => {
        const store = useSettingStore();
        store.enableNIM = {} as IEnableNIMSetting;

        expect(store.isFeatureEnabled(SettingKeys.EnableNIM)).toBe(false);
      });

      it("should prefer settings Map over BFF cache when both exist", () => {
        const store = useSettingStore();
        store.enableNIM = { value: true };
        const setting: TestSetting = {
          key: SettingKeys.EnableNIM,
          value: false,
          label: "Enable NIM",
        };
        store.settings.set(SettingKeys.EnableNIM, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.EnableNIM)).toBe(false);
      });

      it("should fall back to BFF cache when not in settings Map", () => {
        const store = useSettingStore();
        store.enableNIM = { value: true };

        expect(store.isFeatureEnabled(SettingKeys.EnableNIM)).toBe(true);
      });
    });

    describe("edge cases", () => {
      it("should handle undefined setting gracefully", () => {
        const store = useSettingStore();
        const setting: TestSetting = {
          key: SettingKeys.CPUResourcesQuota,
          value: undefined,
          label: "CPU Resources",
        };
        store.settings.set(SettingKeys.CPUResourcesQuota, setting as Setting);

        expect(store.isFeatureEnabled(SettingKeys.CPUResourcesQuota)).toBe(false);
      });

      it("should return false for non-existent feature flag", () => {
        const store = useSettingStore();

        expect(store.isFeatureEnabled("non.existent.key" as SettingKeys)).toBe(false);
      });
    });

    describe("computed getters for migrated node-pool settings", () => {
      it("isCPUResourcesQuotaEnabled reflects the migrated cpuResourcesQuota state", () => {
        const store = useSettingStore();
        store.cpuResourcesQuota = { value: true };

        expect(store.isCPUResourcesQuotaEnabled).toBe(true);
      });

      it("isCPUResourcesQuotaEnabled returns false when disabled", () => {
        const store = useSettingStore();
        store.cpuResourcesQuota = { value: false };

        expect(store.isCPUResourcesQuotaEnabled).toBe(false);
      });

      it("isLimitQuotaOverSubscriptionEnabled reflects the migrated limitQuotaOverSubscription state", () => {
        const store = useSettingStore();
        store.limitQuotaOverSubscription = { value: true };

        expect(store.isLimitQuotaOverSubscriptionEnabled).toBe(true);
      });

      it("isLimitQuotaOverSubscriptionEnabled returns false when disabled", () => {
        const store = useSettingStore();
        store.limitQuotaOverSubscription = { value: false };

        expect(store.isLimitQuotaOverSubscriptionEnabled).toBe(false);
      });

      it("isOverQuotaPriorityEnabled reflects the migrated projectManualOverQuota state", () => {
        const store = useSettingStore();
        store.projectManualOverQuota = { value: true };

        expect(store.isOverQuotaPriorityEnabled).toBe(true);
      });

      it("isOverQuotaPriorityEnabled returns false when disabled", () => {
        const store = useSettingStore();
        store.projectManualOverQuota = { value: false };

        expect(store.isOverQuotaPriorityEnabled).toBe(false);
      });
    });
  });

  describe("loadWorkloadsSettings", () => {
    it("should load workloads settings from service", async () => {
      const store = useSettingStore();
      const mockWorkloadsSettings = {
        enableNIM: {
          value: true,
          displayName: "Enable NIM",
          description: "Enable NVIDIA NIM services",
        },
        deletedWorkloadsView: {
          value: false,
          displayName: "Deleted Workloads View",
          description: "Show deleted workloads",
        },
        allowGPT2Model: {
          value: false,
          displayName: "Allow GPT-2 model",
          description: "Enable to allow GPT-2 model.",
        },
        enableModelCatalog: {
          value: true,
          displayName: "Hugging face models",
          description: "Enable instant deployment of Hugging Face models.",
        },
      };
      vi.mocked(workloadsSettingsService.getWorkloadsSettings).mockResolvedValue(mockWorkloadsSettings);

      await store.loadWorkloadsSettings();

      expect(workloadsSettingsService.getWorkloadsSettings).toHaveBeenCalled();
      expect(store.enableNIM).toEqual(mockWorkloadsSettings.enableNIM);
      expect(store.deletedWorkloadsView).toEqual(mockWorkloadsSettings.deletedWorkloadsView);
      expect(store.allowGPT2Model).toEqual(mockWorkloadsSettings.allowGPT2Model);
      expect(store.enableModelCatalog).toEqual(mockWorkloadsSettings.enableModelCatalog);
    });

    it("should handle error when loading workloads settings", async () => {
      const store = useSettingStore();
      const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      vi.mocked(workloadsSettingsService.getWorkloadsSettings).mockRejectedValue(new Error("Failed to load"));

      await store.loadWorkloadsSettings();

      expect(consoleWarnSpy).toHaveBeenCalledWith("Failed to load workloads settings:", expect.any(Error));
      consoleWarnSpy.mockRestore();
    });
  });

  describe("loadSystemMessage", () => {
    it("should load system message from service", async () => {
      const store = useSettingStore();
      const mockSystemMessage = {
        id: "msg-1",
        msg: "System announcement",
        createdAt: new Date(),
        enableDismiss: true,
      };
      vi.mocked(notificationsSettingsService.getSystemMessage).mockResolvedValue(mockSystemMessage);

      const result = await store.loadSystemMessage();

      expect(notificationsSettingsService.getSystemMessage).toHaveBeenCalled();
      expect(store.systemMessage).toEqual(mockSystemMessage);
      expect(result).toEqual(mockSystemMessage);
    });

    it("should handle error when loading system message", async () => {
      const store = useSettingStore();
      const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      vi.mocked(notificationsSettingsService.getSystemMessage).mockRejectedValue(new Error("Failed to load"));

      const result = await store.loadSystemMessage();

      expect(consoleWarnSpy).toHaveBeenCalledWith("Failed to load system message:", expect.any(Error));
      expect(store.systemMessage).toBeUndefined();
      expect(result).toBeUndefined();
      consoleWarnSpy.mockRestore();
    });
  });

  describe("updateEnableNIM", () => {
    it("should update enableNIM value", async () => {
      const store = useSettingStore();
      store.enableNIM = { value: false };
      vi.mocked(workloadsSettingsService.updateEnableNIM).mockResolvedValue();

      await store.updateEnableNIM(true);

      expect(workloadsSettingsService.updateEnableNIM).toHaveBeenCalledWith(true);
      expect(store.enableNIM.value).toBe(true);
    });

    it("should rollback on error", async () => {
      const store = useSettingStore();
      const original = { value: false, displayName: "Enable NIM" };
      store.enableNIM = { ...original };
      vi.mocked(workloadsSettingsService.updateEnableNIM).mockRejectedValue(new Error("Failed"));

      await expect(store.updateEnableNIM(true)).rejects.toThrow("Failed");

      expect(store.enableNIM).toEqual(original);
    });
  });

  describe("updateEnableModelCatalog", () => {
    it("should update enableModelCatalog value", async () => {
      const store = useSettingStore();
      store.enableModelCatalog = { value: false };
      vi.mocked(workloadsSettingsService.updateEnableModelCatalog).mockResolvedValue();

      await store.updateEnableModelCatalog(true);

      expect(workloadsSettingsService.updateEnableModelCatalog).toHaveBeenCalledWith(true);
      expect(store.enableModelCatalog.value).toBe(true);
    });

    it("should rollback on error", async () => {
      const store = useSettingStore();
      const original = { value: false, displayName: "Hugging face models" };
      store.enableModelCatalog = { ...original };
      vi.mocked(workloadsSettingsService.updateEnableModelCatalog).mockRejectedValue(new Error("Failed"));

      await expect(store.updateEnableModelCatalog(true)).rejects.toThrow("Failed");

      expect(store.enableModelCatalog).toEqual(original);
    });
  });

  describe("updateSetting", () => {
    it("should handle BFF settings update (EnableNIM)", async () => {
      const store = useSettingStore();
      store.enableNIM = { value: false };
      vi.mocked(workloadsSettingsService.updateEnableNIM).mockResolvedValue();

      await store.updateSetting(SettingKeys.EnableNIM, true);

      expect(workloadsSettingsService.updateEnableNIM).toHaveBeenCalledWith(true);
      expect(store.enableNIM.value).toBe(true);
    });

    it("should handle BFF settings update with string value", async () => {
      const store = useSettingStore();
      store.enableNIM = { value: false };
      vi.mocked(workloadsSettingsService.updateEnableNIM).mockResolvedValue();

      await store.updateSetting(SettingKeys.EnableNIM, "true");

      expect(workloadsSettingsService.updateEnableNIM).toHaveBeenCalledWith(true);
      expect(store.enableNIM.value).toBe(true);
    });

    it("should handle BFF settings update (AllowGPT2Model)", async () => {
      const store = useSettingStore();
      store.allowGPT2Model = { value: false };
      vi.mocked(workloadsSettingsService.updateAllowGPT2Model).mockResolvedValue();

      await store.updateSetting(SettingKeys.AllowGPT2Model, true);

      expect(workloadsSettingsService.updateAllowGPT2Model).toHaveBeenCalledWith(true);
      expect(store.allowGPT2Model.value).toBe(true);
    });

    it("should handle BFF settings update (EnableModelCatalog)", async () => {
      const store = useSettingStore();
      store.enableModelCatalog = { value: false };
      vi.mocked(workloadsSettingsService.updateEnableModelCatalog).mockResolvedValue();

      await store.updateSetting(SettingKeys.EnableModelCatalog, true);

      expect(workloadsSettingsService.updateEnableModelCatalog).toHaveBeenCalledWith(true);
      expect(store.enableModelCatalog.value).toBe(true);
    });

    it("should update dedicated state via BFF updater without touching Map", async () => {
      const store = useSettingStore();
      store.enablePolicies = { value: false };
      vi.mocked(workloadsServiceSettingsService.updateEnablePolicies).mockResolvedValue();

      await store.updateSetting(SettingKeys.EnablePolicies, true);

      expect(workloadsServiceSettingsService.updateEnablePolicies).toHaveBeenCalledWith(true);
      expect(store.enablePolicies.value).toBe(true);
      expect(store.settings.has(SettingKeys.EnablePolicies)).toBe(false);
    });

    it("should handle regular settings update", async () => {
      const store = useSettingStore();
      const originalSetting: TestSetting = {
        key: SettingKeys.AllowSwap,
        value: false,
        label: "Allow Swap",
      };
      store.settings.set(SettingKeys.AllowSwap, originalSetting as Setting);
      vi.mocked(settingService.update).mockResolvedValue({} as never);

      await store.updateSetting(SettingKeys.AllowSwap, true);

      expect(settingService.update).toHaveBeenCalledWith(SettingKeys.AllowSwap, true);
      expect(store.settings.get(SettingKeys.AllowSwap)?.value).toBe(true);
    });

    it("should rollback regular settings on error", async () => {
      const store = useSettingStore();
      const originalSetting: TestSetting = {
        key: SettingKeys.AllowSwap,
        value: false,
        label: "Allow Swap",
      };
      store.settings.set(SettingKeys.AllowSwap, originalSetting as Setting);
      vi.mocked(settingService.update).mockRejectedValue(new Error("Failed"));

      await expect(store.updateSetting(SettingKeys.AllowSwap, true)).rejects.toThrow("Failed");

      expect(store.settings.get(SettingKeys.AllowSwap)).toEqual(originalSetting);
    });
  });

  describe("getBffSettingAsOldFormat", () => {
    it("should convert enableNIM to old Setting format", () => {
      const store = useSettingStore();
      store.enableNIM = {
        value: true,
        displayName: "Enable NIM",
        description: "Enable NVIDIA NIM services",
      };

      const result = store.getBffSettingAsOldFormat(SettingKeys.EnableNIM);

      expect(result).toEqual({
        key: SettingKeys.EnableNIM,
        value: "true",
        label: "Enable NIM",
        description: "Enable NVIDIA NIM services",
      });
    });

    it("should convert false value to 'false' string", () => {
      const store = useSettingStore();
      store.enableNIM = { value: false };

      const result = store.getBffSettingAsOldFormat(SettingKeys.EnableNIM);

      expect(result?.value).toBe("false");
    });

    it("should return undefined for non-BFF setting", () => {
      const store = useSettingStore();

      const result = store.getBffSettingAsOldFormat(SettingKeys.AllowSwap);

      expect(result).toBeUndefined();
    });

    it("should return undefined when cached value is not set", () => {
      const store = useSettingStore();
      store.enableNIM = undefined as unknown as IEnableNIMSetting;

      const result = store.getBffSettingAsOldFormat(SettingKeys.EnableNIM);

      expect(result).toBeUndefined();
    });

    it("should convert migrated enablePolicies to old Setting format", () => {
      const store = useSettingStore();
      store.enablePolicies = {
        value: true,
        displayName: "Policies",
        description: "Enable workload policies",
      };

      const result = store.getBffSettingAsOldFormat(SettingKeys.EnablePolicies);

      expect(result).toEqual({
        key: SettingKeys.EnablePolicies,
        value: "true",
        label: "Policies",
        description: "Enable workload policies",
      });
    });

    it("should include stage for gpuOverProvisioning", () => {
      const store = useSettingStore();
      store.gpuOverProvisioning = {
        value: false,
        displayName: "GPU resource optimization",
        description: "Enable GPU optimization",
        stage: "Beta",
      };

      const result = store.getBffSettingAsOldFormat(SettingKeys.GpuOverProvisioning);

      expect(result?.stage).toBe("Beta");
      expect(result?.value).toBe("false");
    });
  });

  describe("loadWorkloadsServiceSettings", () => {
    it("should load settings from workloads service", async () => {
      const store = useSettingStore();
      const mockSettings = {
        enablePolicies: { value: true, displayName: "Policies", description: "desc" },
        enableHybridSubmission: { value: true, displayName: "Flexible submission", description: "desc" },
        enableTemplatesV2: { value: false, displayName: "Templates", description: "desc" },
        gpuOverProvisioning: { value: false, displayName: "GPU opt", description: "desc", stage: "Beta" },
      };
      vi.mocked(workloadsServiceSettingsService.getAll).mockResolvedValue(mockSettings);

      await store.loadWorkloadsServiceSettings();

      expect(workloadsServiceSettingsService.getAll).toHaveBeenCalled();
      expect(store.enablePolicies).toEqual(mockSettings.enablePolicies);
      expect(store.enableHybridSubmission).toEqual(mockSettings.enableHybridSubmission);
      expect(store.enableTemplatesV2).toEqual(mockSettings.enableTemplatesV2);
      expect(store.gpuOverProvisioning).toEqual(mockSettings.gpuOverProvisioning);
    });

    it("should handle error gracefully", async () => {
      const store = useSettingStore();
      const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      vi.mocked(workloadsServiceSettingsService.getAll).mockRejectedValue(new Error("Failed"));

      await store.loadWorkloadsServiceSettings();

      expect(consoleWarnSpy).toHaveBeenCalledWith("Failed to load workloads service settings:", expect.any(Error));
      consoleWarnSpy.mockRestore();
    });
  });

  describe("loadCatalogServiceSettings", () => {
    it("should load settings from catalog service", async () => {
      const store = useSettingStore();
      const mockSettings = {
        enableNgcIntegration: { value: true, displayName: "NGC catalog", description: "desc", stage: "Beta" },
      };
      vi.mocked(catalogServiceSettingsService.getAll).mockResolvedValue(mockSettings);

      await store.loadCatalogServiceSettings();

      expect(catalogServiceSettingsService.getAll).toHaveBeenCalled();
      expect(store.enableNgcIntegration).toEqual(mockSettings.enableNgcIntegration);
    });

    it("should handle error gracefully", async () => {
      const store = useSettingStore();
      const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      vi.mocked(catalogServiceSettingsService.getAll).mockRejectedValue(new Error("Failed"));

      await store.loadCatalogServiceSettings();

      expect(consoleWarnSpy).toHaveBeenCalledWith("Failed to load catalog service settings:", expect.any(Error));
      consoleWarnSpy.mockRestore();
    });
  });

  describe("loadDatavolumesServiceSettings", () => {
    it("should load settings from datavolumes service", async () => {
      const store = useSettingStore();
      const mockSettings = {
        enableDataVolumes: { value: true, displayName: "Data volumes", description: "desc" },
      };
      vi.mocked(datavolumesServiceSettingsService.getAll).mockResolvedValue(mockSettings);

      await store.loadDatavolumesServiceSettings();

      expect(datavolumesServiceSettingsService.getAll).toHaveBeenCalled();
      expect(store.enableDataVolumes).toEqual(mockSettings.enableDataVolumes);
    });

    it("should handle error gracefully", async () => {
      const store = useSettingStore();
      const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      vi.mocked(datavolumesServiceSettingsService.getAll).mockRejectedValue(new Error("Failed"));

      await store.loadDatavolumesServiceSettings();

      expect(consoleWarnSpy).toHaveBeenCalledWith("Failed to load datavolumes service settings:", expect.any(Error));
      consoleWarnSpy.mockRestore();
    });
  });

  describe("migrated settings update actions", () => {
    it("should update enablePolicies and call service", async () => {
      const store = useSettingStore();
      store.enablePolicies = { value: false };
      vi.mocked(workloadsServiceSettingsService.updateEnablePolicies).mockResolvedValue();

      await store.updateEnablePolicies(true);

      expect(workloadsServiceSettingsService.updateEnablePolicies).toHaveBeenCalledWith(true);
      expect(store.enablePolicies.value).toBe(true);
    });

    it("should rollback enablePolicies on error", async () => {
      const store = useSettingStore();
      const original = { value: false, displayName: "Policies" };
      store.enablePolicies = { ...original };
      vi.mocked(workloadsServiceSettingsService.updateEnablePolicies).mockRejectedValue(new Error("Failed"));

      await expect(store.updateEnablePolicies(true)).rejects.toThrow("Failed");
      expect(store.enablePolicies).toEqual(original);
    });

    it("should update enableHybridSubmission via updateSetting", async () => {
      const store = useSettingStore();
      store.enableHybridSubmission = { value: false };
      vi.mocked(workloadsServiceSettingsService.updateEnableHybridSubmission).mockResolvedValue();

      await store.updateSetting(SettingKeys.EnableHybridSubmission, true);

      expect(workloadsServiceSettingsService.updateEnableHybridSubmission).toHaveBeenCalledWith(true);
      expect(store.enableHybridSubmission.value).toBe(true);
    });

    it("should update enableNgcIntegration and call catalog service", async () => {
      const store = useSettingStore();
      store.enableNgcIntegration = { value: false };
      vi.mocked(catalogServiceSettingsService.updateEnableNgcIntegration).mockResolvedValue();

      await store.updateEnableNgcIntegration(true);

      expect(catalogServiceSettingsService.updateEnableNgcIntegration).toHaveBeenCalledWith(true);
      expect(store.enableNgcIntegration.value).toBe(true);
    });

    it("should rollback enableNgcIntegration on error", async () => {
      const store = useSettingStore();
      const original = { value: false, displayName: "NGC catalog" };
      store.enableNgcIntegration = { ...original };
      vi.mocked(catalogServiceSettingsService.updateEnableNgcIntegration).mockRejectedValue(new Error("Failed"));

      await expect(store.updateEnableNgcIntegration(true)).rejects.toThrow("Failed");
      expect(store.enableNgcIntegration).toEqual(original);
    });

    it("should update enableDataVolumes and call datavolumes service", async () => {
      const store = useSettingStore();
      store.enableDataVolumes = { value: false };
      vi.mocked(datavolumesServiceSettingsService.updateEnableDataVolumes).mockResolvedValue();

      await store.updateEnableDataVolumes(true);

      expect(datavolumesServiceSettingsService.updateEnableDataVolumes).toHaveBeenCalledWith(true);
      expect(store.enableDataVolumes.value).toBe(true);
    });

    it("should rollback enableDataVolumes on error", async () => {
      const store = useSettingStore();
      const original = { value: false, displayName: "Data volumes" };
      store.enableDataVolumes = { ...original };
      vi.mocked(datavolumesServiceSettingsService.updateEnableDataVolumes).mockRejectedValue(new Error("Failed"));

      await expect(store.updateEnableDataVolumes(true)).rejects.toThrow("Failed");
      expect(store.enableDataVolumes).toEqual(original);
    });
  });

  describe("refreshLocalSetting with migrated settings", () => {
    it("should refresh enableTemplatesV2 from workloads service", async () => {
      const store = useSettingStore();
      store.enableTemplatesV2 = { value: true };
      const refreshed = { value: false, displayName: "Templates", description: "desc" };
      vi.mocked(workloadsServiceSettingsService.getEnableTemplatesV2).mockResolvedValue(refreshed);

      await store.refreshLocalSetting(SettingKeys.EnableTemplatesV2);

      expect(workloadsServiceSettingsService.getEnableTemplatesV2).toHaveBeenCalled();
      expect(store.enableTemplatesV2).toEqual(refreshed);
    });

    it("should refresh enableDataVolumes from datavolumes service", async () => {
      const store = useSettingStore();
      store.enableDataVolumes = { value: true };
      const refreshed = { value: false, displayName: "Data volumes", description: "desc" };
      vi.mocked(datavolumesServiceSettingsService.getEnableDataVolumes).mockResolvedValue(refreshed);

      await store.refreshLocalSetting(SettingKeys.EnableDataVolumes);

      expect(datavolumesServiceSettingsService.getEnableDataVolumes).toHaveBeenCalled();
      expect(store.enableDataVolumes).toEqual(refreshed);
    });

    it("should fall back to settingService for non-migrated keys", async () => {
      const store = useSettingStore();
      vi.mocked(settingService.getByKey).mockResolvedValue(true);

      await store.refreshLocalSetting(SettingKeys.AllowSwap);

      expect(settingService.getByKey).toHaveBeenCalledWith(SettingKeys.AllowSwap);
    });
  });

  describe("migrated settings getters", () => {
    it("isPolicyManagerEnabled should read from dedicated state", () => {
      const store = useSettingStore();
      store.enablePolicies = { value: true };
      expect(store.isPolicyManagerEnabled).toBe(true);

      store.enablePolicies = { value: false };
      expect(store.isPolicyManagerEnabled).toBe(false);
    });

    it("isGPUOverProvisioningEnabled should read from dedicated state", () => {
      const store = useSettingStore();
      store.gpuOverProvisioning = { value: true };
      expect(store.isGPUOverProvisioningEnabled).toBe(true);

      store.gpuOverProvisioning = { value: false };
      expect(store.isGPUOverProvisioningEnabled).toBe(false);
    });

    it("isHybridSubmissionEnabled should read from dedicated state", () => {
      const store = useSettingStore();
      store.enableHybridSubmission = { value: true };
      expect(store.isHybridSubmissionEnabled).toBe(true);

      store.enableHybridSubmission = { value: false };
      expect(store.isHybridSubmissionEnabled).toBe(false);
    });

    it("isDataVolumeEnabled should read from dedicated state", () => {
      const store = useSettingStore();
      store.enableDataVolumes = { value: true };
      expect(store.isDataVolumeEnabled).toBe(true);

      store.enableDataVolumes = { value: false };
      expect(store.isDataVolumeEnabled).toBe(false);
    });

    it("isNgcIntegrationEnabled should read from dedicated state and check unleash", () => {
      const store = useSettingStore();
      store.enableNgcIntegration = { value: true };
      expect(store.isNgcIntegrationEnabled).toBe(true);

      store.enableNgcIntegration = { value: false };
      expect(store.isNgcIntegrationEnabled).toBe(false);

      vi.mocked(unleashService.showNgcIntegration).mockReturnValue(false);
      store.enableNgcIntegration = { value: true };
      expect(store.isNgcIntegrationEnabled).toBe(false);
      vi.mocked(unleashService.showNgcIntegration).mockReturnValue(true);
    });

    it("isFeatureEnabled should fall back to migrated settings", () => {
      const store = useSettingStore();
      store.enablePolicies = { value: true };

      expect(store.isFeatureEnabled(SettingKeys.EnablePolicies)).toBe(true);
    });

    it("isApplicationSubmissionEnabled should return true when enabled", () => {
      const store = useSettingStore();
      store.applicationSubmissionSetting = { enabled: true };
      expect(store.isApplicationSubmissionEnabled).toBe(true);
    });

    it("isApplicationSubmissionEnabled should return false when disabled", () => {
      const store = useSettingStore();
      store.applicationSubmissionSetting = { enabled: false };
      expect(store.isApplicationSubmissionEnabled).toBe(false);
    });

    it("isApplicationSubmissionEnabled should return null when not loaded", () => {
      const store = useSettingStore();
      store.applicationSubmissionSetting = null;
      expect(store.isApplicationSubmissionEnabled).toBeNull();
    });
  });

  describe("loadApplicationSubmissionSetting", () => {
    it("should load applicationSubmissionSetting and cache it", async () => {
      const store = useSettingStore();
      const applicationSubmissionSetting = { enabled: true };
      vi.mocked(workloadService.getApplicationSubmissionSetting).mockResolvedValue(applicationSubmissionSetting);

      const result = await store.loadApplicationSubmissionSetting();

      expect(workloadService.getApplicationSubmissionSetting).toHaveBeenCalled();
      expect(result).toEqual(applicationSubmissionSetting);
      expect(store.applicationSubmissionSetting).toEqual(applicationSubmissionSetting);
    });
  });

  describe("updateApplicationSubmissionSetting", () => {
    it("should update applicationSubmissionSetting and call service", async () => {
      const store = useSettingStore();
      store.applicationSubmissionSetting = { enabled: true };
      vi.mocked(workloadService.updateApplicationSubmissionSetting).mockResolvedValue();

      await store.updateApplicationSubmissionSetting(false);

      expect(workloadService.updateApplicationSubmissionSetting).toHaveBeenCalledWith(false);
      expect(store.applicationSubmissionSetting?.enabled).toBe(false);
    });

    it("should rollback on error", async () => {
      const store = useSettingStore();
      store.applicationSubmissionSetting = { enabled: true };
      vi.mocked(workloadService.updateApplicationSubmissionSetting).mockRejectedValue(new Error("Failed"));

      await expect(store.updateApplicationSubmissionSetting(false)).rejects.toThrow("Failed");
      expect(store.applicationSubmissionSetting?.enabled).toBe(true);
    });

    it("should no-op when setting is null", async () => {
      const store = useSettingStore();
      store.applicationSubmissionSetting = null;

      const result = await store.updateApplicationSubmissionSetting(false);

      expect(result).toBeUndefined();
      expect(workloadService.updateApplicationSubmissionSetting).not.toHaveBeenCalled();
    });
  });
});
