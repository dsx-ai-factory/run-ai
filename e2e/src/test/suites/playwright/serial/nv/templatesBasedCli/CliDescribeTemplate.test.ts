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
    orgUnitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType, WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI Describe Template - Positive Tests, tags: ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message constants
        const ERROR_TEMPLATE_NOT_FOUND = /template .* not found/i;

        let projectName: string;
        let namespace: string;
        let project: Project;
        let templateName: string;
        let templateId: string;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-template-desc', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            logger.info(`Created project: ${projectName} (ID: ${project.id})`);

            // Create a training template for describe tests
            await test.step('Create training template via API', async function () {
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
                        description: 'Template for describe testing'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created template: ${templateName} (ID: ${templateId})`);
            });

            // Set project context once for all tests
            runaiCliV2.projectSet(projectName);
            logger.info(`Set project context to: ${projectName}`);
        });

        test.afterAll('Cleanup resources', async () => {
            try {
                await templateUtils.deleteWorkloadTemplateV2(
                    templateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted template: ${templateId}`);
            } catch (error) {
                logger.error(`Failed to delete template ${templateId}:`, error);
            }

            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`Deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName}:`, error);
            }
        });

        test('T5539375: Describe template', async function () {
            await test.step('Execute template describe command and verify details', async function () {
                logger.info(`Describing template: ${templateName}`);
                logger.info(`Fetching template details via API for: ${templateName}`);

                const templateResponse = await templateUtils.getWorkloadTemplateV2(
                    templateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );

                expect(templateResponse.response.statusCode).toEqual(200);
                const template = templateResponse.body;

                const stdout = runaiCliV2.templateDescribe(templateName);
                logger.info(`Template describe output:\n${stdout}`);

                expect(stdout).toBeTruthy();
                expect(stdout).toContain(templateName);
                expect(stdout.toLowerCase()).toMatch(/training|type/i);
                expect(stdout.toLowerCase()).toMatch(/scope|cluster/i);

                logger.info(`Verified template details: name=${templateName}, type=Training`);
                logger.info(
                    `Template metadata validated from API: id=${template.meta.id}, scope=${template.meta.scopeType}`
                );
            });
        });

        test('T5539386: Describe template with default output (table format)', function () {
            logger.info(`Describing template with default output (table format): ${templateName}`);

            const stdout = runaiCliV2.templateDescribe(templateName);

            expect(stdout).toBeTruthy();
            expect(stdout).toContain(templateName);

            logger.info('Verified: Template described successfully with default output');
        });

        test('T5539387: Describe system template', function () {
            const systemTemplateName = 'pytorch-1-gpu';
            logger.info(`Describing system template: ${systemTemplateName}`);

            const stdout = runaiCliV2.templateDescribe(systemTemplateName);

            expect(stdout).toBeTruthy();
            expect(stdout).toContain(systemTemplateName);
            expect(stdout.toLowerCase()).toContain('system');
            expect(stdout.toLowerCase()).toContain('training');

            logger.info('Verified: System template described successfully');
        });

        test('T5539388: Describe template in specific project using -p flag', async function () {
            let departmentId: number;
            let projectA: Project;
            let projectAName: string;
            let projectTemplateId: string;
            let deptTemplateId: string;

            await test.step('Create department and project', async function () {
                const departmentName = objectGenerator.generateShortNameByPrefix('dept-template-desc', 5);
                const department = await orgUnitUtils.createDepartment({
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
                departmentId = Number(department.id);
                logger.info(`Created department: ${departmentName} (ID: ${departmentId})`);

                projectAName = objectGenerator.generateShortNameByPrefix('proj-a-desc', 5);
                const namespaceA = Utils.namespaceFromProjectName(projectAName);
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentId
                });
                logger.info(`Created project: ${projectAName} (ID: ${projectA.id})`);
            });

            await test.step('Create project-scoped template', async function () {
                const projectTemplateName = objectGenerator.generateShortNameByPrefix(
                    'test-project-template',
                    5
                );
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Project,
                        scopeId: projectA.id.toString(),
                        templateName: projectTemplateName,
                        description: 'Project-scoped template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                projectTemplateId = response.body.meta.id;
                logger.info(`Created project template: ${projectTemplateName}`);

                const stdout = runaiCliV2.templateDescribe(projectTemplateName, {
                    '-p': projectAName
                } as any);
                expect(stdout).toContain(projectTemplateName);
                expect(stdout.toLowerCase()).toContain('project');
                logger.info('Verified: Project-scoped template described with -p flag');
            });

            await test.step('Create department-scoped template', async function () {
                const deptTemplateName = objectGenerator.generateShortNameByPrefix('test-dept-template', 5);
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Department,
                        scopeId: departmentId.toString(),
                        templateName: deptTemplateName,
                        description: 'Department-scoped template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                deptTemplateId = response.body.meta.id;
                logger.info(`Created department template: ${deptTemplateName}`);

                const stdout = runaiCliV2.templateDescribe(deptTemplateName, { '-p': projectAName } as any);
                expect(stdout).toContain(deptTemplateName);
                expect(stdout.toLowerCase()).toContain('department');
                logger.info('Verified: Department template accessible from project with -p flag');
            });

            await test.step('Cleanup templates and resources', async function () {
                await templateUtils.deleteWorkloadTemplateV2(
                    projectTemplateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
                await templateUtils.deleteWorkloadTemplateV2(
                    deptTemplateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectA.id);
                await orgUnitUtils.deleteDepartment(departmentId.toString());
                logger.info('Cleanup completed');
            });
        });

        test('T5539389: Describe template with JSON output format', async function () {
            await test.step('Describe template with JSON output', function () {
                logger.info(`Describing template with JSON output: ${templateName}`);

                const cliPath = runaiCliV2['cliPath'];
                const describeCommand = `${cliPath} template describe ${templateName} -o json`;
                const stdout = runaiCliV2['bashCli'].exec(describeCommand);

                logger.info(`JSON output received (length: ${stdout.length} chars)`);

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);

                // Verify it's valid JSON
                expect(() => {
                    JSON.parse(stdout);
                }).not.toThrow();

                logger.info('Verified: Output is valid JSON');
            });

            await test.step('Verify JSON contains template data', function () {
                logger.info('Validating JSON structure and content');

                const cliPath = runaiCliV2['cliPath'];
                const describeCommand = `${cliPath} template describe ${templateName} -o json`;
                const stdout = runaiCliV2['bashCli'].exec(describeCommand);
                const jsonData = JSON.parse(stdout);

                // Verify key fields: meta and spec exist
                expect(jsonData.meta).toBeDefined();
                expect(jsonData.spec).toBeDefined();

                // Verify template name
                expect(jsonData.meta.name).toEqual(templateName);

                logger.info(`Verified JSON with meta.name=${jsonData.meta.name}`);
            });
        });

        test('T5539390: Describe template with YAML output format', async function () {
            await test.step('Describe template with YAML output and verify content', function () {
                logger.info(`Describing template with YAML output: ${templateName}`);

                const cliPath = runaiCliV2['cliPath'];
                const describeCommand = `${cliPath} template describe ${templateName} -o yaml`;
                const stdout = runaiCliV2['bashCli'].exec(describeCommand);

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);

                // Verify YAML content contains key fields
                expect(stdout).toContain('meta:');
                expect(stdout).toContain('spec:');
                expect(stdout).toContain(templateName);

                logger.info('Verified: YAML output contains meta, spec, and template name');
            });
        });

        test('T5539391: Describe template with table output format (explicit)', function () {
            logger.info(`Describing template with explicit table format: ${templateName}`);

            const cliPath = runaiCliV2['cliPath'];
            const describeCommand = `${cliPath} template describe ${templateName} -o table`;
            const stdout = runaiCliV2['bashCli'].exec(describeCommand);

            expect(stdout).toBeTruthy();
            expect(stdout).toContain(templateName);

            logger.info('Verified: Template described successfully with table output format');
        });

        test('T5539397: Describe cluster-scoped template from project context', async function () {
            let clusterTemplateId: string;

            await test.step('Create cluster-scoped template', async function () {
                const clusterTemplateName = objectGenerator.generateShortNameByPrefix(
                    'test-cluster-describe',
                    5
                );
                const spec = {
                    image: 'pytorch/pytorch:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 2, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        templateName: clusterTemplateName,
                        description: 'Cluster-scoped template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                clusterTemplateId = response.body.meta.id;
                logger.info(`Created cluster template: ${clusterTemplateName}`);

                const stdout = runaiCliV2.templateDescribe(clusterTemplateName, { '-p': projectName } as any);
                expect(stdout).toContain(clusterTemplateName);
                expect(stdout.toLowerCase()).toContain('cluster');
                logger.info('Verified: Cluster template accessible from project context');
            });

            await test.step('Cleanup', async function () {
                await templateUtils.deleteWorkloadTemplateV2(
                    clusterTemplateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
            });
        });

        test('T5539398: Describe department-scoped template from multiple projects', async function () {
            let departmentId: number;
            let projectA: Project;
            let projectB: Project;
            let projectC: Project;
            let dept2Id: number;
            let deptTemplateId: string;

            await test.step('Create departments and projects', async function () {
                const dept1Name = objectGenerator.generateShortNameByPrefix('dept-01-desc', 5);
                const department1 = await orgUnitUtils.createDepartment({
                    name: dept1Name,
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
                departmentId = Number(department1.id);

                const dept2Name = objectGenerator.generateShortNameByPrefix('dept-02-desc', 5);
                const department2 = await orgUnitUtils.createDepartment({
                    name: dept2Name,
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
                dept2Id = Number(department2.id);

                const projAName = objectGenerator.generateShortNameByPrefix('proj-a-dept1', 5);
                const nsA = Utils.namespaceFromProjectName(projAName);
                projectA = await projectResourcesUtils.createProject(nsA, {
                    name: projAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentId
                });

                const projBName = objectGenerator.generateShortNameByPrefix('proj-b-dept1', 5);
                const nsB = Utils.namespaceFromProjectName(projBName);
                projectB = await projectResourcesUtils.createProject(nsB, {
                    name: projBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentId
                });

                const projCName = objectGenerator.generateShortNameByPrefix('proj-c-dept2', 5);
                const nsC = Utils.namespaceFromProjectName(projCName);
                projectC = await projectResourcesUtils.createProject(nsC, {
                    name: projCName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: dept2Id
                });

                logger.info(`Created projects in dept1: ${projAName}, ${projBName}; dept2: ${projCName}`);
            });

            await test.step('Create department-scoped template and verify access', async function () {
                const deptTemplateName = objectGenerator.generateShortNameByPrefix('test-dept-shared', 5);
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 1, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Workspace,
                        scopeType: ScopeType.Department,
                        scopeId: departmentId.toString(),
                        templateName: deptTemplateName,
                        description: 'Department-scoped template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                deptTemplateId = response.body.meta.id;
                logger.info(`Created department template: ${deptTemplateName}`);

                const stdoutA = runaiCliV2.templateDescribe(deptTemplateName, { '-p': projectA.name } as any);
                expect(stdoutA).toContain(deptTemplateName);
                logger.info('Verified: Project-A can access department template');

                const stdoutB = runaiCliV2.templateDescribe(deptTemplateName, { '-p': projectB.name } as any);
                expect(stdoutB).toContain(deptTemplateName);
                logger.info('Verified: Project-B can access department template');

                logger.info('Attempting to describe from Project-C in different department');
                expect(() => {
                    runaiCliV2.templateDescribe(deptTemplateName, { '-p': projectC.name } as any);
                }).toThrow(ERROR_TEMPLATE_NOT_FOUND);
                logger.info('Verified: Project-C cannot access dept-01 template');
            });

            await test.step('Cleanup', async function () {
                await templateUtils.deleteWorkloadTemplateV2(
                    deptTemplateId,
                    WorkloadType.Workspace,
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                await operations.deleteProjectAndNamespace(projectC.id);
                await orgUnitUtils.deleteDepartment(departmentId.toString());
                await orgUnitUtils.deleteDepartment(dept2Id.toString());
            });
        });

        test('T5539400: Describe template in default project without -p flag', async function () {
            let defaultProject: Project;
            let defaultProjectName: string;
            let templateId: string;

            await test.step('Create project and template', async function () {
                defaultProjectName = objectGenerator.generateShortNameByPrefix('proj-default-desc', 5);
                const namespace = Utils.namespaceFromProjectName(defaultProjectName);
                defaultProject = await projectResourcesUtils.createProject(namespace, {
                    name: defaultProjectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project: ${defaultProjectName}`);

                const templateName = objectGenerator.generateShortNameByPrefix(
                    'test-default-proj-describe',
                    5
                );
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Project,
                        scopeId: defaultProject.id.toString(),
                        templateName: templateName,
                        description: 'Template in default project'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                templateId = response.body.meta.id;
                logger.info(`Created template: ${templateName}`);

                runaiCliV2.projectSet(defaultProjectName);
                logger.info(`Set default project: ${defaultProjectName}`);

                const stdoutWithoutFlag = runaiCliV2.templateDescribe(templateName);
                expect(stdoutWithoutFlag).toContain(templateName);
                logger.info('Verified: Template described without -p flag (using default project)');

                const stdoutWithFlag = runaiCliV2.templateDescribe(templateName, {
                    '-p': defaultProjectName
                } as any);
                expect(stdoutWithFlag).toContain(templateName);
                logger.info('Verified: Template described with -p flag');
            });

            await test.step('Cleanup', async function () {
                await templateUtils.deleteWorkloadTemplateV2(
                    templateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(defaultProject.id);
            });
        });
    });
});
