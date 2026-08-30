// scopeUtil.test.ts
import { describe, test, expect } from "vitest";
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
