/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// scopeUtil.test.ts
import { afterEach, beforeEach, describe, test, expect, vi } from "vitest";
import { AxiosError } from "axios";
import { idpService } from "@/services/control-plane/idp.service/idp.service";
import { idpCommon } from "@/components/settings/sections/security/sso-settings/idps/idp-common/idp-common";

describe("getDefaultScope", () => {
  test("should return ['openid'] when input is an empty array", () => {
    const scope = "";
    const result = idpCommon.getDefaultScope(scope, true);
    expect(result).toEqual(["openid"]);
    const noOpenIDResult = idpCommon.getDefaultScope(scope, false);
    expect(noOpenIDResult).toEqual([]);
  });

  test("should return ['openid', 'profile', 'email'] when input does not contain 'openid'", () => {
    const scope = "profile email";
    const result = idpCommon.getDefaultScope(scope, true);
    expect(result).toEqual(["openid", "profile", "email"]);
    const noOpenIDResult = idpCommon.getDefaultScope(scope, false);
    expect(noOpenIDResult).toEqual(["profile", "email"]);
  });

  test("should return ['openid', 'profile'] when input already contains 'openid'", () => {
    const scope = "openid profile";
    const result = idpCommon.getDefaultScope(scope, true);
    expect(result).toEqual(["openid", "profile"]);
    const noOpenIDResult = idpCommon.getDefaultScope(scope, false);
    expect(noOpenIDResult).toEqual(["openid", "profile"]);
  });

  test("should return ['openid'] when input is undefined", () => {
    const scope = undefined;
    const result = idpCommon.getDefaultScope(scope, true);
    expect(result).toEqual(["openid"]);
    const noOpenIDResult = idpCommon.getDefaultScope(scope, false);
    expect(noOpenIDResult).toEqual([]);
  });
});

describe("isCertificateExpiringSoon", () => {
  const now = new Date("2026-01-01T00:00:00.000Z");
  const daysFromNow = (days: number): Date => new Date(now.getTime() + days * 24 * 60 * 60 * 1000);

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(now);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  test("should return false when the certificate is far from expiring", () => {
    expect(idpCommon.isCertificateExpiringSoon(daysFromNow(61))).toBe(false);
  });

  test("should return true at the 60-day threshold", () => {
    expect(idpCommon.isCertificateExpiringSoon(daysFromNow(60))).toBe(true);
  });

  test("should return true when within the 60-day window", () => {
    expect(idpCommon.isCertificateExpiringSoon(daysFromNow(1))).toBe(true);
  });

  test("should return false once the certificate has already expired", () => {
    expect(idpCommon.isCertificateExpiringSoon(daysFromNow(-1))).toBe(false);
  });
});

describe("validateSamlMetadata", () => {
  const badRequest = (data: unknown) =>
    new AxiosError("Bad Request", "ERR_BAD_REQUEST", undefined, undefined, { status: 400, data } as never);

  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("returns the expiration date from the backend dry run and updates an existing IdP by alias", async () => {
    const spy = vi.spyOn(idpService, "updateIdp").mockResolvedValue({
      alias: "saml",
      samlData: { validateSignature: true, certificateExpirationDate: "2030-05-17T00:00:00.000Z" },
    });
    const addSpy = vi.spyOn(idpService, "addIdp");
    const result = await idpCommon.validateSamlMetadata("xml", "idp.xml", "saml");
    expect(result).toEqual({ expirationDate: new Date("2030-05-17T00:00:00.000Z"), error: "" });
    expect(spy).toHaveBeenCalledWith(
      "saml",
      expect.objectContaining({ samlData: expect.objectContaining({ metadataXmlFile: "xml", fileName: "idp.xml" }) }),
      true,
    );
    expect(addSpy).not.toHaveBeenCalled();
  });

  test("uses the create dry run for a new IdP", async () => {
    const spy = vi.spyOn(idpService, "addIdp").mockResolvedValue({
      alias: "saml",
      samlData: { validateSignature: true, certificateExpirationDate: "2030-05-17T00:00:00.000Z" },
    });
    await idpCommon.validateSamlMetadata("xml", "idp.xml");
    expect(spy).toHaveBeenCalledWith(expect.anything(), true);
  });

  test("keeps the date and drops the error when the backend rejects an expired certificate", async () => {
    vi.spyOn(idpService, "addIdp").mockRejectedValue(
      badRequest({
        alias: "saml",
        samlData: { validateSignature: true, certificateExpirationDate: "2020-01-01T00:00:00.000Z" },
      }),
    );
    const result = await idpCommon.validateSamlMetadata("xml", "idp.xml");
    expect(result.expirationDate).toEqual(new Date("2020-01-01T00:00:00.000Z"));
    expect(result.error).toBe("");
  });

  test("keeps the backend rejection when the certificate is not expired (e.g. not yet valid)", async () => {
    vi.spyOn(idpService, "addIdp").mockRejectedValue(
      badRequest({
        alias: "saml",
        samlData: { validateSignature: true, certificateExpirationDate: "2099-01-01T00:00:00.000Z" },
      }),
    );
    const result = await idpCommon.validateSamlMetadata("xml", "idp.xml");
    expect(result.expirationDate).toEqual(new Date("2099-01-01T00:00:00.000Z"));
    expect(result.error).toBe("Invalid SAML metadata");
  });

  test("treats the backend's zero time as an unknown expiration date", async () => {
    vi.spyOn(idpService, "addIdp").mockResolvedValue({
      alias: "saml",
      samlData: { validateSignature: true, certificateExpirationDate: "0001-01-01T00:00:00Z" },
    });
    const result = await idpCommon.validateSamlMetadata("xml", "idp.xml");
    expect(result.expirationDate).toBeNull();
    expect(result.error).toBe("Could not read the certificate expiration date.");
  });

  test("returns the backend message when there is no date", async () => {
    vi.spyOn(idpService, "addIdp").mockRejectedValue(badRequest({ code: 400, message: "no signing certificate" }));
    expect(await idpCommon.validateSamlMetadata("xml", "idp.xml")).toEqual({
      expirationDate: null,
      error: "no signing certificate",
    });
  });

  test("returns an error instead of throwing when the request fails", async () => {
    vi.spyOn(idpService, "addIdp").mockRejectedValue(new Error("Network Error"));
    expect(await idpCommon.validateSamlMetadata("xml", "idp.xml")).toEqual({
      expirationDate: null,
      error: "Network Error",
    });
  });
});
