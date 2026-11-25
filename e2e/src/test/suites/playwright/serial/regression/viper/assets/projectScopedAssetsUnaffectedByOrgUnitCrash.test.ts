/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { kubeCli } from '../../../../../../../modules/cli';
import {
    credentialUtils,
    dataSourcesUtils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    settings,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { AllureUtils } from '../../../../../../../modules/helpers/utils/AllureUtils';
import { Scope } from '../../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Generic Secret & Pvc should be unaffected by org-unit service crash ${tags.teams.viper}`, () => {
    shouldRunTests(settings.isSelfHosted)(function () {
        test.setTimeout(240 * 1000); // this test might take longer than the default timeout
        test.describe.configure({ retries: 1 });
        let project: Project;
        let allureUtils: AllureUtils;

        // Use beforeAll instead of beforeEach to create project once
        // This prevents 502 errors when the second test starts after org-unit service was killed
        test.beforeAll(async function () {
            allureUtils = new AllureUtils();
            try {
                project = await projectResourcesUtils.createProjectAndNamespace('project-for-data-sources');
            } catch (error) {
                logger.error(`Failed to create project: ${error.message}`);
                logger.error(`Error details: ${JSON.stringify(error, null, 2)}`);

                // Attach error details to Allure report
                const { attachment } = await import('allure-js-commons');
                await attachment(
                    'Project Creation Error Details',
                    JSON.stringify(
                        {
                            message: error.message,
                            statusCode: error.statusCode,
                            url: error.options?.uri || error.options?.url,
                            method: error.options?.method,
                            response: error.body,
                            timestamp: new Date().toISOString()
                        },
                        null,
                        2
                    ),
                    'application/json'
                );

                // Collect backend logs for 502 debugging
                logger.info('Collecting backend service logs for debugging...');
                await allureUtils.attachSystemComponentLogs(
                    'org-unit-service',
                    'Org-Unit Service Logs',
                    'runai-backend',
                    1000
                );
                await allureUtils.attachSystemComponentLogs(
                    'cluster-service',
                    'Cluster Service Logs',
                    'runai-backend',
                    500
                );
                await allureUtils.attachSystemComponentLogs(
                    'backend-nginx',
                    'API Gateway Logs',
                    'runai-backend',
                    500
                );
                throw error;
            }
        });

        // Use afterAll instead of afterEach to delete project once at the end
        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (e) {
                logger.error(`Failed to delete project ${project.name}, error: ${e}`);
            }
            // Wait for org-unit-service pod to be fully running
            // This ensures no "PodInitializing" state that would cause log collection to fail in subsequent tests
            await kubeCli.waitForOrgUnitServiceFullReadiness(180);
        });

        // Collect logs on test failure
        test.afterEach(async function ({}, testInfo) {
            if (testInfo.status === 'failed' || testInfo.status === 'timedOut') {
                try {
                    logger.info('Test failed - collecting backend service logs...');
                    await allureUtils.attachSystemComponentLogs(
                        'org-unit-service',
                        'Org-Unit Service Logs',
                        'runai-backend',
                        1000
                    );
                } catch (error) {
                    logger.warn(`Failed to collect failure logs: ${error.message}`);
                }
            }
        });

        test('PVC is not affected by org-unit service crash', async ({}) => {
            let pvcCreationTimestamp: Date;
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc');
            await test.step('Create project-scoped PVC', async () => {
                const pvcPayload = dataSourcesUtils.getPvcCreationPayload({
                    name: pvcName,
                    scope: Scope.Project,
                    scopeId: project.id.toString()
                });
                await dataSourcesUtils.createPvcAsset(pvcPayload);
                await waitUtils.waitForFullClaimName(pvcName, project.namespace);
                pvcCreationTimestamp = (await kubeCli.listPVCsByNamespace(project.namespace)).items['0']
                    .metadata.creationTimestamp;
            });

            await test.step('Kill org-unit service', async () => {
                await kubeCli.killOrgUnitServicePod();
            });

            await test.step('Wait for org-unit service to recreate', async () => {
                await kubeCli.waitForOrgUnitPodRecreationAndReadiness();
            });

            await test.step('Assert pvc was not deleted', async () => {
                expect(
                    (await kubeCli.listPVCsByNamespace(project.namespace)).items['0'].metadata
                        .creationTimestamp
                ).toEqual(pvcCreationTimestamp);
            });
        });

        test('Generic secret is not affected by org-unit service crash', async ({}) => {
            let genericSecretCreationTimestamp: Date;
            let secretName: string;

            await test.step('Create project-scoped generic secret', async () => {
                secretName =
                    'genericsecret-' +
                    (await credentialUtils.createGenericSecretWithArbitraryFields(project.id.toString())).meta
                        .name;
                await waitUtils.waitForSecret(secretName, project.namespace);
                genericSecretCreationTimestamp = (await kubeCli.getSecret(secretName, project.namespace))
                    .metadata.creationTimestamp;
            });

            await test.step('Kill org-unit service', async () => {
                await kubeCli.killOrgUnitServicePod();
            });

            await test.step('Wait for org-unit service to recreate', async () => {
                await kubeCli.waitForOrgUnitPodRecreationAndReadiness();
            });

            await test.step('Assert pvc was not deleted', async () => {
                expect(
                    (await kubeCli.getSecret(secretName, project.namespace)).metadata.creationTimestamp
                ).toEqual(genericSecretCreationTimestamp);
            });
        });
    });
});
