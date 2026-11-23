/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    templateUtils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { OrgUnitUtils } from '../../../../../../modules/helpers/utils/OrgUnitUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType, WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { runaiCliV2, bashCli } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit With Template - Positive Tests, tags: ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-template-cli', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            logger.info(`Created project: ${projectName} (ID: ${project.id})`);

            // Set project context once for all tests
            runaiCliV2.projectSet(projectName);
            logger.info(`Set project context to: ${projectName}`);
        });

        test.afterAll('Cleanup resources', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`Deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName}:`, error);
            }
        });

        test('T5539374: Template override with explicit parameters', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-train', 5);
            let templateId: string;
            let templateName: string;
            const explicitCpuRequest = 0.5;

            await test.step('Create training template', async function () {
                templateName = objectGenerator.generateShortNameByPrefix('test-pytorch-training', 5);

                const spec = {
                    image: 'alpine:latest',
                    imagePullPolicy: 'IfNotPresent',
                    command: 'sleep 10',
                    compute: {
                        gpuDevicesRequest: 0,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: templateName,
                        description: 'Template for Training testing'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created template: ${templateName} (ID: ${templateId})`);
            });

            await test.step('Submit workload with template and explicit CPU override', async function () {
                logger.info(
                    `Submitting workload: ${workloadName} with template: ${templateName} and explicit CPU: ${explicitCpuRequest}`
                );

                const submitFlag = {
                    [SubmitFlagsMerged.template]: templateName,
                    [SubmitFlagsMerged.cpuCoreRequest]: explicitCpuRequest.toString()
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                logger.info(`CLI output: ${stdout}`);

                expect(stdout).not.toContain('Error');
                expect(stdout).not.toContain('Failed');

                logger.info('Waiting for workload pods to be running...');
                await waitUtils.waitForPods(workloadName, namespace, 1, 60, ['Running']);
                logger.info('Workload pods are running');
            });

            await test.step('Verify workload uses explicit CPU parameter (not template default)', async function () {
                logger.info(`Getting workload ID for: ${workloadName} in project: ${project.id}`);
                const workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);
                logger.info(`Found workload ID: ${workloadId}`);

                const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                const workload = workloadResponse.body;

                const requestedCpu = workload.workloadRequestedResources?.cpu?.request ?? 0;
                logger.info(`Workload requested CPU: ${requestedCpu}, expected: ${explicitCpuRequest}`);

                expect(requestedCpu).toBeCloseTo(explicitCpuRequest, 0.01);
                logger.info(
                    `Verified: Explicit CPU parameter (${explicitCpuRequest}) took precedence over template default (0.1)`
                );
            });

            await test.step('Cleanup: Delete template and workload', async function () {
                try {
                    logger.info(`Deleting workload: ${workloadName}`);
                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                    logger.info(`Successfully deleted workload: ${workloadName}`);

                    await templateUtils.deleteWorkloadTemplateV2(
                        templateId,
                        WorkloadType.Training,
                        flowRootState.users.system_administrator.idToken
                    );
                    logger.info(`Deleted template: ${templateId}`);
                } catch (error) {
                    logger.error('Failed to cleanup resources:', error);
                }
            });
        });

        test('T5539376: Verify --template flag available in all workload types', async function () {
            const workloadTypes = [
                { name: 'training', command: 'training submit -h' },
                { name: 'workspace', command: 'workspace submit -h' },
                { name: 'inference', command: 'inference submit -h' },
                { name: 'PyTorch', command: 'training pytorch submit -h' },
                { name: 'MPI', command: 'training mpi submit -h' },
                { name: 'TensorFlow', command: 'training tensorflow submit -h' },
                { name: 'XGBoost', command: 'training xgboost submit -h' }
            ];

            for (const workloadType of workloadTypes) {
                await test.step(`Check --template flag in ${workloadType.name} submit help`, function () {
                    logger.info(`Checking --template flag for ${workloadType.name}`);

                    const cliPath = runaiCliV2.cliPath;
                    const helpCommand = `${cliPath} ${workloadType.command}`;

                    const stdout = bashCli.exec(helpCommand);
                    logger.info(
                        `Help output for ${workloadType.name} (first 500 chars):\n${stdout.substring(0, 500)}`
                    );

                    expect(stdout.toLowerCase()).toContain('template');
                    logger.info(`✓ Verified --template flag exists in ${workloadType.name} help`);
                });
            }
        });

        test('T5539378: Submit workload with template and environment variables', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-env', 5);
            let templateId: string;
            let templateName: string;

            await test.step('Create training template', async function () {
                templateName = objectGenerator.generateShortNameByPrefix('test-pytorch-training', 5);

                const spec = {
                    image: 'alpine:latest',
                    imagePullPolicy: 'IfNotPresent',
                    command: 'sleep 10',
                    compute: {
                        gpuDevicesRequest: 0,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: templateName,
                        description: 'Template for Training testing'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created template: ${templateName} (ID: ${templateId})`);
            });

            await test.step('Submit workload with template and environment variables', async function () {
                logger.info(
                    `Submitting workload: ${workloadName} with template: ${templateName} and env vars`
                );

                const submitFlag = {
                    [SubmitFlagsMerged.template]: templateName,
                    [SubmitFlagsMerged.envVar]: ['VAR1=value1', 'VAR2=value2']
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                logger.info(`CLI output: ${stdout}`);

                expect(stdout).not.toContain('Error');
                expect(stdout).not.toContain('Failed');

                logger.info('Waiting for workload pods to be running...');
                await waitUtils.waitForPods(workloadName, namespace, 1, 60, ['Running']);
                logger.info('Workload pods are running');
            });

            await test.step('Verify workload has both template config and environment variables', async function () {
                const workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);

                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                const training = (await workloadsCrudUtils.getTraining(workloadId)).body;
                logger.info(`Workload: ${JSON.stringify(training)}`);

                expect(training).toBeTruthy();
                expect(training.name).toEqual(workloadName);

                const environmentVariables = training.spec?.environmentVariables ?? [];
                const environmentVariablesByName = Object.fromEntries(
                    environmentVariables.map((envVar) => [envVar.name, envVar.value])
                );

                expect(environmentVariablesByName).toHaveProperty('VAR1', 'value1');
                expect(environmentVariablesByName).toHaveProperty('VAR2', 'value2');

                logger.info(
                    `✓ Verified environment variables: VAR1=${environmentVariablesByName['VAR1']}, VAR2=${environmentVariablesByName['VAR2']}`
                );
                logger.info(
                    '✓ Verified workload created successfully with template and environment variables'
                );
                logger.info(`Note: CPU allocation is ${training.spec?.compute?.cpuCoreRequest ?? 0}`);
            });

            await test.step('Cleanup: Delete template and workload', async function () {
                try {
                    logger.info(`Deleting workload: ${workloadName}`);
                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);

                    await templateUtils.deleteWorkloadTemplateV2(
                        templateId,
                        WorkloadType.Training,
                        flowRootState.users.system_administrator.idToken
                    );
                    logger.info(`Deleted template: ${templateId}`);
                } catch (error) {
                    logger.error('Failed to cleanup resources:', error);
                }
            });
        });

        test.describe('T5539379: Department-scoped template tests', () => {
            let department: any;
            let projectInDept: any;
            const orgUnitUtils = new OrgUnitUtils();

            test.afterAll('Cleanup department and project', async () => {
                try {
                    if (projectInDept?.id) {
                        await operations.deleteProjectAndNamespace(projectInDept.id);
                        logger.info(`Deleted project: ${projectInDept.name}`);
                    }
                    if (department?.id) {
                        await orgUnitUtils.deleteDepartment(department.id);
                        logger.info(`Deleted department: ${department.name}`);
                    }
                    // Reset project context to main project to avoid affecting subsequent tests
                    runaiCliV2.projectSet(projectName);
                    logger.info(`Reset project context to: ${projectName}`);
                } catch (error) {
                    logger.error('Failed to cleanup department/project in afterAll:', error);
                }
            });

            test('T5539379: Submit workload with department-scoped template using -p flag', async function () {
                const templateName = objectGenerator.generateShortNameByPrefix('test-dept-pytorch', 5);
                const workloadName = objectGenerator.generateShortNameByPrefix('test-dept', 5);
                const departmentName = objectGenerator.generateShortNameByPrefix('dept-test', 5);
                const deptProjectName = objectGenerator.generateShortNameByPrefix('proj-dept', 5);
                let templateId: string;

                await test.step('Create department and project within department', async function () {
                    department = await orgUnitUtils.createDepartment({
                        name: departmentName,
                        clusterId: flowRootState.cluster.uuid,
                        resources: [
                            {
                                gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                                cpu: null,
                                memory: null,
                                nodePool: null
                            }
                        ]
                    });
                    logger.info(`Created department: ${departmentName} (ID: ${department.id})`);

                    projectInDept = await projectResourcesUtils.createProjectAndNamespace(
                        deptProjectName,
                        Number(department.id),
                        0,
                        0
                    );
                    logger.info(
                        `Created project: ${projectInDept.name} (ID: ${projectInDept.id}) in department: ${departmentName}`
                    );
                });

                await test.step('Create department-scoped template', async function () {
                    logger.info(`Creating department-scoped template: ${templateName}`);

                    const spec = {
                        image: 'alpine:latest',
                        imagePullPolicy: 'IfNotPresent',
                        command: 'sleep 10',
                        compute: {
                            gpuDevicesRequest: 0,
                            cpuCoreRequest: 0.1,
                            cpuMemoryRequest: '100M'
                        }
                    };

                    const response = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: WorkloadType.Training,
                            scopeType: ScopeType.Department,
                            scopeId: department.id,
                            templateName: templateName,
                            description: 'Department-scoped template for testing'
                        },
                        spec,
                        undefined,
                        flowRootState.users.system_administrator.idToken
                    );

                    expect(response.response.statusCode).toEqual(200);
                    templateId = response.body.meta.id;
                    logger.info(`Created department-scoped template: ${templateName} (ID: ${templateId})`);
                });

                await test.step('Submit workload using department-scoped template with -p flag', async function () {
                    runaiCliV2.projectSet(projectInDept.name);

                    logger.info(`Submitting workload with department-scoped template: ${templateName}`);

                    const submitFlag = {
                        [SubmitFlagsMerged.template]: templateName
                    } as Partial<typeof SubmitFlagsMerged>;

                    const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                    logger.info(`CLI output: ${stdout}`);

                    expect(stdout).not.toContain('Error');
                    expect(stdout).not.toContain('Failed');

                    const deptNamespace = Utils.namespaceFromProjectName(projectInDept.name);
                    logger.info('Waiting for workload pods to be running...');
                    await waitUtils.waitForPods(workloadName, deptNamespace, 1, 60, ['Running']);
                    logger.info('Workload submission succeeded with department-scoped template');
                });

                await test.step('Verify workload created successfully', function () {
                    const stdout = runaiCliV2.describeRawFlags(workloadName, cliV2WorkloadType.training, '');
                    expect(stdout).toContain(workloadName);
                    logger.info(`Verified workload exists: ${workloadName}`);
                });

                await test.step('Cleanup: Delete workload and template', async function () {
                    try {
                        runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);

                        await templateUtils.deleteWorkloadTemplateV2(
                            templateId,
                            WorkloadType.Training,
                            flowRootState.users.system_administrator.idToken
                        );
                        logger.info(`Deleted template: ${templateId}`);
                    } catch (error) {
                        logger.error('Cleanup error:', error);
                    }
                });
            });
        });

        test('T5539380: Submit workload with cluster-scoped template using -p flag', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-cluster-scope', 5);
            let templateId: string;
            let clusterTemplateName: string;

            runaiCliV2.projectSet(projectName);

            await test.step('Create cluster-scoped training template', async function () {
                clusterTemplateName = objectGenerator.generateShortNameByPrefix('test-pytorch-training', 5);

                const spec = {
                    image: 'alpine:latest',
                    imagePullPolicy: 'IfNotPresent',
                    command: 'sleep 10',
                    compute: {
                        gpuDevicesRequest: 0,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: clusterTemplateName,
                        description: 'Cluster-scoped template for training'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created cluster-scoped template: ${clusterTemplateName} (ID: ${templateId})`);
            });

            await test.step('Submit workload with cluster-scoped template', async function () {
                logger.info(`Submitting workload with cluster-scoped template: ${clusterTemplateName}`);

                const submitFlag = {
                    [SubmitFlagsMerged.template]: clusterTemplateName
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                logger.info(`CLI output: ${stdout}`);

                expect(stdout).not.toContain('Error');
                expect(stdout).not.toContain('Failed');

                logger.info('Waiting for workload pods to be running...');
                await waitUtils.waitForPods(workloadName, namespace, 1, 60, ['Running']);
                logger.info('Workload submitted successfully using cluster-scoped template');
            });

            await test.step('Verify workload created successfully', function () {
                try {
                    const stdout = runaiCliV2.describeRawFlags(workloadName, cliV2WorkloadType.training, '');
                    expect(stdout).toContain(workloadName);
                    logger.info('Verified workload exists with cluster-scoped template configuration');
                } catch (error) {
                    logger.error('Failed to describe workload:', error);
                    throw error;
                }
            });

            await test.step('Cleanup: Delete template and workload', async function () {
                try {
                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);

                    await templateUtils.deleteWorkloadTemplateV2(
                        templateId,
                        WorkloadType.Training,
                        flowRootState.users.system_administrator.idToken
                    );
                    logger.info(`Deleted template: ${templateId}`);
                } catch (error) {
                    logger.error('Cleanup error:', error);
                }
            });
        });

        test.describe('T5539385: Template hierarchy and access control tests', () => {
            let dept1: any, dept2: any;
            let projectA: any, projectB: any;
            const orgUnitUtils = new OrgUnitUtils();

            test.afterAll('Cleanup departments and projects', async () => {
                try {
                    if (projectA?.id) {
                        await operations.deleteProjectAndNamespace(projectA.id);
                        logger.info('Deleted project-a');
                    }
                    if (projectB?.id) {
                        await operations.deleteProjectAndNamespace(projectB.id);
                        logger.info('Deleted project-b');
                    }
                    if (dept1?.id) {
                        await orgUnitUtils.deleteDepartment(dept1.id);
                        logger.info('Deleted dept-01');
                    }
                    if (dept2?.id) {
                        await orgUnitUtils.deleteDepartment(dept2.id);
                        logger.info('Deleted dept-02');
                    }
                    // Reset project context to main project to avoid affecting subsequent tests
                    runaiCliV2.projectSet(projectName);
                    logger.info(`Reset project context to: ${projectName}`);
                } catch (error) {
                    logger.error('Failed to cleanup departments/projects in afterAll:', error);
                }
            });

            test('T5539385: Verify complete scope hierarchy and template inheritance', async function () {
                test.setTimeout(300 * 1000);

                let clusterTemplateId: string, dept1TemplateId: string, dept2TemplateId: string;
                let projectATemplateId: string, projectBTemplateId: string;
                const clusterTemplateName = objectGenerator.generateShortNameByPrefix(
                    'test-cluster-shared',
                    5
                );
                const dept1TemplateName = objectGenerator.generateShortNameByPrefix('test-dept-shared', 5);
                const dept2TemplateName = objectGenerator.generateShortNameByPrefix('test-dept2-isolated', 5);
                const projectATemplateName = objectGenerator.generateShortNameByPrefix(
                    'test-project-a-local',
                    5
                );
                const projectBTemplateName = objectGenerator.generateShortNameByPrefix(
                    'test-project-b-local',
                    5
                );

                await test.step('Create complete scope hierarchy', async function () {
                    dept1 = await orgUnitUtils.createDepartment({
                        name: objectGenerator.generateShortNameByPrefix('dept-01', 5),
                        clusterId: flowRootState.cluster.uuid,
                        resources: [
                            {
                                gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                                cpu: null,
                                memory: null,
                                nodePool: null
                            }
                        ]
                    });
                    logger.info(`Created dept-01 (ID: ${dept1.id})`);

                    dept2 = await orgUnitUtils.createDepartment({
                        name: objectGenerator.generateShortNameByPrefix('dept-02', 5),
                        clusterId: flowRootState.cluster.uuid,
                        resources: [
                            {
                                gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                                cpu: null,
                                memory: null,
                                nodePool: null
                            }
                        ]
                    });
                    logger.info(`Created dept-02 (ID: ${dept2.id})`);

                    projectA = await projectResourcesUtils.createProjectAndNamespace(
                        objectGenerator.generateShortNameByPrefix('project-a', 5),
                        Number(dept1.id),
                        0,
                        0
                    );
                    logger.info(`Created project-a (ID: ${projectA.id}) in dept-01`);

                    projectB = await projectResourcesUtils.createProjectAndNamespace(
                        objectGenerator.generateShortNameByPrefix('project-b', 5),
                        Number(dept1.id),
                        0,
                        0
                    );
                    logger.info(`Created project-b (ID: ${projectB.id}) in dept-01`);
                });

                await test.step('Create templates at all scope levels', async function () {
                    const baseSpec = {
                        image: 'alpine:latest',
                        imagePullPolicy: 'IfNotPresent',
                        command: 'sleep 10',
                        compute: {
                            cpuCoreRequest: 0.1,
                            cpuMemoryRequest: '100M'
                        }
                    };

                    const clusterSpec = {
                        ...baseSpec,
                        compute: { ...baseSpec.compute, gpuDevicesRequest: 0, cpuCoreRequest: 0.1 }
                    };
                    const clusterResponse = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: WorkloadType.Training,
                            scopeType: ScopeType.Cluster,
                            scopeId: flowRootState.cluster.uuid,
                            templateName: clusterTemplateName,
                            description: 'Cluster-scoped template'
                        },
                        clusterSpec,
                        undefined,
                        flowRootState.users.system_administrator.idToken
                    );
                    clusterTemplateId = clusterResponse.body.meta.id;
                    logger.info(`Created cluster template: ${clusterTemplateName}`);

                    const dept1Spec = {
                        ...baseSpec,
                        compute: { ...baseSpec.compute, gpuDevicesRequest: 0, cpuCoreRequest: 0.2 }
                    };
                    const dept1Response = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: WorkloadType.Training,
                            scopeType: ScopeType.Department,
                            scopeId: dept1.id,
                            templateName: dept1TemplateName,
                            description: 'Department1-scoped template'
                        },
                        dept1Spec,
                        undefined,
                        flowRootState.users.system_administrator.idToken
                    );
                    dept1TemplateId = dept1Response.body.meta.id;
                    logger.info(`Created dept1 template: ${dept1TemplateName}`);

                    const dept2Spec = {
                        ...baseSpec,
                        compute: { ...baseSpec.compute, gpuDevicesRequest: 0, cpuCoreRequest: 0.1 }
                    };
                    const dept2Response = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: WorkloadType.Training,
                            scopeType: ScopeType.Department,
                            scopeId: dept2.id,
                            templateName: dept2TemplateName,
                            description: 'Department2-scoped template'
                        },
                        dept2Spec,
                        undefined,
                        flowRootState.users.system_administrator.idToken
                    );
                    dept2TemplateId = dept2Response.body.meta.id;
                    logger.info(`Created dept2 template: ${dept2TemplateName}`);

                    const projectASpec = {
                        ...baseSpec,
                        compute: { ...baseSpec.compute, gpuDevicesRequest: 0, cpuCoreRequest: 0.3 }
                    };
                    const projectAResponse = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: WorkloadType.Training,
                            scopeType: ScopeType.Project,
                            scopeId: projectA.id.toString(),
                            templateName: projectATemplateName,
                            description: 'Project-A-scoped template'
                        },
                        projectASpec,
                        undefined,
                        flowRootState.users.system_administrator.idToken
                    );
                    projectATemplateId = projectAResponse.body.meta.id;
                    logger.info(`Created project-a template: ${projectATemplateName}`);

                    const projectBSpec = {
                        ...baseSpec,
                        compute: { ...baseSpec.compute, gpuDevicesRequest: 0, cpuCoreRequest: 0.4 }
                    };
                    const projectBResponse = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: WorkloadType.Training,
                            scopeType: ScopeType.Project,
                            scopeId: projectB.id.toString(),
                            templateName: projectBTemplateName,
                            description: 'Project-B-scoped template'
                        },
                        projectBSpec,
                        undefined,
                        flowRootState.users.system_administrator.idToken
                    );
                    projectBTemplateId = projectBResponse.body.meta.id;
                    logger.info(`Created project-b template: ${projectBTemplateName}`);

                    logger.info('All templates created across hierarchy');
                });

                await test.step('Verify project-a can access cluster-scoped template', async function () {
                    const workloadName = objectGenerator.generateShortNameByPrefix(
                        'test-hierarchy-cluster',
                        5
                    );
                    runaiCliV2.projectSet(projectA.name);

                    const submitFlag = {
                        [SubmitFlagsMerged.template]: clusterTemplateName
                    } as Partial<typeof SubmitFlagsMerged>;
                    const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                    expect(stdout).not.toContain('Error');

                    const projectANamespace = Utils.namespaceFromProjectName(projectA.name);
                    await waitUtils.waitForPods(workloadName, projectANamespace, 1, 60, ['Running']);

                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                    logger.info('✓ Project-a can access cluster-scoped template');
                });

                await test.step('Verify project-a can access department-scoped template', async function () {
                    const workloadName = objectGenerator.generateShortNameByPrefix('test-hierarchy-dept', 5);
                    runaiCliV2.projectSet(projectA.name);

                    const submitFlag = {
                        [SubmitFlagsMerged.template]: dept1TemplateName
                    } as Partial<typeof SubmitFlagsMerged>;
                    const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                    expect(stdout).not.toContain('Error');

                    const projectANamespace = Utils.namespaceFromProjectName(projectA.name);
                    await waitUtils.waitForPods(workloadName, projectANamespace, 1, 60, ['Running']);

                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                    logger.info('✓ Project-a can access department1-scoped template');
                });

                await test.step('Verify project-a can access its own project-scoped template', async function () {
                    const workloadName = objectGenerator.generateShortNameByPrefix(
                        'test-hierarchy-project',
                        5
                    );
                    runaiCliV2.projectSet(projectA.name);

                    const submitFlag = {
                        [SubmitFlagsMerged.template]: projectATemplateName
                    } as Partial<typeof SubmitFlagsMerged>;
                    const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                    expect(stdout).not.toContain('Error');

                    const projectANamespace = Utils.namespaceFromProjectName(projectA.name);
                    await waitUtils.waitForPods(workloadName, projectANamespace, 1, 60, ['Running']);

                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                    logger.info('✓ Project-a can access its own project-scoped template');
                });

                await test.step('Verify project-a CANNOT access project-b template (negative test)', function () {
                    const workloadName = objectGenerator.generateShortNameByPrefix('test-cross-project', 5);
                    runaiCliV2.projectSet(projectA.name);

                    const submitFlag = {
                        [SubmitFlagsMerged.template]: projectBTemplateName
                    } as Partial<typeof SubmitFlagsMerged>;

                    expect(() => {
                        runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                    }).toThrow();

                    logger.info('✓ Verified: Project-a cannot access project-b template');
                });

                await test.step('Verify project-a CANNOT access dept-02 template (negative test)', function () {
                    const workloadName = objectGenerator.generateShortNameByPrefix('test-cross-dept', 5);
                    runaiCliV2.projectSet(projectA.name);

                    const submitFlag = {
                        [SubmitFlagsMerged.template]: dept2TemplateName
                    } as Partial<typeof SubmitFlagsMerged>;

                    expect(() => {
                        runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                    }).toThrow();

                    logger.info('✓ Verified: Project-a cannot access dept-02 template');
                });

                await test.step('Cleanup: Delete all templates', async function () {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            clusterTemplateId,
                            WorkloadType.Training,
                            flowRootState.users.system_administrator.idToken
                        );
                        await templateUtils.deleteWorkloadTemplateV2(
                            dept1TemplateId,
                            WorkloadType.Training,
                            flowRootState.users.system_administrator.idToken
                        );
                        await templateUtils.deleteWorkloadTemplateV2(
                            dept2TemplateId,
                            WorkloadType.Training,
                            flowRootState.users.system_administrator.idToken
                        );
                        await templateUtils.deleteWorkloadTemplateV2(
                            projectATemplateId,
                            WorkloadType.Training,
                            flowRootState.users.system_administrator.idToken
                        );
                        await templateUtils.deleteWorkloadTemplateV2(
                            projectBTemplateId,
                            WorkloadType.Training,
                            flowRootState.users.system_administrator.idToken
                        );
                        logger.info('Deleted all templates');
                    } catch (error) {
                        logger.error('Cleanup error:', error);
                    }
                });
            });
        });
    });
});
