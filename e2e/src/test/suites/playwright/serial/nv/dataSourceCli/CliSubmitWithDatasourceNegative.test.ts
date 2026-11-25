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
    dataSourcesUtils,
    orgUnitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI Submit with Datasource - Negative Tests, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_INVALID_FORMAT_WITH_HINT =
            /invalid datasource format.*required format is type=TYPE,name=NAME/i;
        const ERROR_INVALID_FORMAT = /invalid datasource format/i;
        const ERROR_DATASOURCE_NOT_FOUND = /failed to submit\. datasource .* of type .* not found/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-ds-neg', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`Deleted project: ${projectName}`);
                }
            });
        });

        test('T5585479: Submit with datasource missing type', function () {
            test.step('Attempt to submit with datasource missing type', function () {
                logger.info('Attempting to submit with datasource missing type');
                const jobName = objectGenerator.generateShortNameByPrefix('job-no-type', 5);

                expect(() => {
                    runaiCliV2['bashCli'].exec(
                        `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                            '--datasource name=test-ds ' +
                            '--image pytorch/pytorch:latest ' +
                            '-g 1'
                    );
                }).toThrow(ERROR_INVALID_FORMAT_WITH_HINT);

                logger.info('Verified: Command failed with appropriate error for missing type');
            });
        });

        test('T5585480: Submit with datasource missing name', function () {
            test.step('Attempt to submit with datasource missing name', function () {
                logger.info('Attempting to submit with datasource missing name');
                const jobName = objectGenerator.generateShortNameByPrefix('job-no-name', 5);

                expect(() => {
                    runaiCliV2['bashCli'].exec(
                        `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                            '--datasource type=pvc ' +
                            '--image pytorch/pytorch:latest ' +
                            '-g 1'
                    );
                }).toThrow(ERROR_INVALID_FORMAT_WITH_HINT);

                logger.info('Verified: Command failed with appropriate error for missing name');
            });
        });

        test('T5585481: Submit with datasource empty type value', async function () {
            const dsName = objectGenerator.generateShortNameByPrefix('ds-test', 5);
            let dataSourceId: string;

            await test.step('Create PVC datasource via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC datasource: ${dsName}`);
            });

            test.step('Attempt to submit with empty type value', function () {
                logger.info('Attempting to submit with empty type value');
                const jobName = objectGenerator.generateShortNameByPrefix('job-empty-type', 5);

                expect(() => {
                    runaiCliV2['bashCli'].exec(
                        `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                            `--datasource type=,name=${dsName} ` +
                            '--image pytorch/pytorch:latest ' +
                            '-g 1'
                    );
                }).toThrow(ERROR_INVALID_FORMAT);

                logger.info('Verified: Command failed with appropriate error for empty type');
            });

            await test.step('Cleanup datasource', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleaned up datasource');
            });
        });

        test('T5585482: Submit with datasource empty name value', function () {
            test.step('Attempt to submit with empty name value', function () {
                logger.info('Attempting to submit with empty name value');
                const jobName = objectGenerator.generateShortNameByPrefix('job-empty-name', 5);

                expect(() => {
                    runaiCliV2['bashCli'].exec(
                        `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                            '--datasource type=pvc,name= ' +
                            '--image pytorch/pytorch:latest ' +
                            '-g 1'
                    );
                }).toThrow(ERROR_INVALID_FORMAT);

                logger.info('Verified: Command failed with appropriate error for empty name');
            });
        });

        test('T5585483: Submit with datasource invalid format (plain name)', async function () {
            const dsName = objectGenerator.generateShortNameByPrefix('ds-test', 5);
            let dataSourceId: string;

            await test.step('Create PVC datasource via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC datasource: ${dsName}`);
            });

            test.step('Attempt to submit with plain name (invalid format)', function () {
                logger.info('Attempting to submit with plain name (invalid format)');
                const jobName = objectGenerator.generateShortNameByPrefix('job-plain-name', 5);

                expect(() => {
                    runaiCliV2['bashCli'].exec(
                        `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                            `--datasource ${dsName} ` +
                            '--image pytorch/pytorch:latest ' +
                            '-g 1'
                    );
                }).toThrow(ERROR_INVALID_FORMAT_WITH_HINT);

                logger.info('Verified: Command failed with appropriate error for plain name');
            });

            await test.step('Cleanup datasource', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleaned up datasource');
            });
        });

        test('T5585484: Submit with datasource empty string', function () {
            test.step('Attempt to submit with empty string', function () {
                logger.info('Attempting to submit with empty string');
                const jobName = objectGenerator.generateShortNameByPrefix('job-empty-str', 5);

                expect(() => {
                    runaiCliV2['bashCli'].exec(
                        `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                            '--datasource "" ' +
                            '--image pytorch/pytorch:latest ' +
                            '-g 1'
                    );
                }).toThrow(ERROR_INVALID_FORMAT);

                logger.info('Verified: Command failed with appropriate error for empty string');
            });
        });

        test('T5585485: Submit with non-existent datasource name', function () {
            test.step('Attempt to submit with non-existent datasource', function () {
                logger.info('Attempting to submit with non-existent datasource');
                const jobName = objectGenerator.generateShortNameByPrefix('job-no-ds', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                        '--datasource type=pvc,name=non-existent-datasource-xyz ' +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1',
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_DATASOURCE_NOT_FOUND);
                logger.info('Verified: Command failed with appropriate error for non-existent datasource');
            });
        });

        test('T5585486: Submit with invalid datasource type', async function () {
            const dsName = objectGenerator.generateShortNameByPrefix('ds-test', 5);
            let dataSourceId: string;

            await test.step('Create PVC datasource via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC datasource: ${dsName}`);
            });

            test.step('Attempt to submit with wrong/invalid type', function () {
                logger.info('Attempting to submit with invalid type');
                const jobName = objectGenerator.generateShortNameByPrefix('job-wrong-type', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                        `--datasource type=invalid-type-xyz,name=${dsName} ` +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1',
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_DATASOURCE_NOT_FOUND);
                logger.info('Verified: Command failed with appropriate error for invalid type');
            });

            await test.step('Cleanup datasource', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleaned up datasource');
            });
        });

        test('T5585487: Submit with datasource from different project (no access)', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-a', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-b', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const namespaceB = Utils.namespaceFromProjectName(projectBName);
            const dsName = objectGenerator.generateShortNameByPrefix('ds-isolated', 5);
            let projectA: Project;
            let projectB: Project;
            let dataSourceId: string;

            await test.step('Create project-a and project-b', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project-a: ${projectAName}`);

                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project-b: ${projectBName}`);
            });

            await test.step('Create datasource in project-b', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, projectB);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC datasource in project-b: ${dsName}`);
            });

            test.step('Attempt to submit from project-a using project-b datasource', function () {
                logger.info('Attempting to submit from project-a using project-b datasource');
                const jobName = objectGenerator.generateShortNameByPrefix('job-cross-proj', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                        `--datasource type=pvc,name=${dsName} ` +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1 ' +
                        `-p ${projectAName}`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_DATASOURCE_NOT_FOUND);
                logger.info('Verified: project-a cannot access project-b datasources');
            });

            await test.step('Cleanup datasource and projects', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                logger.info('Cleaned up all resources');
            });
        });

        test('T5585488: Submit with datasource from different department (no access)', async function () {
            const deptAName = objectGenerator.generateShortNameByPrefix('dept-a', 5);
            const deptBName = objectGenerator.generateShortNameByPrefix('dept-b', 5);
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-in-a', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const dsName = objectGenerator.generateShortNameByPrefix('ds-dept-b', 5);
            let departmentAId: number;
            let departmentBId: number;
            let projectA: Project;
            let projectB: Project;
            let dataSourceId: string;

            await test.step('Create dept-a and dept-b', async function () {
                const deptA = await orgUnitUtils.createDepartment({
                    name: deptAName,
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
                departmentAId = Number(deptA.id);
                logger.info(`Created dept-a: ${deptAName}`);

                const deptB = await orgUnitUtils.createDepartment({
                    name: deptBName,
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
                departmentBId = Number(deptB.id);
                logger.info(`Created dept-b: ${deptBName}`);
            });

            await test.step('Create project-in-dept-a', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2,
                    departmentId: departmentAId
                });
                logger.info(`Created project in dept-a: ${projectAName}`);
            });

            await test.step('Create temporary project in dept-b for datasource creation', async function () {
                const tempProjectBName = objectGenerator.generateShortNameByPrefix('proj-temp-b', 5);
                const tempNamespaceB = Utils.namespaceFromProjectName(tempProjectBName);
                projectB = await projectResourcesUtils.createProject(tempNamespaceB, {
                    name: tempProjectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentBId
                });
                logger.info(`Created temporary project in dept-b: ${tempProjectBName}`);
            });

            await test.step('Create department-scoped datasource in dept-b', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Department, projectB);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created department-scoped datasource in dept-b: ${dsName}`);
            });

            test.step('Attempt to submit from project-in-dept-a using dept-b datasource', function () {
                logger.info('Attempting to submit from dept-a using dept-b datasource');
                const jobName = objectGenerator.generateShortNameByPrefix('job-cross-dept', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${jobName} ` +
                        `--datasource type=pvc,name=${dsName} ` +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1 ' +
                        `-p ${projectAName}`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_DATASOURCE_NOT_FOUND);
                logger.info('Verified: project in dept-a cannot access dept-b datasources');
            });

            await test.step('Cleanup all resources', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                await orgUnitUtils.deleteDepartment(departmentAId.toString());
                await orgUnitUtils.deleteDepartment(departmentBId.toString());
                logger.info('Cleaned up all resources');
            });
        });
    });
});
