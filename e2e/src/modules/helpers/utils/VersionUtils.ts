/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import compareVersions from 'semver-compare';
import { load } from 'js-yaml';

const kubernetesJobResumeFixPatchByMinor = new Map<number, number>([
    [32, 10],
    [33, 6],
    [34, 2]
]);

export class VersionUtils {
    private extractVersion(version: string): string {
        const rcLocation = version.indexOf('-rc');
        if (rcLocation > -1) {
            version = version.slice(0, rcLocation);
        }
        if (version.slice(0, 1) === 'v') {
            version = version.slice(1);
        }
        const stagingLocation = version.indexOf('-staging');
        if (stagingLocation > -1) {
            version = version.slice(0, stagingLocation);
        }
        return version;
    }

    /**
     * Returns true when the given version string represents a master / dev build.
     * Master builds are identified the same way as inside `isVersionSupported`:
     * either they start with the `0.0.0` placeholder version, or their tag
     * contains `master`.
     */
    isMasterBuild(version: string | undefined): boolean {
        if (!version) return false;
        return version.startsWith('0.0.0') || version.includes('master');
    }

    /**
     * @param version - Use flowRootState.cluster.version instead of process.env.OPERATOR_VERSION
     * @param minimumVersion - Minimum version required for support
     */
    isVersionSupported(version: string, minimumVersion: string): boolean {
        if (!version) {
            return false;
        }
        version = this.extractVersion(version);
        minimumVersion = this.extractVersion(minimumVersion);
        return (
            version.startsWith('0.0.0') ||
            version.includes('master') ||
            compareVersions(version, minimumVersion) >= 0
        );
    }

    /**
     * Kubernetes 1.32-1.34 can reject Job status updates after a suspended Job is resumed.
     * Fixed in 1.32.10, 1.33.6, and 1.34.2.
     * See https://github.com/kubernetes/kubernetes/issues/134521.
     */
    isKubernetesVersionAffectedByJobResumeBug(version: string): boolean {
        const versionParts = version.match(/^v?(\d+)\.(\d+)\.(\d+)/);
        if (!versionParts) {
            throw new Error(`Unable to parse Kubernetes version: ${version}`);
        }

        const [, majorPart, minorPart, patchPart] = versionParts;
        const major = Number(majorPart);
        const minor = Number(minorPart);
        const patch = Number(patchPart);
        const fixedPatch = kubernetesJobResumeFixPatchByMinor.get(minor);

        return major === 1 && fixedPatch !== undefined && patch < fixedPatch;
    }

    async getLatestTagVersionFromYAML(): Promise<string> {
        const yamlURL = 'https://helm.ngc.nvidia.com/nvidia/runai-oss/index.yaml';
        const response = await fetch(yamlURL);
        if (!response.ok) {
            throw new Error(`Unable to fetch the public cluster chart index: HTTP ${response.status}`);
        }
        const index = load(await response.text()) as {
            entries?: Record<string, Array<{ version: string }>>;
        };
        const versions = (index?.entries?.['runai-cluster'] ?? [])
            .map((chart) => chart.version)
            .filter((version) => /^\d+\.\d+\.\d+$/.test(version));
        if (versions.length === 0) {
            throw new Error('Public chart index contains no released runai-cluster versions');
        }
        return versions.sort(compareVersions).at(-1);
    }

    async getMajorMinorVersion(): Promise<string> {
        const latestVersionTag = await this.getLatestTagVersionFromYAML();
        const [major, minor] = latestVersionTag.split('.');
        return `${major}.${minor}`;
    }
}
