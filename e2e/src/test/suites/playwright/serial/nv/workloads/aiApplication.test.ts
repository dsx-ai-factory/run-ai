/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    objectGenerator,
    operations,
    tags,
    waitUtils,
    flowRootState,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { orgUnitService } from '../../../../../../modules/api';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { workloadsService } from '../../../../../../modules/api';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { bashCli, kubeCli } from '../../../../../../modules/cli';

test.describe(`Runai  ${tags.tool.ui} ${tags.resource.workloads}  ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        test.setTimeout(300 * 1000);
        const RAG_HELM_CHART_URL =
            'https://helm.ngc.nvidia.com/nvidia/blueprint/charts/nvidia-blueprint-rag-v2.2.0.tgz';
        const RAG_RELEASE_NAME = 'rag';

        const VSS_HELM_CHART_URL =
            'https://helm.ngc.nvidia.com/nvidia/blueprint/charts/nvidia-blueprint-vss-2.3.1.tgz';
        const VSS_RELEASE_NAME = 'vss-blueprint';

        // RAG deployment workload names
        const RAG_WORKLOAD_NAMES = [
            'ingestor-server',
            'milvus-standalone',
            'nv-ingest-paddle',
            'rag-etcd',
            'rag-frontend',
            'rag-minio',
            'rag-nemoretriever-graphic-elements-v1',
            'rag-nemoretriever-page-elements-v2',
            'rag-nemoretriever-table-structure-v1',
            'rag-nim-llm',
            'rag-nvidia-nim-llama-32-nv-embedqa-1b-v2',
            'rag-nv-ingest',
            'rag-redis-master',
            'rag-redis-replicas',
            'rag-server',
            'rag-text-reranking-nim'
        ];

        // VSS deployment workload names
        const VSS_WORKLOAD_NAMES = [
            'etcd-etcd-deployment',
            'milvus-milvus-deployment',
            'minio-minio-deployment',
            'nemo-embedding-embedding-deployment',
            'nemo-rerank-ranking-deployment',
            'neo4j-neo4j-deployment',
            'vss-blueprint',
            'vss-vss-deployment'
        ];

        let projectName: string;
        let namespace: string;
        let project: Project;
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let workloadPage: WorkloadPage;
        let videoName: string;
        let workspaceId: string;
        let trainingId: string;
        const resources = [
            {
                gpu: {
                    deserved: 8,
                    limit: 8
                }
            }
        ];
        const TEMP_DIR = '/tmp/helm-charts';

        test.beforeAll(async function () {
            const projectRequest = {
                name: objectGenerator.generateShortNameByPrefix('ai-application-proj1'),
                clusterId: flowRootState.cluster.uuid,
                resources: resources
            };

            orgUnitService.projectApi.accessToken = flowRootState.users.system_administrator.accessToken;

            const projectResponse = await orgUnitService.projectApi.createProject(projectRequest);
            project = projectResponse.body as any;

            projectName = project.name;
            namespace = Utils.namespaceFromProjectName(projectName);
            // wait until namespace is created
            await waitUtils.waitForCondition(
                async () => {
                    try {
                        const ns = await kubeCli.CoreV1Api.readNamespace({ name: namespace });
                        return ns?.metadata?.name === namespace;
                    } catch (error) {
                        logger.error(`Failed to read namespace ${namespace}: ${error}`);
                        return false;
                    }
                },
                `Wait for namespace ${namespace} to be created`,
                true,
                20,
                2
            );

            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
            workloadsService.workloadsApi.accessToken = flowRootState.users.l1_researcher.accessToken;

            browserEngine = new BrowserEngine();
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);

                // Clean up downloaded chart files
                logger.info(`Cleaning up chart directory: ${TEMP_DIR}`);
                bashCli.execWithOutput(`rm -rf ${TEMP_DIR}`, false);
                logger.info('Cleaned up downloaded chart files');
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        // Helper functions for helm operations

        function ensureChartDownloaded(chartUrl: string): string {
            bashCli.execWithOutput(`mkdir -p ${TEMP_DIR}`, false);

            // Get chart file path
            const chartName = chartUrl.split('/').pop()?.replace('.tgz', '') || 'chart';
            const chartFile = `${TEMP_DIR}/${chartName}.tgz`;

            // Download if not exists
            if (bashCli.execWithOutput(`ls -la ${chartFile}`, false).status !== 0) {
                logger.info(`Downloading chart: ${chartUrl}`);
                // Try helm pull first; fall back to curl if helm is unavailable or fails
                const pullRes = bashCli.execWithOutput(`helm pull ${chartUrl} -d ${TEMP_DIR}`, false);
                if (pullRes.status !== 0) {
                    logger.warn(
                        `helm pull failed (code=${pullRes.status}). Falling back to curl download for ${chartUrl}`
                    );
                    const curlCmd = `curl -fL "${chartUrl}" -o "${chartFile}"`;
                    const curlRes = bashCli.execWithOutput(curlCmd, false);
                    if (curlRes.status !== 0) {
                        throw new Error(
                            `Failed to download chart via helm and curl. helmStatus=${pullRes.status}, curlStatus=${curlRes.status}`
                        );
                    }
                }
            } else {
                logger.info(`Chart ${chartFile} already exists, skipping download`);
            }
            return chartFile;
        }

        function helmReleaseExists(releaseName: string, namespace: string): boolean {
            const listCmd = `helm list -n ${namespace}`;
            const res = bashCli.execWithOutput(listCmd, false);
            return res.stdout.includes(releaseName);
        }

        async function installApplication(
            releaseName: string,
            namespace: string,
            chartUrl: string
        ): Promise<void> {
            const chartFile = ensureChartDownloaded(chartUrl);
            const command = `helm install ${releaseName} -n ${namespace} ${chartFile}`;

            logger.info(`Helm install command: ${command}`);
            for (let attempt = 1; attempt <= 3; attempt++) {
                const result = bashCli.execWithOutput(command, false);
                logger.info(`Helm install attempt ${attempt} exit code: ${result.status}`);
                if (result.status === 0) break;
                if (attempt < 3) await waitUtils.sleep(3000);
            }
            await waitUtils.waitForCondition(
                () => helmReleaseExists(releaseName, namespace),
                `Wait for ${releaseName} to be installed in ${namespace}`,
                true,
                60,
                3
            );
        }

        async function uninstallApplication(releaseName: string, namespace: string): Promise<void> {
            const existedAtStart = helmReleaseExists(releaseName, namespace);
            if (!existedAtStart) {
                logger.info(
                    `Application ${releaseName} not found in namespace ${namespace}, skipping uninstall`
                );
                return;
            }

            logger.info(`Uninstalling application ${releaseName} from namespace ${namespace}`);
            const command = `helm uninstall ${releaseName} -n ${namespace}`;
            for (let attempt = 1; attempt <= 3; attempt++) {
                const result = bashCli.execWithOutput(command, false);
                logger.info(`helm uninstall attempt ${attempt}: exit=${result.status}`);
                if (result.status === 0) {
                    break;
                }
                if (attempt < 3) await waitUtils.sleep(3000);
            }

            await waitUtils.waitForCondition(
                () => !helmReleaseExists(releaseName, namespace),
                `Wait for ${releaseName} to be uninstalled from ${namespace}`,
                true,
                60,
                3
            );
            logger.info(`Confirmed ${releaseName} uninstalled from ${namespace}`);
        }

        // Helper function to wait for AI applications to match expected release name
        async function waitForAiApplications(releaseName: string, description: string): Promise<void> {
            await waitUtils.waitForCondition(
                async () => {
                    const aiApplications = await workloadPage.getWorkloadAiApplications();
                    logger.info(`AI applications: ${JSON.stringify(aiApplications)}`);
                    const allMatch = aiApplications.every((app) => app.trim() === releaseName);
                    logger.info(`All AI applications are '${releaseName}': ${allMatch}`);
                    return allMatch;
                },
                description,
                true,
                30,
                5
            );
        }

        // Helper function to wait for workloads to match expected workload names
        async function waitForWorkloads(
            expectedWorkloadNames: string[],
            description: string,
            exactMatch: boolean = false
        ): Promise<void> {
            await waitUtils.waitForCondition(
                async () => {
                    const workloads = await workloadPage.getWorkloadNames();
                    logger.info(`Workloads: ${JSON.stringify(workloads)}`);
                    const trimmedWorkloads = workloads.map((workload) => workload.trim());
                    const containsAllExpected = expectedWorkloadNames.every((expected) =>
                        trimmedWorkloads.includes(expected)
                    );
                    const noUnexpected = trimmedWorkloads.every((actual) =>
                        expectedWorkloadNames.includes(actual)
                    );
                    const allMatch = exactMatch ? containsAllExpected && noUnexpected : containsAllExpected;
                    logger.info(
                        `Workloads validation -> containsAllExpected: ${containsAllExpected}, noUnexpected: ${noUnexpected}, exactMatch: ${exactMatch}, result: ${allMatch}`
                    );
                    return allMatch;
                },
                description,
                true,
                30,
                5
            );
        }

        test.describe('UI: AI application (non-runai namespace)', function () {
            let nonRunaiNamespace = '';

            test.beforeAll(async function () {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                workloadPage = new WorkloadPage(browserEngine.currentPage);

                await authPage.signIn(
                    flowRootState.users.l1_researcher.email,
                    flowRootState.users.l1_researcher.password
                );
                await workloadPage.goToPage();
                await workloadPage.removeAdvancedFilter();
            });

            test.afterAll(async function ({}, testInfo) {
                if (nonRunaiNamespace) {
                    try {
                        await uninstallApplication(RAG_RELEASE_NAME, nonRunaiNamespace);
                    } catch (e) {
                        logger.warn(
                            `Failed to helm-uninstall ${RAG_RELEASE_NAME} from ${nonRunaiNamespace}: ${e}`
                        );
                    }
                    const del = bashCli.execWithOutput(
                        `kubectl delete namespace ${nonRunaiNamespace} --ignore-not-found --wait=false`,
                        false
                    );
                    logger.info(
                        `Cleaned up non-runai namespace ${nonRunaiNamespace} (delete exit=${del.status})`
                    );
                }
                await browserEngine.stopAndArchive(testInfo, videoName);
            });

            test('T5391584: Detect application under non-runai namespace', async function () {
                videoName = 'template-T5391584-nonrunai';
                nonRunaiNamespace = objectGenerator.generateShortNameByPrefix('test-namespace');
                await test.step('Install rag application', async () => {
                    bashCli.exec(`kubectl create namespace ${nonRunaiNamespace}`);
                    await installApplication(RAG_RELEASE_NAME, nonRunaiNamespace, RAG_HELM_CHART_URL);
                });

                await test.step('Verify applications are empty via API after install', async () => {
                    const response = await workloadsUtilsV2.listApplications();
                    logger.info(`Applications: ${JSON.stringify(response.body.applications)}`);
                    expect(response.body.applications.length).toBe(0);
                });

                await test.step('Go to workloads page and check AI Applications Column', async () => {
                    const workloads = (await workloadPage.getWorkloadNames()).map((workload) =>
                        workload.trim()
                    );
                    logger.info(`Workloads: ${JSON.stringify(workloads)}`);
                    const visibleRagWorkloads = RAG_WORKLOAD_NAMES.filter((workload) =>
                        workloads.includes(workload)
                    );
                    expect(
                        visibleRagWorkloads,
                        'RAG workloads from non-runai namespace should not appear in Workloads page. ' +
                            `Current workloads: ${JSON.stringify(workloads)}`
                    ).toEqual([]);
                    logger.info('Verified: Non-runai namespace workloads are not shown as Run:ai workloads');
                });
            });
        });

        test.describe('AI application', function () {
            let project2: Project;
            let projectName2: string;
            let namespace2: string;
            test.beforeAll(async function () {
                await installApplication(RAG_RELEASE_NAME, namespace, RAG_HELM_CHART_URL);
                await installApplication(VSS_RELEASE_NAME, namespace, VSS_HELM_CHART_URL);
            });

            test.beforeEach(async function () {
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                workloadPage = new WorkloadPage(browserEngine.currentPage);

                await authPage.signIn(
                    flowRootState.users.l1_researcher.email,
                    flowRootState.users.l1_researcher.password
                );
                await workloadPage.goToPage();
                await workloadPage.removeAdvancedFilter();
                await workloadPage.setRowsPerPage(100);
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
                try {
                    if (workspaceId) {
                        await workloadsCrudUtils.deleteWorkspace(workspaceId);
                    }
                } catch (error) {
                    logger.warn(`Failed to delete workspace ${workspaceId}: ${error}`);
                }
                try {
                    if (trainingId) {
                        await workloadsCrudUtils.deleteTraining(trainingId);
                    }
                } catch (error) {
                    logger.warn(`Failed to delete training ${trainingId}: ${error}`);
                }
            });

            test.afterAll(async function () {
                test.setTimeout(300 * 1000);
                // cleanup project2 if created in T5391588
                try {
                    if (namespace2) {
                        await uninstallApplication(VSS_RELEASE_NAME, namespace2);
                    }
                } catch (e) {
                    logger.warn(`Failed to uninstall VSS from ${namespace2}: ${e}`);
                }
                try {
                    if (project2?.id) {
                        await operations.deleteProjectAndNamespace(project2.id);
                    }
                } catch (e) {
                    logger.warn(`Failed to delete project2 ${projectName2} and ns ${namespace2}: ${e}`);
                }
                await uninstallApplication(RAG_RELEASE_NAME, namespace);
                await uninstallApplication(VSS_RELEASE_NAME, namespace);
                await waitUtils.waitForCondition(
                    async () => {
                        const response = await workloadsUtilsV2.listApplications();
                        return response.body.applications.length === 0;
                    },
                    'Wait for applications to be empty after uninstall',
                    true,
                    90,
                    3
                );
            });

            test('T5391582: Detect application under runai namespace', async function () {
                videoName = 'template-T5391582-runai';
                await test.step('Go to workloads page and check AI Applications Column', async () => {
                    const count = await workloadPage.locator(workloadPage.aiApplicationCol).count();
                    expect(count).toBe(0);
                    await workloadPage.showAllColumns();
                    await waitForWorkloads(
                        RAG_WORKLOAD_NAMES,
                        'Wait for all RAG workloads to be in the list'
                    );
                });
            });

            test('T5391583: Filter workload by appliation name', async function () {
                videoName = 'template-T5391583-filter-workload-by-appliation-name';
                const workspaceJobName = objectGenerator.generateShortNameByPrefix('workspace-job');
                const trainingJobName = objectGenerator.generateShortNameByPrefix('training-job');

                await test.step('Submit a training workload and workspace workload', async () => {
                    const { body: workspace } = await workloadsCrudUtils.createWorkspace(
                        workspaceJobName,
                        true,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity'
                        }
                    );
                    workspaceId = workspace.workloadId;

                    const { body: training } = await workloadsCrudUtils.createTraining(
                        trainingJobName,
                        true,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity'
                        }
                    );
                    trainingId = training.workloadId;
                });

                await test.step('Go to workloads page and check AI Applications Column', async () => {
                    const count = await workloadPage.locator(workloadPage.aiApplicationCol).count();
                    expect(count).toBe(0);
                    await workloadPage.showAllColumns();
                });

                await test.step('Filter workload by appliation name', async () => {
                    // filter by rag application
                    await workloadPage.addFilterOfContainandEquals(
                        'AI application',
                        'contains',
                        RAG_RELEASE_NAME
                    );
                    await waitForWorkloads(
                        RAG_WORKLOAD_NAMES,
                        'Wait for all RAG workloads to be in the list',
                        true
                    );

                    await waitForAiApplications(RAG_RELEASE_NAME, 'Wait for all AI applications to be RAG');

                    // filter by vss application
                    await workloadPage.removeAdvancedFilter();
                    await workloadPage.addFilterOfContainandEquals(
                        'AI application',
                        'equals',
                        VSS_RELEASE_NAME
                    );

                    await waitForWorkloads(
                        VSS_WORKLOAD_NAMES,
                        'Wait for all VSS workloads to be in the list',
                        true
                    );

                    await waitForAiApplications(VSS_RELEASE_NAME, 'Wait for all AI applications to be VSS');
                });

                await test.step('Remove filters', async () => {
                    await workloadPage.removeAdvancedFilter();
                });
            });

            test('T5391585: Detect application deoplyed by Nvidia AI BluePrint', async function () {
                videoName = 'template-T5391585-nvidia-ai-blueprint';
                await test.step('Go to workloads page and check AI Applications Column', async () => {
                    const count = await workloadPage.locator(workloadPage.aiApplicationCol).count();
                    expect(count).toBe(0);
                    await workloadPage.showAllColumns();
                    // check if all rag workloads are in the list
                    await waitForWorkloads(
                        VSS_WORKLOAD_NAMES,
                        'Wait for all VSS workloads to be in the list'
                    );
                });
            });

            test('T5391586: List applications', async function () {
                videoName = 'template-T5391586-list-applications';
                await test.step('List applications via API', async () => {
                    const response = await workloadsUtilsV2.listApplications();
                    logger.info(`Full JSON response: ${JSON.stringify(response.body, null, 2)}`);

                    const applications = response.body.applications;
                    logger.info(
                        `Applications array with full details: ${JSON.stringify(applications, null, 2)}`
                    );

                    // Verify response contains two applications
                    expect(applications.length).toBe(2);

                    // Verify both applications exist
                    const ragApp = applications.find((app) => app.name === 'rag');
                    const vssApp = applications.find((app) => app.name === 'vss-blueprint');

                    expect(ragApp).toBeDefined();
                    expect(vssApp).toBeDefined();
                });
            });

            test('T5391587: Get application by ID', async function () {
                videoName = 'template-T5391587-get-application-by-id';
                test.setTimeout(300 * 1000);
                // Helper function to validate application
                const validateApplication = async (
                    appName: string,
                    expectedType: string,
                    expectedWorkloads: string[]
                ): Promise<void> => {
                    const response = await workloadsUtilsV2.listApplications();
                    const applications = response.body.applications;

                    logger.info(`Looking for application: ${appName}`);
                    logger.info(`Available applications: ${JSON.stringify(applications.map((a) => a.name))}`);

                    const app = applications.find((a) => a.name === appName);
                    expect(app).toBeDefined();

                    // Wait for application and workloads to be ready
                    await waitUtils.waitForCondition(
                        async () => {
                            const appDetails = (await workloadsUtilsV2.getApplication(app.id)).body;

                            // Check phase and workloads
                            const phaseReady = appDetails.phase?.trim();
                            const resources = Array.isArray(appDetails?.resources)
                                ? appDetails.resources
                                : [];
                            if (resources.length === 0) {
                                logger.info(`${appName}: resources not ready yet`);
                                return false;
                            }
                            const externalWorkloads = resources
                                .filter((r) => r.kind === 'ExternalWorkload')
                                .map((r) => r.name);
                            const workloadsMatch = expectedWorkloads.every((expected) =>
                                externalWorkloads.some((actual) => actual.includes(expected))
                            );

                            logger.info(
                                `${appName}: phase=${phaseReady}, workloads=${workloadsMatch ? 'matched' : 'pending'}`
                            );
                            return phaseReady && workloadsMatch;
                        },
                        `Wait for ${appName} to be ready`,
                        true,
                        180,
                        5
                    );

                    const appResponse = await workloadsUtilsV2.getApplication(app.id);
                    const appDetails = appResponse.body;

                    // Verify basic fields
                    expect(appDetails.source).toBe('helm');
                    expect(appDetails.projectName).toContain('ai-application-proj');
                    expect(appDetails.name).toBe(appName);
                    expect(appDetails.type).toBe(expectedType);
                    expect(appDetails.phase).toBeDefined();
                    expect(appDetails.resources.length).toBeGreaterThan(0);

                    logger.info(`✅ ${appName} ready`);
                };

                await test.step('Get RAG application by ID', async () => {
                    await validateApplication('vss-blueprint', 'nvidia-blueprint-vss', VSS_WORKLOAD_NAMES);
                });
            });

            test('T5391588: List applications under different projects', async function () {
                test.setTimeout(300 * 1000);
                videoName = 'template-T5391588-filter-by-project';
                await test.step('Prepare: uninstall VSS from project1 namespace', async () => {
                    await uninstallApplication(VSS_RELEASE_NAME, namespace);
                    await waitUtils.waitForCondition(
                        async () => {
                            const res = await workloadsUtilsV2.listApplications({
                                projectId: project.id.toString()
                            });
                            const apps = res.body.applications ?? [];
                            const vss = apps.find((a) => a.name === VSS_RELEASE_NAME);
                            return !vss;
                        },
                        'Wait for VSS application to disappear from project1',
                        true,
                        180,
                        5
                    );
                });

                await test.step('Create project2 and install VSS there', async () => {
                    const projectRequest = {
                        name: objectGenerator.generateShortNameByPrefix('ai-application-proj2'),
                        clusterId: flowRootState.cluster.uuid,
                        resources: resources
                    };

                    orgUnitService.projectApi.accessToken =
                        flowRootState.users.system_administrator.accessToken;

                    const projectResponse = await orgUnitService.projectApi.createProject(projectRequest);
                    project2 = projectResponse.body as any;

                    projectName2 = project2.name;
                    namespace2 = Utils.namespaceFromProjectName(projectName2);

                    await waitUtils.waitForCondition(
                        async () => {
                            try {
                                const ns = await kubeCli.CoreV1Api.readNamespace({ name: namespace2 });
                                return ns?.metadata?.name === namespace2;
                            } catch (error) {
                                logger.error(`Failed to read namespace ${namespace2}: ${error}`);
                                return false;
                            }
                        },
                        `Wait for namespace ${namespace2} to be created`,
                        true,
                        20,
                        2
                    );

                    await installApplication(VSS_RELEASE_NAME, namespace2, VSS_HELM_CHART_URL);
                });

                await test.step('List applications under project 1', async () => {
                    const response = await workloadsUtilsV2.listApplications({
                        projectId: project.id.toString()
                    });
                    const applications = response.body.applications;
                    const ragApp = applications.find((app) => app.name === 'rag');
                    const vssApp = applications.find((app) => app.name === 'vss-blueprint');
                    expect(ragApp).toBeDefined();
                    expect(vssApp).not.toBeDefined();
                });

                await test.step('List applications under project 2', async () => {
                    const response = await workloadsUtilsV2.listApplications({
                        projectId: project2.id.toString()
                    });
                    const applications = response.body.applications;
                    const ragApp = applications.find((app) => app.name === 'rag');
                    const vssApp = applications.find((app) => app.name === 'vss-blueprint');
                    expect(ragApp).not.toBeDefined();
                    expect(vssApp).toBeDefined();
                });
            });
        });
    });
});
