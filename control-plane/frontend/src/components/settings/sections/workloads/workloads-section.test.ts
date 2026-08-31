/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import WorkloadsSection from "./workloads-section.vue";
import type { IWorkloadsSectionSettings } from "@/models/setting.model";
import { useClusterStore } from "@/stores/cluster.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import type { YamlSubmissionSetting } from "@/swagger-models/workloads-manager-client";
import type { IApplicationSubmissionSetting } from "@/services/cluster/workload.service/workload.service";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { SettingKeys } from "@/models/setting.model";

vi.mock("@/services/infra/unleash.service/unleash.service", () => ({
  unleashService: {
    isFeatureEnabled: vi.fn(() => true),
  },
}));

interface CreateWrapperOptions {
  yamlSubmissionEnabled?: boolean;
  showYamlSubmission?: boolean;
  applicationSubmissionEnabled?: boolean;
  showApplicationSubmission?: boolean;
  showNgcRegistry?: boolean;
  hasRegistriesRead?: boolean;
  hasSettingsPerm?: boolean;
}

const createWrapper = (workloadSectionSettings: IWorkloadsSectionSettings, options?: CreateWrapperOptions) => {
  const pinia = createTestingPinia({ createSpy: vi.fn });

  const clusterStore = useClusterStore(pinia);
  const showYaml = options?.showYamlSubmission ?? false;
  const showApp = options?.showApplicationSubmission ?? false;
  const showNgc = options?.showNgcRegistry ?? false;
  clusterStore.isAtleastOneClusterSufficientForVersion = vi.fn().mockImplementation((version: string) => {
    if (version === "2.23") return showYaml || showApp || showNgc;
    return showYaml;
  });

  const permissionStore = usePermissionStore(pinia);
  const hasRegistries = options?.hasRegistriesRead ?? false;
  permissionStore.hasPermission = vi.fn().mockImplementation((resource: ResourceType, action: Action) => {
    if (resource === ResourceType.Registries && action === Action.Read) return hasRegistries;
    return false;
  });
  const hasSettings = options?.hasSettingsPerm ?? true;
  permissionStore.hasSomePermittedActions = vi.fn().mockImplementation((resource: ResourceType) => {
    if (resource === ResourceType.Settings) return hasSettings;
    return true;
  });

  const settingStore = useSettingStore(pinia);
  settingStore.yamlSubmissionSetting =
    options?.yamlSubmissionEnabled !== undefined
      ? ({ enabled: options.yamlSubmissionEnabled } as YamlSubmissionSetting)
      : null;
  settingStore.applicationSubmissionSetting =
    options?.applicationSubmissionEnabled !== undefined
      ? ({ enabled: options.applicationSubmissionEnabled } as IApplicationSubmissionSetting)
      : null;

  return mount(WorkloadsSection, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        RunaiExpansionItem: {
          template: "<div><slot /></div>",
        },
        ToggleSetting: {
          name: "ToggleSetting",
          props: {
            setting: { type: Object, default: () => ({}) },
            resourceType: { type: String },
          },
          template: "<div></div>",
        },
        NgcPrivateRegistrySection: true,
        YamlSubmissionSettingSection: true,
        ApplicationSubmissionSettingSection: true,
      },
    },
    props: { workloadSectionSettings },
  });
};

describe("WorkloadsSection - sectionSubHeader counts", () => {
  it("counts all enabled settings correctly", () => {
    const wrapper = createWrapper({
      policies: { key: "policies", value: "true" },
      models: { key: "models", value: "true" },
      allowGPT2Model: { key: "allowGPT2Model", value: "true" },
      nim: { key: "nim", value: "true" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (4) / Disabled (0)");
  });

  it("includes allowGPT2Model in count when models value is string 'false' (truthy string)", () => {
    const wrapper = createWrapper({
      policies: { key: "policies", value: "true" },
      models: { key: "models", value: "false" },
      allowGPT2Model: { key: "allowGPT2Model", value: "true" },
      nim: { key: "nim", value: "true" },
    });
    // Boolean("false") === true, so allowGPT2Model IS counted
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (3) / Disabled (1)");
  });

  it("includes allowGPT2Model in count when models is enabled", () => {
    const wrapper = createWrapper({
      policies: { key: "policies", value: "false" },
      models: { key: "models", value: "true" },
      allowGPT2Model: { key: "allowGPT2Model", value: "true" },
      nim: { key: "nim", value: "false" },
    });
    // allowGPT2Model IS counted since models is enabled
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (2) / Disabled (2)");
  });

  it("excludes allowGPT2Model from count when models is absent", () => {
    const wrapper = createWrapper({
      policies: { key: "policies", value: "true" },
      allowGPT2Model: { key: "allowGPT2Model", value: "true" },
    });
    // models is undefined (falsy), so allowGPT2Model should NOT be counted
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (1) / Disabled (0)");
  });

  it("shows Enabled (0) / Disabled (0) when no settings are present", () => {
    const wrapper = createWrapper({});
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (0) / Disabled (0)");
  });

  it("counts mixed enabled and disabled settings correctly", () => {
    const wrapper = createWrapper({
      policies: { key: "policies", value: "true" },
      models: { key: "models", value: "false" },
      deletedWorkloadsView: { key: "deletedWorkloadsView", value: "false" },
      nim: { key: "nim", value: "true" },
      hybridSubmission: { key: "hybridSubmission", value: "false" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (2) / Disabled (3)");
  });

  it("includes yamlSubmission as enabled when setting is enabled and cluster supports it", () => {
    const wrapper = createWrapper(
      {
        policies: { key: "policies", value: "true" },
        yamlSubmission: { key: "yamlSubmission" },
      },
      { showYamlSubmission: true, yamlSubmissionEnabled: true },
    );
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (2) / Disabled (0)");
  });

  it("includes yamlSubmission as disabled when setting is disabled and cluster supports it", () => {
    const wrapper = createWrapper(
      {
        policies: { key: "policies", value: "true" },
        yamlSubmission: { key: "yamlSubmission" },
      },
      { showYamlSubmission: true, yamlSubmissionEnabled: false },
    );
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (1) / Disabled (1)");
  });

  it("excludes yamlSubmission from count when cluster does not support it", () => {
    const wrapper = createWrapper(
      {
        policies: { key: "policies", value: "true" },
        yamlSubmission: { key: "yamlSubmission" },
      },
      { showYamlSubmission: false, yamlSubmissionEnabled: true },
    );
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (1) / Disabled (0)");
  });

  it("excludes dataVolumes and templatesV2 from count when hybridSubmission is disabled", () => {
    const wrapper = createWrapper({
      hybridSubmission: { key: "hybridSubmission", value: "false" },
      dataVolumes: { key: "dataVolumes", value: "false" },
      templatesV2: { key: "templatesV2", value: "false" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (0) / Disabled (1)");
  });

  it("includes dataVolumes and templatesV2 in count when hybridSubmission is enabled", () => {
    const wrapper = createWrapper({
      hybridSubmission: { key: "hybridSubmission", value: "true" },
      dataVolumes: { key: "dataVolumes", value: "false" },
      templatesV2: { key: "templatesV2", value: "true" },
    });
    // hybridSubmission(enabled) + templatesV2(enabled) = 2, dataVolumes(disabled) = 1
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (2) / Disabled (1)");
  });

  it("includes applicationSubmission as enabled when setting is enabled and feature flag is on", () => {
    const wrapper = createWrapper(
      {
        policies: { key: "policies", value: "true" },
        applicationSubmission: { key: "applicationSubmission" },
      },
      { showApplicationSubmission: true, applicationSubmissionEnabled: true },
    );
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (2) / Disabled (0)");
  });

  it("includes applicationSubmission as disabled when setting is disabled", () => {
    const wrapper = createWrapper(
      {
        policies: { key: "policies", value: "true" },
        applicationSubmission: { key: "applicationSubmission" },
      },
      { showApplicationSubmission: true, applicationSubmissionEnabled: false },
    );
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (1) / Disabled (1)");
  });

  it("excludes applicationSubmission from count when feature flag is off", () => {
    vi.mocked(unleashService.isFeatureEnabled).mockReturnValueOnce(false);

    const wrapper = createWrapper(
      {
        policies: { key: "policies", value: "true" },
        applicationSubmission: { key: "applicationSubmission" },
      },
      { showApplicationSubmission: false, applicationSubmissionEnabled: true },
    );
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (1) / Disabled (0)");
  });
});

describe("WorkloadsSection - NGC integration toggle and registry form gating", () => {
  const ngcOn = { key: SettingKeys.EnableNgcIntegration, value: "true" };
  const ngcOff = { key: SettingKeys.EnableNgcIntegration, value: "false" };

  it("filters EnableNgcIntegration out of the generic toggle loop via shouldShowSetting", () => {
    const wrapper = createWrapper({ ngcIntegration: ngcOn });
    expect(wrapper.vm.shouldShowSetting(SettingKeys.EnableNgcIntegration)).toBe(false);
  });

  it("hides toggle row and form when user has no Registries Read permission", () => {
    const wrapper = createWrapper({ ngcIntegration: ngcOn }, { hasRegistriesRead: false, showNgcRegistry: true });
    expect(wrapper.vm.showNgcPrivateRegistries).toBe(false);
    expect(wrapper.findComponent({ name: "NgcPrivateRegistrySection" }).exists()).toBe(false);
  });

  it("hides toggle row and form when no cluster meets the NGC API key version", () => {
    const wrapper = createWrapper({ ngcIntegration: ngcOn }, { hasRegistriesRead: true, showNgcRegistry: false });
    expect(wrapper.vm.showNgcPrivateRegistries).toBe(false);
    expect(wrapper.findComponent({ name: "NgcPrivateRegistrySection" }).exists()).toBe(false);
  });

  it("renders the registry form when user has Registries Read, cluster supports NGC, and toggle is on", () => {
    const wrapper = createWrapper({ ngcIntegration: ngcOn }, { hasRegistriesRead: true, showNgcRegistry: true });
    expect(wrapper.vm.showNgcPrivateRegistries).toBe(true);
    expect(wrapper.vm.isNgcIntegrationEnabled).toBe(true);
    expect(wrapper.findComponent({ name: "NgcPrivateRegistrySection" }).exists()).toBe(true);
  });

  it("hides the registry form (but keeps the toggle row) when toggle is off", () => {
    const wrapper = createWrapper({ ngcIntegration: ngcOff }, { hasRegistriesRead: true, showNgcRegistry: true });
    expect(wrapper.vm.showNgcPrivateRegistries).toBe(true);
    expect(wrapper.vm.isNgcIntegrationEnabled).toBe(false);
    expect(wrapper.findComponent({ name: "NgcPrivateRegistrySection" }).exists()).toBe(false);
  });

  it("returns the ngcIntegration setting from workloadSectionSettings via the computed", () => {
    const wrapper = createWrapper({ ngcIntegration: ngcOn });
    expect(wrapper.vm.ngcIntegrationSetting).toEqual(ngcOn);
  });

  it("treats a missing ngcIntegration setting as disabled", () => {
    const wrapper = createWrapper({}, { hasRegistriesRead: true, showNgcRegistry: true });
    expect(wrapper.vm.ngcIntegrationSetting).toBeUndefined();
    expect(wrapper.vm.isNgcIntegrationEnabled).toBe(false);
    expect(wrapper.findComponent({ name: "NgcPrivateRegistrySection" }).exists()).toBe(false);
  });
});

describe("WorkloadsSection - canViewSettingsSection vs NGC independence", () => {
  it("canViewSettingsSection is true when the user has any Settings action", () => {
    const wrapper = createWrapper({ policies: { key: "policies", value: "true" } }, { hasSettingsPerm: true });
    expect(wrapper.vm.canViewSettingsSection).toBe(true);
  });

  it("canViewSettingsSection is false when the user has no Settings actions", () => {
    const wrapper = createWrapper({ policies: { key: "policies", value: "true" } }, { hasSettingsPerm: false });
    expect(wrapper.vm.canViewSettingsSection).toBe(false);
  });

  it("hides non-NGC toggles when canViewSettingsSection is false", () => {
    const wrapper = createWrapper(
      { policies: { key: "policies", value: "true" }, models: { key: "models", value: "true" } },
      { hasSettingsPerm: false },
    );
    expect(wrapper.findAllComponents({ name: "ToggleSetting" })).toHaveLength(0);
  });

  it("still shows the NGC toggle + form when Settings perm is absent but Registries is present", () => {
    const wrapper = createWrapper(
      { ngcIntegration: { key: SettingKeys.EnableNgcIntegration, value: "true" } },
      { hasSettingsPerm: false, hasRegistriesRead: true, showNgcRegistry: true },
    );
    expect(wrapper.vm.canViewSettingsSection).toBe(false);
    expect(wrapper.vm.showNgcPrivateRegistries).toBe(true);
    expect(wrapper.findComponent({ name: "NgcPrivateRegistrySection" }).exists()).toBe(true);

    const toggles = wrapper.findAllComponents({ name: "ToggleSetting" });
    expect(toggles).toHaveLength(1);
    expect(toggles[0].props("resourceType")).toBe("Registries");
  });

  it("hides everything when both Settings and Registries are absent", () => {
    const wrapper = createWrapper(
      { ngcIntegration: { key: SettingKeys.EnableNgcIntegration, value: "true" } },
      { hasSettingsPerm: false, hasRegistriesRead: false, showNgcRegistry: true },
    );
    expect(wrapper.vm.canViewSettingsSection).toBe(false);
    expect(wrapper.vm.showNgcPrivateRegistries).toBe(false);
    expect(wrapper.findAllComponents({ name: "ToggleSetting" })).toHaveLength(0);
    expect(wrapper.findComponent({ name: "NgcPrivateRegistrySection" }).exists()).toBe(false);
  });

  it("NGC integration toggle uses resourceType='Registries' (not 'Settings')", () => {
    const wrapper = createWrapper(
      { ngcIntegration: { key: SettingKeys.EnableNgcIntegration, value: "true" } },
      { hasSettingsPerm: true, hasRegistriesRead: true, showNgcRegistry: true },
    );

    const toggles = wrapper.findAllComponents({ name: "ToggleSetting" });
    const ngcToggle = toggles.find((t) => t.props("resourceType") === "Registries");
    expect(ngcToggle).toBeDefined();
  });

  it("non-NGC toggles do not override resourceType (fall back to toggle-setting's 'Settings' default)", () => {
    const wrapper = createWrapper(
      {
        policies: { key: "policies", value: "true" },
        models: { key: "models", value: "true" },
      },
      { hasSettingsPerm: true },
    );

    const toggles = wrapper.findAllComponents({ name: "ToggleSetting" });
    const nonNgcToggles = toggles.filter((t) => t.props("resourceType") !== "Registries");
    expect(nonNgcToggles.length).toBeGreaterThan(0);
    expect(nonNgcToggles.every((t) => t.props("resourceType") === undefined)).toBe(true);
  });
});
