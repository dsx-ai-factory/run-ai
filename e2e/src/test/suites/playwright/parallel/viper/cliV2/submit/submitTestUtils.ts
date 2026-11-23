/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from '@playwright/test';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    DeleteFlagsMerged,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    flowRootState,
    logger,
    waitUtils,
    workloadsUtilsV2,
    workloadsCrudUtils,
    workloadsManagerUtils
} from '../../../../../../../modules/helpers';
import { Phase } from '../../../../../../../modules/api/apis/assetsApi';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

export interface PvcValidation {
    claimName: string;
    path: string;
    existingPvc?: boolean;
    readOnly?: boolean;
}

export interface NfsValidation {
    server: string;
    path: string;
    mountPath: string;
    readOnly?: boolean;
}

export interface GitValidation {
    repository: string;
    path: string;
    branch?: string;
}

export interface S3Validation {
    bucket: string;
    path: string;
    url?: string;
}

export interface HostPathValidation {
    path: string;
    mountPath: string;
    readOnly?: boolean;
}

export interface ConfigMapValidation {
    configMap: string;
    mountPath: string;
}

export interface SecretVolumeValidation {
    secret?: string;
    mountPath: string;
}

export interface DataVolumeValidation {
    id: string;
    mountPath?: string;
}

export interface StorageValidation {
    pvc?: PvcValidation[];
    nfs?: NfsValidation[];
    git?: GitValidation[];
    s3?: S3Validation[];
    hostPath?: HostPathValidation[];
    configMapVolume?: ConfigMapValidation[];
    secretVolume?: SecretVolumeValidation[];
    dataVolume?: DataVolumeValidation[];
}

type ActualStorage = {
    pvc?: any[];
    nfs?: any[];
    git?: any[];
    s3?: any[];
    hostPath?: any[];
    configMapVolume?: any[];
    secretVolume?: any[];
    dataVolume?: any[];
};

const validatePvcStorage = (expected: PvcValidation[], actual: ActualStorage | undefined): void => {
    expect(actual?.pvc, 'PVC storage not found in workload').toBeDefined();
    for (const expectedPvc of expected) {
        const foundPvc = actual?.pvc?.find((p) => p.claimName === expectedPvc.claimName);
        expect(foundPvc, `PVC with claimName ${expectedPvc.claimName} not found`).toBeDefined();
        expect(foundPvc?.path).toBe(expectedPvc.path);
        if (expectedPvc.existingPvc !== undefined) {
            expect(foundPvc?.existingPvc).toBe(expectedPvc.existingPvc);
        }
        if (expectedPvc.readOnly !== undefined) {
            expect(foundPvc?.readOnly).toBe(expectedPvc.readOnly);
        }
    }
};

const validateNfsStorage = (expected: NfsValidation[], actual: ActualStorage | undefined): void => {
    expect(actual?.nfs, 'NFS storage not found in workload').toBeDefined();
    for (const expectedNfs of expected) {
        const foundNfs = actual?.nfs?.find((n) => n.mountPath === expectedNfs.mountPath);
        expect(foundNfs, `NFS with mountPath ${expectedNfs.mountPath} not found`).toBeDefined();
        expect(foundNfs?.server).toBe(expectedNfs.server);
        expect(foundNfs?.path).toBe(expectedNfs.path);
        if (expectedNfs.readOnly !== undefined) {
            expect(foundNfs?.readOnly).toBe(expectedNfs.readOnly);
        }
    }
};

const validateGitStorage = (expected: GitValidation[], actual: ActualStorage | undefined): void => {
    expect(actual?.git, 'Git storage not found in workload').toBeDefined();
    for (const expectedGit of expected) {
        const foundGit = actual?.git?.find((g) => g.path === expectedGit.path);
        expect(foundGit, `Git with path ${expectedGit.path} not found`).toBeDefined();
        expect(foundGit?.repository).toBe(expectedGit.repository);
        if (expectedGit.branch !== undefined) {
            expect(foundGit?.branch).toBe(expectedGit.branch);
        }
    }
};

const validateS3Storage = (expected: S3Validation[], actual: ActualStorage | undefined): void => {
    expect(actual?.s3, 'S3 storage not found in workload').toBeDefined();
    for (const expectedS3 of expected) {
        const foundS3 = actual?.s3?.find((s) => s.path === expectedS3.path);
        expect(foundS3, `S3 with path ${expectedS3.path} not found`).toBeDefined();
        expect(foundS3?.bucket).toBe(expectedS3.bucket);
        if (expectedS3.url !== undefined) {
            expect(foundS3?.url).toBe(expectedS3.url);
        }
    }
};

const validateHostPathStorage = (expected: HostPathValidation[], actual: ActualStorage | undefined): void => {
    expect(actual?.hostPath, 'HostPath storage not found in workload').toBeDefined();
    for (const expectedHostPath of expected) {
        const foundHostPath = actual?.hostPath?.find((h) => h.mountPath === expectedHostPath.mountPath);
        expect(
            foundHostPath,
            `HostPath with mountPath ${expectedHostPath.mountPath} not found`
        ).toBeDefined();
        expect(foundHostPath?.path).toBe(expectedHostPath.path);
        if (expectedHostPath.readOnly !== undefined) {
            expect(foundHostPath?.readOnly).toBe(expectedHostPath.readOnly);
        }
    }
};

const validateConfigMapStorage = (
    expected: ConfigMapValidation[],
    actual: ActualStorage | undefined
): void => {
    expect(actual?.configMapVolume, 'ConfigMap storage not found in workload').toBeDefined();
    for (const expectedCm of expected) {
        const foundCm = actual?.configMapVolume?.find((c) => c.mountPath === expectedCm.mountPath);
        expect(foundCm, `ConfigMap with mountPath ${expectedCm.mountPath} not found`).toBeDefined();
        expect(foundCm?.configMap).toBe(expectedCm.configMap);
    }
};

const validateSecretVolumeStorage = (
    expected: SecretVolumeValidation[],
    actual: ActualStorage | undefined
): void => {
    expect(actual?.secretVolume, 'SecretVolume storage not found in workload').toBeDefined();
    for (const expectedSecret of expected) {
        const foundSecret = actual?.secretVolume?.find((s) => s.mountPath === expectedSecret.mountPath);
        expect(
            foundSecret,
            `SecretVolume with mountPath ${expectedSecret.mountPath} not found`
        ).toBeDefined();
        if (expectedSecret.secret !== undefined) {
            expect(foundSecret?.secret).toBe(expectedSecret.secret);
        }
    }
};

const validateDataVolumeStorage = (
    expected: DataVolumeValidation[],
    actual: ActualStorage | undefined
): void => {
    expect(actual?.dataVolume, 'DataVolume storage not found in workload').toBeDefined();
    for (const expectedDv of expected) {
        const foundDv = actual?.dataVolume?.find((dv) => dv.id === expectedDv.id);
        expect(foundDv, `DataVolume with id ${expectedDv.id} not found`).toBeDefined();
        if (expectedDv.mountPath !== undefined) {
            expect(foundDv?.mountPath).toBe(expectedDv.mountPath);
        }
    }
};

const validateStorageFromSpec = (
    expectedStorage: StorageValidation,
    spec: { storage?: ActualStorage } | undefined
): void => {
    const actualStorage = spec?.storage;

    if (expectedStorage.pvc) {
        validatePvcStorage(expectedStorage.pvc, actualStorage);
    }
    if (expectedStorage.nfs) {
        validateNfsStorage(expectedStorage.nfs, actualStorage);
    }
    if (expectedStorage.git) {
        validateGitStorage(expectedStorage.git, actualStorage);
    }
    if (expectedStorage.s3) {
        validateS3Storage(expectedStorage.s3, actualStorage);
    }
    if (expectedStorage.hostPath) {
        validateHostPathStorage(expectedStorage.hostPath, actualStorage);
    }
    if (expectedStorage.configMapVolume) {
        validateConfigMapStorage(expectedStorage.configMapVolume, actualStorage);
    }
    if (expectedStorage.secretVolume) {
        validateSecretVolumeStorage(expectedStorage.secretVolume, actualStorage);
    }
    if (expectedStorage.dataVolume) {
        validateDataVolumeStorage(expectedStorage.dataVolume, actualStorage);
    }
};

export interface FieldsValidation {
    image?: string;
    environmentVariables?: Record<string, string>;
    environmentVariableContains?: Record<string, string>;
    cpuCoreRequest?: number;
    gpuDevicesRequest?: number;
    imagePullSecretNames?: string[];
}

type ManifestContainer = {
    command?: string[];
    env?: Array<{ name?: string; value?: string }>;
    image?: string;
    name?: string;
    resources?: {
        requests?: Record<string, number | string>;
    };
    volumeMounts?: Array<{ mountPath?: string; name?: string; readOnly?: boolean }>;
};

type ManifestPodSpec = {
    containers?: ManifestContainer[];
    initContainers?: ManifestContainer[];
    volumes?: Array<{
        name?: string;
        configMap?: { name?: string };
        hostPath?: { path?: string };
        nfs?: { path?: string; readOnly?: boolean; server?: string };
        persistentVolumeClaim?: { claimName?: string; readOnly?: boolean };
        secret?: { secretName?: string };
    }>;
};

function podSpecsFromManifest(manifest: { [key: string]: any }): ManifestPodSpec[] {
    const podSpecs: ManifestPodSpec[] = [];
    const visit = (value: unknown): void => {
        if (!value || typeof value !== 'object') {
            return;
        }
        const object = value as Record<string, unknown>;
        if (Array.isArray(object.containers)) {
            podSpecs.push(object as ManifestPodSpec);
            return;
        }
        for (const child of Object.values(object)) {
            if (Array.isArray(child)) {
                child.forEach(visit);
            } else {
                visit(child);
            }
        }
    };

    visit(manifest);
    return podSpecs;
}

function mountForVolume(
    podSpec: ManifestPodSpec,
    volumeName: string | undefined,
    mountPath: string
): { mountPath?: string; name?: string; readOnly?: boolean } | undefined {
    return podSpec.containers
        ?.flatMap((container) => container.volumeMounts ?? [])
        .find((mount) => mount.name === volumeName && mount.mountPath === mountPath);
}

function envValue(container: ManifestContainer, name: string): string | undefined {
    return container.env?.find((entry) => entry.name === name)?.value;
}

function quantityToNumber(quantity: number | string | undefined): number | undefined {
    if (quantity === undefined) {
        return undefined;
    }
    if (typeof quantity === 'number') {
        return quantity;
    }

    const suffixes: Record<string, number> = {
        m: 1e-3,
        u: 1e-6,
        n: 1e-9
    };
    const match = quantity.match(/^([+-]?(?:\d+(?:\.\d*)?|\.\d+))([mun]?)$/);
    if (!match) {
        return undefined;
    }
    return Number(match[1]) * (suffixes[match[2]] ?? 1);
}

const validateWorkloadFieldsFromManifest = (
    expectedFields: FieldsValidation,
    manifest: { [key: string]: any }
): void => {
    const containers = podSpecsFromManifest(manifest).flatMap((podSpec) => podSpec.containers ?? []);
    expect(containers.length, 'No workload containers found in workload manifest').toBeGreaterThan(0);

    if (expectedFields.image) {
        expect(
            containers.some((container) => container.image === expectedFields.image),
            `Image ${expectedFields.image} not found in workload manifest`
        ).toBe(true);
    }

    if (expectedFields.environmentVariables) {
        for (const [name, value] of Object.entries(expectedFields.environmentVariables)) {
            const actualValues = containers
                .flatMap((container) => container.env ?? [])
                .filter((entry) => entry.name === name)
                .map((entry) => entry.value);
            expect(
                actualValues.includes(value),
                `Expected env var ${name}=${value}, actual values: ${JSON.stringify(actualValues)}`
            ).toBe(true);
        }
    }

    if (expectedFields.environmentVariableContains) {
        for (const [name, substring] of Object.entries(expectedFields.environmentVariableContains)) {
            const actualValues = containers
                .flatMap((container) => container.env ?? [])
                .filter((entry) => entry.name === name)
                .map((entry) => entry.value ?? '');
            expect(
                actualValues.some((value) => value.toLowerCase().includes(substring.toLowerCase())),
                `Expected env var ${name} to contain ${substring}, actual values: ${JSON.stringify(actualValues)}`
            ).toBe(true);
        }
    }

    const matchesResourceRequest = (resourceName: string, expected: number): boolean => {
        const actualValues = containers
            .map((container) => quantityToNumber(container.resources?.requests?.[resourceName]))
            .filter((value): value is number => value !== undefined);
        return (expected === 0 && actualValues.length === 0) || actualValues.includes(expected);
    };

    if (expectedFields.cpuCoreRequest !== undefined) {
        expect(
            matchesResourceRequest('cpu', expectedFields.cpuCoreRequest),
            `CPU request ${expectedFields.cpuCoreRequest} not found in workload manifest`
        ).toBe(true);
    }
    if (expectedFields.gpuDevicesRequest !== undefined) {
        expect(
            matchesResourceRequest('nvidia.com/gpu', expectedFields.gpuDevicesRequest),
            `GPU request ${expectedFields.gpuDevicesRequest} not found in workload manifest`
        ).toBe(true);
    }
};

const validateStorageFromManifest = (
    expectedStorage: StorageValidation,
    manifest: { [key: string]: any }
): void => {
    const podSpecs = podSpecsFromManifest(manifest);
    expect(podSpecs.length, 'No pod specifications found in workload manifest').toBeGreaterThan(0);

    for (const expected of expectedStorage.pvc ?? []) {
        const match = podSpecs.some((podSpec) => {
            const volume = podSpec.volumes?.find(
                (candidate) => candidate.persistentVolumeClaim?.claimName === expected.claimName
            );
            const mount = mountForVolume(podSpec, volume?.name, expected.path);
            const readOnly = mount?.readOnly ?? volume?.persistentVolumeClaim?.readOnly ?? false;
            // The rendered pod records the resulting claim and mount, not whether the claim was pre-existing.
            return !!volume && !!mount && (expected.readOnly === undefined || readOnly === expected.readOnly);
        });
        expect(match, `PVC ${expected.claimName} mounted at ${expected.path} not found`).toBe(true);
    }

    for (const expected of expectedStorage.nfs ?? []) {
        const match = podSpecs.some((podSpec) => {
            const volume = podSpec.volumes?.find(
                (candidate) =>
                    candidate.nfs?.server === expected.server && candidate.nfs?.path === expected.path
            );
            const mount = mountForVolume(podSpec, volume?.name, expected.mountPath);
            const readOnly = mount?.readOnly ?? volume?.nfs?.readOnly ?? false;
            return !!volume && !!mount && (expected.readOnly === undefined || readOnly === expected.readOnly);
        });
        expect(match, `NFS ${expected.server}:${expected.path} at ${expected.mountPath} not found`).toBe(
            true
        );
    }

    for (const expected of expectedStorage.hostPath ?? []) {
        const match = podSpecs.some((podSpec) => {
            const volume = podSpec.volumes?.find((candidate) => candidate.hostPath?.path === expected.path);
            const mount = mountForVolume(podSpec, volume?.name, expected.mountPath);
            return (
                !!volume &&
                !!mount &&
                (expected.readOnly === undefined || !!mount.readOnly === expected.readOnly)
            );
        });
        expect(match, `HostPath ${expected.path} at ${expected.mountPath} not found`).toBe(true);
    }

    for (const expected of expectedStorage.configMapVolume ?? []) {
        const match = podSpecs.some((podSpec) => {
            const volume = podSpec.volumes?.find(
                (candidate) => candidate.configMap?.name === expected.configMap
            );
            return !!volume && !!mountForVolume(podSpec, volume.name, expected.mountPath);
        });
        expect(match, `ConfigMap ${expected.configMap} at ${expected.mountPath} not found`).toBe(true);
    }

    for (const expected of expectedStorage.secretVolume ?? []) {
        const match = podSpecs.some((podSpec) => {
            const volume = podSpec.volumes?.find(
                (candidate) =>
                    candidate.secret !== undefined &&
                    (expected.secret === undefined || candidate.secret.secretName === expected.secret)
            );
            return !!volume && !!mountForVolume(podSpec, volume.name, expected.mountPath);
        });
        expect(match, `Secret volume at ${expected.mountPath} not found`).toBe(true);
    }

    for (const expected of expectedStorage.git ?? []) {
        const match = podSpecs.some((podSpec) => {
            const workloadVolumeNames = new Set(
                (podSpec.containers ?? [])
                    .flatMap((container) => container.volumeMounts ?? [])
                    .filter((mount) => mount.mountPath === expected.path)
                    .map((mount) => mount.name)
            );
            return (podSpec.initContainers ?? []).some((container) => {
                const sharesWorkloadVolume = container.volumeMounts?.some((mount) =>
                    workloadVolumeNames.has(mount.name)
                );
                return (
                    container.name?.startsWith('git-sync') &&
                    envValue(container, 'GITSYNC_REPO') === expected.repository &&
                    (expected.branch === undefined ||
                        envValue(container, 'GITSYNC_REF') === expected.branch) &&
                    sharesWorkloadVolume
                );
            });
        });
        expect(match, `Git repository ${expected.repository} at ${expected.path} not found`).toBe(true);
    }

    for (const expected of expectedStorage.s3 ?? []) {
        const match = podSpecs.some((podSpec) => {
            const workloadVolumeNames = new Set(
                (podSpec.containers ?? [])
                    .flatMap((container) => container.volumeMounts ?? [])
                    .filter((mount) => mount.mountPath === expected.path)
                    .map((mount) => mount.name)
            );
            return (podSpec.initContainers ?? []).some((container) => {
                const command = container.command ?? [];
                return (
                    container.name?.startsWith('s3-fs') &&
                    command.includes(expected.bucket) &&
                    (expected.url === undefined || command.includes(expected.url)) &&
                    container.volumeMounts?.some((mount) => workloadVolumeNames.has(mount.name))
                );
            });
        });
        expect(match, `S3 bucket ${expected.bucket} at ${expected.path} not found`).toBe(true);
    }

    expect(
        expectedStorage.dataVolume ?? [],
        'DataVolume storage cannot yet be validated from a rendered workload manifest'
    ).toHaveLength(0);
};

export interface AssetValidation {
    environmentAssetName?: string;
    computeAssetName?: string;
    datasourceAssetName?: string;
}

type WorkloadWithAssets = {
    environments?: { name: string }[];
    computes?: { name: string }[];
    datasources?: { name: string }[];
};

const validateEnvironmentAsset = (expectedName: string, workload: WorkloadWithAssets): void => {
    const envAssets = workload.environments || [];
    logger.info(`Environment assets from API (full): ${JSON.stringify(envAssets)}`);
    const found = envAssets.some((e) => e.name === expectedName);
    expect(
        found,
        `Environment asset ${expectedName} not found. Available: ${JSON.stringify(envAssets)}`
    ).toBe(true);
};

const validateComputeAsset = (expectedName: string, workload: WorkloadWithAssets): void => {
    const computeAssets = workload.computes || [];
    const found = computeAssets.some((c) => c.name === expectedName);
    expect(
        found,
        `Compute asset ${expectedName} not found. Available: ${JSON.stringify(computeAssets.map((c) => c.name))}`
    ).toBe(true);
};

const validateDatasourceAsset = (expectedName: string, workload: WorkloadWithAssets): void => {
    const dsAssets = workload.datasources || [];
    const found = dsAssets.some((d) => d.name === expectedName);
    expect(
        found,
        `Datasource asset ${expectedName} not found. Available: ${JSON.stringify(dsAssets.map((d) => d.name))}`
    ).toBe(true);
};

const validateAssets = (expectedAssets: AssetValidation, workload: WorkloadWithAssets): void => {
    if (expectedAssets.environmentAssetName) {
        validateEnvironmentAsset(expectedAssets.environmentAssetName, workload);
    }
    if (expectedAssets.computeAssetName) {
        validateComputeAsset(expectedAssets.computeAssetName, workload);
    }
    if (expectedAssets.datasourceAssetName) {
        validateDatasourceAsset(expectedAssets.datasourceAssetName, workload);
    }
};

type WorkloadWithFields = {
    images?: string[];
    environmentVariables?: Record<string, string>;
    workloadRequestedResources?: {
        cpu?: { request?: number };
        gpu?: { request?: number };
    };
};

const validateWorkloadFields = (expectedFields: FieldsValidation, workload: WorkloadWithFields): void => {
    if (expectedFields.image) {
        expect(workload.images).toContain(expectedFields.image);
    }
    if (expectedFields.environmentVariables) {
        for (const [key, value] of Object.entries(expectedFields.environmentVariables)) {
            expect(
                workload.environmentVariables?.[`${key}`],
                `Expected env var ${key}=${value}, actual: ${workload.environmentVariables?.[`${key}`]}`
            ).toBe(value);
        }
    }
    if (expectedFields.environmentVariableContains) {
        for (const [key, substring] of Object.entries(expectedFields.environmentVariableContains)) {
            const actual = workload.environmentVariables?.[key];
            expect(actual, `Env var ${key} should exist`).toBeDefined();
            expect(
                actual!.toLowerCase(),
                `Env var ${key} should contain "${substring}". Actual: ${actual}`
            ).toContain(substring.toLowerCase());
        }
    }
    if (expectedFields.cpuCoreRequest !== undefined) {
        expect(workload.workloadRequestedResources?.cpu?.request).toBe(expectedFields.cpuCoreRequest);
    }
    if (expectedFields.gpuDevicesRequest !== undefined) {
        expect(workload.workloadRequestedResources?.gpu?.request).toBe(expectedFields.gpuDevicesRequest);
    }
};

export interface WorkloadValidation {
    fields?: FieldsValidation;
    assets?: AssetValidation;
    storage?: StorageValidation;
    runningPhaseTimeoutSec?: number;
}

export const submitWorkload = (
    workloadName: string,
    submitFlags: Partial<typeof SubmitFlagsMerged>,
    workloadType: cliV2WorkloadType
): void => {
    const result = runaiCliV2.submit(submitFlags, workloadName, workloadType);
    expect(result).toContain(`Creating ${workloadType} ${workloadName}`);
};

export const validateWorkload = async (
    workloadName: string,
    projectId: number,
    expectedValues: WorkloadValidation,
    workloadType: cliV2WorkloadType = cliV2WorkloadType.training,
    shouldWaitForRunningPhase = false
): Promise<void> => {
    if (shouldWaitForRunningPhase) {
        await waitUtils.waitForBackendWorkloadPhase(workloadName, projectId, Phase.Running, 300);
    }

    // Wait for workload to appear in control plane before validation
    const workloadId = await workloadsUtilsV2.waitForWorkloadIdByWorkloadNameAndProjectId(
        workloadName,
        projectId.toString()
    );
    const workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
    let submissionType = AnyWorkloadSubmissionType.Distributed;
    if (workloadType === cliV2WorkloadType.workspace) {
        submissionType = AnyWorkloadSubmissionType.Workspace;
    } else if (workloadType === cliV2WorkloadType.training) {
        submissionType = AnyWorkloadSubmissionType.Training;
    } else if (
        workloadType === cliV2WorkloadType.inference ||
        workloadType === cliV2WorkloadType.inferenceNim
    ) {
        submissionType = AnyWorkloadSubmissionType.Inference;
    }
    const redirectsToWorkloadsManager =
        workloadType !== cliV2WorkloadType.inference &&
        (await isAnyWorkloadSubmissionEnabled(submissionType));
    let workloadManifest: { [key: string]: any } | undefined;
    if (redirectsToWorkloadsManager && (expectedValues.fields || expectedValues.storage)) {
        expect(workload.sourceApi).toBe('WorkloadsV2');
        const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
        workloadManifest = getResult.body.workload.workloadManifest;
    }

    logger.info(`validating workload ${workloadName} expected fields: ${JSON.stringify(expectedValues)}`);

    if (expectedValues.fields) {
        if (workloadManifest) {
            validateWorkloadFieldsFromManifest(expectedValues.fields, workloadManifest);
        } else {
            validateWorkloadFields(expectedValues.fields, workload);
        }
    }

    if (expectedValues.assets) {
        validateAssets(expectedValues.assets, workload);
    }

    const distributedTrainingTypes: cliV2WorkloadType[] = [
        cliV2WorkloadType.mpi,
        cliV2WorkloadType.pyTorch,
        cliV2WorkloadType.tensorFlow,
        cliV2WorkloadType.xgBoost,
        cliV2WorkloadType.jax
    ];

    if (expectedValues.storage) {
        let storageToValidateFromSpec = expectedValues.storage;
        if (workloadManifest) {
            const { dataVolume, ...renderedStorage } = expectedValues.storage;
            validateStorageFromManifest(renderedStorage, workloadManifest);
            if (!dataVolume?.length) {
                return;
            }
            // Translation replaces the DataVolume ID with a resolved PVC in the rendered manifest.
            // Validate that mount above, then validate the original ID from the persisted source request.
            storageToValidateFromSpec = { dataVolume };
        }

        workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        let spec: { storage?: ActualStorage } | undefined;
        if (distributedTrainingTypes.includes(workloadType)) {
            spec = (await workloadsCrudUtils.getDistributed(workloadId)).body.spec;
        } else if (workloadType === cliV2WorkloadType.training) {
            spec = (await workloadsCrudUtils.getTraining(workloadId)).body.spec;
        } else if (workloadType === cliV2WorkloadType.workspace) {
            spec = (await workloadsCrudUtils.getWorkspace(workloadId)).body.spec;
        } else if (workloadType === cliV2WorkloadType.inference) {
            spec = (await workloadsCrudUtils.getInference(workloadId)).body.spec;
        }

        validateStorageFromSpec(storageToValidateFromSpec, spec);
    }
};

export const deleteWorkload = (
    workloadName: string,
    projectName: string,
    workloadType: cliV2WorkloadType
): void => {
    logger.info(`deleting workload ${workloadName}`);
    runaiCliV2.delete(
        { [DeleteFlagsMerged.project]: projectName } as Partial<typeof DeleteFlagsMerged>,
        workloadName,
        workloadType
    );
};
