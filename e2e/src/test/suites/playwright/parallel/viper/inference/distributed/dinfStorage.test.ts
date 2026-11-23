/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    credentialUtils,
    dataSourcesUtils,
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
import { test } from '@playwright/test';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { createDistributedInference } from '../inferenceUtils.js';
import { getBaseWorkloadSpec } from './dinfUtils';
import { SecretType } from '../../../../../../../modules/api/apis/k8sObjectsTrackerApi/generated/model/secretType';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { DistributedInferenceLeaderWorkerSpecCompute } from '../../../../../../../modules/api/apis/workloadsApi/generated/api';

supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe(
        `distributed inference workload storage tests: ${tags.resource.distInferenceWorkload} ${tags.tool.api} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit}`,
        { tag: [tags.resource.distInferenceWorkload] },
        function () {
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let workloadSpec: DistributedInferenceSpecSpec;

            // this is the command that the inference hello world image is using
            // we run it after our custom commands in order to get the server up and running
            const HELLO_WORLD_SERVER_COMMAND = '/server';

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

                workloadSpec = getBaseWorkloadSpec();
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

            test('creates a workload with large shm', async function () {
                workloadSpec.leader.command = `/bin/sh -c "echo hello > /dev/shm/hello && df /dev/shm && ${HELLO_WORLD_SERVER_COMMAND}"`;
                workloadSpec.leader.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.leader.compute.largeShmRequest = true;
                workloadSpec.worker = workloadSpec.leader;

                // create the workload and wait for it to be in phase Running
                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            test('creates a workload with existing pvc', async function () {
                const pvcName = objectGenerator.generateShortNameByPrefix(`pvc-dinf-${workloadName}`);
                const pvc = await dataSourcesUtils.createClusterPvc(pvcName, project.status.namespace);

                workloadSpec.leader.command = `/bin/sh -c "ls /path && ${HELLO_WORLD_SERVER_COMMAND}"`;
                workloadSpec.leader.storage = {
                    pvc: [
                        {
                            name: 'pvc1',
                            claimName: pvc.metadata.name,
                            path: '/path',
                            existingPvc: true
                        }
                    ]
                };
                workloadSpec.worker = workloadSpec.leader;

                // create the workload and wait for it to be in phase Running
                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            test('creates a workload with new pvc', async function () {
                const claimName = workloadName + '-claim';

                workloadSpec.leader.storage = {
                    pvc: [
                        {
                            name: 'vol1',
                            claimName: claimName,
                            path: '/data',
                            existingPvc: false,
                            claimInfo: {
                                // didn't specify storage class, so it will use the default one
                                size: '1Gi',
                                accessModes: { readWriteOnce: true }
                            }
                        }
                    ]
                };
                workloadSpec.worker = workloadSpec.leader;

                // create the workload and wait for it to be in phase Running
                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info(`workload ${workloadName} is in phase Running ${workloadId}`);

                // check that the pvc was created
                await waitUtils.getPvc(claimName, project.status.namespace);
                logger.info(`pvc ${claimName} was created`);

                // delete the workload and check that the pvc is deleted
                await workloadsCrudUtils.deleteDistributedInference(workloadId);
                await waitUtils.waitForPvc(claimName, project.status.namespace, 90, false);
                logger.info(`pvc ${claimName} was deleted`);
            });

            test('creates a workload with configmap', async function () {
                const cmName = objectGenerator.generateShortNameByPrefix(`cm-dinf-${workloadName}`);
                const cm = await dataSourcesUtils.createClusterConfigMap(cmName, project.status.namespace);

                workloadSpec.leader.command = `/bin/sh -c "ls /path && ${HELLO_WORLD_SERVER_COMMAND}"`;
                workloadSpec.leader.storage = {
                    configMapVolume: [
                        {
                            name: 'cm1',
                            configMap: cm.metadata.name,
                            mountPath: '/path'
                        }
                    ]
                };
                workloadSpec.worker = workloadSpec.leader;

                // create the workload and wait for it to be in phase Running
                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            test('creates a workload with secret', async function () {
                const secretName = objectGenerator.generateShortNameByPrefix(`secret-dinf-${workloadName}`);
                const secret = await credentialUtils.createExistingSecret(
                    SecretType.AccessKey,
                    false,
                    undefined,
                    project.status.namespace
                );

                workloadSpec.leader.command = `/bin/sh -c "ls /path && ${HELLO_WORLD_SERVER_COMMAND}"`;
                workloadSpec.leader.storage = {
                    secretVolume: [
                        {
                            name: 'secret1',
                            secret: secret.name,
                            mountPath: '/path'
                        }
                    ]
                };
                workloadSpec.worker = workloadSpec.leader;

                // create the workload and wait for it to be in phase Running
                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            test('creates a workload with emptydir', async function () {
                workloadSpec.leader.command = `/bin/sh -c "ls /path && ${HELLO_WORLD_SERVER_COMMAND}"`;
                workloadSpec.leader.storage = {
                    emptyDirVolume: [
                        {
                            name: 'emptydir1',
                            path: '/path',
                            medium: 'Memory',
                            sizeLimit: '1Gi'
                        }
                    ]
                };
                workloadSpec.worker = workloadSpec.leader;

                // create the workload and wait for it to be in phase Running
                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });
        }
    );
});
