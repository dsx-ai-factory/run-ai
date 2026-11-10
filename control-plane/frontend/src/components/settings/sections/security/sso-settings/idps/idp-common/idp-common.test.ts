/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// scopeUtil.test.ts
import { afterEach, beforeEach, describe, test, expect, vi } from "vitest";
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
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("returns the expiration date from the backend dry run and forwards the alias", async () => {
    const spy = vi.spyOn(idpService, "validateSamlCertificate").mockResolvedValue({
      samlData: { validateSignature: true, certificateExpirationDate: "2030-05-17T00:00:00.000Z" },
    });
    const result = await idpCommon.validateSamlMetadata("xml", "idp.xml", "saml");
    expect(result).toEqual({ expirationDate: new Date("2030-05-17T00:00:00.000Z"), error: "" });
    expect(spy).toHaveBeenCalledWith(expect.objectContaining({ metadataXmlFile: "xml", fileName: "idp.xml" }), "saml");
  });

  test("keeps the date and drops the error when the backend rejects an expired certificate", async () => {
    vi.spyOn(idpService, "validateSamlCertificate").mockResolvedValue({
      samlData: { validateSignature: true, certificateExpirationDate: "2020-01-01T00:00:00.000Z" },
      errorMessage: "invalid saml metadata",
    });
    const result = await idpCommon.validateSamlMetadata("xml", "idp.xml");
    expect(result.expirationDate).toEqual(new Date("2020-01-01T00:00:00.000Z"));
    expect(result.error).toBe("");
  });

  test("returns the backend message when there is no date", async () => {
    vi.spyOn(idpService, "validateSamlCertificate").mockResolvedValue({ errorMessage: "no signing certificate" });
    expect(await idpCommon.validateSamlMetadata("xml", "idp.xml")).toEqual({
      expirationDate: null,
      error: "no signing certificate",
    });
  });

  test("returns an error instead of throwing when the request fails", async () => {
    vi.spyOn(idpService, "validateSamlCertificate").mockRejectedValue(new Error("Network Error"));
    expect(await idpCommon.validateSamlMetadata("xml", "idp.xml")).toEqual({
      expirationDate: null,
      error: "Network Error",
    });
  });
});
