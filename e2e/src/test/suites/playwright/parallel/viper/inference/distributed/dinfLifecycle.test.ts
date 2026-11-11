/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isLwsInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { test, TestInfo } from '@playwright/test';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { distributedInferenceUtils } from '../../../../../../../modules/helpers';
import {
    Probe,
    Probes,
    ProbeHandler,
    ProbeHandlerHttpGet,
    DistributedInferenceLeaderWorkerSpecCompute
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe(
        `distributed inference workload lifecycle tests: ${tags.resource.distInferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`,
        { tag: [tags.resource.distInferenceWorkload] },
        function () {
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let workloadSpec: DistributedInferenceSpecSpec;

            test.beforeAll(async function () {
                if (!(await isLwsInstalled())) {
                    logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('workload');
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

                workloadSpec = distributedInferenceUtils.getBaseWorkloadSpec();
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.status.namespace}; error: ${error}`
                    );
                }
            });

            test.afterEach(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads of project ${project.status.namespace}; error: ${error}`
                    );
                }
            });

            test('creates a workload that becomes ready', async function () {
                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
                // wait for the workload to be in phase Running
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    expectedPods
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
            });

            test('creates a workload that is failed', async function () {
                workloadSpec.leader.command = 'no-such-command';

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Failed
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Failed,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a workload that is pending', async function () {
                workloadSpec.leader.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.leader.compute.gpuDevicesRequest = 256;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Pending
                // verify it's not just a temporary phase as with any workload by sleeping for a while and checking again
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Pending, 120);
                await waitUtils.sleep(10);
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Pending,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a workload that is initializing due to non-existing image', async function () {
                workloadSpec.leader.image = 'non-existing-image';
                workloadSpec.worker.image = 'non-existing-image';

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Initializing
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Initializing,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a workload that is initializing (with a server that doesnt open a port)', async function () {
                workloadSpec.leader.image = config.quickStartImage.ubuntu;
                workloadSpec.leader.command = 'sleep 100s';

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Initializing
                // verify it's not just a temporary phase as with any workload by sleeping for a while and checking again
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Initializing, 120);
                await waitUtils.sleep(10);
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Initializing,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a workload that becomes ready (with a server that takes time to initialize)', async function () {
                // sleep a little bit and then run the server
                workloadSpec.leader.command = '/bin/sh -c';
                workloadSpec.leader.args = '"sleep 10s && /server"';

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Running
                // (the workload is expected to be in phase Initializing for a while, then Running, but we don't test it
                // specifically to avoid flakiness)
                const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
                // wait for the workload to be in phase Running
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    expectedPods
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
            });

            test('creates a workload that becomes ready (with a non-default readiness probe)', async function ({}, testInfo: TestInfo) {
                workloadSpec.leader.probes = new Probes();
                workloadSpec.leader.probes.readiness = new Probe();
                workloadSpec.leader.probes.readiness.handler = new ProbeHandler();
                workloadSpec.leader.probes.readiness.handler.httpGet = new ProbeHandlerHttpGet();
                workloadSpec.leader.probes.readiness.handler.httpGet.path = '/';
                workloadSpec.leader.probes.readiness.handler.httpGet.port = 8080;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
                // wait for the workload to be in phase Running
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    expectedPods
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
            });

            test('creates a workload that is degraded (ready but the custom url has issues)', async function () {
                workloadSpec.servingPort.exposedUrl = 'http://underscore_is_invalid_in_urls.example.com';

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Degraded
                const workload = await waitUtils.waitForBackendWorkloadPhase(
                    workloadName,
                    project.id,
                    Phase.Degraded,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a multi-replica workload that is failed (all replicas are failed)', async function () {
                workloadSpec.replicas = 2;
                workloadSpec.leader.command = 'no-such-command';
                workloadSpec.worker.command = 'no-such-command';

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Failed
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Failed,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a multi-replica workload that is pending (all replicas are pending)', async function () {
                workloadSpec.replicas = 2;
                workloadSpec.leader.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.leader.compute.gpuDevicesRequest = 256;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Pending
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Pending,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a multi-replica workload that is degraded (one replica is ready, the other is failed)', async function () {
                workloadSpec.replicas = 2;
                // in order for one replica to succeed and the other to fail, we'll use the
                // HOSTNAME env var, which looks like: workloadname-0, workloadname-1, etc.
                // we'll fail if the hostname doesn't end with -0
                workloadSpec.leader.command =
                    '/bin/sh -c "if [[ \\"$HOSTNAME\\" == *\\"-0\\" ]]; then /server; else exit 1; fi"';

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Degraded
                const workload = await waitUtils.waitForBackendWorkloadPhase(
                    workloadName,
                    project.id,
                    Phase.Degraded,
                    120
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec, false);
            });

            test('creates a multi-replica workload that is initializing (one replica is ready, the other is pending)', async function ({}, testInfo: TestInfo) {
                workloadSpec.replicas = 2;
                // we have 2 gpus in the project, so we'll request 2 per replica to make sure one is pending.
                // note: this test relies on LWS to be gang-scheduled, i.e. each replica should have its own podgroup and
                // be scheduled as a gang. if this test fails, it might be because LWS is not gang-scheduled! (and then,
                // the 2 gpus are consumed by the two leaders, instead of by the first replica's leader and worker).
                workloadSpec.leader.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.leader.compute.gpuDevicesRequest = 1;
                workloadSpec.worker.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.worker.compute.gpuDevicesRequest = 1;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Initializing, then wait for the workload to have 1 ready replica
                // (as it might be in phase Initializing even before the first replica is ready). it's important for
                // resource validation to first make sure we have exactly 1 ready replica
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Initializing, 120);
                const workload = await distributedInferenceUtils.waitForReadyReplicas(workloadId, 1, 120);

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
            });
        }
    );
});
