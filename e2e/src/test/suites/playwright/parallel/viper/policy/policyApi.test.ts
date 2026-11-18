/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { policyApi as policyApiHandler } from '../../../../../../modules/api';
import {
    CleanPodPolicy,
    DistributedPolicyChangeRequestV2,
    ScopeType,
    WorkspacePolicyChangeRequestV2,
    WorkspacePolicyOverwriteRequestV2,
    WorkspacePolicyV2
} from '../../../../../../modules/api/apis/policyApi';
import { kubeCli } from '../../../../../../modules/cli';
import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    versionUtils,
    workloadsCrudUtils,
    workloadsManagerUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import http from 'http';
import { test, TestInfo } from '@playwright/test';
import { DistributedSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedSpecSpec';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

test.describe(
    `Policy API, tags: ${tags.resource.policy} ${tags.tool.api} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.services.policy, tags.runOn.parallelCI] },
    function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let policy: WorkspacePolicyV2;
        let policyHttpResponse: http.IncomingMessage;
        const policyName = 'demo-policy';
        const shouldValidate = false;

        test.describe('CRUD', function () {
            test.beforeAll(async function () {
                policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
                projectName = objectGenerator.generateNameByPrefix('project-api-for-policy');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                });
            });

            test.afterAll('deleting project and namespace', async () => {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} an namespace ${namespace}`);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                    throw error;
                }
            });
            test.beforeEach(async function ({}, testInfo: TestInfo) {
                // Create Policy
                const policyRequest = {
                    meta: {
                        projectId: project.id,
                        scope: ScopeType.Project,
                        name: policyName
                    },
                    policy: {
                        defaults: {
                            args: 'x=1',
                            command: 'test-command',
                            image: 'ubuntu'
                        }
                    }
                } as WorkspacePolicyOverwriteRequestV2;
                const policyResponse = await policyApiHandler.policyApi.overwriteWorkspacePolicyV2(
                    false,
                    shouldValidate,
                    policyRequest
                );
                policy = policyResponse.body;
                policyHttpResponse = policyResponse.response;
            });
            test.afterEach(async function () {
                // CRUD cleanup
                try {
                    await policyApiHandler.policyApi.deleteWorkspacePolicy(
                        'project',
                        undefined,
                        project.id.toString()
                    );
                } catch (e) {
                    if (e.statusCode !== 404) {
                        throw e;
                    }
                }
            });

            test('Should create successfully', function () {
                expect(policyHttpResponse.statusCode).to.equal(StatusCodes.OK);
                expect(policy.effective.defaults.args).to.equal('x=1');
                expect(policy.effective.defaults.command).to.equal('test-command');
                expect(policy.effective.defaults.image).to.equal('ubuntu');
            });

            test('Should update successfully', async function () {
                const policyUpdateRequest = {
                    meta: {
                        projectId: project.id,
                        scope: ScopeType.Project,
                        name: policyName
                    },
                    policy: {
                        defaults: {
                            args: 'x=2',
                            command: 'test-command-2',
                            image: 'alpine'
                        }
                    }
                } as WorkspacePolicyChangeRequestV2;
                const policyResponse = await policyApiHandler.policyApi.overwriteWorkspacePolicyV2(
                    false,
                    shouldValidate,
                    policyUpdateRequest
                );
                expect(policyResponse.response.statusCode).to.equal(StatusCodes.OK);
                expect(policyResponse.body.effective.defaults.args).to.equal('x=2');
                expect(policyResponse.body.effective.defaults.command).to.equal('test-command-2');
                expect(policyResponse.body.effective.defaults.image).to.equal('alpine');
            });

            test('Should delete successfully', async function () {
                await policyApiHandler.policyApi.deleteWorkspacePolicy(
                    'project',
                    undefined,
                    project.id.toString()
                );
                const policyResponse = await policyApiHandler.policyApi.getWorkspacePolicyV2(
                    'project',
                    undefined,
                    project.id.toString()
                );
                const policy = policyResponse.body;
                const emptyPolicyName = 'demo-policy';

                expect(policy.meta.name).to.be.equal(emptyPolicyName);
            });

            test('Create distributed policy and verify its applied', async function () {
                const expectedCleanPodPolicy = CleanPodPolicy.All;
                const policyUpdateRequest = {
                    meta: {
                        projectId: project.id,
                        scope: ScopeType.Project,
                        name: 'dist-policy-demo'
                    },
                    policy: {
                        defaults: {
                            worker: {
                                cleanPodPolicy: expectedCleanPodPolicy
                            }
                        }
                    }
                } as DistributedPolicyChangeRequestV2;

                const policyResponse = await policyApiHandler.policyApi.overwriteDistributedPolicyV2(
                    false,
                    false,
                    policyUpdateRequest
                );
                expect(policyResponse.response.statusCode).to.equal(StatusCodes.OK);

                const workloadNamePrefix = 'dist-clean-pod-policy';
                const distSpec: DistributedSpecSpec = {
                    numWorkers: 1,
                    distributedFramework: DistributedFramework.PyTorch,
                    image: 'ubuntu',
                    command: 'sleep infinity'
                };
                const createResult = await workloadsCrudUtils.createDistributed(
                    workloadNamePrefix,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    distSpec,
                    true
                );
                expect(createResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                const workloadId = createResult.body.workloadId;

                if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Distributed)) {
                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(
                        getResult.body.workload.workloadManifest['spec']['runPolicy']['cleanPodPolicy']
                    ).to.equal(expectedCleanPodPolicy);
                } else {
                    const getResult = await workloadsCrudUtils.getDistributed(workloadId);
                    expect(getResult.body.spec.cleanPodPolicy).to.equal(expectedCleanPodPolicy);
                }
            });
        });
    }
);
