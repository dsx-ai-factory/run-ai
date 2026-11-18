/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//internal modules
import { environmentUtils, orgUnitUtils, workloadsUtilsV2 } from '../../../../../../modules/helpers';
import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    tags,
    config
} from '../../../../../../modules/helpers';

//pages
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { InferencePage } from '../../../../../../modules/ui/runaiV2/pages/InferencePage';
import { expect } from 'chai';
import { pageFilterUtils } from '../../../../../../modules/ui/utils/pageFilterUtils';
import { test } from '@playwright/test';
import { InferenceFlatUIUtil } from '../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import {
    aiPractitionerRoleName,
    projectManagerRoleName
} from '../../../../../../modules/helpers/utils/AuthorizationUtils';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';
import { IInferenceType } from '../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';

interface CreateWorkloadOptions {
    clusterName: string;
    workloadName: string;
    environmentSelector?: string;
    projectId: number;
    projectName: string;
    environmentData?: {
        name: string;
        image?: string;
        containerPort?: number;
    };
    computeResourceData?: {
        name: string;
    };
}
test.describe(`Inference workloads in Workloads UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`, () => {
    let browserEngine: BrowserEngine;
    let authPage: AuthPage;
    let workloadTable: WorkloadPage;
    let project: Project;
    let projectName: string;
    let inferencePage: InferencePage;
    let inferenceFlatUIUtil: InferenceFlatUIUtil;
    let videoName;
    const roles = [PROPERTY_TO_ROLE_MAPPING.system_administrator, aiPractitionerRoleName, projectManagerRoleName];

    roles.forEach((role) => {
        test.describe(`as ${role}`, () => {
            const startSession = async (): Promise<void> => {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                workloadTable = new WorkloadPage(browserEngine.currentPage);
                inferencePage = new InferencePage(browserEngine.currentPage);
                inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
                await authPage.loginAsRole(role);
            };

            test.beforeEach(async () => {
                await startSession();
                await workloadTable.navigateToWorkloads();
            });

            test.afterEach('stop browser and archive session', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, 'workloads-' + videoName);
            });

            const createInference = async (options: CreateWorkloadOptions): Promise<void> => {
                await pageFilterUtils.selectCluster(workloadTable.page, options.clusterName);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: options.clusterName,
                    projectName: options.projectName,
                    workloadName: options.workloadName,
                    inferenceType: IInferenceType.Custom,
                    environmentSection: {
                        newEnvironmentResource: {
                            assetName: options.environmentData.name,
                            projectName: options.projectName,
                            image: options.environmentData.image,
                            servingPort: {
                                container: options.environmentData.containerPort
                            }
                        }
                    },
                    computeSection: {
                        assetName: options.computeResourceData.name
                    }
                });
                logger.debug('inference created successfully');

                await workloadsUtilsV2.waitForStatusInWorkload(
                    options.workloadName,
                    Phase.Running,
                    120
                );
                await workloadTable.validateCurrentPageIsWorkload();
            };

            test('should copy an inference workload and create a new one via Copy & Edit', async () => {
                test.setTimeout(240_000);
                videoName = 'workload-copy-inference';
                const createWorkloadOptions: CreateWorkloadOptions = {
                    clusterName: flowRootState.cluster.name,
                    workloadName: objectGenerator.generateShortNameByPrefix('origin-inf'),
                    projectId: Number(project.id),
                    projectName: project.name,
                    environmentData: {
                        name: objectGenerator.generateNameByPrefix('env-for-copy'),
                        image: config.quickStartImage.inferenceHelloWorld,
                        containerPort: 3000
                    },
                    computeResourceData: {
                        name: 'half-gpu'
                    }
                };

                const copiedWorkloadName = `${objectGenerator.generateShortNameByPrefix('copied-inf')}`;
                await createInference(createWorkloadOptions);

                // copy
                await workloadTable.toggleWorkload(createWorkloadOptions.workloadName);
                await workloadTable.clickOnCopyAndEditWorkload();

                // edit and create
                await inferencePage.fillInferenceName(copiedWorkloadName);
                await inferencePage.clickOnCreateInferenceBtn();
                logger.debug('inference copied and created successfully');

                // clean up — pure API: fire-and-forget batch delete of workloads, then API env
                // delete (its built-in 412 retry handles the "asset in use" race while workloads
                // are Terminating).
                logger.debug('test done successfully, starting cleanup');
                const envId = await environmentUtils.getId(createWorkloadOptions.environmentData.name);
                await workloadsUtilsV2.deleteWorkloadsByProjectId(
                    String(project.id),
                    undefined,
                    undefined,
                    false
                );
                await environmentUtils.delete(envId);
                logger.debug('copy and edit inference test done successfully');
            });

            test('should delete an inference workload from the Workloads table', async () => {
                test.setTimeout(240_000);
                videoName = 'delete-inference';
                const createWorkloadOptions: CreateWorkloadOptions = {
                    clusterName: flowRootState.cluster.name,
                    workloadName: objectGenerator.generateShortNameByPrefix('delete-inference'),
                    projectName: project.name,
                    projectId: Number(project.id),
                    environmentData: {
                        name: objectGenerator.generateNameByPrefix('env-for-inference'),
                        image: config.quickStartImage.inferenceHelloWorld,
                        containerPort: 3000
                    },
                    computeResourceData: {
                        name: 'half-gpu'
                    }
                };

                await createInference(createWorkloadOptions);
                logger.debug('inference created successfully');

                logger.debug('start deleting inference');
                await workloadTable.deleteWorkload(createWorkloadOptions.workloadName);
                logger.debug('inference deleted successfully');

                // cleanup — UI delete above already waited for the workload row to disappear,
                // so the env can be removed straight away via API.
                logger.debug('test done successfully, starting cleanup');
                const envId = await environmentUtils.getId(createWorkloadOptions.environmentData.name);
                await environmentUtils.delete(envId);
                logger.debug('delete inference test done successfully');
            });

            test('should show the most recently created inference as the first row in the Workloads table', async () => {
                test.setTimeout(240_000);
                videoName = 'workload-last-created-inference';
                const createWorkloadOptions: CreateWorkloadOptions = {
                    clusterName: flowRootState.cluster.name,
                    workloadName: objectGenerator.generateShortNameByPrefix('first-inference'),
                    projectId: Number(project.id),
                    projectName: project.name,
                    environmentData: {
                        name: objectGenerator.generateNameByPrefix('env-for-last-created'),
                        image: config.quickStartImage.inferenceHelloWorld,
                        containerPort: 3000
                    },
                    computeResourceData: {
                        name: 'cpu-only'
                    }
                };

                const copiedWorkloadName = `${objectGenerator.generateShortNameByPrefix('copied-inference')}`;
                await createInference(createWorkloadOptions);

                // copy
                await workloadTable.toggleWorkload(createWorkloadOptions.workloadName);
                await workloadTable.clickOnCopyAndEditWorkload();

                // edit and create
                await inferencePage.fillInferenceName(copiedWorkloadName);
                await inferencePage.clickOnCreateInferenceBtn();
                logger.debug('inference copied and created successfully');

                await workloadTable.waitForWorkloadToAppear(copiedWorkloadName);

                const workloadNames: string[] = await workloadTable.getWorkloadNames();
                expect(workloadNames[0].trim()).to.equal(copiedWorkloadName);
                logger.debug('the last created inference is the first');

                // clean up — pure API: fire-and-forget batch delete of workloads, then API env
                // delete (its built-in 412 retry handles the "asset in use" race while workloads
                // are Terminating).
                const envId = await environmentUtils.getId(createWorkloadOptions.environmentData.name);
                await workloadsUtilsV2.deleteWorkloadsByProjectId(
                    String(project.id),
                    undefined,
                    undefined,
                    false
                );
                await environmentUtils.delete(envId);
                logger.debug('copy and edit inference test done successfully');
            });
        });
    });

    const createProject = async (): Promise<void> => {
        projectName = objectGenerator.generateShortNameByPrefix('p-0');

        project = await orgUnitUtils.createProject({
            name: projectName,
            clusterId: flowRootState.cluster.uuid,
            resources: [
                {
                    gpu: {
                        deserved: 16,
                        limit: 16
                    }
                }
            ]
        });
        logger.info(`created project ${project.name}`);
    };

    test.beforeAll(async () => {
        await createProject();
    });

    test.afterAll(async () => {
        await operations.deleteProjectAndNamespace(Number(project.id));
    });
});
