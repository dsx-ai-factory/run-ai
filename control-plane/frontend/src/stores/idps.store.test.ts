/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useIdpsStore } from "./idps.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { idpService } from "@/services/control-plane/idp.service/idp.service";
import { IdpTypeEnum, type Idp } from "@/swagger-models/identity-manager-client";
import { Action, ResourceType, ResourceTypeGroupId } from "@/swagger-models/authorization-client";

vi.mock("@/services/control-plane/idp.service/idp.service", () => ({
  idpService: { getIdps: vi.fn() },
}));

function samlIdp(validateSignature?: boolean): Idp {
  return {
    alias: "saml",
    type: IdpTypeEnum.Saml,
    samlData: { entityId: "entity", validateSignature },
  } as unknown as Idp;
}

function grantSecuritySettings(actions: Action[]): void {
  usePermissionStore().setPermissions([
    {
      resourceType: ResourceType.SecuritySettings,
      displayName: ResourceType.SecuritySettings,
      groupId: ResourceTypeGroupId.Organization,
      actions,
    },
  ]);
}

describe("idps.store - SAML signature validation", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("isSamlSignatureValidationDisabled", () => {
    it.each([
      { name: "SAML with validateSignature=false", idp: samlIdp(false), expected: true },
      { name: "SAML with validateSignature=true", idp: samlIdp(true), expected: false },
      { name: "SAML without the field (older backend)", idp: samlIdp(undefined), expected: false },
      {
        name: "OIDC even if the field is false",
        idp: { alias: "oidc", type: IdpTypeEnum.Oidc, samlData: { validateSignature: false } } as unknown as Idp,
        expected: false,
      },
    ])("$name → $expected", ({ idp, expected }) => {
      const store = useIdpsStore();
      store.idps = [idp];

      expect(store.isSamlSignatureValidationDisabled).toBe(expected);
    });

    it("no IdP configured → false", () => {
      expect(useIdpsStore().isSamlSignatureValidationDisabled).toBe(false);
    });
  });

  describe("shouldShowSamlSignatureBanner", () => {
    it("validation disabled + SecuritySettings Update → true", () => {
      grantSecuritySettings([Action.Read, Action.Update]);
      const store = useIdpsStore();
      store.idps = [samlIdp(false)];

      expect(store.shouldShowSamlSignatureBanner).toBe(true);
    });

    it("validation disabled but only SecuritySettings Read → false", () => {
      grantSecuritySettings([Action.Read]);
      const store = useIdpsStore();
      store.idps = [samlIdp(false)];

      expect(store.shouldShowSamlSignatureBanner).toBe(false);
    });

    it("SecuritySettings Update but validation enabled → false", () => {
      grantSecuritySettings([Action.Read, Action.Update]);
      const store = useIdpsStore();
      store.idps = [samlIdp(true)];

      expect(store.shouldShowSamlSignatureBanner).toBe(false);
    });

    it("banner goes away once the IdP is reloaded with validation enabled", async () => {
      grantSecuritySettings([Action.Read, Action.Update]);
      const store = useIdpsStore();
      store.idps = [samlIdp(false)];
      expect(store.shouldShowSamlSignatureBanner).toBe(true);

      vi.mocked(idpService.getIdps).mockResolvedValue([samlIdp(true)]);
      await store.loadIdps();

      expect(store.shouldShowSamlSignatureBanner).toBe(false);
    });
  });
});
