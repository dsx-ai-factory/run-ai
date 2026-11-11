/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { versionUtils } from '../index';

export enum SupportedVersions {
    V2_9 = '2.9',
    V2_10 = '2.10',
    V2_11 = '2.11',
    V2_12 = '2.12',
    V2_13 = '2.13',
    V2_14 = '2.14',
    V2_15 = '2.15',
    V2_16 = '2.16',
    V2_17 = '2.17',
    V2_18 = '2.18',
    V2_19 = '2.19',
    V2_20 = '2.20',
    V2_23 = '2.23',
    V2_24 = '2.24',
    V2_25 = '2.25',
    V2_26 = '2.26',
    V3_0 = '3.0'
}

export enum Platform {
    OPENSHIFT = 'openshift',
    RKE2 = 'rke2',
    EKS = 'eks',
    GKE = 'gke',
    AKS = 'aks',
    KIND = 'kind',
    GCP = 'gcp'
}

function isPlatformFit(unsupportedPlatforms?: Platform[]): boolean {
    return process.env.CLUSTER_PLATFORM
        ? !unsupportedPlatforms?.includes(process.env.CLUSTER_PLATFORM as Platform)
        : true;
}

export function supportedFrom(minSupportedVersion: string, excludePlatforms?: Platform[]) {
    return function (testFunction: () => void): void {
        const clusterVersion = process.env.OPERATOR_VERSION;
        const clusterVersionFit = clusterVersion
            ? versionUtils.isVersionSupported(clusterVersion, minSupportedVersion)
            : true;
        if (clusterVersionFit && isPlatformFit(excludePlatforms)) {
            testFunction();
        }
    };
}

export function areNativeWorkloadCrdsSupported(): boolean {
    // Test registration happens before flowRootState is initialized, so use the version supplied
    // by the test runner. Master environments may leave OPERATOR_VERSION empty, so missing versions
    // fail closed; explicit master versions are also treated as newer than the CRD retirement.
    const clusterVersion = process.env.OPERATOR_VERSION;
    return clusterVersion ? !versionUtils.isVersionSupported(clusterVersion, SupportedVersions.V3_0) : false;
}

export function unsupportedPlatforms(excludePlatforms?: Platform[]) {
    return function (testFunction: () => void): void {
        if (isPlatformFit(excludePlatforms)) {
            testFunction();
        }
    };
}

/**
 * Async-safe version of unsupportedPlatforms that properly awaits async test functions.
 * Use this when your test function is async to avoid orphaned promises.
 */
export function unsupportedPlatformsAsync(excludePlatforms?: Platform[]) {
    return async function (testFunction: () => Promise<void>): Promise<void> {
        if (isPlatformFit(excludePlatforms)) {
            await testFunction();
        }
    };
}

export function shouldRunTests(condition: boolean) {
    return function (testFunction: () => void): void {
        if (condition) {
            testFunction();
        }
    };
}

export function shouldRunNativeWorkloadCrdTests(testFunction: () => void): void {
    shouldRunTests(areNativeWorkloadCrdsSupported())(testFunction);
}

export function supportedPlatforms(platforms?: Platform[]) {
    return function (testFunction: () => void): void {
        const clusterPlatformFit = process.env.CLUSTER_PLATFORM
            ? platforms?.includes(process.env.CLUSTER_PLATFORM as Platform)
            : true;
        if (clusterPlatformFit) {
            testFunction();
        }
    };
}
