/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    flowRootState,
    inferenceClientCreator,
    InteractiveWorkload,
    TrainingWorkload,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    orgUnitUtils,
    projectResourcesUtils,
    projectUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2,
    versionUtils
} from '../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../modules/cli';
import { test, expect } from '@playwright/test';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { validateDeploymentCreatedSuccessfullyInCluster } from '../../../../../../../modules/helpers/Validations';
import { KubernetesObject } from '@kubernetes/client-node';
import { Pod } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/pod';
import { Source } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/source';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { GpuRequestType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/gpuRequestType';
import { PhaseReason } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phaseReason';
import { OrgType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/orgType';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`workload api advanced tests:  ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.viper}`, () => {
    const projectDeservedGpus = 2;

    let project: Project;
    let namespace: string;

    let projectId: string;
    let defaultNodePoolId: string;

    test.beforeAll(async () => {
        try {
            await operations.forceCleanup();
        } catch (error) {
            logger.error(error);
        }
        const projectName = objectGenerator.generateNameByPrefix('p');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: projectDeservedGpus,
            maxAllowedGpus: projectDeservedGpus
        });

        project = await projectUtils.waitForProjectFromBackendNonEmptyStatus(
            projectName,
            project.id.toString()
        );

        projectId = project.id.toString();
        defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);
    });

    test.afterAll(async () => {
        if (project) {
            await operations.deleteProjectAndNamespace(project.id);
        }

        try {
            await operations.forceCleanup();
        } catch (error) {
            logger.error('Force Cleanup failed', error);
        }
    });

    test.describe('Deployments', () => {
        let deployment: KubernetesObject;

        test.beforeEach(async () => {
            const deploymentName = objectGenerator.generateNameByPrefix('d');

            deployment = await kubeCli.createDeployment(
                inferenceClientCreator.getDeployment(
                    namespace,
                    project.name,
                    deploymentName,
                    1,
                    config.quickStartImage.inference
                )
            );

            validateDeploymentCreatedSuccessfullyInCluster(
                deployment,
                deploymentName,
                namespace,
                project.name,
                1,
                config.quickStartImage.inference
            );

            await waitUtils.waitForPods(
                deploymentName,
                namespace,
                1,
                waitUtils.waitForPodTimeout,
                ['Running'],
                true
            );

            await workloadsUtilsV2.waitForWorkloadPodsByNameAndProjectName(deploymentName, project.name, 60);
        });

        test('quick data reflection', async () => {
            let cpDeployment = (await workloadsUtilsV2.getWorkloadByWorkloadId(deployment.metadata.uid)).body;
            logger.debug(`Got the control plane workload of deployment ${cpDeployment.name}`);
            logger.debug('Deployment:');
            logger.debug(`${JSON.stringify(cpDeployment)}`);

            const inferenceName = versionUtils.isVersionSupported(
                flowRootState.cluster.version,
                SupportedVersions.V2_23
            )
                ? 'very-high'
                : versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_14)
                  ? 'inference'
                  : 'Inference';

            expect(cpDeployment.tenantId).toBe(flowRootState.cluster.tenantId);
            expect(cpDeployment.runningPods).toBe(1);
            expect(cpDeployment.phaseUpdatedAt).toBeTruthy();
            expect(cpDeployment.k8sPhaseUpdatedAt).toBeTruthy();
            expect(cpDeployment.updatedAt).toBeTruthy();
            expect(cpDeployment.source).toBe(Source.Other);
            expect(cpDeployment.deletedAt).toBeFalsy();
            expect(cpDeployment.type).toBe('Deployment');
            expect(cpDeployment.name).toBe(deployment.metadata.name);
            expect(cpDeployment.id).toBe(deployment.metadata.uid);
            expect(cpDeployment.priorityClassName).toBe(inferenceName);
            expect(cpDeployment.submittedBy).toBeFalsy();
            expect(cpDeployment.clusterId).toBe(flowRootState.cluster.uuid);
            expect(cpDeployment.projectName).toBe(project.name);
            expect(cpDeployment.projectId).toBe(project.id.toString());
            expect(cpDeployment.namespace).toBe(deployment.metadata.namespace);
            expect(cpDeployment.createdAt).toBeTruthy();
            expect(cpDeployment.workloadRequestedResources.gpuRequestType).toBe(GpuRequestType.Portion);
            expect(cpDeployment.workloadRequestedResources.gpu.limit).toBe(1);
            expect(cpDeployment.podsRequestedResources.gpuRequestType).toBe(GpuRequestType.Portion);
            expect(cpDeployment.podsRequestedResources.gpu.limit).toBe(1);
            expect(cpDeployment.allocatedResources.gpu).toBe(1);
            expect(cpDeployment.phase).toBe(Phase.Running);
            expect(cpDeployment.conditions.length).toBeGreaterThan(0);
            expect(cpDeployment.requestedPods.number).toBe(1);
            expect(cpDeployment.requestedPods.min).toBeFalsy();
            expect(cpDeployment.requestedPods.max).toBeFalsy();
            expect(cpDeployment.requestedPods.parallelism).toBeFalsy();
            expect(cpDeployment.requestedPods.completions).toBeFalsy();
            expect(cpDeployment.requestedNodePools.length).toBe(1);
            expect(cpDeployment.requestedNodePools[0]).toBe('default');
            expect(cpDeployment.currentNodePools.length).toBe(1);
            expect(cpDeployment.currentNodePools[0]).toBe('default');
            expect(cpDeployment.completedAt).toBeFalsy();
            expect(cpDeployment.images.length).toBe(1);
            expect(cpDeployment.images[0]).toBe(config.quickStartImage.inference);

            logger.debug(
                `Control plane workload ${cpDeployment.name} validated, patching new replicas count`
            );
            const patch = {
                spec: {
                    replicas: 3
                }
            };
            await kubeCli.patchDeployment(deployment.metadata.name, namespace, patch);

            await waitUtils.waitForCondition(
                async () => {
                    cpDeployment = (await workloadsUtilsV2.getWorkloadByWorkloadId(deployment.metadata.uid))
                        .body;
                    return cpDeployment.requestedPods?.number === 3;
                },
                `wait for deployment ${deployment.metadata.name} requestedPods to reflect 3 replicas`,
                true,
                20,
                2
            );

            logger.debug(`Refreshed the control plane workload of deployment ${cpDeployment.name}`);
            expect(cpDeployment.requestedPods.number).toBe(3);
        });

        test('pods', async () => {
            const deploymentPods = (await workloadsUtilsV2.getWorkloadPods(deployment.metadata.uid)).body;

            expect(deploymentPods.next).toBeFalsy();
            expect(deploymentPods.pods.length).toBe(1);
            const pod = deploymentPods.pods[0];
            expect(pod.workloadId).toBe(deployment.metadata.uid);
            expect(pod.tenantId).toBe(flowRootState.cluster.tenantId);
            expect(pod.clusterId).toBe(flowRootState.cluster.uuid);
            expect(pod.projectId).toBe(project.id.toString());
            expect(pod.nodeName).toBeTruthy();
            expect(pod.createdAt).toBeTruthy();
            expect(pod.completedAt).toBeFalsy();
            expect(pod.currentNodePool).toBe('default');
            expect(pod.requestedNodePools.length).toBe(1);
            expect(pod.requestedNodePools[0]).toBe('default');
            expect(pod.updatedAt).toBeTruthy();
            expect(pod.k8sPhaseUpdatedAt).toBeTruthy();
            expect(pod.deletedAt).toBeFalsy();
            expect(pod.k8sPhase).toBe('Running');
        });

        test('events', async () => {
            logger.debug('Wait 10 seconds for events to sync');

            let deploymentPods: Array<Pod>;
            await Promise.all([
                workloadsUtilsV2.getWorkloadPods(deployment.metadata.uid).then((response) => {
                    deploymentPods = response.body.pods;
                }),
                waitUtils.sleep(10000)
            ]);

            const deploymentEvents = (await workloadsUtilsV2.getWorkloadEvents(deployment.metadata.uid)).body
                .events;
            logger.debug(`Got the control plane events of deployment ${deployment.metadata.name}`);

            expect(deploymentEvents.length).toBeGreaterThanOrEqual(deploymentPods.length);

            expect(
                deploymentEvents.some(
                    (e) =>
                        e.involvedObject.uid === deployment.metadata.uid &&
                        e.involvedObject.kind === 'Deployment'
                )
            ).toBe(true);
            expect(deploymentEvents.some((e) => e.involvedObject.kind === 'ReplicaSet')).toBe(true);
            for (const pod of deploymentPods) {
                expect(
                    deploymentEvents.some(
                        (e) =>
                            e.involvedObject.name === pod.name &&
                            e.involvedObject.uid === pod.id &&
                            e.involvedObject.kind === 'Pod' &&
                            e.involvedObject.namespace === deployment.metadata.namespace
                    )
                ).toBe(true);
            }
        });

        test('deletion', async () => {
            await kubeCli.deleteDeployment(deployment.metadata.name, deployment.metadata.namespace);
            const cpDeployment = await workloadsUtilsV2.waitForDeletedWorkload(
                deployment.metadata.uid,
                project.id
            );

            expect(cpDeployment.deletedAt).toBeTruthy();

            deployment = null;
        });

        test.afterEach(async () => {
            if (deployment) {
                await kubeCli.deleteDeployment(deployment.metadata.name, deployment.metadata.namespace);
                deployment = null;
            }
        });
    });

    supportedFrom(SupportedVersions.V2_19)(() => {
        shouldRunNativeWorkloadCrdTests(() =>
        test.describe('training pending (OverLimit) and then running', () => {
            let workloadName: string;
            let workload: TrainingWorkload;
            let workloadId: string;

            test.beforeAll(() => {
                workloadName = objectGenerator.generateNameByPrefix('job');

                workload = new TrainingWorkload(config.quickStartImage.ubuntu, namespace, workloadName);
                workload.arguments = 'sleep infinity';
                workload.cpu = '1';
                workload.createHomeDir = true;
                workload.gpu = '1';
                workload.memory = '100M';
            });

            test.afterAll(async () => {
                try {
                    await operations.deleteWorkloads([workload]);
                } catch (error) {
                    logger.error(
                        `Failed to delete workload ${workloadName} in project ${project.name}; error: ${error}`
                    );
                }

                await orgUnitUtils.patchProjectResources(projectId, [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: projectDeservedGpus, limit: projectDeservedGpus }
                    }
                ]);
                await waitUtils.sleep(5000);
                await orgUnitUtils.waitForProjectFromBackendToBeReady(project.name, projectId);
            });

            test('submit training workload and wait for it to be pending', async () => {
                workload.gpu = String(projectDeservedGpus + 1);
                await kubeCli.createWorkload(workload);

                    const trainingWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(
                        workloadName,
                        namespace
                    );
                workloadId = trainingWorkload.metadata.uid;
                    await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                        workloadId,
                        Phase.Pending,
                        workloadName
                    );
            });

            test('validate phaseReason and pendingSchedulingMessages', async () => {
                const apiWorkload = await workloadsUtilsV2.waitForApiWorkloadWithPhaseReason(
                    workloadId,
                    PhaseReason.OverLimit,
                    workloadName
                );

                expect(apiWorkload.phaseReason).toBe(PhaseReason.OverLimit);
                expect(apiWorkload.pendingSchedulingMessages).toHaveLength(1);
                expect(apiWorkload.pendingSchedulingMessages[0].nodePool).toBe('default');
                expect(apiWorkload.pendingSchedulingMessages[0].phaseReason).toBe(PhaseReason.OverLimit);
                expect(apiWorkload.pendingSchedulingMessages[0].reason).toBe(
                    'The project is not allowed to go over quota'
                );
                expect(apiWorkload.pendingSchedulingMessages[0].orgType).toBe(OrgType.Project);
                expect(apiWorkload.pendingSchedulingMessages[0].userMessage).toBe(
                    'GPU quota exceeded.\nRequested: 3 devices\nProject quota: 2 devices, with 0 devices already allocated'
                );

                    const apiWorkloadFromList = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId))
                        .body;
                expect(apiWorkloadFromList.phaseReason).toBe(PhaseReason.OverLimit);
            });

            test('increase the project quota and wait for workload to be running', async () => {
                await orgUnitUtils.patchProjectResources(projectId, [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: projectDeservedGpus * 2, limit: projectDeservedGpus * 2 }
                    }
                ]);
                await orgUnitUtils.waitForProjectFromBackendToBeReady(project.name, projectId);

                await waitUtils.waitForJobToBeRunning(
                    workload.name,
                    namespace,
                    crds.runaiTrainingWorkloadCrd.kind
                );
            });

            test('validate phaseReason and pendingSchedulingMessages are nil', async () => {
                    await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                        workloadId,
                        Phase.Running,
                        workloadName
                    );

                const apiWorkload = await workloadsUtilsV2.waitForApiWorkloadWithPhaseReason(
                    workloadId,
                    null,
                    workloadName
                );

                expect(apiWorkload.phaseReason).toBeUndefined();
                expect(apiWorkload.pendingSchedulingMessages).toBeUndefined();

                    const apiWorkloadFromList = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId))
                        .body;
                expect(apiWorkloadFromList.phaseReason).toBeUndefined();
            });

            test.describe('interactive pending (NonPreemtibleOverQuota) and then running', () => {
                let workloadName: string;
                let workload: InteractiveWorkload;
                let workloadId: string;

                test.beforeAll(() => {
                    workloadName = objectGenerator.generateNameByPrefix('job');

                    workload = new InteractiveWorkload(
                        config.quickStartImage.ubuntu,
                        namespace,
                        workloadName
                    );
                    workload.arguments = 'sleep infinity';
                    workload.cpu = '1';
                    workload.createHomeDir = true;
                    workload.gpu = '1';
                    workload.memory = '100M';
                });

                test.afterAll(async () => {
                    try {
                        await operations.deleteWorkloads([workload]);
                    } catch (error) {
                        logger.error(
                            `Failed to delete workload ${workloadName} in project ${project.name}; error: ${error}`
                        );
                    }

                    await orgUnitUtils.patchProjectResources(projectId, [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: projectDeservedGpus, limit: projectDeservedGpus }
                        }
                    ]);
                    await orgUnitUtils.waitForProjectFromBackendToBeReady(project.name, projectId);
                });

                test('patch the project quota to allow over quota', async () => {
                    await orgUnitUtils.patchProjectResources(projectId, [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: projectDeservedGpus, limit: -1 }
                        }
                    ]);
                    await waitUtils.sleep(10000);
                    await orgUnitUtils.waitForProjectFromBackendToBeReady(project.name, projectId);
                });

                test('submit interactive workload and wait for it to be pending', async () => {
                    workload.gpu = String(projectDeservedGpus + 1);
                    await kubeCli.createWorkload(workload);

                        const interactiveWorkload = await kubeCli.getInteractiveWorkload(
                            workloadName,
                            namespace
                        );
                    workloadId = interactiveWorkload.metadata.uid;
                    await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                        workloadId,
                        Phase.Pending,
                        workloadName
                    );
                });

                test('validate phaseReason and pendingSchedulingMessages', async () => {
                    const apiWorkload = await workloadsUtilsV2.waitForApiWorkloadWithPhaseReason(
                        workloadId,
                        PhaseReason.NonPreemptibleOverQuota,
                        workloadName
                    );

                    expect(apiWorkload.phaseReason).toBe(PhaseReason.NonPreemptibleOverQuota);
                    expect(apiWorkload.pendingSchedulingMessages).toHaveLength(1);
                    expect(apiWorkload.pendingSchedulingMessages[0].nodePool).toBe('default');
                    expect(apiWorkload.pendingSchedulingMessages[0].phaseReason).toBe(
                        PhaseReason.NonPreemptibleOverQuota
                    );
                    expect(apiWorkload.pendingSchedulingMessages[0].reason).toBe(
                        'Non-preemptible workload is over quota'
                    );
                    expect(apiWorkload.pendingSchedulingMessages[0].orgType).toBe(OrgType.Project);
                    expect(apiWorkload.pendingSchedulingMessages[0].userMessage).toBe(
                        'GPU quota exceeded for non-preemptible workloads.\nRequested: 3 devices\nProject quota: 2 devices, with 0 devices already allocated for non-preemptible'
                    );

                        const apiWorkloadFromList = (
                            await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)
                        ).body;
                    expect(apiWorkloadFromList.phaseReason).toBe(PhaseReason.NonPreemptibleOverQuota);
                });

                test('increase the project quota and wait for workload to be running', async () => {
                    await orgUnitUtils.patchProjectResources(projectId, [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: projectDeservedGpus * 2, limit: projectDeservedGpus * 2 }
                        }
                    ]);
                    await orgUnitUtils.waitForProjectFromBackendToBeReady(project.name, projectId);

                    await waitUtils.waitForJobToBeRunning(
                        workload.name,
                        namespace,
                        crds.runaiInteractiveWorkloadCrd.kind
                    );
                });

                test('validate phaseReason and pendingSchedulingMessages are nil', async () => {
                    await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                        workloadId,
                        Phase.Running,
                        workloadName
                    );

                    const apiWorkload = await workloadsUtilsV2.waitForApiWorkloadWithPhaseReason(
                        workloadId,
                        null,
                        workloadName
                    );

                    expect(apiWorkload.phaseReason).toBeUndefined();
                    expect(apiWorkload.pendingSchedulingMessages).toBeUndefined();

                        const apiWorkloadFromList = (
                            await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)
                        ).body;
                    expect(apiWorkloadFromList.phaseReason).toBeUndefined();
                });
            });
            })
        );
    });
});
