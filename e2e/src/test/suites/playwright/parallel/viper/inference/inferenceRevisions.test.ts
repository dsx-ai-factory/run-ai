/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';

import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { createInference, updateInferenceSpec } from './inferenceUtils';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { workloadsService } from '../../../../../../modules/api';
import { HistoryRecordType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/historyRecordType';

shouldRunNativeWorkloadCrdTests(() => {
supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe.serial(
        `inference workload revisions lifecycle tests, tags: ${tags.runOn.nightly} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
        { tag: [tags.resource.inferenceWorkload] },
        function () {
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let workloadSpec: InferenceSpecSpec;
            let revisionId: string;
            let revisionName: string;

            test.beforeAll(async function () {
                if (!(await isKnativeInstalled())) {
                    logger.warn('knative is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });

                // Initialize workload name and spec once for all tests
                workloadName = objectGenerator.generateNameByPrefix('workload');

                workloadSpec = {
                    image: config.quickStartImage.inferenceHelloWorld,
                    servingPort: {
                        container: 8080
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 2,
                        metric: 'throughput',
                        metricThreshold: 100
                    }
                };

                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                workloadsService.revisionsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                    workloadsService.podsApi.accessToken =
                        flowRootState.users.system_administrator.accessToken;
                    workloadsService.eventsApi.accessToken =
                        flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadUtils.clearProjectWorkloadsAndPods(project.name, 120);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and pods of project ${project.name} namespace ${project.status.namespace}`
                    );
                }

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}`
                    );
                }
            });

            test('create inference workload', async function () {
                logger.debug('create inference workload');
                [workloadId, workloadName] = await createInference(
                    workloadName,
                    workloadSpec,
                    project,
                    false
                );

                // wait for the workload to be in phase Running
                logger.debug('wait for workload to be Running');
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            test('update inference workload', async function () {
                expect(workloadId, 'workloadId should be set from previous test').to.not.be.undefined;

                // update the workload
                workloadSpec.image = config.quickStartImage.inferenceAutoScale;
                workloadSpec.autoscaling.maxReplicas = 4;
                workloadSpec.compute = {
                    cpuCoreRequest: 0.2,
                    cpuMemoryRequest: '123M'
                };
                await updateInferenceSpec(workloadId, workloadSpec);

                // wait for the new revision to be in phase Running
                revisionName = workloadName + '-00002';
                const revisionFromGetRevisions = await waitUtils.waitForBackendRevisionPhase(
                    revisionName,
                    workloadId,
                    Phase.Running,
                    120
                );
                expect(revisionFromGetRevisions).to.not.be.undefined;
                revisionId = revisionFromGetRevisions.id;
            });

            test('check get revision by id', async function () {
                expect(revisionId, 'revisionId should be set from previous test').to.not.be.undefined;
                expect(revisionName, 'revisionName should be set from previous test').to.not.be.undefined;

                logger.debug('get revision by id');
                const revision = (await workloadsService.revisionsApi.getRevision(revisionId)).body;
                expect(revision).to.not.be.undefined;
                expect(revision.name).to.equal(revisionName);
                expect(revision.phase).to.equal(Phase.Running);
                expect(revision.workloadId).to.equal(workloadId);
                expect(revision.images).to.have.lengthOf(1);
                expect(revision.images[0]).to.equal(workloadSpec.image);
                expect(revision.requestedPods.min).to.equal(workloadSpec.autoscaling.minReplicas);
                expect(revision.requestedPods.max).to.equal(workloadSpec.autoscaling.maxReplicas);
                expect(revision.requestedPods.number).to.equal(1);
            });

            test('check get inference workload revisions', async function () {
                expect(workloadId, 'workloadId should be set from previous test').to.not.be.undefined;
                expect(workloadName, 'workloadName should be set from previous test').to.not.be.undefined;

                logger.debug('get inference revisions by workload id');
                const revisions = (
                    await workloadsService.revisionsApi.getInferenceWorkloadRevisions(workloadId)
                ).body.revisions;
                expect(revisions).to.have.lengthOf(2);
                expect(revisions.map((r) => r.name)).to.include(workloadName + '-00001');
                expect(revisions.map((r) => r.name)).to.include(workloadName + '-00002');
            });

            test('check count inference workload revisions', async function () {
                expect(workloadId, 'workloadId should be set from previous test').to.not.be.undefined;

                logger.debug('count inference revisions by workload id');
                const count = (
                    await workloadsService.revisionsApi.countInferenceWorkloadRevisions(workloadId)
                ).body.count;
                expect(count).to.equal(2);
            });

            test('check get revision pods', async function () {
                expect(revisionId, 'revisionId should be set from previous test').to.not.be.undefined;
                expect(workloadId, 'workloadId should be set from previous test').to.not.be.undefined;
                expect(workloadName, 'workloadName should be set from previous test').to.not.be.undefined;

                logger.debug('get revision pods');
                const pods = await waitUtils.waitForBackendRevisionPodsPhase(revisionId, 'Running');
                expect(pods).to.have.lengthOf(1);
                const pod = pods[0];
                expect(pod.name).to.have.string(revisionName);
                expect(pod.revisionId).to.equal(revisionId);
                expect(pod.workloadId).to.equal(workloadId);
                expect(pod.k8sPhase).to.equal('Running');
                expect(pod.requestedResources.cpu.request).to.greaterThanOrEqual(
                    workloadSpec.compute.cpuCoreRequest
                ); // there's some overhead beyond the requested value
                expect(pod.requestedResources.cpuMemory.request).to.equal(
                    workloadSpec.compute.cpuMemoryRequest
                );
            });

            test('check get revision events', async function () {
                expect(revisionId, 'revisionId should be set from previous test').to.not.be.undefined;

                logger.debug('get revision events');
                const events = await waitUtils.waitForBackendRevisionEvents(revisionId, 120, 2, 1);
                expect(events).to.not.be.undefined;
                expect(events).to.have.lengthOf.at.least(1);
                // we expect at least one event - on the revision
                let foundRevisionEvent = false;
                for (const event of events) {
                    if (event.involvedObject.kind === 'Revision') {
                        expect(event.involvedObject.uid).to.equal(revisionId);
                        foundRevisionEvent = true;
                    }
                }
                expect(foundRevisionEvent).to.be.true;
            });

            test('check get revision history', async function () {
                expect(revisionId, 'revisionId should be set from previous test').to.not.be.undefined;

                logger.debug('get revision history');
                const history = (await workloadsService.eventsApi.getRevisionHistory(revisionId)).body
                    .records;
                // we expect at least three history entries - an event and two phase change
                expect(history).to.have.lengthOf.at.least(2);
                let foundEvent = false;
                let numberOfPhaseChanges = 0;
                for (const record of history) {
                    if (record.meta.type === HistoryRecordType.Event) {
                        if (record.spec.event.involvedObject.kind === 'Revision') {
                            expect(record.spec.event.involvedObject.uid).to.equal(revisionId);
                            foundEvent = true;
                        }
                    } else if (record.meta.type === HistoryRecordType.PhaseUpdate) {
                        if (
                            record.spec.phaseUpdate.phase === Phase.Running ||
                            record.spec.phaseUpdate.phase === Phase.Pending ||
                            record.spec.phaseUpdate.phase === Phase.Initializing
                        ) {
                            numberOfPhaseChanges++;
                        }
                    }
                }
                expect(foundEvent).to.be.true;
                expect(numberOfPhaseChanges).to.be.greaterThan(2);
            });
        }
    );
});
});
