/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    logger,
    tags,
    objectGenerator,
    projectUtils,
    operations,
    Utils,
    orgUnitUtils,
    nodePoolUtils,
    waitUtils,
    crds,
    apiUtils,
    authorizationUtils,
    config
} from '../../../../../../../modules/helpers';
import { kubeCli, runaiCliV2 } from '../../../../../../../modules/cli';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    DeleteFlagsMerged,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { orgUnitService } from '../../../../../../../modules/api';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';

test.describe(
    `Projects Basic API v2, tags:  ${tags.resource.projects} ${tags.operations.submit} ${tags.teams.viper}  ${tags.runOn.k8sFlavCertKit}`,
    { tag: [tags.services.orgUnitService, tags.services.projectController, tags.runOn.serialCI] },
    function () {

 const runaiProjectOnNamespaceLabelKey = 'runai/queue';
 const DefaultResourcePriority = 'MediumLow';

class ProjectStatusCondition {
    'status': string;

    'type': string;

    'reason'?: string;

    'message'?: string;
}
        let defaultNodePoolId: string;

        test.beforeAll(async function () {
            // delete all existing projects in the cluster before starting tests
            await orgUnitUtils.deleteAllProjectsInCluster(flowRootState.cluster.uuid, flowRootState.users.system_administrator.accessToken);

            defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);
            await apiUtils.disableLimitOverQuotaSubscription();

            // update the default department to have null in scheduling rules (interactiveJobMaxIdleDurationSeconds, etc..)
            await orgUnitUtils.validateAndSetDefaultDepartmentGpu();
            const defaultDepartment = await orgUnitUtils.getDefaultDepartment();
            await orgUnitUtils.updateDepartment(defaultDepartment.id, {
                resources: defaultDepartment.resources,
                schedulingRules: {
                    interactiveJobTimeLimitSeconds: null,
                    interactiveJobMaxIdleDurationSeconds: null,
                    interactiveJobPreemptIdleDurationSeconds: null,
                    trainingJobMaxIdleDurationSeconds: null,
                    trainingJobTimeLimitSeconds: null
                }
            });
        });

        test.afterAll(async function () {
            try {
                await operations.forceCleanup();
            } catch (error) {
                logger.error('Failed to delete projects created in test', error);
            }
        });

        test.describe('Basic project CRUD validations', function () {
            let projectName: string;
            let projectId: string;
            let namespaceName: string;
            let rankName: string;

            test('CRUD project and validate crd+backend project status fields', async function () {
                await test.step('Create project and validate', async function () {
                    logger.info('Create project and validate');
                    projectName = objectGenerator.generateNameByPrefix('p');
                    namespaceName = Utils.namespaceFromProjectName(projectName);
                    rankName = 'High';
                    const res = await orgUnitUtils.createProjectAndValidate({
                        schedulingRules: {
                            interactiveJobTimeLimitSeconds: 2,
                            interactiveJobMaxIdleDurationSeconds: 3,
                            interactiveJobPreemptIdleDurationSeconds: 4,
                            trainingJobMaxIdleDurationSeconds: 5,
                            trainingJobTimeLimitSeconds: 6
                        },
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 1 },
                                rank: rankName
                            }
                        ],
                        name: projectName,
                        clusterId: flowRootState.cluster.uuid
                    });
                    projectId = res.id;

                    const projGet = await orgUnitUtils.getProject(projectId);

                    for (const proj of [res, projGet]) {
                        orgUnitUtils.validateProjectResponse(
                            { body: proj },
                            projectName,
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 1 },
                                rank: rankName
                            },
                            '',
                            1,
                            'default',
                            2,
                            3,
                            4,
                            6,
                            5
                        );
                    }

                    // validate project crd+backend fields
                    await orgUnitUtils.waitForProjectFromBackendWithCorrectNamespace(
                        projectName,
                        String(res.id),
                        namespaceName
                    );
                    const projectCrd = await projectUtils.waitForProjectCrdNonEmptyStatus(
                        projectName,
                        'Ready'
                    );
                    expect(projectCrd['status'].namespace).toEqual(namespaceName);
                    expect(projectCrd['status'].phase).toEqual('Ready');
                    const conditions = projectCrd['status'].conditions as Array<ProjectStatusCondition>;
                    expect(conditions).toBeDefined();
                    const namespaceCondition = conditions.find(
                        (condition) => condition.type === 'NamespaceReady'
                    );
                    expect(namespaceCondition.status).toEqual('True');

                    const backendProject = await orgUnitUtils.waitForProjectFromBackendToBeReady(
                        projectName,
                        projectId
                    );
                    expect(backendProject.status.namespace).toEqual(namespaceName);
                    await projectUtils.validateProjectCrd(
                        projectName,
                        res.parent.name,
                        'default',
                        30,
                        Number(res.id)
                    );
                    await orgUnitUtils.validateProjectQueueCrd(projectName, res);
                });

                await test.step('Update project and validate', async function () {
                    logger.info('Update project and validate');
                    const rankNameUpdate = 'Low';

                    const projUpdateRes = await orgUnitUtils.updateProject(projectId, {
                        schedulingRules: {
                            interactiveJobTimeLimitSeconds: 4,
                            interactiveJobMaxIdleDurationSeconds: 6,
                            interactiveJobPreemptIdleDurationSeconds: 8,
                            trainingJobMaxIdleDurationSeconds: 10,
                            trainingJobTimeLimitSeconds: 12
                        },
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 5 },
                                rank: rankNameUpdate
                            }
                        ]
                    });
                    const projGet = await orgUnitUtils.getProject(projectId);

                    for (const proj of [projUpdateRes, projGet]) {
                        orgUnitUtils.validateProjectResponse(
                            { body: proj },
                            projectName,
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 5 },
                                rank: rankNameUpdate
                            },
                            '',
                            1,
                            'default',
                            4,
                            6,
                            8,
                            12,
                            10
                        );
                    }

                    // wait for project to be Ready again
                    await projectUtils.waitForProjectCrdNonEmptyStatus(projectName, 'Ready');
                    await orgUnitUtils.waitForProjectFromBackendToBeReady(projectName, projectId);
                });

                await test.step('Patch project resources and validate', async function () {
                    logger.info('Patch project resources and validate');
                    const rankNamePatch = 'Medium';

                    const projPatchRes = await orgUnitUtils.patchProjectResources(projectId, [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 3 },
                            rank: rankNamePatch
                        }
                    ]);
                    orgUnitUtils.validateResources(1, projPatchRes, {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 3 },
                        rank: rankNamePatch
                    });
                });

                await test.step('Update project resources and validate', async function () {
                    logger.info('Update project resources and validate');
                    const rankNameUpdate = 'Low';

                    const projUpdateRes = await orgUnitUtils.updateProjectResources(projectId, [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 8 },
                            rank: rankNameUpdate
                        }
                    ]);
                    orgUnitUtils.validateResources(1, projUpdateRes, {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 8 },
                        rank: rankNameUpdate
                    });
                });

                await test.step('Delete project and validate deleted', async function () {
                    logger.info('Delete project and validate deleted');
                    let projects = await orgUnitUtils.getProjectsByClusterId(flowRootState.cluster.uuid);
                    expect(projects).toHaveLength(1);
                    const proj = projects[0];

                    const rankNameDelete = 'Low';

                    orgUnitUtils.validateProjectResponse(
                        { body: { ...proj } as Project },
                        projectName,
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 8 },
                            rank: rankNameDelete
                        },
                        '',
                        1,
                        'default',
                        4,
                        6,
                        8,
                        12,
                        10
                    );

                    await orgUnitUtils.deleteProject(projectId);
                    await waitUtils.waitForCrd(crds.queueCrd, proj.name, undefined, 60, false);

                    projects = await orgUnitUtils.getProjectsByClusterId(flowRootState.cluster.uuid);
                    expect(projects).toHaveLength(0);
                });
            });

            test('CRUD project and validate researcher role', async function () {
                await test.step('Create project and validate', async function () {
                    logger.info('Create project and validate');
                    projectName = objectGenerator.generateNameByPrefix('p');
                    namespaceName = Utils.namespaceFromProjectName(projectName);

                    const res = await orgUnitUtils.createProjectAndValidate({
                        schedulingRules: {
                            interactiveJobTimeLimitSeconds: 2,
                            interactiveJobMaxIdleDurationSeconds: 3,
                            interactiveJobPreemptIdleDurationSeconds: 4,
                            trainingJobMaxIdleDurationSeconds: 5,
                            trainingJobTimeLimitSeconds: 6
                        },
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 1 }
                            }
                        ],
                        name: projectName,
                        clusterId: flowRootState.cluster.uuid
                    });
                    projectId = res.id;

                    const researcher = flowRootState.users.l1_researcher;
                    await authorizationUtils.createAccessRuleWithDefaults(
                        researcher.email,
                        SubjectType.User,
                        (flowRootState.users.l1_researcher.role.id),
                        projectId,
                        ScopeType.Project
                    );
                    logger.info('Get org tree from backend service');
                    orgUnitService.projectApi.accessToken = researcher.idToken;
                    const projGet = await orgUnitUtils.getProjectsByClusterId(
                        flowRootState.cluster.uuid,
                        researcher.idToken
                    );
                    expect(projGet).toHaveLength(1);

                    const resCount = await orgUnitService.projectApi.countProjects([
                        `clusterId==${flowRootState.cluster.uuid}`
                    ]);
                    expect(resCount.body.count).toBe(1);
                });
            });

            test.describe('Create project with existing namespace and validate', function () {
                test('Create project with existing namespace and validate', async function () {
                    projectName = objectGenerator.generateNameByPrefix('p');
                    namespaceName = objectGenerator.generateNameByPrefix('n');

                    await kubeCli.CoreV1Api.createNamespace({
                        body: {
                            metadata: {
                                name: namespaceName,
                                labels: { [runaiProjectOnNamespaceLabelKey]: projectName }
                            }
                        }
                    });

                    const res = await orgUnitUtils.createProjectAndValidate({
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 1 }
                            }
                        ],
                        name: projectName,
                        requestedNamespace: namespaceName,
                        clusterId: flowRootState.cluster.uuid
                    });

                    orgUnitUtils.validateProjectResponse(
                        { body: res },
                        projectName,
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 1 },
                            rank: DefaultResourcePriority
                        },
                        namespaceName
                    );

                    // validate project crd+backend fields
                    const projectCrd = await projectUtils.waitForProjectCrdNonEmptyStatus(
                        projectName,
                        'Ready'
                    );
                    expect(projectCrd['status'].namespace).toEqual(namespaceName);
                    expect(projectCrd['status'].phase).toEqual('Ready');
                    const conditions = projectCrd['status'].conditions as Array<ProjectStatusCondition>;
                    expect(conditions).toBeDefined();
                    const namespaceCondition = conditions.find(
                        (condition) => condition.type === 'NamespaceReady'
                    );
                    expect(namespaceCondition.status).toEqual('True');

                    await orgUnitUtils.waitForProjectFromBackendWithCorrectNamespace(
                        projectName,
                        String(res.id),
                        namespaceName
                    );
                    const backendProject = await orgUnitUtils.waitForProjectFromBackendToBeReady(
                        projectName,
                        String(res.id)
                    );

                    expect(backendProject.status.namespace).toEqual(namespaceName);

                    await projectUtils.validateProjectCrd(
                        projectName,
                        res.parent.name,
                        'default',
                        30,
                        Number(res.id),
                        namespaceName
                    );
                    await orgUnitUtils.validateProjectQueueCrd(projectName, res);
                });
            });
        });

      
        test.describe('Create namespace and project and "manually" link between them', function () {
                let projectName: string;
                let projectId: string;
                let namespaceName: string;
                let jobName: string;

                test.beforeAll(function () {
                    cliLoginAsSystemAdmin();
                });

                test.afterAll(async function () {
                    try {
                        cliLoginAsSystemAdmin();
                        const deleteFlags = {
                            [DeleteFlagsMerged.project]: projectName
                        } as Partial<typeof DeleteFlagsMerged>;
                        runaiCliV2.delete(deleteFlags, jobName, cliV2WorkloadType.workspace);
                    } catch (error) {
                        logger.error(`Failed to delete job ${jobName} in project ${projectName}`);
                        logger.debug(`error: ${JSON.stringify(error)}`);
                    }
                    try {
                        await operations.forceCleanup();
                    } catch (error) {
                        logger.error(
                            `Failed to delete project ${projectName} and namespace ${namespaceName}`
                        );
                        logger.debug(`error: ${JSON.stringify(error)}`);
                    }
                });

                test('Submit job to project with manually linked namespace', async function () {
                    projectName = objectGenerator.generateNameByPrefix('p');
                    namespaceName = objectGenerator.generateNameByPrefix('n');

                    // Create namespace without project label first
                    await kubeCli.CoreV1Api.createNamespace({
                        body: {
                            metadata: {
                                name: namespaceName
                            }
                        }
                    });
                    const namespaces = await kubeCli.CoreV1Api.listNamespace();
                    const foundNamespaces = namespaces.items.filter(
                        (namespace) => namespace.metadata.name === namespaceName
                    );
                    expect(foundNamespaces.length).toEqual(1);

                    // Label namespace to link it to the project
                    await kubeCli.labelNamespace(namespaceName, runaiProjectOnNamespaceLabelKey, projectName);

                    // Create project using V2 API with manually created namespace
                    const res = await orgUnitUtils.createProjectAndValidate({
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 1, limit: 1 }
                            }
                        ],
                        name: projectName,
                        requestedNamespace: namespaceName,
                        clusterId: flowRootState.cluster.uuid
                    });
                    projectId = res.id;

                    // Validate project CRD and backend fields
                    const projectCrd = await projectUtils.waitForProjectCrdNonEmptyStatus(projectName);
                    expect(projectCrd['status'].namespace).toEqual(namespaceName);
                    const backendProject = await orgUnitUtils.waitForProjectFromBackendToBeReady(
                        projectName,
                        projectId
                    );
                    expect(backendProject.status.namespace).toEqual(namespaceName);

                    // Submit an interactive job to verify the project works
                    jobName = objectGenerator.generateNameByPrefix('job');
                    const submitFlag = {
                        [SubmitFlagsMerged.image]: config.quickStartImage.interactive,
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.gpuRequest]: '1'
                    } as Partial<typeof SubmitFlagsMerged>;
                    runaiCliV2.submit(submitFlag, jobName, cliV2WorkloadType.workspace);
                    await waitUtils.waitForPods(jobName, namespaceName, 1, 60, ['Running']);
                    await waitUtils.waitForBackendWorkload(jobName, projectId);

                    // Submit another interactive job - should stay Pending as it exceeds project quota
                    const anotherJobName = objectGenerator.generateNameByPrefix('job');
                    runaiCliV2.submit(submitFlag, anotherJobName, cliV2WorkloadType.workspace);
                    await waitUtils.waitForUnschedulableInNamespace(
                        anotherJobName,
                        namespaceName,
                        60,
                        '',
                        true
                    );
                    await waitUtils.waitForBackendWorkload(anotherJobName, projectId);
                });
            });
        });
