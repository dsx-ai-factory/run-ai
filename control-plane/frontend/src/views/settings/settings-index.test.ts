/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import SettingsIndex from "./settings-index.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAppStore } from "@/stores/app.store";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";
import { SettingKeys, type Setting } from "@/models/setting.model";

vi.mock("@/services/infra/unleash.service/unleash.service", () => ({
  unleashService: {
    showNgcIntegration: vi.fn(() => true),
    isFeatureEnabled: vi.fn(() => false),
  },
}));

const ALL_CRUD = [Action.Create, Action.Read, Action.Update, Action.Delete];

function permission(resourceType: ResourceType, actions: Action[]): Permission {
  return {
    resourceType,
    displayName: resourceType,
    groupId: ResourceTypeGroupId.Organization,
    actions,
  };
}

function buildSettingsMap(): Map<string, Setting> {
  return new Map<string, Setting>([
    [SettingKeys.AdvancedGPUMetrics, { key: SettingKeys.AdvancedGPUMetrics, value: "false" }],
    [SettingKeys.EnablePolicies, { key: SettingKeys.EnablePolicies, value: "false" }],
    [SettingKeys.LimitQuotaOverSubscription, { key: SettingKeys.LimitQuotaOverSubscription, value: "false" }],
  ]);
}

interface CreateOptions {
  permissions?: Permission[];
  clusterSupportsNgc?: boolean;
}

function createWrapper({ permissions = [], clusterSupportsNgc = false }: CreateOptions = {}) {
  const pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });

  const permissionStore = usePermissionStore(pinia);
  permissionStore.setPermissions(permissions);

  const settingStore = useSettingStore(pinia);
  settingStore.settings = buildSettingsMap();
  settingStore.getBffSettingAsOldFormat = vi.fn().mockReturnValue(undefined);

  const clusterStore = useClusterStore(pinia);
  clusterStore.isAtleastOneClusterSufficientForVersion = vi.fn().mockReturnValue(clusterSupportsNgc);

  const appStore = useAppStore(pinia);
  appStore.setPageLoading = vi.fn();

  const wrapper = mount(SettingsIndex, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        AnalyticsSection: { name: "AnalyticsSection", template: "<div data-test='analytics'></div>" },
        ResourcesSection: { name: "ResourcesSection", template: "<div data-test='resources'></div>" },
        WorkloadsSection: { name: "WorkloadsSection", template: "<div data-test='workloads'></div>" },
        SecuritySection: { name: "SecuritySection", template: "<div data-test='security'></div>" },
        NotificationsSection: { name: "NotificationsSection", template: "<div data-test='notifications'></div>" },
        ClusterAuthenticationsSection: {
          name: "ClusterAuthenticationsSection",
          template: "<div data-test='cluster-auth'></div>",
        },
        BrandingSection: { name: "BrandingSection", template: "<div data-test='branding'></div>" },
      },
    },
  });

  return { wrapper, permissionStore, settingStore };
}

function visibleSections(wrapper: ReturnType<typeof createWrapper>["wrapper"]) {
  return {
    analytics: wrapper.findComponent({ name: "AnalyticsSection" }).exists(),
    resources: wrapper.findComponent({ name: "ResourcesSection" }).exists(),
    workloads: wrapper.findComponent({ name: "WorkloadsSection" }).exists(),
    security: wrapper.findComponent({ name: "SecuritySection" }).exists(),
    notifications: wrapper.findComponent({ name: "NotificationsSection" }).exists(),
    clusterAuth: wrapper.findComponent({ name: "ClusterAuthenticationsSection" }).exists(),
    branding: wrapper.findComponent({ name: "BrandingSection" }).exists(),
  };
}

describe("settings-index.vue - permission-driven section visibility", () => {
  it("resolves setting documentation links without mutating the setting store", () => {
    const { wrapper, settingStore } = createWrapper();
    const originalDescription =
      'See <a href="https://run-ai-docs.nvidia.com/self-hosted/2.24/platform-management/policies">policies</a>.';
    const setting = {
      key: SettingKeys.EnablePolicies,
      value: "false",
      description: originalDescription,
    } as Setting;
    settingStore.settings.set(SettingKeys.EnablePolicies, setting);

    const resolved = wrapper.vm.getSettingValue(SettingKeys.EnablePolicies);

    expect(resolved?.description).toBe(
      'See <a href="https://run-ai-docs.nvidia.com/self-hosted/platform-management/policies">policies</a>.',
    );
    expect(setting.description).toBe(originalDescription);
  });

  it("resolves documentation links in BFF-backed settings", () => {
    const { wrapper, settingStore } = createWrapper();
    settingStore.getBffSettingAsOldFormat = vi.fn().mockReturnValue({
      key: SettingKeys.EnableNIM,
      value: "false",
      description: 'See <a href="https://run-ai-docs.nvidia.com/self-hosted/workloads-in-nvidia-run-ai">workloads</a>.',
    });

    const resolved = wrapper.vm.getSettingValue(SettingKeys.EnableNIM);

    expect(resolved?.description).toBe(
      'See <a href="https://run-ai-docs.nvidia.com/self-hosted/workloads-in-nvidia-run-ai">workloads</a>.',
    );
  });

  it("all 5 resources = CRUD → all 7 sections visible", () => {
    const { wrapper } = createWrapper({
      permissions: [
        permission(ResourceType.Settings, ALL_CRUD),
        permission(ResourceType.SecuritySettings, ALL_CRUD),
        permission(ResourceType.BrandingSettings, ALL_CRUD),
        permission(ResourceType.NotificationSettings, ALL_CRUD),
        permission(ResourceType.Registries, ALL_CRUD),
      ],
      clusterSupportsNgc: true,
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: true,
      resources: true,
      workloads: true,
      security: true,
      notifications: true,
      clusterAuth: true,
      branding: true,
    });
  });

  it("zero permissions on all 5 resources → no sections visible", () => {
    const { wrapper } = createWrapper({ permissions: [] });

    expect(visibleSections(wrapper)).toEqual({
      analytics: false,
      resources: false,
      workloads: false,
      security: false,
      notifications: false,
      clusterAuth: false,
      branding: false,
    });
  });

  it("only Settings = CRUD → general sections visible; Security/Branding/Notifications hidden", () => {
    const { wrapper } = createWrapper({
      permissions: [permission(ResourceType.Settings, ALL_CRUD)],
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: true,
      resources: true,
      workloads: true,
      security: false,
      notifications: false,
      clusterAuth: true,
      branding: false,
    });
  });

  it("only SecuritySettings = CRUD → only Security visible", () => {
    const { wrapper } = createWrapper({
      permissions: [permission(ResourceType.SecuritySettings, ALL_CRUD)],
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: false,
      resources: false,
      workloads: false,
      security: true,
      notifications: false,
      clusterAuth: false,
      branding: false,
    });
  });

  it("only BrandingSettings = CRUD → only Branding visible", () => {
    const { wrapper } = createWrapper({
      permissions: [permission(ResourceType.BrandingSettings, ALL_CRUD)],
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: false,
      resources: false,
      workloads: false,
      security: false,
      notifications: false,
      clusterAuth: false,
      branding: true,
    });
  });

  it("only NotificationSettings = CRUD → only Notifications visible", () => {
    const { wrapper } = createWrapper({
      permissions: [permission(ResourceType.NotificationSettings, ALL_CRUD)],
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: false,
      resources: false,
      workloads: false,
      security: false,
      notifications: true,
      clusterAuth: false,
      branding: false,
    });
  });

  it("only Registries = CRUD (cluster supports NGC) → Workloads visible via canViewRegistriesSection", () => {
    const { wrapper } = createWrapper({
      permissions: [permission(ResourceType.Registries, ALL_CRUD)],
      clusterSupportsNgc: true,
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: false,
      resources: false,
      workloads: true,
      security: false,
      notifications: false,
      clusterAuth: false,
      branding: false,
    });
  });

  it("only Registries = CRUD but cluster does NOT support NGC → Workloads hidden", () => {
    const { wrapper } = createWrapper({
      permissions: [permission(ResourceType.Registries, ALL_CRUD)],
      clusterSupportsNgc: false,
    });

    expect(visibleSections(wrapper).workloads).toBe(false);
  });

  it("SecuritySettings = READ + Settings = CRUD → Security visible and general sections visible", () => {
    const { wrapper } = createWrapper({
      permissions: [
        permission(ResourceType.SecuritySettings, [Action.Read]),
        permission(ResourceType.Settings, ALL_CRUD),
      ],
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: true,
      resources: true,
      workloads: true,
      security: true,
      notifications: false,
      clusterAuth: true,
      branding: false,
    });
  });

  it("BrandingSettings = READ only (other resources empty) → only Branding visible", () => {
    const { wrapper } = createWrapper({
      permissions: [permission(ResourceType.BrandingSettings, [Action.Read])],
    });

    expect(visibleSections(wrapper)).toEqual({
      analytics: false,
      resources: false,
      workloads: false,
      security: false,
      notifications: false,
      clusterAuth: false,
      branding: true,
    });
  });

  it("granting Settings CRUD at runtime reflows the visible sections (Pinia reactivity)", async () => {
    const { wrapper, permissionStore } = createWrapper({ permissions: [] });

    expect(visibleSections(wrapper)).toEqual({
      analytics: false,
      resources: false,
      workloads: false,
      security: false,
      notifications: false,
      clusterAuth: false,
      branding: false,
    });

    permissionStore.setPermissions([permission(ResourceType.Settings, ALL_CRUD)]);
    await wrapper.vm.$nextTick();

    expect(visibleSections(wrapper)).toEqual({
      analytics: true,
      resources: true,
      workloads: true,
      security: false,
      notifications: false,
      clusterAuth: true,
      branding: false,
    });
  });
});
