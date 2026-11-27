/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    test as baseTest,
    TestInfo,
    PlaywrightTestArgs,
    PlaywrightTestOptions,
    PlaywrightWorkerArgs,
    PlaywrightWorkerOptions
} from '@playwright/test';
import { logger } from '../../modules/helpers';
import { AllureUtils } from '../../modules/helpers/utils/AllureUtils';
import { crds } from '../../modules/helpers';

enum ResourceType {
    SYSTEM_LOGS = 'system-logs',
    CRD_YAML = 'crd-yaml',
    PODS_YAML = 'pods-yaml',
    PODS_LOGS = 'pods-logs'
}

interface ResourceCollectionRequest {
    type: ResourceType;
    params: any;
    collector: () => Promise<void>;
}
export const SYSTEM_COMPONENTS = {
    WORKLOAD_CONTROLLER: {
        name: 'workload-controller',
        displayName: 'Workload Controller Logs'
    },
    AGENT: {
        name: 'agent',
        displayName: 'Agent Logs'
    },
    CLUSTER_SYNC: {
        name: 'cluster-sync',
        displayName: 'Cluster Sync Logs'
    },
    WORKLOAD_OVERSEER: {
        name: 'workload-overseer',
        displayName: 'Workload Overseer Logs'
    },
    RUNAI_JOB_CONTROLLER: {
        name: 'runai-job-controller',
        displayName: 'RunAI Job Controller Logs'
    },
    MPI_OPERATOR: {
        name: 'mpi-operator',
        displayName: 'MPI Operator Logs'
    },
    TRAINING_OPERATOR: {
        name: 'training-operator',
        displayName: 'Training Operator Logs'
    },
    WORKLOADS_SERVICE: {
        name: 'workloads-service',
        displayName: 'Workloads Service Logs'
    },
    INFERENCE_WORKLOAD_CONTROLLER: {
        name: 'inference-workload-controller',
        displayName: 'Inference Workload Controller Logs'
    },
    POD_GROUP_ASSIGNER: {
        name: 'pod-group-assigner',
        displayName: 'Pod Group Assigner Logs'
    },
    NODEPOOL_CONTROLLER: {
        name: 'nodepool-controller',
        displayName: 'NodePool Controller Logs'
    }
} as const;

// Basic/Atomic debug tags - each collects a single type of resource
export const BASIC_DEBUG_TAGS = {
    // Individual system component logs
    WORKLOAD_CONTROLLER_LOGS: '@debug-workload-controller-logs',
    AGENT_LOGS: '@debug-agent-logs',
    CLUSTER_SYNC_LOGS: '@debug-cluster-sync-logs',
    WORKLOAD_OVERSEER_LOGS: '@debug-workload-overseer-logs',
    RUNAI_JOB_CONTROLLER_LOGS: '@debug-runai-job-controller-logs',
    MPI_OPERATOR_LOGS: '@debug-mpi-operator-logs',
    TRAINING_OPERATOR_LOGS: '@debug-training-operator-logs',
    WORKLOADS_SERVICE_LOGS: '@debug-workloads-service-logs',
    INFERENCE_WORKLOAD_CONTROLLER_LOGS: '@debug-inference-workload-controller-logs',
    POD_GROUP_ASSIGNER_LOGS: '@debug-pod-group-assigner-logs',
    NODEPOOL_CONTROLLER_LOGS: '@debug-nodepool-controller-logs',

    // Individual CRD collections
    RUNAI_JOBS: '@debug-runai-jobs',
    PODS_YAML: '@debug-pods-yaml',
    PODS_LOGS: '@debug-pods-logs',
    TRAINING_WORKLOADS: '@debug-training-workloads',
    INTERACTIVE_WORKLOADS: '@debug-interactive-workloads',
    DISTRIBUTED_WORKLOADS: '@debug-distributed-workloads',
    INFERENCE_WORKLOADS: '@debug-inference-workloads',
    XGBOOST_JOBS: '@debug-xgboost-jobs',
    TF_JOBS: '@debug-tf-jobs',
    MPI_JOBS: '@debug-mpi-jobs',
    PYTORCH_JOBS: '@debug-pytorch-jobs',
    JAX_JOBS: '@debug-jax-jobs',
    KNATIVE_SERVICES: '@debug-knative-services'
} as const;

// Composite debug tags as string constants
export const COMPOSITE_DEBUG_TAGS = {
    CORE_SYSTEM_LOGS: '@debug-core-system-logs',
    EXTENDED_SYSTEM_LOGS: '@debug-extended-system-logs',
    SYSTEM_LOGS: '@debug-system-logs',
    PODS: '@debug-pods',
    TRAINING_WORKLOAD_CRDS: '@debug-training-workload-crds',
    INTERACTIVE_WORKLOAD_CRDS: '@debug-interactive-workload-crds',
    DISTRIBUTED_WORKLOAD_CRDS: '@debug-distributed-workload-crds',
    INFERENCE_WORKLOAD_CRDS: '@debug-inference-workload-crds',
    ALL_KUBEFLOW_JOBS: '@debug-all-kubeflow-jobs',
    ALL_CRDS: '@debug-all-crds',
    FULL_DEBUG: '@debug-full-debug'
} as const;

// Unified debug tags export - contains all available debug tags
export const DEBUG_TAGS = {
    ...BASIC_DEBUG_TAGS,
    ...COMPOSITE_DEBUG_TAGS
} as const;

// Composite debug tags - defined as lists of basic tags
export const COMPOSITE_DEBUG_TAGS_DEF = {
    // System logs combinations
    CORE_SYSTEM_LOGS: [
        BASIC_DEBUG_TAGS.WORKLOAD_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.AGENT_LOGS,
        BASIC_DEBUG_TAGS.CLUSTER_SYNC_LOGS
    ],
    EXTENDED_SYSTEM_LOGS: [
        BASIC_DEBUG_TAGS.WORKLOAD_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.AGENT_LOGS,
        BASIC_DEBUG_TAGS.CLUSTER_SYNC_LOGS,
        BASIC_DEBUG_TAGS.WORKLOAD_OVERSEER_LOGS,
        BASIC_DEBUG_TAGS.RUNAI_JOB_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.POD_GROUP_ASSIGNER_LOGS,
        BASIC_DEBUG_TAGS.NODEPOOL_CONTROLLER_LOGS
    ],
    SYSTEM_LOGS: [
        BASIC_DEBUG_TAGS.WORKLOAD_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.AGENT_LOGS,
        BASIC_DEBUG_TAGS.CLUSTER_SYNC_LOGS,
        BASIC_DEBUG_TAGS.WORKLOAD_OVERSEER_LOGS,
        BASIC_DEBUG_TAGS.RUNAI_JOB_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.MPI_OPERATOR_LOGS,
        BASIC_DEBUG_TAGS.TRAINING_OPERATOR_LOGS,
        BASIC_DEBUG_TAGS.WORKLOADS_SERVICE_LOGS,
        BASIC_DEBUG_TAGS.INFERENCE_WORKLOAD_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.POD_GROUP_ASSIGNER_LOGS,
        BASIC_DEBUG_TAGS.NODEPOOL_CONTROLLER_LOGS
    ],

    // Pods combination
    PODS: [BASIC_DEBUG_TAGS.PODS_YAML, BASIC_DEBUG_TAGS.PODS_LOGS],

    // Workload-specific CRD combinations
    TRAINING_WORKLOAD_CRDS: [
        BASIC_DEBUG_TAGS.TRAINING_WORKLOADS,
        BASIC_DEBUG_TAGS.RUNAI_JOBS,
        BASIC_DEBUG_TAGS.PODS_YAML,
        BASIC_DEBUG_TAGS.PODS_LOGS
    ],
    INTERACTIVE_WORKLOAD_CRDS: [
        BASIC_DEBUG_TAGS.INTERACTIVE_WORKLOADS,
        BASIC_DEBUG_TAGS.RUNAI_JOBS,
        BASIC_DEBUG_TAGS.PODS_YAML,
        BASIC_DEBUG_TAGS.PODS_LOGS
    ],
    DISTRIBUTED_WORKLOAD_CRDS: [
        BASIC_DEBUG_TAGS.DISTRIBUTED_WORKLOADS,
        BASIC_DEBUG_TAGS.XGBOOST_JOBS,
        BASIC_DEBUG_TAGS.TF_JOBS,
        BASIC_DEBUG_TAGS.MPI_JOBS,
        BASIC_DEBUG_TAGS.PYTORCH_JOBS,
        BASIC_DEBUG_TAGS.JAX_JOBS,
        BASIC_DEBUG_TAGS.PODS_YAML,
        BASIC_DEBUG_TAGS.PODS_LOGS
    ],
    INFERENCE_WORKLOAD_CRDS: [BASIC_DEBUG_TAGS.INFERENCE_WORKLOADS, BASIC_DEBUG_TAGS.KNATIVE_SERVICES],

    // Kubeflow jobs group
    ALL_KUBEFLOW_JOBS: [
        BASIC_DEBUG_TAGS.XGBOOST_JOBS,
        BASIC_DEBUG_TAGS.TF_JOBS,
        BASIC_DEBUG_TAGS.MPI_JOBS,
        BASIC_DEBUG_TAGS.PYTORCH_JOBS,
        BASIC_DEBUG_TAGS.JAX_JOBS
    ],

    // Composite collections
    ALL_CRDS: [
        BASIC_DEBUG_TAGS.TRAINING_WORKLOADS,
        BASIC_DEBUG_TAGS.INTERACTIVE_WORKLOADS,
        BASIC_DEBUG_TAGS.DISTRIBUTED_WORKLOADS,
        BASIC_DEBUG_TAGS.INFERENCE_WORKLOADS,
        BASIC_DEBUG_TAGS.RUNAI_JOBS,
        BASIC_DEBUG_TAGS.XGBOOST_JOBS,
        BASIC_DEBUG_TAGS.TF_JOBS,
        BASIC_DEBUG_TAGS.MPI_JOBS,
        BASIC_DEBUG_TAGS.PYTORCH_JOBS,
        BASIC_DEBUG_TAGS.JAX_JOBS,
        BASIC_DEBUG_TAGS.KNATIVE_SERVICES
    ],

    // Full debug collection - all system logs + all CRDs + pods
    FULL_DEBUG: [
        // All system logs
        BASIC_DEBUG_TAGS.WORKLOAD_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.AGENT_LOGS,
        BASIC_DEBUG_TAGS.CLUSTER_SYNC_LOGS,
        BASIC_DEBUG_TAGS.WORKLOAD_OVERSEER_LOGS,
        BASIC_DEBUG_TAGS.RUNAI_JOB_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.MPI_OPERATOR_LOGS,
        BASIC_DEBUG_TAGS.TRAINING_OPERATOR_LOGS,
        BASIC_DEBUG_TAGS.WORKLOADS_SERVICE_LOGS,
        BASIC_DEBUG_TAGS.INFERENCE_WORKLOAD_CONTROLLER_LOGS,
        BASIC_DEBUG_TAGS.POD_GROUP_ASSIGNER_LOGS,
        BASIC_DEBUG_TAGS.NODEPOOL_CONTROLLER_LOGS,
        // All CRDs
        BASIC_DEBUG_TAGS.TRAINING_WORKLOADS,
        BASIC_DEBUG_TAGS.INTERACTIVE_WORKLOADS,
        BASIC_DEBUG_TAGS.DISTRIBUTED_WORKLOADS,
        BASIC_DEBUG_TAGS.INFERENCE_WORKLOADS,
        BASIC_DEBUG_TAGS.RUNAI_JOBS,
        BASIC_DEBUG_TAGS.XGBOOST_JOBS,
        BASIC_DEBUG_TAGS.TF_JOBS,
        BASIC_DEBUG_TAGS.MPI_JOBS,
        BASIC_DEBUG_TAGS.PYTORCH_JOBS,
        BASIC_DEBUG_TAGS.JAX_JOBS,
        BASIC_DEBUG_TAGS.KNATIVE_SERVICES,
        // Pods
        BASIC_DEBUG_TAGS.PODS_YAML,
        BASIC_DEBUG_TAGS.PODS_LOGS
    ]
} as const;

function createSystemLogRequests(
    components: string[],
    logTailLines: number = 500
): ResourceCollectionRequest[] {
    return components.map((component) => ({
        type: ResourceType.SYSTEM_LOGS,
        params: { component, logTailLines },
        collector: async (): Promise<void> => {
            try {
                const allureUtils = new AllureUtils();
                const componentInfo = Object.values(SYSTEM_COMPONENTS).find(
                    (comp) => comp.name === component
                );
                const displayName = componentInfo?.displayName;
                await allureUtils.attachSystemComponentLogs(component, displayName, 'runai', logTailLines);
            } catch (error) {
                logger.error(`Failed to collect ${component} logs:`, error);
            }
        }
    }));
}

function createCRDRequests(crdList: any[]): ResourceCollectionRequest[] {
    return crdList.map((crd) => ({
        type: ResourceType.CRD_YAML,
        params: { kind: crd.kind, group: crd.group || 'core' },
        collector: async (): Promise<void> => {
            try {
                const allureUtils = new AllureUtils();
                await allureUtils.attachCrdYaml(crd);
            } catch (error) {
                logger.error(`Failed to collect ${crd.kind} CRD:`, error);
            }
        }
    }));
}

interface DebugCollectionStrategy {
    getResourceRequests: () => ResourceCollectionRequest[];
}

// Basic debug collection strategies - only for atomic/basic tags
const BASIC_DEBUG_COLLECTION_STRATEGIES: Record<string, DebugCollectionStrategy> = {
    // Individual system components
    [BASIC_DEBUG_TAGS.WORKLOAD_CONTROLLER_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.WORKLOAD_CONTROLLER.name])
    },
    [BASIC_DEBUG_TAGS.AGENT_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.AGENT.name])
    },
    [BASIC_DEBUG_TAGS.CLUSTER_SYNC_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.CLUSTER_SYNC.name])
    },
    [BASIC_DEBUG_TAGS.WORKLOAD_OVERSEER_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.WORKLOAD_OVERSEER.name])
    },
    [BASIC_DEBUG_TAGS.RUNAI_JOB_CONTROLLER_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.RUNAI_JOB_CONTROLLER.name])
    },
    [BASIC_DEBUG_TAGS.MPI_OPERATOR_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.MPI_OPERATOR.name])
    },
    [BASIC_DEBUG_TAGS.TRAINING_OPERATOR_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.TRAINING_OPERATOR.name])
    },
    [BASIC_DEBUG_TAGS.WORKLOADS_SERVICE_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.WORKLOADS_SERVICE.name])
    },
    [BASIC_DEBUG_TAGS.INFERENCE_WORKLOAD_CONTROLLER_LOGS]: {
        getResourceRequests: () =>
            createSystemLogRequests([SYSTEM_COMPONENTS.INFERENCE_WORKLOAD_CONTROLLER.name])
    },
    [BASIC_DEBUG_TAGS.POD_GROUP_ASSIGNER_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.POD_GROUP_ASSIGNER.name])
    },
    [BASIC_DEBUG_TAGS.NODEPOOL_CONTROLLER_LOGS]: {
        getResourceRequests: () => createSystemLogRequests([SYSTEM_COMPONENTS.NODEPOOL_CONTROLLER.name])
    },

    // Individual CRDs
    [BASIC_DEBUG_TAGS.RUNAI_JOBS]: {
        getResourceRequests: () => createCRDRequests([crds.runaiJobCrd])
    },
    [BASIC_DEBUG_TAGS.TRAINING_WORKLOADS]: {
        getResourceRequests: () => createCRDRequests([crds.runaiTrainingWorkloadCrd])
    },
    [BASIC_DEBUG_TAGS.INTERACTIVE_WORKLOADS]: {
        getResourceRequests: () => createCRDRequests([crds.runaiInteractiveWorkloadCrd])
    },
    [BASIC_DEBUG_TAGS.DISTRIBUTED_WORKLOADS]: {
        getResourceRequests: () => createCRDRequests([crds.runaiDistributedWorkloadCrd])
    },
    [BASIC_DEBUG_TAGS.INFERENCE_WORKLOADS]: {
        getResourceRequests: () => createCRDRequests([crds.runaiInferenceWorkloadCrd])
    },
    [BASIC_DEBUG_TAGS.XGBOOST_JOBS]: {
        getResourceRequests: () => createCRDRequests([crds.xgboostJobCrd])
    },
    [BASIC_DEBUG_TAGS.TF_JOBS]: {
        getResourceRequests: () => createCRDRequests([crds.tfJobCrd])
    },
    [BASIC_DEBUG_TAGS.MPI_JOBS]: {
        getResourceRequests: () => createCRDRequests([crds.mpiJobCrd])
    },
    [BASIC_DEBUG_TAGS.PYTORCH_JOBS]: {
        getResourceRequests: () => createCRDRequests([crds.pyTorchJobCrd])
    },
    [BASIC_DEBUG_TAGS.JAX_JOBS]: {
        getResourceRequests: () => createCRDRequests([crds.jaxJobCrd])
    },
    [BASIC_DEBUG_TAGS.KNATIVE_SERVICES]: {
        getResourceRequests: () => createCRDRequests([crds.knativeService])
    },

    // Pods
    [BASIC_DEBUG_TAGS.PODS_YAML]: {
        getResourceRequests: () => [
            {
                type: ResourceType.PODS_YAML,
                params: {},
                collector: async (): Promise<void> => {
                    try {
                        const allureUtils = new AllureUtils();
                        await allureUtils.attachPodsYaml();
                    } catch (error) {
                        logger.error('Failed to collect pods information:', error);
                    }
                }
            }
        ]
    },
    [BASIC_DEBUG_TAGS.PODS_LOGS]: {
        getResourceRequests: () => [
            {
                type: ResourceType.PODS_LOGS,
                params: {},
                collector: async (): Promise<void> => {
                    try {
                        const allureUtils = new AllureUtils();
                        await allureUtils.attachPodsLogsBySelector();
                    } catch (error) {
                        logger.error('Failed to collect pods logs:', error);
                    }
                }
            }
        ]
    }
};

function expandTags(inputTags: string[]): string[] {
    const expandedTagsSet = new Set<string>();

    for (const tag of inputTags) {
        const compositeTagKey = Object.keys(COMPOSITE_DEBUG_TAGS).find(
            (key) => COMPOSITE_DEBUG_TAGS[key as keyof typeof COMPOSITE_DEBUG_TAGS] === tag
        );

        if (compositeTagKey) {
            const basicTags =
                COMPOSITE_DEBUG_TAGS_DEF[compositeTagKey as keyof typeof COMPOSITE_DEBUG_TAGS_DEF];
            basicTags.forEach((basicTag) => expandedTagsSet.add(basicTag));
        } else {
            expandedTagsSet.add(tag);
        }
    }

    return Array.from(expandedTagsSet);
}

function generateResourceRequests(tags: string[]): ResourceCollectionRequest[] {
    const requests: ResourceCollectionRequest[] = [];

    for (const tag of tags) {
        const strategy = BASIC_DEBUG_COLLECTION_STRATEGIES[tag];
        if (strategy) {
            requests.push(...strategy.getResourceRequests());
        }
    }

    return requests;
}

function deduplicateRequests(requests: ResourceCollectionRequest[]): ResourceCollectionRequest[] {
    const seen = new Map<string, ResourceCollectionRequest>();

    for (const request of requests) {
        const key = `${request.type}-${JSON.stringify(request.params)}`;
        if (!seen.has(key)) {
            seen.set(key, request);
        }
    }

    return Array.from(seen.values());
}

async function executeResourceCollection(requests: ResourceCollectionRequest[]): Promise<void> {
    for (const request of requests) {
        await request.collector();
    }
}

function parseDebugTagsFromTitle(title: string): string[] {
    const debugTagsMatches = title.matchAll(/debugTags:\s*([^›]+?)(?=\s*›|$)/g);
    const matchesArray = Array.from(debugTagsMatches);

    const allDebugTags: string[] = [];

    for (const match of matchesArray) {
        const debugTagString = match[1].trim();

        const individualTags = debugTagString
            .split(/[,\s]+/)
            .map((tag) => tag.trim())
            .filter((tag) => tag.length > 0);

        for (const tag of individualTags) {
            if (tag.startsWith('@debug-')) {
                allDebugTags.push(tag);
                continue;
            }

            const debugTagKey = tag.split('.').pop();

            if (debugTagKey && DEBUG_TAGS[debugTagKey as keyof typeof DEBUG_TAGS]) {
                const foundTag = DEBUG_TAGS[debugTagKey as keyof typeof DEBUG_TAGS];
                allDebugTags.push(foundTag);
            }
        }
    }

    return allDebugTags;
}

function getFullTestPath(testInfo: TestInfo): string {
    if (testInfo.titlePath) {
        return testInfo.titlePath.join(' › ');
    }

    const pathParts: string[] = [];
    let currentTest = testInfo;

    while (currentTest) {
        pathParts.unshift(currentTest.title);
        currentTest = (currentTest as any).parent;
    }

    return pathParts.join(' › ');
}

export async function collectTestDebugInfo(testInfo: TestInfo): Promise<void> {
    if (process.env.COLLECT_DEBUG_INFO_ENABLED !== 'true') {
        logger.info('Debug info collection disabled by environment variable');
        return;
    }

    logger.info(`collectTestDebugInfo called for test: "${testInfo.title}" with status: ${testInfo.status}`);

    try {
        const fullTestPath = getFullTestPath(testInfo);
        const debugTags = parseDebugTagsFromTitle(fullTestPath);

        if (debugTags.length === 0) {
            return;
        }

        const expandedTags = expandTags(debugTags);
        const allRequests = generateResourceRequests(expandedTags);
        const uniqueRequests = deduplicateRequests(allRequests);

        await executeResourceCollection(uniqueRequests);
    } catch (debugError) {
        logger.error('Failed to collect debug information:', debugError);
    }
}

export const test = baseTest;

const originalAfterEach = test.afterEach;

test.afterEach = function (fnOrTitle: any, fn?: any): void {
    const wrapWithDebugCollection = (originalFn: any) => {
        return async function ({}: any, testInfo: any): Promise<void> {
            if (testInfo.status === 'failed') {
                try {
                    await baseTest.step('Debug Info Collection', async () => {
                        await collectTestDebugInfo(testInfo);
                    });
                } catch (debugError) {
                    logger.error('Failed to collect debug information:', debugError);
                }
            }

            if (originalFn) {
                await originalFn({}, testInfo);
            }
        };
    };

    if (typeof fnOrTitle === 'string') {
        return originalAfterEach(fnOrTitle, wrapWithDebugCollection(fn));
    } else {
        return originalAfterEach(wrapWithDebugCollection(fnOrTitle));
    }
} as any;

export { expect } from '@playwright/test';
