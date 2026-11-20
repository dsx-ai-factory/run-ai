/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { versionUtils } from '..';
import { SupportedVersions } from '../utils/SupporetedEnvsUtils';

export enum Scope {
    NAMESPACE,
    CLUSTER
}

export const KAI_ORG_UNIT_GROUP = 'kai.resources';

export const KAI_ORG_UNIT_VERSION = 'v1alpha1';

/**
 * NodePool moved from `nodepools.run.ai` to `nodepools.kai.resources`. Unlike the org
 * units, the cutover has no cluster-version boundary to gate on - both CRDs are served by 3.0
 * clusters, and which one holds the node pools depends on the build. The installed CRD is therefore
 * detected at runtime, see NodePoolCrdFields.resolveNodePoolCrd.
 */
export const LEGACY_NODE_POOL_GROUP = 'run.ai';

export const KAI_NODE_POOL_GROUP = KAI_ORG_UNIT_GROUP;

export interface CustomResourceDefinition {
    group: string;
    version: string;
    plural: string;
    kind: string;
    scope: Scope;
    /**
     * Optional function to get the phase/state field path from the CRD status object.
     * If not provided, defaults to 'status.phase'.
     * Example: For DGD which uses 'status.state', return 'status.state'
     */
    getPhasePath?(): string;
}

const kaiConfigCrd: CustomResourceDefinition = {
    group: 'kai.scheduler',
    version: 'v1',
    plural: 'configs',
    kind: 'Config',
    scope: Scope.CLUSTER
};

const topologyCrd: CustomResourceDefinition = {
    group: 'kai.scheduler',
    version: 'v1alpha1',
    plural: 'topologies',
    kind: 'Topology',
    scope: Scope.CLUSTER
};

const kueueTopologyCrd: CustomResourceDefinition = {
    group: 'kai.scheduler',
    version: 'v1alpha1',
    plural: 'topologies',
    kind: 'Topology',
    scope: Scope.CLUSTER
};

// From cluster 3.0 (the KAI cutover) SchedulingShard is served by kai.scheduler/v1.
// Older clusters serve it from the Engine group - see initCrds below.
const runaiSchedulingShardCrd: CustomResourceDefinition = {
    group: 'kai.scheduler',
    version: 'v1',
    plural: 'schedulingshards',
    kind: 'SchedulingShard',
    scope: Scope.CLUSTER
};

const runaiConfigCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1',
    plural: 'runaiconfigs',
    kind: 'RunaiConfig',
    scope: Scope.NAMESPACE
};

const departmentCrd: CustomResourceDefinition = {
    group: KAI_ORG_UNIT_GROUP,
    version: KAI_ORG_UNIT_VERSION,
    plural: 'departments',
    kind: 'Department',
    scope: Scope.CLUSTER
};

const queueCrd: CustomResourceDefinition = {
    group: 'scheduling.run.ai',
    version: 'v2',
    plural: 'queues',
    kind: 'Queue',
    scope: Scope.CLUSTER
};

const runaiJobCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1',
    plural: 'runaijobs',
    kind: 'RunaiJob',
    scope: Scope.NAMESPACE
};

const podGroupsCrd: CustomResourceDefinition = {
    group: 'scheduling.run.ai',
    version: 'v2alpha2',
    plural: 'podgroups',
    kind: 'PodGroup',
    scope: Scope.NAMESPACE
};

// The group is overwritten by NodePoolCrdFields.resolveNodePoolCrd once the cluster has been probed;
// keep the legacy group as the default until cluster discovery identifies a supported replacement.
const nodePoolCrd: CustomResourceDefinition = {
    group: LEGACY_NODE_POOL_GROUP,
    version: 'v1alpha1',
    plural: 'nodepools',
    kind: 'NodePool',
    scope: Scope.CLUSTER
};

const accessRuleCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'accessrules',
    kind: 'AccessRule',
    scope: Scope.CLUSTER
};

const datavolumeCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'datavolumes',
    kind: 'DataVolume',
    scope: Scope.CLUSTER
};

const runaiTrainingWorkloadCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v2alpha1',
    plural: 'trainingworkloads',
    kind: 'TrainingWorkload',
    scope: Scope.NAMESPACE
};

const runaiInteractiveWorkloadCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v2alpha1',
    plural: 'interactiveworkloads',
    kind: 'InteractiveWorkload',
    scope: Scope.NAMESPACE
};

const runaiInferenceWorkloadCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v2alpha1',
    plural: 'inferenceworkloads',
    kind: 'InferenceWorkload',
    scope: Scope.NAMESPACE
};

const runaiDistributedInferenceWorkloadCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'distributedinferenceworkloads',
    kind: 'DistributedInferenceWorkload',
    scope: Scope.NAMESPACE
};

const runaiDistributedWorkloadCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v2alpha1',
    plural: 'distributedworkloads',
    kind: 'DistributedWorkload',
    scope: Scope.NAMESPACE
};

const runaiPyTorchJobCrd: CustomResourceDefinition = {
    group: 'kubeflow.org',
    version: 'v1',
    plural: 'pytorchjobs',
    kind: 'PyTorchJob',
    scope: Scope.NAMESPACE
};

const runaiProjectCrd: CustomResourceDefinition = {
    group: KAI_ORG_UNIT_GROUP,
    version: KAI_ORG_UNIT_VERSION,
    plural: 'projects',
    kind: 'Project',
    scope: Scope.CLUSTER
};

const knativeServiceCrd: CustomResourceDefinition = {
    group: 'serving.knative.dev',
    version: 'v1',
    plural: 'services',
    kind: 'Service',
    scope: Scope.NAMESPACE
};

const knativeRevisionCrd: CustomResourceDefinition = {
    group: 'serving.knative.dev',
    version: 'v1',
    plural: 'revisions',
    kind: 'Revision',
    scope: Scope.NAMESPACE
};

const lwsCrd: CustomResourceDefinition = {
    group: 'leaderworkerset.x-k8s.io',
    version: 'v1',
    plural: 'leaderworkersets',
    kind: 'LeaderWorkerSet',
    scope: Scope.NAMESPACE
};

const tfJobCrd: CustomResourceDefinition = {
    group: 'kubeflow.org',
    version: 'v1',
    plural: 'tfjobs',
    kind: 'TFJob',
    scope: Scope.NAMESPACE
};

const pyTorchJobCrd: CustomResourceDefinition = {
    group: 'kubeflow.org',
    version: 'v1',
    plural: 'pytorchjobs',
    kind: 'PyTorchJob',
    scope: Scope.NAMESPACE
};

const xgboostJobCrd: CustomResourceDefinition = {
    group: 'kubeflow.org',
    version: 'v1',
    plural: 'xgboostjobs',
    kind: 'XGBoostJob',
    scope: Scope.NAMESPACE
};
const mpiJobCrd: CustomResourceDefinition = {
    group: 'kubeflow.org',
    version: 'v2beta1',
    plural: 'mpijobs',
    kind: 'MPIJob',
    scope: Scope.NAMESPACE
};

const jaxJobCrd: CustomResourceDefinition = {
    group: 'kubeflow.org',
    version: 'v1',
    plural: 'jaxjobs',
    kind: 'JAXJob',
    scope: Scope.NAMESPACE
};

const rayClusterCrd: CustomResourceDefinition = {
    group: 'ray.io',
    version: 'v1alpha1',
    plural: 'rayclusters',
    kind: 'RayCluster',
    scope: Scope.NAMESPACE
};

const nimServiceCrd: CustomResourceDefinition = {
    group: 'apps.nvidia.com',
    version: 'v1alpha1',
    plural: 'nimservices',
    kind: 'NIMService',
    scope: Scope.NAMESPACE
};

const dynamoGraphDeploymentCrd: CustomResourceDefinition = {
    group: 'nvidia.com',
    version: 'v1alpha1',
    plural: 'dynamographdeployments',
    kind: 'DynamoGraphDeployment',
    scope: Scope.NAMESPACE,
    getPhasePath: () => 'status.state' // DGD uses 'state' instead of 'phase'
};

const dynamoGraphDeploymentV1Beta1Crd: CustomResourceDefinition = {
    group: 'nvidia.com',
    version: 'v1beta1',
    plural: 'dynamographdeployments',
    kind: 'DynamoGraphDeployment',
    scope: Scope.NAMESPACE,
    getPhasePath: () => 'status.state'
};

const kartaCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'kartas',
    kind: 'Karta',
    scope: Scope.CLUSTER
};

const externalWorkloadCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'externalworkloads',
    kind: 'ExternalWorkload',
    scope: Scope.NAMESPACE,
    getPhasePath: () => 'status.observedStatus.phase'
};

const workloadRunnerCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'workloadrunners',
    kind: 'WorkloadRunner',
    scope: Scope.NAMESPACE,
    getPhasePath: () => 'status.phase'
};

const computeDomainCrd: CustomResourceDefinition = {
    group: 'resource.nvidia.com',
    version: 'v1beta1',
    plural: 'computedomains',
    kind: 'ComputeDomain',
    scope: Scope.NAMESPACE
};

const httpRouteCrd: CustomResourceDefinition = {
    group: 'gateway.networking.k8s.io',
    version: 'v1',
    plural: 'httproutes',
    kind: 'HTTPRoute',
    scope: Scope.NAMESPACE
};

const applicationsCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'applications',
    kind: 'Application',
    scope: Scope.NAMESPACE
};

const adviceCrd: CustomResourceDefinition = {
    group: 'heimdall.nvidia.com',
    version: 'v1alpha1',
    plural: 'advices',
    kind: 'Advice',
    scope: Scope.NAMESPACE
};

const restartRequestCrd: CustomResourceDefinition = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'restartrequests',
    kind: 'RestartRequest',
    scope: Scope.NAMESPACE
};

export const crds = {
    runaiConfigCrd: runaiConfigCrd,
    kaiConfigCrd: kaiConfigCrd,
    departmentCrd: departmentCrd,
    queueCrd: queueCrd,
    nodePoolCrd: nodePoolCrd,
    runaiSchedulingShardCrd: runaiSchedulingShardCrd,
    accessRuleCrd: accessRuleCrd,
    datavolumeCrd: datavolumeCrd,
    runaiJobCrd: runaiJobCrd,
    podGroupsCrd: podGroupsCrd,
    runaiTrainingWorkloadCrd: runaiTrainingWorkloadCrd,
    runaiDistributedWorkloadCrd: runaiDistributedWorkloadCrd,
    runaiPyTorchJobCrd: runaiPyTorchJobCrd,
    runaiProjectCrd: runaiProjectCrd,
    runaiInteractiveWorkloadCrd: runaiInteractiveWorkloadCrd,
    runaiInferenceWorkloadCrd: runaiInferenceWorkloadCrd,
    runaiDistributedInferenceWorkloadCrd: runaiDistributedInferenceWorkloadCrd,
    knativeService: knativeServiceCrd,
    knativeRevision: knativeRevisionCrd,
    lwsCrd: lwsCrd,
    tfJobCrd: tfJobCrd,
    pyTorchJobCrd: pyTorchJobCrd,
    jaxJobCrd: jaxJobCrd,
    xgboostJobCrd: xgboostJobCrd,
    mpiJobCrd: mpiJobCrd,
    rayClusterCrd: rayClusterCrd,
    topologyCrd: topologyCrd,
    nimServiceCrd: nimServiceCrd,
    dynamoGraphDeploymentCrd: dynamoGraphDeploymentCrd,
    dynamoGraphDeploymentV1Beta1Crd: dynamoGraphDeploymentV1Beta1Crd,
    kartaCrd: kartaCrd,
    externalWorkloadCrd: externalWorkloadCrd,
    workloadRunnerCrd: workloadRunnerCrd,
    computeDomainCrd: computeDomainCrd,
    httpRouteCrd: httpRouteCrd,
    applicationsCrd: applicationsCrd,
    adviceCrd: adviceCrd,
    restartRequestCrd: restartRequestCrd,

    initCrds(clusterVersion: string): void {
        if (!versionUtils.isVersionSupported(clusterVersion, SupportedVersions.V2_14)) {
            queueCrd.version = 'v2alpha2';
        }

        if (!versionUtils.isVersionSupported(clusterVersion, SupportedVersions.V2_24)) {
            crds.topologyCrd.group = 'kueue.x-k8s.io';
        }

        // set both ways - the upgrade suites re-run initCrds once the cluster is upgraded
        if (!versionUtils.isVersionSupported(clusterVersion, SupportedVersions.V3_0)) {
            runaiSchedulingShardCrd.group = 'engine.run.ai';
            runaiSchedulingShardCrd.version = 'v1alpha1';
            runaiProjectCrd.group = 'run.ai';
            runaiProjectCrd.version = 'v2';
            departmentCrd.group = 'scheduling.run.ai';
            departmentCrd.version = 'v2alpha2';
        }
    }
};
