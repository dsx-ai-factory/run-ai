/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import SamlSignatureStatusBar from "./saml-signature-status-bar.vue";
import { useIdpsStore } from "@/stores/idps.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { IdpTypeEnum, type Idp } from "@/swagger-models/identity-manager-client";
import { Action, ResourceType, ResourceTypeGroupId } from "@/swagger-models/authorization-client";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";

const mockRoute = vi.hoisted(() => ({ name: "some-page" as string }));

vi.mock("vue-router", async (importOriginal) => ({
  ...(await importOriginal<Record<string, unknown>>()),
  useRoute: () => mockRoute,
}));

interface CreateOptions {
  validateSignature?: boolean;
  type?: IdpTypeEnum;
  actions?: Action[];
  routeName?: string;
}

function createWrapper({
  validateSignature = false,
  type = IdpTypeEnum.Saml,
  actions = [Action.Read, Action.Update],
  routeName = "some-page",
}: CreateOptions = {}) {
  const pinia = createTestingPinia({ stubActions: false });
  usePermissionStore(pinia).setPermissions([
    {
      resourceType: ResourceType.SecuritySettings,
      displayName: ResourceType.SecuritySettings,
      groupId: ResourceTypeGroupId.Organization,
      actions,
    },
  ]);
  useIdpsStore(pinia).idps = [{ alias: "idp", type, samlData: { validateSignature } } as unknown as Idp];
  mockRoute.name = routeName;

  return mount(SamlSignatureStatusBar, {
    global: {
      plugins: [Quasar, pinia],
      stubs: {
        RouterLink: { name: "RouterLink", props: ["to"], template: "<a data-test='link'><slot /></a>" },
      },
    },
  });
}

const bar = (wrapper: ReturnType<typeof createWrapper>) => wrapper.find("[aid='saml-signature-status-bar']");

describe("saml-signature-status-bar.vue", () => {
  it("shows the banner for an admin with SecuritySettings Update when validation is disabled", () => {
    const wrapper = createWrapper();

    expect(bar(wrapper).exists()).toBe(true);
    expect(wrapper.text()).toContain("Your SAML connection isn't validating identity provider signatures,");
    expect(wrapper.text()).toContain("which may impact user verification. To enable validation,");
    expect(wrapper.text()).toContain("reconfigure SAML in your SSO settings.");
  });

  it("uses the error variant", () => {
    expect(createWrapper().find(".notification-alert-bar").classes()).toContain("variant-error");
  });

  it("links to general settings with the security section expanded", () => {
    const link = createWrapper().findComponent({ name: "RouterLink" });

    expect(link.props("to")).toEqual({ name: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX, query: { section: "security" } });
  });

  it("is hidden when validation is enabled", () => {
    expect(bar(createWrapper({ validateSignature: true })).exists()).toBe(false);
  });

  it("is hidden for a user without SecuritySettings Update", () => {
    expect(bar(createWrapper({ actions: [Action.Read] })).exists()).toBe(false);
  });

  it("is hidden when the IdP is not SAML", () => {
    expect(bar(createWrapper({ type: IdpTypeEnum.Oidc })).exists()).toBe(false);
  });

  it("is hidden on the general settings page, which shows the warning inline", () => {
    expect(bar(createWrapper({ routeName: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX })).exists()).toBe(false);
  });
});
