import { describe, it, expect } from "vitest";
import { isNewerVersion, isVersionValid, getVersionWithoutPatch } from "./version.util";

describe("version.util", () => {
  it("isNewerVersion", () => {
    expect(isNewerVersion("2.16", "2.15")).true;
    expect(isNewerVersion("2.16.8", "2.15")).true;
    expect(isNewerVersion("2.15", "2.15")).true;
    expect(isNewerVersion("2.15.0", "2.15")).true;
    expect(isNewerVersion("2.16", "2.15.9")).true;
    expect(isNewerVersion("2.14", "2.15")).false;
    expect(isNewerVersion("2.14.1", "2.15")).false;
    expect(isNewerVersion("2.14.8", "2.15.9")).false;
  });

  it("should return true for valid semantic version strings", () => {
    const validVersions = ["1.0.0", "1.0.0-alpha", "1.0.0-alpha.1", "2.1.0"];

    validVersions.forEach((version) => {
      expect(isVersionValid(version)).toBe(true);
    });
  });

  it("should return false for invalid semantic version strings", () => {
    const invalidVersions = ["1.0", "1.0.0.1", "1.0.0-", "v1.0.0"];

    invalidVersions.forEach((version) => {
      expect(isVersionValid(version)).toBe(false);
    });
  });

  it("should return the major and minor version without patch", () => {
    expect(getVersionWithoutPatch("1.2.3")).toBe("1.2");
    expect(getVersionWithoutPatch("10.20.30")).toBe("10.20");
    expect(getVersionWithoutPatch("1.0.0")).toBe("1.0");
    expect(getVersionWithoutPatch("2.5.9")).toBe("2.5");
    expect(getVersionWithoutPatch("1.2.3-alpha+001")).toBe("1.2");
    expect(getVersionWithoutPatch("1.2.3+build.5678")).toBe("1.2");
  });

  it("should return undefined for invalid version strings", () => {
    expect(getVersionWithoutPatch("1.2")).toBeUndefined();
    expect(getVersionWithoutPatch("1.2.3.4")).toBeUndefined();
    expect(getVersionWithoutPatch("v1.2.3")).toBeUndefined();
    expect(getVersionWithoutPatch("1.2.3-alpha-beta")).toBeUndefined();
    expect(getVersionWithoutPatch("abc.def.ghi")).toBeUndefined();
  });

  it("should handle edge cases", () => {
    expect(getVersionWithoutPatch("0.0.0")).toBe("0.0");
    expect(getVersionWithoutPatch("1.0.0-alpha")).toBe("1.0");
    expect(getVersionWithoutPatch("2.3.4+build.1234")).toBe("2.3");
  });
});
