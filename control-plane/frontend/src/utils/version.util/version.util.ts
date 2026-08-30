import compareVersions from "semver-compare";
import { CLUSTER_PR_TO_MASTER_VERSION_PREFIX, TEST_ENV_VERSION } from "@/models/cluster.model";

export function isNewerVersion(versionToTest: string, baseVersion: string): boolean {
  try {
    if (versionToTest?.includes(TEST_ENV_VERSION) || versionToTest.includes(CLUSTER_PR_TO_MASTER_VERSION_PREFIX))
      return true;

    // This is a duplicate code from frontend-v1 versionUtil.ts
    if (versionToTest.slice(0, 1) == "v") versionToTest = versionToTest.slice(1);
    if (versionToTest.includes("-")) versionToTest = versionToTest.slice(0, versionToTest.search("-"));
    if (baseVersion.slice(0, 1) == "v") baseVersion = baseVersion.slice(1);

    const compareVersionsResult = compareVersions(versionToTest, baseVersion);
    return compareVersionsResult >= 0;
  } catch (e) {
    console.warn("version check failed, err: ", e);
    return false;
  }
}

export function isVersionBetween(version: string, minVersion: string, maxVersion: string): boolean {
  if (!version) {
    return false;
  }
  version = _extractVersion(version);
  return compareVersions(version, minVersion) >= 0 && compareVersions(version, maxVersion) < 0;
}

export function isVersionValid(version: string) {
  const semverRegex = /^(\d+)\.(\d+)\.(\d+)(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?(\+[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?$/;
  return semverRegex.test(version);
}

export function getVersionWithoutPatch(version: string): string | undefined {
  if (isVersionValid(version)) {
    const parts = version.split(".");
    return `${parts[0]}.${parts[1]}`;
  }
}

function _extractVersion(version: string): string {
  const rcLocation = version.indexOf("-rc");
  if (rcLocation > -1) {
    version = version.slice(0, rcLocation);
  }
  if (version.slice(0, 1) == "v") {
    version = version.slice(1);
  }
  return version;
}
