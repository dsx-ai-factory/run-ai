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
    tags,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { kubeCli, bashCli } from '../../../../../../modules/cli';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';

test.describe(`Storage data sources E2E Tests, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.teams.loong} ${tags.runOn.dgxc}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let training: TrainingPage;
        let workspace: WorkspacePage;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let dataSourcePage: DataSourcePage;
        let videoName: string;
        const claimSize = '1';
        const claimUnit = 'GB';
        const storageClass = 'standard';

        test.beforeAll(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            auth = new AuthPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterAll(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test.describe('test custom container path', function () {
            test('T3942077: L1 researcher can update the data source mount path when create the job', async function () {
                test.setTimeout(300 * 1000); // 5 minutes timeout
                videoName = 'T3942077-custom-container-path';
                const pvcName = objectGenerator.generateShortNameByPrefix('pvc');
                const MOUNT_PATH = '/data1';
                const workspaceName = objectGenerator.generateShortNameByPrefix('workspace');
                const workloadContainerPath = '/data2';
                let podName: string;

                await test.step('Create a PVC of standard storage class', async () => {
                    await dataSourcePage.createPVC({
                        name: pvcName,
                        scope: Scope.Project,
                        scopeEntityName: projectName,
                        storageClass: storageClass,
                        accessMode: 'Read-write by one node',
                        claimSize: claimSize,
                        claimUnit: claimUnit,
                        volumeMode: 'Filesystem',
                        containerPath: MOUNT_PATH
                    });
                    await dataSourcesUtils.waitForDataSourceStatusInTable(
                        pvcName,
                        'No issues found',
                        dataSourcePage
                    );
                });

                await test.step('create workspace workload with the data source and update container path', async () => {
                    await workspace.goToPage();
                    await workspace.submitWorkspaceWorkloadHybridUI(projectName, workspaceName, {
                        environment: 'jupyter-scipy',
                        computeResource: 'one-gpu',
                        dataSource: {
                            dsName: pvcName,
                            containerPath: workloadContainerPath
                        }
                    });
                    await workspace.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);
                });

                await test.step('check PVC is now in Bound state with RWO mode', async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const pvcList = await kubeCli.listPVCsByNamespace(project.namespace);
                            const pvc = pvcList.items.find((p) => p.metadata?.name?.includes(pvcName));
                            return (
                                pvc !== undefined &&
                                pvc.status?.phase === 'Bound' &&
                                pvc.spec?.accessModes?.includes('ReadWriteOnce') === true
                            );
                        },
                        'waiting for PVC to be in Bound state with ReadWriteOnce mode',
                        true,
                        30
                    );
                });

                await test.step('verify mount path in workspace pod', async () => {
                    const pods = await kubeCli.listPodsOfJob(`runai-${projectName}`, workspaceName);
                    podName = pods[0].metadata.name;
                    const mountCheckCmd = `kubectl exec -it ${podName} -n runai-${projectName} -- df -h | grep '${workloadContainerPath}'`;
                    const mountOutput = bashCli.exec(mountCheckCmd).trim();
                    expect(mountOutput).toContain(workloadContainerPath);
                });

                await test.step('test file operations in PVC storage', async () => {
                    await dataSourcesUtils.verifyDataSourceRW(podName, namespace, workloadContainerPath);
                });

                await test.step('delete workspace workload', async () => {
                    await training.deleteWorkloads([workspaceName]);
                    await waitUtils.waitForElementToNotExist(
                        training,
                        `td:has-text("${workspaceName}")`,
                        120
                    );
                });

                await test.step('delete data source', async () => {
                    await dataSourcePage.goToPage();
                    await dataSourcePage.deleteDataSource(pvcName);
                    await waitUtils.waitForElementToNotExist(
                        dataSourcePage,
                        `td:has-text("${pvcName}")`,
                        120
                    );
                });

                await test.step('verify PVC does not exist anymore', async () => {
                    const pvcList = await kubeCli.listPVCsByNamespace(project.namespace);
                    const pvc = pvcList.items.find((p) => p.metadata?.name?.includes(pvcName));
                    expect(pvc).toBeUndefined();
                });
            });
        });
    });
});
