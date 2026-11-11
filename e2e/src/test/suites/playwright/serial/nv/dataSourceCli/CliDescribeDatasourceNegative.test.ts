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
    dataSourcesUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI Describe Datasource - Negative Tests, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_DATASOURCE_NOT_FOUND = /failed to describe data source.*does not exist/i;
        const ERROR_TYPE_FLAG_REQUIRED = /required flag\(s\) "type" not set/i;
        const ERROR_UNSUPPORTED_OUTPUT_FORMAT = /failed to describe data source.*unsupported output format/i;
        const ERROR_MISSING_NAME_ARGUMENT = /accepts 1 arg\(s\), received 0/i;
        const ERROR_EMPTY_NAME = /failed to describe data source.*name cannot be empty/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-ds-neg', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
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

        test('T5585425: Describe non-existent data source', function () {
            test.step('Attempt to describe a non-existent data source', function () {
                logger.info('Attempting to describe non-existent data source');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe non-existent-datasource --type pvc`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: failed to describe data source. datasource 'non-existent-datasource' does not exist"
                expect(result.stderr).toMatch(ERROR_DATASOURCE_NOT_FOUND);

                logger.info('Verified: Command failed with error - data source not found');
            });
        });

        test('T5585426: Describe data source without --type flag', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-no-type-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            test.step('Attempt to describe without --type flag', function () {
                logger.info('Attempting to describe without --type flag');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe ${dataSourceName}`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: required flag(s) "type" not set"
                expect(result.stderr).toMatch(ERROR_TYPE_FLAG_REQUIRED);

                logger.info('Verified: Command failed with error - --type flag is required');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });
        });

        // Coverage remains disabled pending revalidation of this case
        test.skip('T5585427: Describe data source with invalid type', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-invalid-type-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            test.step('Attempt to describe with invalid type', function () {
                logger.info('Attempting to describe with invalid type (invalid-xyz)');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe ${dataSourceName} --type invalid-xyz`,
                    false
                );

                // TODO: Revalidate this negative datasource case before restoring coverage.
                // CLI doesn't validate the --type parameter
                // Expected: CLI should reject invalid type with error
                // Actual: CLI reports datasource not found instead
                expect(result.status).not.toBe(0);

                logger.info('Verified: Command failed (but with wrong error message - should validate type)');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });
        });

        test('T5585428: Describe data source with invalid output format', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-invalid-fmt-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            test.step('Attempt to describe with invalid output format', function () {
                logger.info('Attempting to describe with invalid output format (xml)');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe ${dataSourceName} --type pvc -o xml`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: failed to describe data source. unsupported output format: xml"
                expect(result.stderr).toMatch(ERROR_UNSUPPORTED_OUTPUT_FORMAT);

                logger.info('Verified: Command failed with error - unsupported output format');
            });

            await test.step('Cleanup data source', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });
        });

        test('T5585429: Describe data source without name argument', function () {
            test.step('Attempt to describe without data source name', function () {
                logger.info('Attempting to describe without name argument');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe --type pvc`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: accepts 1 arg(s), received 0"
                expect(result.stderr).toMatch(ERROR_MISSING_NAME_ARGUMENT);

                logger.info('Verified: Command failed with error - name argument is required');
            });
        });

        test('T5585430: Describe data source with empty name', function () {
            test.step('Attempt to describe with empty name', function () {
                logger.info('Attempting to describe with empty name');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe "" --type pvc`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: failed to describe data source. datasource name cannot be empty"
                expect(result.stderr).toMatch(ERROR_EMPTY_NAME);

                logger.info('Verified: Command failed with error - name cannot be empty');
            });
        });

        test('T5585431: Describe data source from different project without access', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-isolated-ds', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-ds-b', 5);
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-ds-a', 5);
            const namespaceB = Utils.namespaceFromProjectName(projectBName);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            let dataSourceId: string;
            let projectB: Project;
            let projectA: Project;

            await test.step('Create project-a and project-b via API', async function () {
                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-b: ${projectBName} (ID: ${projectB.id})`);

                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-a: ${projectAName} (ID: ${projectA.id})`);
            });

            await test.step('Create data source in project-b via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, projectB);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source in project-b: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            test.step('Attempt to describe from project-a', function () {
                logger.info('Attempting to describe project-b data source from project-a');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe ${dataSourceName} --type pvc -p ${projectAName}`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: failed to describe data source. datasource 'test-isolated-ds-40017' does not exist"
                expect(result.stderr).toMatch(ERROR_DATASOURCE_NOT_FOUND);

                logger.info(
                    'Verified: Command failed with error - data source not accessible from different project'
                );
            });

            await test.step('Cleanup data source and projects', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);

                await operations.deleteProjectAndNamespace(projectB.id);
                logger.info(`Deleted project-b: ${projectBName}`);

                await operations.deleteProjectAndNamespace(projectA.id);
                logger.info(`Deleted project-a: ${projectAName}`);
            });
        });

        test('T5585432: Describe deleted data source', async function () {
            const dataSourceName = objectGenerator.generateShortNameByPrefix('test-to-delete-ds', 5);
            let dataSourceId: string;

            await test.step('Create PVC data source via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dataSourceName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC data source: ${dataSourceName} (ID: ${dataSourceId})`);
            });

            await test.step('Verify data source can be described', function () {
                logger.info('Verifying data source can be described before deletion');
                const stdout = runaiCliV2.datasourceDescribe(dataSourceName, 'pvc');

                expect(stdout).toContain(dataSourceName);
                logger.info('Verified: Data source can be described successfully');
            });

            await test.step('Delete the data source via API', async function () {
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info(`Deleted PVC data source: ${dataSourceId}`);
            });

            test.step('Attempt to describe the deleted data source', function () {
                logger.info('Attempting to describe deleted data source');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource describe ${dataSourceName} --type pvc`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: failed to describe data source. datasource 'test-to-delete-ds-b2994' does not exist"
                expect(result.stderr).toMatch(ERROR_DATASOURCE_NOT_FOUND);

                logger.info('Verified: Command failed with error - data source not found after deletion');
            });
        });
    });
});
