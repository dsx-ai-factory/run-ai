/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    departmentUtils,
    tags,
    Utils,
    apiUtils,
    datavolumeUtils,
    waitUtils,
    workloadsUtilsV2,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { kubeCli, bashCli } from '../../../../../../modules/cli';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    supportedFrom,
    supportedPlatforms,
    SupportedVersions,
    Platform
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { DataVolumePage } from '../../../../../../modules/ui/runaiV2/pages/DataVolumePage';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { workloadsCrudService } from '../../../../../../modules/api';

test.describe(`Storage data sources E2E Tests, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let dataSourcePage: DataSourcePage;
        let trainingPage: TrainingPage;
        let loginUIPage: LoginUi;
        let dataVolumePage: DataVolumePage;
        let hybridForm: HybridForm;
        let videoName: string;
        let department1Id: number;
        let department2Id: number;
        let projectNameList: Array<string>;
        let namespaceList: Array<string>;
        let projectIdList: Array<number>;
        let dataSourceName: string;
        let dataVolumeName: string;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            await apiUtils.enableDataVolumes();
            const [dept1, dept2] = await Promise.all([
                departmentUtils.createBackendDepartmentWithName(
                    objectGenerator.generateShortNameByPrefix('d1')
                ),
                departmentUtils.createBackendDepartmentWithName(
                    objectGenerator.generateShortNameByPrefix('d2')
                )
            ]);
            department1Id = Number(dept1.body.id);
            department2Id = Number(dept2.body.id);
            projectNameList = Array.from({ length: 5 }, () =>
                objectGenerator.generateShortNameByPrefix('project-e2e-for-pvc-ds')
            );
            namespaceList = projectNameList.map((name) => Utils.namespaceFromProjectName(name));
            // create 1 project under dept1, 2 projects under dept2, 2 projects under the 'default'
            const projectsToCreate = [
                { namespace: namespaceList[0], name: projectNameList[0], departmentId: department1Id },
                { namespace: namespaceList[1], name: projectNameList[1], departmentId: department2Id },
                { namespace: namespaceList[2], name: projectNameList[2], departmentId: department2Id },
                { namespace: namespaceList[3], name: projectNameList[3] },
                { namespace: namespaceList[4], name: projectNameList[4] }
            ];
            const projResponse = await Promise.all(
                projectsToCreate.map((config) =>
                    projectResourcesUtils.createProject(config.namespace, {
                        name: config.name,
                        deservedGpus: 0,
                        maxAllowedGpus: 0,
                        ...(config.departmentId && { departmentId: config.departmentId })
                    })
                )
            );
            projectIdList = projResponse.map((proj) => proj.id);
        });

        test.afterAll('deleting project and namespace', async () => {
            await apiUtils.disableDataVolumes();
            // delete all projects and ns
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
            // delete all departments
            await Promise.all(
                [department1Id, department2Id].map((deptId) => departmentUtils.deleteDepartment(deptId))
            );
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            loginUIPage = new LoginUi(browserEngine.currentPage);
            dataVolumePage = new DataVolumePage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        async function checkStorageNotVisible(): Promise<void> {
            await hybridForm.page.waitForLoadState('load');
            await hybridForm.openDataSourcesSection();
            await hybridForm.openDataSourcesDrawer();
            // TO-DO temp workaround: Eslint and Prettier rules have a conflict with single quote in a string
            await expect(hybridForm.locator('text=Looks like you don')).toBeVisible();
            await expect(hybridForm.locator('text=have any setups yet')).toBeVisible();
            await hybridForm.closeDrawer();
            await hybridForm.closeWorkloadForm();
        }

        function verifyStorageRW(
            podName: string,
            namespace: string,
            containerPath: string,
            expectedSize: string,
            canWrite = true
        ): void {
            const dataToWrite = 'test123';
            if (canWrite) {
                // write data
                bashCli.exec(
                    `kubectl exec -it ${podName} -n ${namespace} -- /bin/sh -c 'echo ${dataToWrite} > ${containerPath}/tmp.txt'`
                );
                // read data
                const data = bashCli
                    .exec(`kubectl exec -it ${podName} -n ${namespace} -- cat ${containerPath}/tmp.txt`)
                    .trim();
                expect(data).toEqual(dataToWrite);
            } else {
                // write data（expect failure）
                try {
                    bashCli.exec(
                        `kubectl exec -it ${podName} -n ${namespace} -- /bin/sh -c 'echo ${dataToWrite} > ${containerPath}/tmp.txt'`
                    );
                } catch (error) {
                    expect(String(error)).toContain('command terminated with exit code');
                }
            }
        }

        test.describe('test storage data source', function () {
            let workloadIds: Array<string> = [];

            test.beforeEach(function () {
                dataSourceName = objectGenerator.generateNameByPrefix('ds');
                dataVolumeName = objectGenerator.generateNameByPrefix('data-volume');
                workloadsCrudService.trainingsApi.accessToken = flowRootState.users.l1_researcher.accessToken;
            });

            test.afterEach(async () => {
                await Promise.all(
                    workloadIds.map(async (workloadId) => {
                        try {
                            await workloadsCrudUtils.deleteTraining(workloadId);
                            workloadIds = workloadIds.filter((id) => id !== workloadId);
                        } catch (error) {
                            logger.error(`failed to delete workload ${workloadId}`);
                        }
                    })
                );
            });

            const dataSets = [
                {
                    testCase: 'test Kind standard data volume',
                    platform: [Platform.KIND],
                    claimSize: '1',
                    claimUnit: 'GiB',
                    storageClass: 'standard',
                    accessMode: 'Read-write by one node',
                    testEnable: true,
                    videoName: 'kind-standard-data-volume'
                },
                {
                    testCase: 'test EKS gp2 data volume',
                    platform: [Platform.EKS],
                    claimSize: '1',
                    claimUnit: 'GiB',
                    storageClass: 'gp2',
                    accessMode: 'Read-write by one node',
                    testEnable: true,
                    videoName: 'eks-gp2-data-volume'
                },
                {
                    testCase: 'test GCP standard-rwo data volume',
                    platform: [Platform.GCP],
                    claimSize: '1',
                    claimUnit: 'GiB',
                    storageClass: 'standard-rwo',
                    accessMode: 'Read-write by one node',
                    testEnable: true,
                    videoName: 'gcp-standard-rwo-data-volume'
                }
            ];

            dataSets.forEach((caseData) => {
                supportedPlatforms(caseData.platform)(function () {
                    if (!caseData.testEnable) return;
                    test(caseData.testCase, async () => {
                        // PVC provisioning plus the five workloads below outlast the default timeout
                        test.setTimeout(25 * 60 * 1000);
                        videoName = caseData.videoName;
                        const trainingNames = Array.from({ length: 5 }, () =>
                            objectGenerator.generateShortNameByPrefix('training')
                        );
                        const containerPath = '/testdata';

                        await test.step('signin with L1 researcher', async () => {
                            await authPage.signIn(
                                flowRootState.users.l1_researcher.email,
                                flowRootState.users.l1_researcher.password
                            );
                        });

                        await test.step('create a project scoped data source under department 1', async () => {
                            await dataSourcePage.createPVC({
                                name: dataSourceName,
                                scope: Scope.Project,
                                scopeEntityName: projectNameList[0],
                                storageClass: caseData.storageClass,
                                accessMode: caseData.accessMode,
                                claimSize: caseData.claimSize,
                                claimUnit: caseData.claimUnit,
                                volumeMode: 'Filesystem',
                                containerPath: containerPath
                            });
                            await dataSourcesUtils.validateExpectedStatus(dataSourceName, 'No issues found');
                        });

                        await test.step('create training workload with the data source', async () => {
                            await trainingPage.createFlexibleTraining(projectNameList[0], trainingNames[0], {
                                dsName: dataSourceName,
                                containerPath: containerPath
                            });
                            await trainingPage.waitForWorkloadUIStatus(
                                trainingNames[0],
                                EWorkloadStatus.Running
                            );
                            const workloadId = await workloadsUtilsV2.getWorkloadIdByName(
                                trainingNames[0],
                                projectIdList[0]
                            );
                            workloadIds.push(workloadId);
                        });

                        await test.step('logout and login with data volume administrator', async () => {
                            await loginUIPage.logoutToLoginPage();
                            await authPage.signIn(
                                flowRootState.users.data_volume_administrator.email,
                                flowRootState.users.data_volume_administrator.password
                            );
                        });

                        await test.step('data volume administrator creates a project scoped data volume from the existing PVC', async () => {
                            // shared data volume for a project in depmt 2 and depmt "default"
                            await dataVolumePage.goToPage();
                            await dataVolumePage.createDataVolume(
                                dataVolumeName,
                                {
                                    projectName: projectNameList[0],
                                    pvcName: dataSourceName
                                },
                                {
                                    scopeType: ScopeType.Project,
                                    scopeNameList: [projectNameList[1], projectNameList[3]]
                                }
                            );
                            await datavolumeUtils.waitForCrdDatavolumeToBeInPhase(dataVolumeName, 'Ready');
                        });

                        await test.step('logout and login with L1 researcher', async () => {
                            await loginUIPage.logoutToLoginPage();
                            await authPage.signIn(
                                flowRootState.users.l1_researcher.email,
                                flowRootState.users.l1_researcher.password
                            );
                        });

                        await test.step('create workload using shared data volume under department 2', async () => {
                            await trainingPage.createFlexibleTraining(projectNameList[1], trainingNames[1], {
                                dsName: dataVolumeName,
                                containerPath: containerPath
                            });
                            await trainingPage.waitForWorkloadUIStatus(
                                trainingNames[1],
                                EWorkloadStatus.Running
                            );
                            const workloadId = await workloadsUtilsV2.getWorkloadIdByName(
                                trainingNames[1],
                                projectIdList[1]
                            );
                            workloadIds.push(workloadId);
                        });

                        await test.step('create workload using shared data volume under the default department', async () => {
                            await trainingPage.createFlexibleTraining(projectNameList[3], trainingNames[3], {
                                dsName: dataVolumeName,
                                containerPath: containerPath
                            });
                            await trainingPage.waitForWorkloadUIStatus(
                                trainingNames[3],
                                EWorkloadStatus.Running
                            );
                            const workloadId = await workloadsUtilsV2.getWorkloadIdByName(
                                trainingNames[3],
                                projectIdList[3]
                            );
                            workloadIds.push(workloadId);
                        });

                        await test.step('verify not able to see shared data volume out of scope under department 2', async () => {
                            await trainingPage.fillFirstNewTrainingPage(
                                projectNameList[2],
                                trainingNames[2],
                                false,
                                undefined,
                                undefined,
                                true
                            );
                            await checkStorageNotVisible();
                        });

                        await test.step('verify not able to see shared data volume out of scope under the default department', async () => {
                            await trainingPage.fillFirstNewTrainingPage(
                                projectNameList[4],
                                trainingNames[4],
                                false,
                                undefined,
                                undefined,
                                true
                            );
                            await checkStorageNotVisible();
                        });

                        await test.step('verify pvc volume for training0 is R/W', async () => {
                            const pods = await kubeCli.listPodsOfJob(
                                `runai-${projectNameList[0]}`,
                                trainingNames[0]
                            );
                            verifyStorageRW(
                                pods[0].metadata.name,
                                `runai-${projectNameList[0]}`,
                                containerPath,
                                caseData.claimSize + caseData.claimUnit[0],
                                true
                            );
                        });

                        await test.step('verify data volume for training1 and training3 is read only', async () => {
                            for (const idx of [1, 3]) {
                                const pods = await kubeCli.listPodsOfJob(
                                    `runai-${projectNameList[idx]}`,
                                    `${trainingNames[idx]}`
                                );
                                verifyStorageRW(
                                    pods[0].metadata.name,
                                    `runai-${projectNameList[idx]}`,
                                    containerPath,
                                    caseData.claimSize + caseData.claimUnit[0],
                                    false
                                );
                            }
                        });

                        await test.step('delete training worloads', async () => {
                            await Promise.all(
                                workloadIds.map(async (workloadId) => {
                                    try {
                                        await workloadsCrudUtils.deleteTraining(workloadId);
                                        workloadIds = workloadIds.filter((id) => id !== workloadId);
                                    } catch (error) {
                                        logger.error(`failed to delete workload ${workloadId}`);
                                    }
                                })
                            );
                        });

                        await test.step('data volume administrator can delete data volume', async () => {
                            await loginUIPage.logoutToLoginPage();
                            await authPage.signIn(
                                flowRootState.users.data_volume_administrator.email,
                                flowRootState.users.data_volume_administrator.password
                            );
                            await dataVolumePage.goToPage();
                            await dataVolumePage.deleteDataVolume(dataVolumeName);
                            await waitUtils.waitForElementToNotExist(
                                dataVolumePage,
                                `td:has-text("${dataVolumeName}")`,
                                120
                            );
                        });

                        await test.step('L1 research user can delete data source', async () => {
                            await loginUIPage.logoutToLoginPage();
                            await authPage.signIn(
                                flowRootState.users.l1_researcher.email,
                                flowRootState.users.l1_researcher.password
                            );
                            await dataSourcePage.goToPage();
                            await dataSourcePage.deleteDataSource(dataSourceName);
                            await waitUtils.waitForElementToNotExist(
                                dataVolumePage,
                                `td:has-text("${dataSourceName}")`,
                                120
                            );
                        });
                    });
                });
            });
        });
    });
});
