/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import SsoSettings from "./sso-settings.vue";
import { useIdpsStore } from "@/stores/idps.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import { useSecuritySettingsStore } from "@/stores/security-settings.store";
import type { Idp, SamlData } from "@/swagger-models/identity-manager-client";
import { IdpTypeEnum, SettingsKeyEnum } from "@/swagger-models/identity-manager-client";
import { securitySettingsService } from "@/services/control-plane/security-settings.service/security-settings.service";
import { settingsUtil } from "@/utils/settings.util";
import { DOCS } from "@/constants/docs";
import type { MockedFunction } from "vitest";

vi.mock("@/services/control-plane/security-settings.service/security-settings.service");

vi.mock("console", () => ({
  error: vi.fn(),
}));

describe("sso-settings.vue", () => {
  const mockSamlData: SamlData = {
    entityId: "test-entity",
    singleSignOnServiceUrl: "http://test.com/sso",
    signingCertificate: "test-cert",
    validateSignature: true,
  };

  const factory = (idp?: Idp) => {
    const mockNotify = vi.fn();

    // Mock store states before mounting
    setActivePinia(createTestingPinia({ createSpy: vi.fn }));
    const idpStore = useIdpsStore();
    const authStore = useAuthStore();
    const appStore = useAppStore();
    const securitySettingsStore = useSecuritySettingsStore();

    // Mock store methods and properties
    Object.assign(idpStore, {
      currentIdp: idp,
      idpMappers: [],
    });

    Object.assign(appStore, {
      config: {
        authHost: "http://test.com",
      },
    });

    Object.assign(securitySettingsStore, {
      updateSetting: vi.fn().mockResolvedValue(undefined) as MockedFunction<typeof securitySettingsStore.updateSetting>,
    });

    // Mock idmSettingsService
    vi.mocked(securitySettingsService.getSettingByKey).mockResolvedValue({
      autoRedirectSSO: {
        value: {
          enabled: false,
          idpAlias: idp?.alias || "",
        },
      },
    });

    const wrapper = mount(SsoSettings, {
      global: {
        plugins: [Quasar],
        stubs: {
          // Stub only the components we use
          SamlSettings: {
            name: "saml-settings",
            template: '<div><slot name="box-footer"></slot></div>',
          },
          OidcSettings: true,
          OcpSettings: true,
          RunaiTooltip: true,
          SettingReadonlyInput: {
            template: '<div><slot :readonly="false"></slot></div>',
          },
          ToggleSetting: true,
          // Simple stubs for Quasar components
          "q-btn": true,
          "q-menu": true,
          "q-list": true,
          "q-item": true,
          "q-item-section": true,
          "q-separator": true,
        },
        mocks: {
          $q: {
            notify: mockNotify,
          },
        },
      },
      props: {
        settings: {
          autoRedirectSSO: {
            value: {
              enabled: false,
              idpAlias: idp?.alias || "",
            },
          },
        },
      },
    });

    return {
      wrapper,
      idpStore,
      authStore,
      appStore,
      securitySettingsStore,
      mockNotify,
    };
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders properly", () => {
    const { wrapper } = factory();
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.find(".sso-settings").exists()).toBe(true);
  });

  it("shows add button when no IDP is configured", async () => {
    const { wrapper } = factory();
    await wrapper.setData({
      isNewIdp: false,
      idpNewAlias: "",
    });
    expect(wrapper.find("[aid='add-idp-btn']").exists()).toBe(true);
  });

  it("hides add button when IDP is configured", async () => {
    const mockIdp: Idp = {
      alias: IdpTypeEnum.Saml,
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper } = factory(mockIdp);
    await wrapper.vm.$nextTick();
    expect(wrapper.find("[aid='add-idp-btn']").exists()).toBe(false);
  });

  it("shows SAML settings when SAML IDP is selected", async () => {
    const { wrapper } = factory();
    await wrapper.vm.handleSelectIDP(IdpTypeEnum.Saml);
    expect(wrapper.findComponent({ name: "saml-settings" }).exists()).toBe(true);
  });

  it("shows OIDC settings when OIDC IDP is selected", async () => {
    const { wrapper } = factory();
    await wrapper.vm.handleSelectIDP(IdpTypeEnum.Oidc);
    expect(wrapper.findComponent({ name: "oidc-settings" }).exists()).toBe(true);
  });

  it("shows OpenShift settings when OpenShift IDP is selected", async () => {
    const { wrapper } = factory();
    await wrapper.vm.handleSelectIDP(IdpTypeEnum.OpenshiftV4);
    expect(wrapper.findComponent({ name: "ocp-settings" }).exists()).toBe(true);
  });

  it("shows OIDC settings when alias is auto-generated (type != alias)", async () => {
    const mockIdp: Idp = {
      alias: "oidc-a1b2c3",
      type: IdpTypeEnum.Oidc,
      redirectUri: "http://test.com",
    };
    const { wrapper } = factory(mockIdp);
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent({ name: "oidc-settings" }).exists()).toBe(true);
    expect(wrapper.find("[aid='add-idp-btn']").exists()).toBe(false);
  });

  it("shows SAML settings when alias is auto-generated (type != alias)", async () => {
    const mockIdp: Idp = {
      alias: "saml-a1b2c3",
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper } = factory(mockIdp);
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent({ name: "saml-settings" }).exists()).toBe(true);
    expect(wrapper.find("[aid='add-idp-btn']").exists()).toBe(false);
  });

  it("shows auto-redirect toggle when IDP is configured", async () => {
    const mockIdp: Idp = {
      alias: IdpTypeEnum.Saml,
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper, idpStore } = factory(mockIdp);

    Object.assign(idpStore, { currentIdp: mockIdp });

    await wrapper.setData({
      isEditMode: false,
      isNewIdp: false,
      idpNewAlias: "",
    });
    await wrapper.vm.$nextTick();

    const toggle = wrapper.findComponent({ name: "toggle-setting" });
    expect(toggle.exists()).toBe(true);
  });

  it("hides auto-redirect toggle when in edit mode", async () => {
    const mockIdp: Idp = {
      alias: IdpTypeEnum.Saml,
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper, authStore, idpStore } = factory(mockIdp);

    Object.assign(authStore, { isCurrentUserSSO: true });
    Object.assign(idpStore, { currentIdp: mockIdp });

    await wrapper.setData({
      isEditMode: true,
      isNewIdp: false,
      idpNewAlias: "",
    });
    await wrapper.vm.$nextTick();

    const toggle = wrapper.findComponent({ name: "toggle-setting" });
    expect(toggle.exists()).toBe(false);
  });

  it("hides auto-redirect toggle when no IDP is configured", async () => {
    const { wrapper, authStore } = factory();

    Object.assign(authStore, { isCurrentUserSSO: true });

    await wrapper.setData({
      isEditMode: false,
      isNewIdp: false,
      idpNewAlias: "",
    });
    await wrapper.vm.$nextTick();

    const toggle = wrapper.findComponent({ name: "toggle-setting" });
    expect(toggle.exists()).toBe(false);
  });

  it("shows help text when not in edit mode and no new IDP", async () => {
    const mockIdp: Idp = {
      alias: IdpTypeEnum.Saml,
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper } = factory(mockIdp);
    await wrapper.setData({
      isEditMode: false,
      isNewIdp: false,
      idpNewAlias: "",
    });
    await wrapper.vm.$nextTick();
    expect(wrapper.find(`a[href="${DOCS.AUTHENTICATION.OVERVIEW}"]`).exists()).toBe(true);
  });

  it("hides help text when in edit mode", async () => {
    const mockIdp: Idp = {
      alias: IdpTypeEnum.Saml,
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper } = factory(mockIdp);
    await wrapper.setData({ isEditMode: true });
    await wrapper.vm.$nextTick();
    expect(wrapper.find(`a[href="${DOCS.AUTHENTICATION.OVERVIEW}"]`).exists()).toBe(false);
  });

  it("updates auto-redirect setting when toggle is changed", async () => {
    const mockIdp: Idp = {
      alias: IdpTypeEnum.Saml,
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper, idpStore, securitySettingsStore } = factory(mockIdp);

    Object.assign(idpStore, { currentIdp: mockIdp });

    await wrapper.setData({
      isEditMode: false,
      isNewIdp: false,
      idpNewAlias: "",
    });

    await wrapper.vm.updateRedirectSSO("autoRedirectSSO", true);
    await flushPromises();

    expect(securitySettingsStore.updateSetting).toHaveBeenCalledWith(SettingsKeyEnum.AutoRedirectSso, {
      enabled: true,
      idpAlias: mockIdp.alias,
    });
  });

  it("handles error when updating auto-redirect setting", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const mockIdp: Idp = {
      alias: IdpTypeEnum.Saml,
      type: IdpTypeEnum.Saml,
      redirectUri: "http://test.com",
      samlData: mockSamlData,
    };
    const { wrapper, idpStore, securitySettingsStore, mockNotify } = factory(mockIdp);

    Object.assign(idpStore, { currentIdp: mockIdp });

    // Mock the store to reject with an error
    (
      securitySettingsStore.updateSetting as MockedFunction<typeof securitySettingsStore.updateSetting>
    ).mockRejectedValueOnce(new Error("Update failed"));

    await wrapper.setData({
      isEditMode: false,
      isNewIdp: false,
      idpNewAlias: "",
    });

    await wrapper.vm.updateRedirectSSO("autoRedirectSSO", true);
    await flushPromises();

    expect(securitySettingsStore.updateSetting).toHaveBeenCalledWith(SettingsKeyEnum.AutoRedirectSso, {
      enabled: true,
      idpAlias: mockIdp.alias,
    });
    expect(mockNotify).toHaveBeenCalledWith(settingsUtil.getSettingErrorMessage());
    errorSpy.mockRestore();
  });
});
