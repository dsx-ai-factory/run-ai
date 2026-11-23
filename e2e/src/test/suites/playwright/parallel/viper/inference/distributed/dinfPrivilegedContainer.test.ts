/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    policyUtils,
    tags,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { StatusCodes } from 'http-status-codes';
import { workloadsCrudService, policyApi as policyApiHandler } from '../../../../../../../modules/api';
import { kubeCli } from '../../../../../../../modules/cli';
import { getBaseWorkloadSpec } from './dinfUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { ScopeType } from '../../../../../../../modules/api/apis/policyApi';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/policyUtils';
import { isLwsInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';

test.describe(
    `distributed inference - security container configuration ${tags.resource.distInferenceWorkload} ${tags.operations.submit} ${tags.teams.viper}`,
    { tag: [tags.resource.distInferenceWorkload] },
    () => {
        supportedFrom(SupportedVersions.V2_23)(function () {
            let project: Project;
            let policyName: string;

            test.beforeAll(async () => {
                if (!(await isLwsInstalled())) {
                    logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                    test.skip();
                }
            });

            test.beforeEach(async () => {
                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('p'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 1, limit: 1 } }]
                });
                workloadsCrudService.distributedInferencesApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
                policyName = `project-${project.id}`;
            });

            test.afterEach(async () => {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            });

            test('Should fail to submit privileged distributed inference workload without policy', async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('i');
                const workloadSpec: DistributedInferenceSpecSpec = getBaseWorkloadSpec();
                workloadSpec.leader.security = { privileged: true };

                await expect(
                    workloadsCrudUtils.createDistributedInference(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        workloadSpec
                    )
                ).rejects.toMatchObject({
                    response: {
                        body: {
                            complianceIssues: [
                                {
                                    field: 'leader.security.privileged',
                                    details: 'the administrator prohibited changing the value of this field'
                                }
                            ]
                        }
                    }
                });
            });

            test('Should succeed to submit privileged distributed inference workload when policy allows it', async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('i');
                const workloadSpec: DistributedInferenceSpecSpec = getBaseWorkloadSpec();
                workloadSpec.leader.security = { privileged: true };

                await test.step('Create workload policy with privileged.canEdit=true', async () => {
                    const rules = {
                        leader: {
                            security: {
                                privileged: { canEdit: true }
                            }
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        WorkloadType.DistributedInference,
                        undefined,
                        rules,
                        ScopeType.Project
                    );
                });

                await test.step('Submit privileged distributed inference workload', async () => {
                    const result = await workloadsCrudUtils.createDistributedInference(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        workloadSpec
                    );
                    expect(result.response.statusCode).toBe(StatusCodes.ACCEPTED);

                    const pods = await waitUtils.waitForPods(
                        result.body.name,
                        project.status.namespace,
                        2,
                        60,
                        ['Running'],
                        true
                    );
                    expect(pods.length).toBe(2);
                    expect(pods[0].spec.containers[0].securityContext?.privileged).toBeTruthy();

                    logger.info('Verifying privileged behavior by running "mount" inside the container');
                    const mountCheck = await kubeCli.execInPod(
                        pods[0].metadata.name,
                        project.status.namespace,
                        ['sh', '-c', 'mkdir -p /mnt/test && mount -t tmpfs tmpfs /mnt/test']
                    );
                    expect(mountCheck.stderr).toBe('');
                });

                await test.step('Cleanup workload policy', async () => {
                    await policyUtils.deleteWorkloadPolicyV2(WorkloadType.DistributedInference, {
                        scope: ScopeType.Project.toString(),
                        scopeId: project.id.toString()
                    });
                });
            });

            supportedFrom(SupportedVersions.V2_25)(function () {
                test('Should submit distributed inference with hostIpc and hostNetwork on leader and worker', async () => {
                    const workloadName = objectGenerator.generateShortNameByPrefix('i');
                    const workloadSpec: DistributedInferenceSpecSpec = getBaseWorkloadSpec();
                    workloadSpec.leader.security = { hostIpc: true, hostNetwork: true };
                    workloadSpec.worker.security = { hostIpc: true, hostNetwork: true };

                    const result = await workloadsCrudUtils.createDistributedInference(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        workloadSpec
                    );
                    expect(result.response.statusCode).toBe(StatusCodes.ACCEPTED);

                    const pods = await waitUtils.waitForPods(
                        result.body.name,
                        project.status.namespace,
                        2,
                        60,
                        ['Running'],
                        true
                    );
                    expect(pods.length).toBe(2);

                    logger.info('Verifying hostIPC and hostNetwork on pods');
                    for (const pod of pods) {
                        expect(pod.spec.hostIPC).toBeTruthy();
                        expect(pod.spec.hostNetwork).toBeTruthy();
                    }
                });
            });
        });
    }
);
