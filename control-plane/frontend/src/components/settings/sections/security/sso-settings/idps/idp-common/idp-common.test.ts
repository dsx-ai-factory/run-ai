/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// scopeUtil.test.ts
import { afterEach, beforeEach, describe, test, expect, vi } from "vitest";
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

describe("isSamlSigningDisabled", () => {
  test("is true when WantAssertionsSigned is false", () => {
    const xml = `<SPSSODescriptor WantAssertionsSigned="false"></SPSSODescriptor>`;
    expect(idpCommon.isSamlSigningDisabled(xml)).toBe(true);
  });

  test("is true when AuthnRequestsSigned is false", () => {
    const xml = `<SPSSODescriptor AuthnRequestsSigned="false"></SPSSODescriptor>`;
    expect(idpCommon.isSamlSigningDisabled(xml)).toBe(true);
  });

  test("is true when WantAuthnRequestsSigned is false", () => {
    const attribute = "WantAuthnRequestsSigned";
    const xml = `<IDPSSODescriptor ${attribute}="false"></IDPSSODescriptor>`;
    expect(idpCommon.isSamlSigningDisabled(xml)).toBe(true);
  });

  test("is false when signing is required", () => {
    const xml = `<SPSSODescriptor WantAssertionsSigned="true" AuthnRequestsSigned="true"></SPSSODescriptor>`;
    expect(idpCommon.isSamlSigningDisabled(xml)).toBe(false);
  });

  test("is false for metadata that does not mention signing", () => {
    expect(idpCommon.isSamlSigningDisabled("<EntityDescriptor></EntityDescriptor>")).toBe(false);
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
