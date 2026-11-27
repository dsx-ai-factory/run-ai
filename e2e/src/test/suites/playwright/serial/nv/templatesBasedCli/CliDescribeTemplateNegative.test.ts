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
    orgUnitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType, WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI Describe Template - Negative Tests, tags: ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message constants
        const ERROR_TEMPLATE_NOT_FOUND = /template .* not found/i;
        const ERROR_TEMPLATE_NAME_EMPTY = /template name cannot be empty/i;
        const ERROR_MISSING_TEMPLATE_ARG = /accepts 1 arg\(s\), received 0/i;
        const ERROR_UNSUPPORTED_OUTPUT_FORMAT = /unsupported output format: xml/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-template-desc-neg', 5);
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

        test('T5539392: Describe non-existent template', async function () {
            const nonExistentTemplate = 'non-existent-template-xyz';

            await test.step('Attempt to describe non-existent template', function () {
                logger.info(`Attempting to describe non-existent template: ${nonExistentTemplate}`);

                expect(() => {
                    runaiCliV2.templateDescribe(nonExistentTemplate);
                }).toThrow(ERROR_TEMPLATE_NOT_FOUND);

                logger.info('Verified: Describe non-existent template fails with appropriate error');
            });
        });

        test('T5539393: Describe template from different project scope without access', async function () {
            let departmentId: number;
            let projectA: Project;
            let projectB: Project;
            let projectAName: string;
            let projectBName: string;
            let templateId: string;

            await test.step('Create department and two projects', async function () {
                const departmentName = objectGenerator.generateShortNameByPrefix('dept-template-neg', 5);
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

                projectAName = objectGenerator.generateShortNameByPrefix('proj-a-neg', 5);
                const namespaceA = Utils.namespaceFromProjectName(projectAName);
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentId
                });

                projectBName = objectGenerator.generateShortNameByPrefix('proj-b-neg', 5);
                const namespaceB = Utils.namespaceFromProjectName(projectBName);
                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentId
                });

                logger.info(`Created projects: ${projectAName}, ${projectBName}`);
            });

            await test.step('Create template in project-b', async function () {
                const templateName = objectGenerator.generateShortNameByPrefix('test-project-isolated', 5);
                const spec = {
                    image: 'alpine:latest',
                    command: 'sleep 10',
                    compute: { gpuDevicesRequest: 0, cpuCoreRequest: 0.1, cpuMemoryRequest: '100M' }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Project,
                        scopeId: projectB.id.toString(),
                        templateName: templateName,
                        description: 'Project-b isolated template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                templateId = response.body.meta.id;
                logger.info(`Created template in project-b: ${templateName}`);

                logger.info('Attempting to describe from project-a (should fail)');
                expect(() => {
                    runaiCliV2.templateDescribe(templateName, { '-p': projectAName } as any);
                }).toThrow(ERROR_TEMPLATE_NOT_FOUND);

                logger.info('Verified: Cannot access project-b template from project-a');
            });

            await test.step('Cleanup', async function () {
                await templateUtils.deleteWorkloadTemplateV2(
                    templateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                await orgUnitUtils.deleteDepartment(departmentId.toString());
            });
        });

        test('T5539394: Describe template with empty template name', async function () {
            await test.step('Attempt to describe template with empty name', function () {
                logger.info('Attempting to describe template with empty name (empty string)');

                const cliPath = runaiCliV2['cliPath'];
                const describeCommand = `${cliPath} template describe ""`;

                expect(() => {
                    runaiCliV2['bashCli'].exec(describeCommand);
                }).toThrow(ERROR_TEMPLATE_NAME_EMPTY);

                logger.info('Verified: Command failed with error - template name cannot be empty');
            });
        });

        test('T5539395: Describe template without template name argument', async function () {
            await test.step('Attempt to describe without template name', function () {
                logger.info('Attempting to run describe command without template name argument');

                const cliPath = runaiCliV2['cliPath'];
                const describeCommand = `${cliPath} template describe`;

                expect(() => {
                    runaiCliV2['bashCli'].exec(describeCommand);
                }).toThrow(ERROR_MISSING_TEMPLATE_ARG);

                logger.info('Verified: Command failed with error - accepts 1 arg(s), received 0');
            });
        });

        test('T5539396: Describe template with invalid output format', async function () {
            await test.step('Attempt to describe with invalid output format', function () {
                const templateName = 'pytorch-1-gpu'; // Use existing system template

                logger.info(
                    `Attempting to describe template with invalid output format (xml): ${templateName}`
                );

                const cliPath = runaiCliV2['cliPath'];
                const describeCommand = `${cliPath} template describe ${templateName} -o xml`;

                expect(() => {
                    runaiCliV2['bashCli'].exec(describeCommand);
                }).toThrow(ERROR_UNSUPPORTED_OUTPUT_FORMAT);

                logger.info(
                    'Verified: Invalid output format (xml) is rejected with error - unsupported output format'
                );
            });
        });

        test('T5539399: Describe deleted template', async function () {
            const templateName = objectGenerator.generateShortNameByPrefix('test-to-delete', 5);
            let templateId: string;

            await test.step('Create template', async function () {
                logger.info(`Creating template: ${templateName}`);

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
                        description: 'Template to be deleted for describe testing'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created template: ${templateName} (ID: ${templateId})`);
            });

            await test.step('Verify template can be described before deletion', function () {
                try {
                    const stdout = runaiCliV2.templateDescribe(templateName);
                    expect(stdout).toContain(templateName);
                    logger.info('✓ Template can be described before deletion');
                } catch (error) {
                    logger.error('Failed to describe template before deletion:', error);
                    throw error;
                }
            });

            await test.step('Delete template via API', async function () {
                logger.info(`Deleting template: ${templateId}`);

                await templateUtils.deleteWorkloadTemplateV2(
                    templateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );

                logger.info('Template deleted successfully');

                // Wait for template to be actually deleted
                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            await templateUtils.getWorkloadTemplateV2(
                                templateId,
                                WorkloadType.Training,
                                flowRootState.users.system_administrator.idToken
                            );
                            return false; // Template still exists
                        } catch {
                            // Template not found - deletion completed
                            return true;
                        }
                    },
                    'Template deletion to complete',
                    true,
                    30,
                    1
                );
                logger.info('Template deletion confirmed via API');
            });

            await test.step('Attempt to describe deleted template', function () {
                logger.info(`Attempting to describe deleted template: ${templateName}`);

                expect(() => {
                    runaiCliV2.templateDescribe(templateName);
                }).toThrow(ERROR_TEMPLATE_NOT_FOUND);

                logger.info('Verified: Describe deleted template fails with appropriate error');
            });
        });
    });
});
