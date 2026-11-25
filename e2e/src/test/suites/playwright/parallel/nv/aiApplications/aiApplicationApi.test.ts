/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    objectGenerator,
    projectResourcesUtils,
    tags,
    logger,
    operations,
    waitUtils,
    aiApplicationUtils,
    Utils,
    crds
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

const CUSTOM_URL_CHART =
    'https://github.com/prometheus-community/helm-charts/releases/download/alertmanager-1.33.1/alertmanager-1.33.1.tgz';

test.describe(`AI Application API E2E Tests, tags: ${tags.tool.api} ${tags.resource.aiApplications} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        // Shared resources
        let projectNameList: string[];
        let namespaceList: string[];
        let projectIdList: number[];

        // Per-test resource tracking for cleanup
        let aiApplicationIds: string[];

        test.beforeAll(async () => {
            projectNameList = Array.from({ length: 2 }, () =>
                objectGenerator.generateShortNameByPrefix('api-app', 5)
            );
            namespaceList = projectNameList.map((name) => Utils.namespaceFromProjectName(name));

            const projectsToCreate = [
                { namespace: namespaceList[0], name: projectNameList[0] },
                { namespace: namespaceList[1], name: projectNameList[1] }
            ];
            const projResponse = await Promise.all(
                projectsToCreate.map((config) =>
                    projectResourcesUtils.createProject(config.namespace, {
                        name: config.name,
                        deservedGpus: 2,
                        maxAllowedGpus: -1
                    })
                )
            );
            projectIdList = projResponse.map((proj) => proj.id);
        });

        test.beforeEach(function () {
            aiApplicationIds = [];
        });

        test.afterEach('cleanup AI applications', async () => {
            await Promise.all(
                aiApplicationIds.map(async (appId) => {
                    try {
                        await aiApplicationUtils.deleteAiApplication(appId);
                    } catch (e) {
                        logger.error(`Failed to delete AI application ${appId}`, e);
                    }
                })
            );
        });

        test.afterAll('cleanup projects', async () => {
            await Promise.all(
                projectIdList.map(async (projectId, i) => {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                        logger.info(
                            `deleted project ${projectNameList[i]} and namespace ${namespaceList[i]}`
                        );
                    } catch (error) {
                        logger.error(
                            `failed to delete project ${projectNameList[i]} and namespace ${namespaceList[i]}`,
                            error
                        );
                    }
                })
            );
        });

        test.describe('AI Application CRUD API tests', () => {
            test('T5857350: API: Create AI Application via chartRef', async () => {
                const appName = objectGenerator.generateShortNameByPrefix('api-create', 5);
                const namespace = namespaceList[0];

                await test.step('Create AI application via API', async () => {
                    const resp = await aiApplicationUtils.createAiApplication({
                        name: appName,
                        projectId: projectIdList[0].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                    });

                    expect(resp.response.statusCode).toBe(200);
                    expect(resp.body.id).toBeDefined();
                    expect(resp.body.name).toBe(appName);
                    expect(resp.body.desiredPhase).toBe('deployed');
                    aiApplicationIds.push(resp.body.id);
                });

                await test.step('Poll until phase becomes deployed', async () => {
                    await waitUtils.waitForAiApplicationPhase(aiApplicationIds[0], 'deployed');
                });

                await test.step('Verify application CR exists in cluster', async () => {
                    await waitUtils.waitForApplicationCrdToAppear(appName, namespace);
                    await waitUtils.waitForApplicationCRStatus(appName, namespace, 'completed', 'deployed');
                });
            });

            test('T5857351: API: Create AI Application with custom values override', async () => {
                const appName = objectGenerator.generateShortNameByPrefix('api-override', 5);
                const namespace = namespaceList[0];
                const overrides = { replicaCount: 2 };

                await test.step('Create AI application with overrides', async () => {
                    const resp = await aiApplicationUtils.createAiApplication({
                        name: appName,
                        projectId: projectIdList[0].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: {
                            chartRef: { chartURL: CUSTOM_URL_CHART },
                            values: overrides
                        }
                    });

                    expect(resp.response.statusCode).toBe(200);
                    aiApplicationIds.push(resp.body.id);
                });

                await test.step('Wait for application to be deployed', async () => {
                    await waitUtils.waitForApplicationCRStatus(appName, namespace, 'completed', 'deployed');
                });

                await test.step('Verify overrides are applied in the CR', async () => {
                    const cr = await kubeCli.getCrd(crds.applicationsCrd, appName, namespace);
                    const crValues = (cr as any)?.spec?.helm?.values;
                    expect(crValues).toBeDefined();
                    expect(crValues).toMatchObject(overrides);
                });
            });

            test('T5857353: API: Get single AI Application', async () => {
                const appName = objectGenerator.generateShortNameByPrefix('api-get', 5);

                await test.step('Create an AI application', async () => {
                    const resp = await aiApplicationUtils.createAiApplication({
                        name: appName,
                        projectId: projectIdList[0].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                    });
                    aiApplicationIds.push(resp.body.id);
                });

                await test.step('Wait for application to have a phase', async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const resp = await aiApplicationUtils.getAiApplication(aiApplicationIds[0]);
                            return resp.body.phase !== undefined && resp.body.phase !== null;
                        },
                        `Wait for AI application ${appName} to have a phase`,
                        true,
                        60,
                        3
                    );
                });

                await test.step('GET application by ID and verify all fields', async () => {
                    const resp = await aiApplicationUtils.getAiApplication(aiApplicationIds[0]);

                    expect(resp.response.statusCode).toBe(200);
                    expect(resp.body.id).toBe(aiApplicationIds[0]);
                    expect(resp.body.name).toBe(appName);
                    expect(resp.body.source).toBeDefined();
                    expect(resp.body.type).toBeDefined();
                    expect(resp.body.phase).toBeDefined();
                    expect(resp.body.projectId).toBe(projectIdList[0].toString());
                    expect(resp.body.projectName).toBe(projectNameList[0]);
                    expect(resp.body.desiredPhase).toBeDefined();
                    expect(resp.body.spec).toBeDefined();
                    expect(resp.body.createdAt).toBeDefined();
                    expect(resp.body.updateCount).toBeDefined();
                });

                await test.step('GET with non-existent UUID returns error', async () => {
                    const fakeId = '00000000-0000-0000-0000-000000000000';
                    try {
                        await aiApplicationUtils.getAiApplication(fakeId);
                        expect(true).toBe(false); // Should not reach here
                    } catch (e: any) {
                        expect(e.statusCode).toBe(404);
                    }
                });
            });

            test('T5857354: API: Delete AI Application', async () => {
                const appName = objectGenerator.generateShortNameByPrefix('api-delete', 5);
                const namespace = namespaceList[0];

                await test.step('Create and wait for deployment', async () => {
                    const resp = await aiApplicationUtils.createAiApplication({
                        name: appName,
                        projectId: projectIdList[0].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                    });
                    aiApplicationIds.push(resp.body.id);

                    await waitUtils.waitForApplicationCRStatus(appName, namespace, 'completed', 'deployed');
                });

                await test.step('DELETE application and verify response', async () => {
                    const resp = await aiApplicationUtils.deleteAiApplication(aiApplicationIds[0]);
                    expect(resp.response.statusCode).toBe(202);
                    aiApplicationIds.pop();
                });

                await test.step('Verify application removed from API list', async () => {
                    await waitUtils.waitForAiApplicationToBeRemoved(appName);
                });

                await test.step('Verify Application CR removed from cluster', async () => {
                    await kubeCli.waitForCrdToBeDeleted(crds.applicationsCrd, appName, namespace);
                });
            });
        });

        test.describe('AI Application List API tests', () => {
            const listAppNames: string[] = [];
            const listAppIds: string[] = [];

            test.beforeAll(async () => {
                // Create 3 apps: 2 in project 1, 1 in project 2
                const names = Array.from({ length: 3 }, (_, i) =>
                    objectGenerator.generateShortNameByPrefix(`api-list-${i}`, 5)
                );
                const configs = [
                    { name: names[0], projectId: projectIdList[0].toString() },
                    { name: names[1], projectId: projectIdList[0].toString() },
                    { name: names[2], projectId: projectIdList[1].toString() }
                ];

                const responses = await Promise.all(
                    configs.map((cfg) =>
                        aiApplicationUtils.createAiApplication({
                            name: cfg.name,
                            projectId: cfg.projectId,
                            clusterId: flowRootState.cluster.uuid,
                            spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                        })
                    )
                );

                for (const resp of responses) {
                    listAppNames.push(resp.body.name);
                    listAppIds.push(resp.body.id);
                }

                // Wait for all to be deployed
                await Promise.all([
                    waitUtils.waitForApplicationCRStatus(names[0], namespaceList[0], 'completed', 'deployed'),
                    waitUtils.waitForApplicationCRStatus(names[1], namespaceList[0], 'completed', 'deployed'),
                    waitUtils.waitForApplicationCRStatus(names[2], namespaceList[1], 'completed', 'deployed')
                ]);
            });

            test.afterAll(async () => {
                await Promise.all(
                    listAppIds.map(async (appId) => {
                        try {
                            await aiApplicationUtils.deleteAiApplication(appId);
                            logger.info(`Deleted list test AI application ${appId}`);
                        } catch (e) {
                            logger.error(`Failed to delete list test AI application ${appId}`, e);
                        }
                    })
                );
            });

            test('T5857352: API: List AI Applications with pagination, sorting, filtering, and search', async () => {
                await test.step('Pagination: offset=0, limit=1 returns 1 result', async () => {
                    const resp = await aiApplicationUtils.listAiApplications({ offset: 0, limit: 1 });
                    expect(resp.body.applications.length).toBe(1);
                });

                await test.step('Sorting: sortBy name asc returns alphabetical order', async () => {
                    const resp = await aiApplicationUtils.listAiApplications({
                        sortOrder: 'asc',
                        sortBy: 'name' as any
                    });
                    const names = resp.body.applications.map((a: any) => a.name);
                    const sorted = [...names].sort((a, b) => a.localeCompare(b));
                    expect(names).toEqual(sorted);
                });

                await test.step('Filtering: filterBy projectId returns only apps in that project', async () => {
                    const resp = await aiApplicationUtils.listAiApplications({
                        filterBy: [`projectId==${projectIdList[0]}`]
                    });
                    for (const app of resp.body.applications) {
                        expect(app.projectId).toBe(projectIdList[0].toString());
                    }
                    // Should have at least 2 apps in project 1
                    expect(resp.body.applications.length).toBeGreaterThanOrEqual(2);
                });

                await test.step('Search: search by substring returns matching results', async () => {
                    const searchTerm = listAppNames[0].substring(0, 12);
                    const resp = await aiApplicationUtils.listAiApplications({ search: searchTerm });
                    expect(resp.body.applications.length).toBeGreaterThan(0);
                    for (const app of resp.body.applications) {
                        expect(app.name).toContain(searchTerm);
                    }
                });
            });
        });

        test.describe('AI Application Auth API tests', () => {
            test('T5857355: API: Authentication and authorization (401/403)', async () => {
                await test.step('GET without token returns 401', async () => {
                    try {
                        await aiApplicationUtils.listAiApplications({ accessToken: '' });
                        expect(true).toBe(false);
                    } catch (e: any) {
                        expect(e.statusCode).toBe(401);
                    }
                });

                await test.step('POST without token returns 401', async () => {
                    try {
                        await aiApplicationUtils.createAiApplication(
                            {
                                name: 'unauthorized-app',
                                projectId: projectIdList[0].toString(),
                                clusterId: flowRootState.cluster.uuid,
                                spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                            },
                            ''
                        );
                        expect(true).toBe(false);
                    } catch (e: any) {
                        expect(e.statusCode).toBe(401);
                    }
                });

                await test.step('POST with viewer role returns 403', async () => {
                    try {
                        await aiApplicationUtils.createAiApplication(
                            {
                                name: 'viewer-app',
                                projectId: projectIdList[0].toString(),
                                clusterId: flowRootState.cluster.uuid,
                                spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                            },
                            flowRootState.users.viewer.accessToken
                        );
                        expect(true).toBe(false);
                    } catch (e: any) {
                        expect(e.statusCode).toBe(403);
                    }
                });

                await test.step('DELETE with viewer role returns 403', async () => {
                    // Create an app first as admin
                    const appName = objectGenerator.generateShortNameByPrefix('auth-test', 5);
                    const resp = await aiApplicationUtils.createAiApplication({
                        name: appName,
                        projectId: projectIdList[0].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                    });
                    aiApplicationIds.push(resp.body.id);

                    try {
                        await aiApplicationUtils.deleteAiApplication(
                            resp.body.id,
                            flowRootState.users.viewer.accessToken
                        );
                        expect(true).toBe(false);
                    } catch (e: any) {
                        expect(e.statusCode).toBe(403);
                    }
                });
            });
        });

        test.describe('AI Application CR verification tests', () => {
            let crAppName: string;
            let crAppId: string;
            let crNamespace: string;

            test.beforeAll(async () => {
                crAppName = objectGenerator.generateShortNameByPrefix('api-cr', 5);
                crNamespace = namespaceList[0];

                const resp = await aiApplicationUtils.createAiApplication({
                    name: crAppName,
                    projectId: projectIdList[0].toString(),
                    clusterId: flowRootState.cluster.uuid,
                    spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                });
                crAppId = resp.body.id;

                await waitUtils.waitForApplicationCRStatus(crAppName, crNamespace, 'completed', 'deployed');
            });

            test.afterAll(async () => {
                try {
                    await aiApplicationUtils.deleteAiApplication(crAppId);
                    logger.info(`Deleted CR test AI application ${crAppId}`);
                } catch (e) {
                    logger.error(`Failed to delete CR test AI application ${crAppId}`, e);
                }
            });

            test('T5857356: Verify Application CR annotations and labels for Run:ai-created app', async () => {
                await test.step('Get application CR from cluster', async () => {
                    const cr = await kubeCli.getCrd(crds.applicationsCrd, crAppName, crNamespace);
                    const metadata = (cr as any)?.metadata;
                    const annotations = metadata?.annotations;
                    const labels = metadata?.labels;
                    const spec = (cr as any)?.spec;

                    await test.step('Verify annotations', () => {
                        expect(annotations).toBeDefined();
                        expect(annotations['run.ai/application-update-counter']).toBeDefined();
                        expect(annotations['run.ai/application-id']).toBe(crAppId);
                        expect(annotations['run.ai/application']).toBe('true');
                    });

                    await test.step('Verify labels', () => {
                        expect(labels).toBeDefined();
                        expect(labels['run.ai/project']).toBe(projectNameList[0]);
                        expect(labels['run.ai/project-id']).toBeDefined();
                    });

                    await test.step('Verify spec contains helm chartRef and values', () => {
                        expect(spec).toBeDefined();
                        expect(spec.helm).toBeDefined();
                        expect(spec.helm.chartUrl || spec.helm.chartRef).toBeDefined();
                    });

                    await test.step('Verify spec source and type', () => {
                        expect(spec.source).toBe('runai');
                        expect(spec.type).toBe('application');
                    });
                });
            });

            test('T5857357: Verify Application CR status after successful deployment', async () => {
                await test.step('Get application CR and verify status', async () => {
                    const cr = await kubeCli.getCrd(crds.applicationsCrd, crAppName, crNamespace);
                    const status = (cr as any)?.status;

                    await test.step('Verify action status', () => {
                        expect(status.action).toBeDefined();
                        expect(status.action.type).toBe('install');
                        expect(status.action.phase).toBe('completed');
                    });

                    await test.step('Verify helm status', () => {
                        expect(status.helmStatus).toBeDefined();
                        expect(status.helmStatus.helmPhase).toBe('deployed');
                    });

                    await test.step('Verify resources list exists', () => {
                        expect(status.resources).toBeDefined();
                        expect(Array.isArray(status.resources)).toBe(true);
                        expect(status.resources.length).toBeGreaterThan(0);
                        logger.info(`Application CR has ${status.resources.length} resources`);
                    });
                });
            });
        });
    });
});
