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
    waitUtils
} from '../../../../../../modules/helpers';
import { OrgUnitUtils } from '../../../../../../modules/helpers/utils/OrgUnitUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType, WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit With Template - Negative Tests, tags: ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-template-neg', 5);
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

        test('T5539373: Submit training workload with non-existent template', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-train', 5);
            const nonExistentTemplate = 'non-existent-template';

            await test.step('Attempt to submit workload with non-existent template', function () {
                logger.info(
                    `Attempting to submit workload: ${workloadName} with non-existent template: ${nonExistentTemplate}`
                );

                const submitFlag = {
                    [SubmitFlagsMerged.template]: nonExistentTemplate,
                    [SubmitFlagsMerged.image]: 'ubuntu'
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow(/does not support.*workloads/i);

                logger.info('Verified: Command failed as expected - template does not exist');
            });
        });

        test('T5539377: Submit with template and missing required parameters', async function () {
            const templateName = objectGenerator.generateShortNameByPrefix('test-minimal-incomplete', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-missing', 5);
            let templateId: string;

            await test.step('Create template without image configuration', async function () {
                logger.info(`Creating incomplete template: ${templateName}`);

                const spec = {
                    command: 'sleep 10',
                    compute: {
                        gpuDevicesRequest: 1,
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
                        description: 'Incomplete template for testing missing parameters'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created incomplete template: ${templateName} (ID: ${templateId})`);
            });

            await test.step('Attempt to submit workload with incomplete template', function () {
                logger.info(`Attempting to submit workload with incomplete template: ${templateName}`);

                const submitFlag = {
                    [SubmitFlagsMerged.template]: templateName
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow();

                logger.info('Verified: Missing required parameters causes submission to fail');
            });

            await test.step('Cleanup: Delete incomplete template', async function () {
                try {
                    await templateUtils.deleteWorkloadTemplateV2(
                        templateId,
                        WorkloadType.Training,
                        flowRootState.users.system_administrator.idToken
                    );
                    logger.info(`Deleted incomplete template: ${templateId}`);
                } catch (error) {
                    logger.error('Failed to delete template:', error);
                }
            });
        });

        // KNOWN BUG: Error message is misleading
        test('T5539382: Submit workload with mismatched template workload type', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-mismatch', 5);
            const templateName = objectGenerator.generateShortNameByPrefix('test-training-template', 5);
            let templateId: string;

            await test.step('Create training template', async function () {
                const spec = {
                    image: 'alpine:latest',
                    imagePullPolicy: 'IfNotPresent',
                    command: 'sleep 10',
                    compute: {
                        gpuDevicesRequest: 1,
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
                        description: 'Training template for testing type mismatch'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created training template: ${templateName} (ID: ${templateId})`);
            });

            await test.step('Attempt to submit workspace workload with training template', function () {
                logger.info(
                    `Attempting to submit workspace workload with training template: ${templateName}`
                );

                const submitFlag = {
                    [SubmitFlagsMerged.template]: templateName
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.workspace);
                }).toThrow(/does not support.*workloads/i);

                logger.info('Verified: Cannot use training template for workspace workload');
            });

            await test.step('Attempt to submit inference workload with training template', function () {
                logger.info(
                    `Attempting to submit inference workload with training template: ${templateName}`
                );

                const submitFlag = {
                    [SubmitFlagsMerged.template]: templateName,
                    [SubmitFlagsMerged.image]: 'alpine:latest',
                    [SubmitFlagsMerged.servingPort]: '8080'
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.inference);
                }).toThrow(/does not support.*workloads/i);

                logger.info('Verified: Cannot use training template for inference workload');
            });

            await test.step('Cleanup: Delete template', async function () {
                try {
                    await templateUtils.deleteWorkloadTemplateV2(
                        templateId,
                        WorkloadType.Training,
                        flowRootState.users.system_administrator.idToken
                    );
                    logger.info(`Deleted template: ${templateId}`);
                } catch (error) {
                    logger.error('Failed to delete template:', error);
                }
            });
        });

        test('T5539383: Submit workload with empty template name', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-empty', 5);

            await test.step('Attempt to submit workload with empty template name', function () {
                logger.info('Attempting to submit workload with empty template name');

                const submitFlag = {
                    [SubmitFlagsMerged.template]: '',
                    [SubmitFlagsMerged.image]: 'ubuntu'
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow();

                logger.info('Verified: Command failed as expected - empty template name was rejected');
            });
        });

        test('T5539384: Submit with deleted template', async function () {
            const templateName = objectGenerator.generateShortNameByPrefix('test-to-be-deleted', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-deleted', 5);
            let templateId: string;

            await test.step('Create a template to be deleted', async function () {
                logger.info(`Creating template: ${templateName}`);

                const spec = {
                    image: 'alpine:latest',
                    imagePullPolicy: 'IfNotPresent',
                    command: 'sleep 10',
                    compute: {
                        gpuDevicesRequest: 1,
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
                        description: 'Template to be deleted for testing'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created template: ${templateName} (ID: ${templateId})`);
            });

            await test.step('Verify template exists via CLI', function () {
                logger.info(`Verifying template exists: ${templateName}`);
                logger.info(`Using project: ${projectName}`);

                try {
                    const stdout = runaiCliV2.templateDescribe(templateName);
                    expect(stdout).toContain(templateName);
                    logger.info('✓ Template exists and can be described');
                } catch (error) {
                    logger.error('Failed to describe template:', error);
                    logger.error(`Project name: ${projectName}`);
                    logger.error(`Template name: ${templateName}`);
                    throw error;
                }
            });

            await test.step('Delete the template via API', async function () {
                logger.info(`Deleting template via API: ${templateId}`);

                await templateUtils.deleteWorkloadTemplateV2(
                    templateId,
                    WorkloadType.Training,
                    flowRootState.users.system_administrator.idToken
                );

                logger.info('Template deleted successfully');
                await waitUtils.sleep(2000);
            });

            await test.step('Attempt to submit workload with deleted template', function () {
                logger.info(`Attempting to submit workload with deleted template: ${templateName}`);

                const submitFlag = {
                    [SubmitFlagsMerged.template]: templateName
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow(/does not support.*workloads/i);

                logger.info('Verified: Cannot submit workload with deleted template');
            });
        });

        test('T5539381: Submit workload with template from different department scope', async function () {
            const templateName = objectGenerator.generateShortNameByPrefix('test-dept2-pytorch', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-cross-dept', 5);
            const dept1Name = objectGenerator.generateShortNameByPrefix('dept1', 5);
            const dept2Name = objectGenerator.generateShortNameByPrefix('dept2', 5);
            const proj1Name = objectGenerator.generateShortNameByPrefix('proj-dept1', 5);
            let templateId: string;
            let department1: any;
            let department2: any;
            let projectInDept1: any;
            const orgUnitUtils = new OrgUnitUtils();

            await test.step('Create two departments and project in department1', async function () {
                department1 = await orgUnitUtils.createDepartment({
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
                logger.info(`Created department1: ${dept1Name} (ID: ${department1.id})`);

                department2 = await orgUnitUtils.createDepartment({
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
                logger.info(`Created department2: ${dept2Name} (ID: ${department2.id})`);

                projectInDept1 = await projectResourcesUtils.createProjectAndNamespace(
                    proj1Name,
                    Number(department1.id),
                    0,
                    0
                );
                logger.info(`Created project: ${projectInDept1.name} in department1`);
            });

            await test.step('Create template in department2 scope', async function () {
                logger.info(`Creating template in department2: ${templateName}`);

                const spec = {
                    image: 'alpine:latest',
                    imagePullPolicy: 'IfNotPresent',
                    command: 'sleep 10',
                    compute: {
                        gpuDevicesRequest: 1,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    }
                };

                const response = await templateUtils.createWorkloadTemplateV2(
                    {
                        workloadType: WorkloadType.Training,
                        scopeType: ScopeType.Department,
                        scopeId: department2.id,
                        templateName: templateName,
                        description: 'Department2-scoped template'
                    },
                    spec,
                    undefined,
                    flowRootState.users.system_administrator.idToken
                );

                expect(response.response.statusCode).toEqual(200);
                templateId = response.body.meta.id;
                logger.info(`Created template in department2: ${templateName} (ID: ${templateId})`);
            });

            await test.step('Attempt to use department2 template from project in department1 (should fail)', function () {
                runaiCliV2.projectSet(projectInDept1.name);

                logger.info('Attempting to use department2 template from department1 project');

                const submitFlag = {
                    [SubmitFlagsMerged.template]: templateName
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow(/does not support.*workloads/i);

                logger.info('Verified: Cannot use template from different department');
            });

            await test.step('Cleanup: Delete template, projects and departments', async function () {
                try {
                    await templateUtils.deleteWorkloadTemplateV2(
                        templateId,
                        WorkloadType.Training,
                        flowRootState.users.system_administrator.idToken
                    );
                    logger.info(`Deleted template: ${templateId}`);

                    await operations.deleteProjectAndNamespace(projectInDept1.id);
                    logger.info(`Deleted project: ${projectInDept1.name}`);

                    await orgUnitUtils.deleteDepartment(department1.id);
                    logger.info(`Deleted department1: ${dept1Name}`);

                    await orgUnitUtils.deleteDepartment(department2.id);
                    logger.info(`Deleted department2: ${dept2Name}`);
                } catch (error) {
                    logger.error('Cleanup error:', error);
                }
            });
        });
    });
});
