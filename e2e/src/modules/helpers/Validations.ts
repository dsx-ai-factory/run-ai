/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { V1Deployment, V1Pod, V1PodList } from '@kubernetes/client-node';
import { DeploymentCondition } from '../api/apis/backendClusterApi';
import { bashCli, kubeCli } from '../cli';
import { assert, expect } from 'chai';
import { flowRootState, logger, waitUtils, workloadsUtilsV2 } from '.';
import { StatusCodes } from 'http-status-codes';
import {
    AssignedResources,
    AssignedResourcesCpu,
    AssignedResourcesGpu,
    AssignedResourcesMemory,
    Project
} from '../api/apis/projectsDepartmentsApi';
import { SoftAssert } from './SoftAssert';
import {
    Resources,
    Project as ProjectV2,
    ResourcesGpu,
    ResourcesCpu,
    ResourcesMemory
} from '../api/apis/orgUnitApi';
import { Workload } from '../api/apis/workloadsApi/generated/model/workload';
import { Pod as PodV2 } from '../api/apis/workloadsApi/generated/model/pod';

export function validateDeploymentCreatedSuccessfullyInCluster(
    deployment: V1Deployment,
    deploymentName: string,
    namespaceName: string,
    projectName: string,
    replicas: number,
    image: string,
    nodePoolName?: string
): void {
    logger.debug(`Validating deployment ${deploymentName} was created successfully in cluster`);
    expect(deployment.kind).to.equal('Deployment');
    expect(deployment.apiVersion).to.equal('apps/v1');
    expect(deployment.metadata.name).to.equal(deploymentName);
    expect(deployment.metadata.namespace).to.equal(namespaceName);
    expect(deployment.metadata.labels['project']).to.equal(projectName);
    expect(deployment.metadata.labels['app']).to.equal('test-app');
    expect(deployment.spec.replicas).to.equal(replicas);
    expect(deployment.spec.template.spec.schedulerName).to.equal('runai-scheduler');
    expect(deployment.spec.template.spec.containers[0].name).to.equal('ubuntu');
    expect(deployment.spec.template.spec.containers[0].image).to.equal(image);
    expect(deployment.spec.template.spec.containers[0].command.length).to.equal(1);
    expect(deployment.spec.template.spec.containers[0].command[0]).to.equal('sleep');
    expect(deployment.spec.template.spec.containers[0].args.length).to.equal(1);
    expect(deployment.spec.template.spec.containers[0].args[0]).to.equal('infinity');
    if (nodePoolName)
        expect(deployment.spec.template.metadata.labels['runai/node-pool']).to.equal(nodePoolName);
    logger.debug(`Deployment ${deploymentName} was created successfully in cluster`);
}

export function validatePodsCreatedSuccessfullyInCluster(
    podList: V1PodList,
    namespaceName: string,
    projectName: string,
    image: string
): void {
    podList.items.forEach((pod: V1Pod) => {
        expect(pod.metadata.namespace).to.equal(namespaceName);
        expect(pod.metadata.labels['project']).to.equal(projectName);
        expect(pod.metadata.labels['app']).to.equal('test-app');
        expect(pod.spec.schedulerName).to.equal('runai-scheduler');
        expect(pod.spec.containers[0].name).to.equal('ubuntu');
        expect(pod.spec.containers[0].image).to.equal(image);
        expect(pod.spec.containers[0].command[0]).to.equal('sleep');
        expect(pod.spec.containers[0].args[0]).to.equal('infinity');
    });
}

export function validatePodsMatch(podsFromCluster: V1PodList, podsFromBackend: PodV2[]): void {
    for (let i = 0; i < podsFromBackend.length; i++) {
        const podFromBackend: PodV2 = podsFromBackend[i];
        const podFromCluster: V1Pod = podsFromCluster.items[i];
        expect(podFromBackend.name).to.equal(podFromCluster.metadata.name);
        expect(podFromBackend.id).to.equal(podFromCluster.metadata.uid);
        expect(podFromBackend.containers[0].image).to.equal(podFromCluster.spec.containers[0].image);
        expect(podFromBackend.requestedResources.gpu.request).to.equal(1);
        expect(podFromBackend.nodeName).not.to.be.undefined;
        expect(podFromBackend.createdAt).not.to.be.undefined;
        expect(podFromBackend.updatedAt).not.to.be.undefined;
        expect(podFromBackend.completedAt).to.be.undefined;
    }
}

async function validateDeploymentInCP(clusterId: string, deployment: V1Deployment): Promise<Workload> {
    logger.debug(`Validating deployment ${deployment.metadata.name} was created successfully in backend`);
    await workloadsUtilsV2.waitForWorkloadInControlPlane(deployment.metadata.uid);
    const listWorkloadsResponse = await workloadsUtilsV2.getWorkloadsByClusterId(clusterId);
    expect(listWorkloadsResponse.response.statusCode).to.equal(StatusCodes.OK);
    const deploymentsFromBackend: Workload[] = listWorkloadsResponse.body.workloads.filter(
        (workload) => workload.type === 'Deployment'
    );
    expect(deploymentsFromBackend.length).to.be.above(0);
    const deploymentFromBackend = deploymentsFromBackend.find(
        (value) => value.name === deployment.metadata.name
    );

    const countWorkloadsResponse = await workloadsUtilsV2.countWorkloadsByClusterId(clusterId);
    expect(countWorkloadsResponse.response.statusCode).to.equal(StatusCodes.OK);
    const countDeploymentsFromBackend = countWorkloadsResponse.body;
    expect(countDeploymentsFromBackend.count).to.equal(listWorkloadsResponse.body.workloads.length);

    return deploymentFromBackend;
}

function validateBackendDeploymentStatusCreated(backendDeployment: Workload): void {
    expect(backendDeployment.conditions.length).to.greaterThan(0);
    const firstType = backendDeployment.conditions[0].type;
    expect(firstType).to.oneOf([
        DeploymentCondition.TypeEnum.Progressing,
        DeploymentCondition.TypeEnum.Available
    ]);
    if (firstType === DeploymentCondition.TypeEnum.Progressing.toString()) {
        expect(backendDeployment.conditions[0].status).to.be.equal(
            'True',
            `Condition fail reason: ${backendDeployment.conditions[0].reason}`
        );
    } else {
        expect(backendDeployment.conditions[0].status).to.oneOf(['True', 'False']);
    }
    expect(backendDeployment.conditions[0].reason).to.exist;

    // In case we have 2 conditions
    if (backendDeployment.conditions.length > 1) {
        expect(backendDeployment.conditions[1].type).to.equal(
            firstType === DeploymentCondition.TypeEnum.Progressing.toString()
                ? DeploymentCondition.TypeEnum.Available
                : DeploymentCondition.TypeEnum.Progressing
        );
        if (firstType === DeploymentCondition.TypeEnum.Progressing.toString()) {
            expect(backendDeployment.conditions[1].status).to.oneOf(['True', 'False']);
        } else {
            expect(backendDeployment.conditions[1].status).to.be.equal(
                'True',
                `Condition fail reason: ${backendDeployment.conditions[1].reason}`
            );
        }
        expect(backendDeployment.conditions[1].reason).to.exist;
    }
}

function validateClusterAndBackendDeploymentMatch(
    deployment: V1Deployment,
    backendDeployment: Workload
): void {
    expect(backendDeployment.id).to.equal(deployment.metadata.uid);
    expect(backendDeployment.name).to.equal(deployment.metadata.name);
    expect(backendDeployment.requestedPods.number).to.equal(deployment.spec.replicas);
    expect(backendDeployment.projectName).to.equal(deployment.metadata.labels['project']);
    expect(backendDeployment.createdAt).not.to.be.undefined;

    expect(backendDeployment.podsRequestedResources.gpu.request).to.equal(1);
    expect(backendDeployment.images.length).to.equal(1);
    expect(backendDeployment.images[0]).to.equal(deployment.spec.template.spec.containers[0].image);
}

export async function validateClusterAndBackendDeploymentPodsMatch(
    deployment: V1Deployment,
    backendDeployment: Workload,
    replicas: number
): Promise<void> {
    logger.debug(`Validation deployment ${deployment.metadata.name} pods`);
    await waitUtils.waitForPodsByOwnerPrefix(
        deployment.metadata.name,
        deployment.metadata.namespace,
        replicas
    );
    const podsInNamespace = await kubeCli.CoreV1Api.listNamespacedPod({
        namespace: backendDeployment.namespace
    });
    const podsFromCluster: V1PodList = podsInNamespace;
    const deploymentPods = new V1PodList();
    deploymentPods.items = new Array<V1Pod>(0);
    for (const podInNamespace of podsFromCluster.items) {
        if (podInNamespace.metadata.name.startsWith(deployment.metadata.name)) {
            deploymentPods.items.push(podInNamespace);
        }
    }
    expect(deploymentPods.items.length).to.equal(replicas);
    validatePodsCreatedSuccessfullyInCluster(
        deploymentPods,
        backendDeployment.namespace,
        backendDeployment.projectName,
        backendDeployment.images[0]
    );

    const podsFromBackendResponse = await workloadsUtilsV2.getWorkloadPods(
        backendDeployment.id,
        flowRootState.users.ml_engineer.accessToken
    );
    expect(podsFromBackendResponse.response.statusCode).to.equal(StatusCodes.OK);
    expect(podsFromBackendResponse.body.pods.length).to.equal(replicas);
    const podsFromBackend: PodV2[] = podsFromBackendResponse.body.pods;
    validatePodsMatch(deploymentPods, podsFromBackend);
}

function validateDeploymentCreatedSuccessfullyInBackendFn(
    clusterId: string,
    deployment: V1Deployment,
    replicas: number
): () => Promise<boolean> {
    return async (): Promise<boolean> => {
        try {
            const deploymentFromBackend = await validateDeploymentInCP(clusterId, deployment);

            validateClusterAndBackendDeploymentMatch(deployment, deploymentFromBackend);
            validateBackendDeploymentStatusCreated(deploymentFromBackend);

            await validateClusterAndBackendDeploymentPodsMatch(deployment, deploymentFromBackend, replicas);
            logger.debug(`Deployment ${deployment.metadata.name} was created successfully in backend`);
        } catch (e) {
            logger.debug(`Failed validating deployment in backend was created successfully, error: ${e}`);
            return false;
        }
        return true;
    };
}

export async function waitAndValidateDeploymentCreatedSuccessfullyInBackend(
    clusterId: string,
    deployment: V1Deployment,
    replicas: number
): Promise<void> {
    await waitUtils.waitForCondition(
        validateDeploymentCreatedSuccessfullyInBackendFn(clusterId, deployment, replicas),
        `wait for deployment from backend ${deployment.metadata.name} in namespace ${deployment.metadata.namespace} with ${replicas} replicas to be created in backend successfully`,
        true,
        80
    );
}

function validateBackendDeploymentStatusAvailable(backendDeployment: Workload): void {
    expect(backendDeployment.conditions.length).to.equal(2);
    const firstType = backendDeployment.conditions[0].type;
    const firstReason = backendDeployment.conditions[0].reason;

    expect(firstType).to.be.oneOf(['Progressing', 'Available']);
    expect(backendDeployment.conditions[0].status).to.equal(
        'True',
        `Deployment condition fail reason: ${backendDeployment.conditions[0].reason}`
    );
    expect(firstReason).to.be.oneOf(['NewReplicaSetAvailable', 'MinimumReplicasAvailable']);

    expect(backendDeployment.conditions[1].type).to.equal(
        firstType === DeploymentCondition.TypeEnum.Progressing.toString() ? 'Available' : 'Progressing'
    );
    expect(backendDeployment.conditions[1].status).to.equal(
        'True',
        `Deployment condition fail reason: ${backendDeployment.conditions[1].reason}`
    );
    expect(backendDeployment.conditions[1].reason).to.equal(
        firstReason === 'NewReplicaSetAvailable' ? 'MinimumReplicasAvailable' : 'NewReplicaSetAvailable'
    );
}

function validateDeploymentConditions(deploymentFromBackend: Workload, deployment: V1Deployment): void {
    expect(deploymentFromBackend.conditions.length).to.equal(deployment.status.conditions.length);
    for (let i = 0; i < deployment.status.conditions.length; i++) {
        expect(deploymentFromBackend.conditions[i].reason).to.equal(deployment.status.conditions[i].reason);
        expect(deploymentFromBackend.conditions[i].status).to.equal(deployment.status.conditions[i].status);
        expect(deploymentFromBackend.conditions[i].type).to.equal(deployment.status.conditions[i].type);
        expect(deploymentFromBackend.conditions[i].message).to.equal(deployment.status.conditions[i].message);
    }
}

function validateDeploymentInBackendStatusAvailableFn(
    clusterId: string,
    deployment: V1Deployment
): () => Promise<boolean> {
    return async (): Promise<boolean> => {
        try {
            const deploymentFromBackend = await validateDeploymentInCP(clusterId, deployment);

            validateClusterAndBackendDeploymentMatch(deployment, deploymentFromBackend);

            validateDeploymentConditions(deploymentFromBackend, deployment);
            validateBackendDeploymentStatusAvailable(deploymentFromBackend);
        } catch (e) {
            logger.debug(`Failed validating deployment in backend in status available, error: ${e}`);
            return false;
        }
        return true;
    };
}

export async function waitAndValidateDeploymentInBackendStatusAvailable(
    clusterId: string,
    deployment: V1Deployment,
    timeout = 120
): Promise<void> {
    await waitUtils.waitForCondition(
        validateDeploymentInBackendStatusAvailableFn(clusterId, deployment),
        `wait for deployment from backend ${deployment.metadata.name} in namespace ${deployment.metadata.namespace} to be available`,
        true,
        timeout
    );
}

export async function validateDeploymentsDeleted(workloadId: string, projectId: string): Promise<void> {
    const deployment = await workloadsUtilsV2.waitForDeletedWorkload(
        workloadId,
        projectId,
        flowRootState.users.ml_engineer.accessToken
    );
    expect(deployment.deletedAt).to.not.be.null;
}

export function validateBackendProject(
    project: Project,
    name: string,
    departmentId: number,
    expectedResources: AssignedResources,
    nodePoolResourcesMap: any
): void {
    logger.debug(`Validating backend project ${project.name} properties`);
    const softAssert = new SoftAssert();
    softAssert.assert(
        project.name === name,
        `Expected project.name to be ${name}, got instead ${project.name}`
    );
    softAssert.assert(
        project.departmentId === departmentId,
        `Expected project.departmentId to be ${departmentId}, got instead ${project.departmentId}`
    );

    logger.debug(`Validating backend project ${project.name} node pool resources`);
    project.nodePoolsResources.forEach((nodePoolResources) => {
        logger.debug(`nodePoolResources: ${JSON.stringify(nodePoolResources)}`);
        const expectedNodePoolResources = nodePoolResourcesMap[nodePoolResources.nodePool.name];
        if (expectedNodePoolResources) {
            logger.debug('validating node pool ' + nodePoolResources.nodePool.name + ' resources');
            logger.debug(`expected node pool resources: ${JSON.stringify(expectedNodePoolResources)}`);
            softAssert.assertDeepNonStrictEqual<AssignedResourcesGpu>(
                expectedNodePoolResources.gpu,
                nodePoolResources.gpu,
                'Expected GPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<AssignedResourcesCpu>(
                expectedNodePoolResources.cpu,
                nodePoolResources.cpu,
                'Expected CPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<AssignedResourcesMemory>(
                expectedNodePoolResources.memory,
                nodePoolResources.memory,
                'Expected memory resources to match'
            );
            logger.debug(`Done validating node pool ${nodePoolResources.nodePool.name}resources`);
        }
    });
    softAssert.throwErrorIfFailed();
}

export function validateOrgUnitProject(
    project: ProjectV2,
    name: string,
    departmentId: string,
    expectedResources: Map<string, Resources>
): void {
    logger.debug(`Validating backend project ${project.name} properties`);
    const softAssert = new SoftAssert();
    softAssert.assert(
        project.name === name,
        `Expected project.name to be ${name}, got instead ${project.name}`
    );
    softAssert.assert(
        project.parentId === departmentId,
        `Expected project.name to be ${departmentId}, got instead ${project.parentId}`
    );

    logger.debug(`Validating org unit project ${project.name} resources`);
    project.resources.forEach((resources) => {
        logger.debug(`nodePoolResources: ${JSON.stringify(resources)}`);
        const expectedNodePoolResources = expectedResources.get(resources.nodePool.name);
        if (expectedNodePoolResources) {
            logger.debug('validating node pool ' + resources.nodePool.name + ' resources');
            logger.debug(`expected node pool resources: ${JSON.stringify(expectedNodePoolResources)}`);
            softAssert.assertDeepNonStrictEqual<ResourcesGpu>(
                resources.gpu,
                expectedNodePoolResources.gpu,
                'Expected GPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<ResourcesCpu>(
                resources.cpu,
                expectedNodePoolResources.cpu,
                'Expected CPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<ResourcesMemory>(
                resources.memory,
                expectedNodePoolResources.memory,
                'Expected memory resources to match'
            );
            logger.debug(`Done validating node pool ${resources.nodePool.name}resources`);
        }
    });
    softAssert.throwErrorIfFailed();
}

export function validateJobTimeLimitForProject(
    project: Project,
    interactiveJobTimeLimitSecs: number | undefined
): void {
    logger.debug(`Validating interactive job time limit for backend project ${project.name}`);
    expect(project.interactiveJobTimeLimitSecs).to.be.equal(interactiveJobTimeLimitSecs);
}

export function validateJobMaxIdleDurationsForProject(
    project: Project,
    interactiveJobMaxIdleDurationSecs?: number,
    interactivePreemptibleJobMaxIdleDurationSecs?: number,
    trainingJobMaxIdleDurationSecs?: number
): void {
    logger.debug(`Validating job max idle time for backend project ${project.name}`);
    expect(project.interactiveJobMaxIdleDurationSecs).to.be.equal(interactiveJobMaxIdleDurationSecs);
    expect(project.interactivePreemptibleJobMaxIdleDurationSecs).to.be.equal(
        interactivePreemptibleJobMaxIdleDurationSecs
    );
    expect(project.trainingJobMaxIdleDurationSecs).to.be.equal(trainingJobMaxIdleDurationSecs);
}

export async function verifyNodeExclusion(
    excludedNodeNames: string[],
    namespace: string,
    workloadName: string
): Promise<void> {
    const pods = await kubeCli.listPodsByNamepsace(namespace);
    // only get workload pods in current test
    const workloadPods = pods.filter((pod) => pod.metadata.name.includes(workloadName));
    expect(workloadPods.length).to.greaterThan(0);
    for (const pod of workloadPods) {
        expect(pod.spec?.nodeName).to.exist;
        if (excludedNodeNames.includes(pod.spec.nodeName)) {
            const oWideCmd = `kubectl get pod -n ${namespace} -o wide`;
            const affinityCmd = `kubectl get pod ${pod.metadata.name} -n ${namespace} -o yaml`;
            logger.info(`-o wide cmd: \n${oWideCmd}`);
            logger.warn(`kgp -o wide: \n${bashCli.exec(oWideCmd)}`);
            logger.info(`get node affinity cmd: \n${affinityCmd}`);
            logger.warn(`node affinity: \n${bashCli.exec(affinityCmd)}`);
            assert.fail(`pod is running on excluded node ${pod.spec?.nodeName}`);
        }
    }
}

export async function verifyNodeInclusion(
    includeNodeNames: string[],
    namespace: string,
    workloadName: string
): Promise<void> {
    const pods = await kubeCli.listPodsByNamepsace(namespace);
    // only get workload pods in current test
    const workloadPods = pods.filter((pod) => pod.metadata.name.includes(workloadName));
    expect(workloadPods.length).to.greaterThan(0);
    for (const pod of workloadPods) {
        expect(pod.spec?.nodeName).to.exist;
        if (!includeNodeNames.includes(pod.spec.nodeName)) {
            const oWideCmd = `kubectl get pod -n ${namespace} -o wide`;
            const affinityCmd = `kubectl get pod ${pod.metadata.name} -n ${namespace} -o yaml`;
            logger.info(`-o wide cmd: \n${oWideCmd}`);
            logger.warn(`kgp -o wide: \n${bashCli.exec(oWideCmd)}`);
            logger.info(`get node affinity cmd: \n${affinityCmd}`);
            logger.warn(`node affinity: \n${bashCli.exec(affinityCmd)}`);
            assert.fail(
                `pod is running on node ${pod.spec?.nodeName} and suppose it should have run on ${includeNodeNames}`
            );
        }
    }
}
