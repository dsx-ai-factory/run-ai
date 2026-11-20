/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    clusterApiK8sObjectUtils,
    config,
    crds,
    inferenceClientCreator,
    logger,
    objectGenerator,
    operations,
    projectUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { createProject } from '../../../../../../../modules/helpers/models/projects';
import { submitTrainWorkload } from '../../../../../../../modules/helpers/models/workloads';
import { TrainingWorkload } from '../../../../../../../modules/helpers/models/trainingWorkload';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../../modules/cli';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ObjectWrapper } from '../../../../../../../modules/helpers/utils/ClusterApiK8sObjectUtils';
import { WorkloadElement } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadElement';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';

test.describe(
    `Cluster-api k8s-object endpoints, tags: ${tags.teams.viper} ${tags.resource.workloads}`,
    {
        tag: [tags.services.clusterApi, tags.services.workloads, tags.services.externalWorkloadIntegrator]
    },
    function () {
        supportedFrom(SupportedVersions.V2_26)(function () {
            const RETRIEVED_AT_MAX_SKEW_MS = 5 * 60 * 1000;
            const POD_READY_SECONDS = 60;
            const ELEMENT_PROPAGATION_SECONDS = 180;

            function validateWrapperShape(body: ObjectWrapper): void {
                expect(body).to.be.an('object');
                expect(body.object, 'wrapper.object should be present').to.be.an('object');
                expect(body.retrievedAt, 'wrapper.retrievedAt should be present').to.be.a('string');

                const retrievedAtMs = Date.parse(body.retrievedAt);
                expect(retrievedAtMs, `retrievedAt should parse as a date: ${body.retrievedAt}`).to.not.be
                    .NaN;
                const skewMs = Math.abs(Date.now() - retrievedAtMs);
                expect(
                    skewMs,
                    `retrievedAt should be within ${RETRIEVED_AT_MAX_SKEW_MS}ms of now (got skew ${skewMs}ms)`
                ).to.be.lessThan(RETRIEVED_AT_MAX_SKEW_MS);

                const metadata = body.object?.metadata as Record<string, any> | undefined;
                expect(
                    metadata?.managedFields,
                    'metadata.managedFields must be stripped from the returned k8s object'
                ).to.be.undefined;
                expect(
                    metadata?.annotations?.['kubectl.kubernetes.io/last-applied-configuration'],
                    'kubectl.kubernetes.io/last-applied-configuration annotation must be stripped'
                ).to.be.undefined;

                expect(metadata?.resourceVersion, 'metadata.resourceVersion must be present').to.be.a(
                    'string'
                );
                expect(metadata?.creationTimestamp, 'metadata.creationTimestamp must be present').to.be.a(
                    'string'
                );
                expect(body.object.spec, 'object.spec must be present').to.be.an('object');
            }

            function validateObjectKindAndName(
                body: ObjectWrapper,
                expectedKind: string,
                expectedName: string
            ): void {
                expect(body.object.kind, `object.kind should be ${expectedKind}`).to.equal(expectedKind);
                expect(body.object.metadata?.name, `object.metadata.name should be ${expectedName}`).to.equal(
                    expectedName
                );
            }

            let project: Project;
            let namespace: string;

            test.beforeAll(async function () {
                const options = { maxAllowedGpus: -1, deservedGpus: 1 };
                project = await createProject('object-endpoints', options);
                namespace = Utils.namespaceFromProjectName(project.name);
            });

            test.afterAll(async function () {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (error) {
                    logger.error(`failed to delete project: ${project.name}`, error);
                }
                await projectUtils.waitForProjectCrdToBeDeleted(project.name);
            });

            shouldRunNativeWorkloadCrdTests(() =>
            test.describe('TrainingWorkload', function () {
                let trainingWorkload: TrainingWorkload;
                let workloadId: string;
                let podId: string;
                let podName: string;

                test.beforeAll(async function () {
                    trainingWorkload = await submitTrainWorkload(
                        'obj-train',
                        project.name,
                        config.quickStartImage.training
                    );
                    workloadId = await workloadsUtilsV2.getWorkloadIdByWorkloadNameAndNamespace(
                        crds.runaiTrainingWorkloadCrd,
                        trainingWorkload.name,
                        namespace
                    );
                    await workloadsUtilsV2.waitForWorkloadInControlPlane(workloadId);

                    await waitUtils.waitForCondition(
                        async () => {
                            const res = await workloadsUtilsV2.getWorkloadPods(workloadId);
                            return (res.body.pods?.length ?? 0) > 0;
                        },
                        'wait for at least one pod for the training workload',
                        true,
                        POD_READY_SECONDS
                    );

                    const podsRes = await workloadsUtilsV2.getWorkloadPods(workloadId);
                    const firstPod = podsRes.body.pods[0];
                    podId = firstPod.id;
                    podName = firstPod.name;
                    logger.info(
                        `training workload ${trainingWorkload.name} id=${workloadId} pod=${podName} podId=${podId}`
                    );
                });

                test.afterAll(async function () {
                    if (trainingWorkload?.name) {
                        try {
                            await kubeCli.deleteCrd(
                                crds.runaiTrainingWorkloadCrd,
                                trainingWorkload.name,
                                namespace
                            );
                        } catch (error) {
                            logger.debug(
                                `failed to delete training workload ${trainingWorkload.name}: ${error}`
                            );
                        }
                    }
                    trainingWorkload = undefined;
                });

                test('GET workloads/{id}/object returns wrapped TrainingWorkload CR', async function () {
                    const res = await clusterApiK8sObjectUtils.getWorkloadObject(workloadId);
                    expect(res.status, `body: ${JSON.stringify(res.data)}`).to.equal(StatusCodes.OK);
                    validateWrapperShape(res.data);
                    validateObjectKindAndName(res.data, 'TrainingWorkload', trainingWorkload.name);
                    expect(res.data.object.apiVersion).to.match(/^run\.ai\//);
                    expect(res.data.object.metadata?.namespace).to.equal(namespace);
                });

                test('GET workloads/{id}/pods/{podId}/object returns wrapped Pod manifest', async function () {
                    const res = await clusterApiK8sObjectUtils.getPodObject(workloadId, podId);
                    expect(res.status, `body: ${JSON.stringify(res.data)}`).to.equal(StatusCodes.OK);
                    validateWrapperShape(res.data);
                    validateObjectKindAndName(res.data, 'Pod', podName);
                    expect(res.data.object.apiVersion).to.equal('v1');
                    expect(res.data.object.metadata?.namespace).to.equal(namespace);
                });

                test('GET workloads/{id}/pods/{podId}/object with non-existent podId returns 404', async function () {
                    const fakePodId = '00000000-0000-0000-0000-000000000000';
                    const res = await clusterApiK8sObjectUtils.getPodObject(workloadId, fakePodId);
                        expect(res.status, `body: ${JSON.stringify(res.data)}`).to.equal(
                            StatusCodes.NOT_FOUND
                        );
                });

                test('GET workloads/{id}/object with valid-but-unknown workloadId returns 403', async function () {
                    const unknownButValidUuid = '11111111-1111-1111-1111-111111111111';
                    const res = await clusterApiK8sObjectUtils.getWorkloadObject(unknownButValidUuid);
                        expect(res.status, `body: ${JSON.stringify(res.data)}`).to.equal(
                            StatusCodes.FORBIDDEN
                        );
                });

                test('GET workloads/{id}/object with malformed workloadId returns 400', async function () {
                    const res = await clusterApiK8sObjectUtils.getWorkloadObject('not-a-uuid');
                        expect(res.status, `body: ${JSON.stringify(res.data)}`).to.equal(
                            StatusCodes.BAD_REQUEST
                        );
            });
                })
            );

            test.describe('External Deployment workload (kubernetesRef flow)', function () {
                const deploymentName = objectGenerator.generateNameByPrefix('obj-dep');
                let deploymentWorkloadId: string;
                let element: WorkloadElement;

                test.beforeAll(async function () {
                    const deployment = await kubeCli.createDeployment(
                        inferenceClientCreator.getDeployment(
                            namespace,
                            project.name,
                            deploymentName,
                            1,
                            config.quickStartImage.inference
                        )
                    );

                    deploymentWorkloadId = deployment.metadata.uid;
                    await workloadsUtilsV2.waitForWorkloadInControlPlane(deploymentWorkloadId);

                    await waitUtils.waitForCondition(
                        async () => {
                            const res = await workloadsUtilsV2.getWorkloadElements(deploymentWorkloadId);
                            return (res.body.elements?.length ?? 0) > 0;
                        },
                        'structureElements visible via workloads-service',
                        true,
                        ELEMENT_PROPAGATION_SECONDS
                    );
                    const elemRes = await workloadsUtilsV2.getWorkloadElements(deploymentWorkloadId);
                    element = elemRes.body.elements[0];
                    logger.info(`external workload elementId=${element.id}`);
                });

                test.afterAll(async function () {
                    try {
                        await kubeCli.deleteDeployment(deploymentName, namespace);
                    } catch (error) {
                        logger.warn(`failed to delete deployment ${deploymentName}: ${error}`);
                    }
                });

                test('GET workloads/{deploymentWorkloadId}/object returns the wrapped Deployment manifest', async function () {
                    const res = await clusterApiK8sObjectUtils.getWorkloadObject(deploymentWorkloadId);
                    expect(res.status, `body: ${JSON.stringify(res.data)}`).to.equal(StatusCodes.OK);
                    validateWrapperShape(res.data);
                    validateObjectKindAndName(res.data, 'Deployment', deploymentName);
                    expect(res.data.object.apiVersion).to.equal('apps/v1');
                    expect(res.data.object.metadata?.namespace).to.equal(namespace);
                });

                test('GET workloads/{id}/elements/{elementId}/object returns wrapped element CR', async function () {
                    const res = await clusterApiK8sObjectUtils.getElementObject(
                        deploymentWorkloadId,
                        element.id
                    );
                    expect(res.status, `body: ${JSON.stringify(res.data)}`).to.equal(StatusCodes.OK);
                    validateWrapperShape(res.data);

                    // Cluster-api must resolve to the same object workloads-service indexed.
                    const gvk = element.kubernetesRef?.gvk;
                    const expectedApiVersion = gvk?.group ? `${gvk.group}/${gvk.version}` : gvk?.version;
                    expect(res.data.object.kind).to.equal(gvk?.kind);
                    expect(res.data.object.apiVersion).to.equal(expectedApiVersion);
                    expect(res.data.object.metadata?.name).to.equal(element.kubernetesRef?.name);
                });
            });
        });
    }
);
