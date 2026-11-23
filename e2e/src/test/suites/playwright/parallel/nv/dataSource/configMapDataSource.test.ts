/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import {
    objectGenerator,
    flowRootState,
    projectResourcesUtils,
    operations,
    dataSourcesUtils,
    tags,
    logger,
    departmentUtils,
    trainingUtils,
    workloadsUtilsV2,
    waitUtils
} from '../../../../../../modules/helpers';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { cliV2WorkloadType } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';

supportedFrom(SupportedVersions.V2_18)(function () {
    test.describe(`Create configmap data source, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.runOn.dgxc} ${tags.teams.loong}`, () => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let dataSource: DataSourcePage;
        let workloadPage: WorkloadPage;
        let project: any;
        let trainingPage: TrainingPage;
        let videoName: string;
        const MOUNT_PATH = '/mnt/data';
        const CONFIG_KEY1 = 'key1';
        const CONFIG_KEY2 = 'key2';
        const CONFIG_VALUE1 = 'value1';
        const CONFIG_VALUE2 = 'value2';

        test.beforeAll(async () => {
            const projectName = objectGenerator.generateNameByPrefix('project-cm');
            project = await projectResourcesUtils.createProjectAndNamespace(projectName);
            runaiCliV2.loginAsUser(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            runaiCliV2.projectSet(project.name);
        });

        test.afterAll(async () => {
            await operations.deleteProjectAndNamespace(project.id);
        });

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            dataSource = new DataSourcePage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            await auth.loginAsRole(PROPERTY_TO_ROLE_MAPPING.l1_researcher);
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T4050113: Create configmap data source in project scope', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'configmap-datasource-project';
            const configMapName = objectGenerator.generateNameByPrefix('e2e-configmap');
            const dataSourceName = objectGenerator.generateNameByPrefix('e2e-ds-cm');
            const workloadName = objectGenerator.generateNameByPrefix('e2e-wl-cm');

            await test.step('Create ConfigMap in Kubernetes', async () => {
                const configMapYaml = `apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: ${configMapName}\n  namespace: ${project.namespace}\n  labels:\n    run.ai/resource: feli-config-map\ndata:\n  ${CONFIG_KEY1}: ${CONFIG_VALUE1}\n  ${CONFIG_KEY2}: ${CONFIG_VALUE2}\n`;
                await kubeCli.createFromYaml(configMapYaml);
            });

            await test.step('Verify ConfigMap exists', async () => {
                const cm = await kubeCli.getConfigMap(configMapName, project.namespace);
                expect(cm).toBeTruthy();
            });

            await test.step('Create data source in project scope', async () => {
                await dataSource.createConfigMap({
                    name: dataSourceName,
                    scope: Scope.Project,
                    scopeEntityName: project.name,
                    configMapName: configMapName,
                    mountPath: MOUNT_PATH
                });
            });

            await test.step('Wait for data source to be ready', async () => {
                await dataSourcesUtils.waitForDataSourceStatusInTable(
                    dataSourceName,
                    'No issues found',
                    dataSource
                );
            });

            await test.step('Create training workload with data source', async () => {
                await trainingPage.createFlexibleTraining(project.name, workloadName, {
                    dsName: dataSourceName,
                    containerPath: MOUNT_PATH
                });
            });

            await test.step('Wait for workload to be running', async () => {
                await workloadPage.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
            });

            await test.step('Verify ConfigMap data is mounted correctly', async () => {
                const lsOutput = await waitUtils.waitForRunaiExecNonEmptyOutput(
                    workloadName,
                    `ls ${MOUNT_PATH}`
                );
                expect(lsOutput).toContain(CONFIG_KEY1);
                expect(lsOutput).toContain(CONFIG_KEY2);

                const value1 = await waitUtils.waitForRunaiExecNonEmptyOutput(
                    workloadName,
                    `cat ${MOUNT_PATH}/${CONFIG_KEY1}`
                );
                const value2 = await waitUtils.waitForRunaiExecNonEmptyOutput(
                    workloadName,
                    `cat ${MOUNT_PATH}/${CONFIG_KEY2}`
                );
                expect(value1.trim()).toBe(CONFIG_VALUE1);
                expect(value2.trim()).toBe(CONFIG_VALUE2);
            });

            await test.step('Test data source scope limitation with another project', async () => {
                const anotherProjectName = objectGenerator.generateNameByPrefix('project-cm-other');
                const anotherProject =
                    await projectResourcesUtils.createProjectAndNamespace(anotherProjectName);
                const anotherWorkloadName = objectGenerator.generateNameByPrefix('e2e-wl-cm-other');
                await browserEngine.currentPage.reload();
                await trainingPage.fillFirstNewTrainingPage(
                    anotherProject.name,
                    anotherWorkloadName,
                    false,
                    undefined,
                    undefined,
                    true
                );

                const hybridForm = new HybridForm(browserEngine.currentPage);
                await hybridForm.openDataSourcesSection();
                await hybridForm.openDataSourcesDrawer();
                const dsLocator = hybridForm.page.locator(`text=${dataSourceName}`);
                const exists = (await dsLocator.count()) > 0;
                expect(exists).toBe(false);

                await dataSource.leavePageWithConfirmation();
                await operations.deleteProjectAndNamespace(anotherProject.id);
            });

            await test.step('Clean up resources', async () => {
                runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                await dataSource.goToPage();
                await dataSource.deleteDataSource(dataSourceName);
                await kubeCli.deleteConfigMap({ name: configMapName, namespace: project.namespace });
            });
        });
    });
});

supportedFrom(SupportedVersions.V2_19)(function () {
    test.describe(`Create configmap data source API tests, tags: ${tags.tool.api} ${tags.resource.datasource} ${tags.teams.loong}`, () => {
        let project: any;
        let newDepartment: any;
        let projectInNewDepartment: any;
        let anotherProjectInDefaultDept: any;
        const MOUNT_PATH = '/mnt/data';
        const CONFIG_KEY1 = 'key1';
        const CONFIG_KEY2 = 'key2';
        const CONFIG_VALUE1 = 'value1';
        const CONFIG_VALUE2 = 'value2';
        let createdWorkloadIds: string[] = [];
        let createdDataSourceIds: string[] = [];
        let createdConfigMapNames: string[] = [];

        test.beforeAll(async () => {
            // Create main project
            const projectName = objectGenerator.generateNameByPrefix('project-cm-api');
            project = await projectResourcesUtils.createProjectAndNamespace(projectName);

            // Create new department and project in that department
            const departmentResponse = await departmentUtils.createBackendDepartmentAndValidate(0, 0);
            newDepartment = {
                id: departmentResponse.departmentId,
                name: departmentResponse.departmentName
            };

            // Wait for department to be available in backend
            await departmentUtils.waitForDepartmentInBackend(newDepartment.name);

            const projectInNewDeptName = objectGenerator.generateNameByPrefix('project-cm-new-dept');
            projectInNewDepartment = await projectResourcesUtils.createProjectAndNamespace(
                projectInNewDeptName,
                newDepartment.id
            );

            // Create another project in default department
            const anotherProjectName = objectGenerator.generateNameByPrefix('project-cm-default-dept');
            anotherProjectInDefaultDept =
                await projectResourcesUtils.createProjectAndNamespace(anotherProjectName);

            runaiCliV2.loginAsUser(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            runaiCliV2.projectSet(project.name);
        });

        test.afterAll(async () => {
            await operations.deleteProjectAndNamespace(project.id);
            await operations.deleteProjectAndNamespace(projectInNewDepartment.id);
            await operations.deleteProjectAndNamespace(anotherProjectInDefaultDept.id);
            await departmentUtils.deleteDepartment(newDepartment.id);
        });

        test.afterEach(async () => {
            await Promise.all(
                createdWorkloadIds.map(async (workloadId) => {
                    try {
                        await workloadsUtilsV2.deleteWorkload(workloadId, WorkloadType.TRAINING);
                    } catch (error) {
                        logger.error(`Failed to delete workload ${workloadId}`, error);
                    }
                })
            );

            await Promise.all(
                createdDataSourceIds.map(async (dataSourceId) => {
                    try {
                        await dataSourcesUtils.deleteConfigMapAsset(
                            dataSourceId,
                            flowRootState.users.l1_researcher.accessToken
                        );
                    } catch (error) {
                        logger.error(`Failed to delete data source ${dataSourceId}`, error);
                    }
                })
            );

            await Promise.all(
                createdConfigMapNames.map(async (configMapName) => {
                    try {
                        await kubeCli.deleteConfigMap({ name: configMapName, namespace: 'runai' });
                    } catch (error) {
                        logger.error(`Failed to delete configmap ${configMapName}`, error);
                    }
                })
            );

            createdWorkloadIds = [];
            createdDataSourceIds = [];
            createdConfigMapNames = [];
        });

        // Helper functions to reduce code duplication
        async function createConfigMap(
            configMapName: string,
            scope: 'cluster' | 'department',
            departmentId?: string
        ): Promise<void> {
            const labels =
                scope === 'cluster'
                    ? 'run.ai/resource: feli-config-map\n    run.ai/cluster-wide: "true"'
                    : `run.ai/resource: feli-config-map\n    run.ai/department: "${departmentId}"`;

            const configMapYaml = `apiVersion: v1
kind: ConfigMap
metadata:
  name: ${configMapName}
  namespace: runai
  labels:
    ${labels}
data:
  ${CONFIG_KEY1}: ${CONFIG_VALUE1}
  ${CONFIG_KEY2}: ${CONFIG_VALUE2}`;

            await kubeCli.createFromYaml(configMapYaml);
            createdConfigMapNames.push(configMapName);
        }

        async function verifyConfigMapExists(configMapName: string): Promise<void> {
            const cm = await kubeCli.getConfigMap(configMapName, 'runai');
            expect(cm).toBeTruthy();
        }

        async function verifyConfigMapNotVisibleInScope(
            configMapName: string,
            scope: 'project' | 'department' | 'cluster',
            scopeId: string
        ): Promise<void> {
            const configMaps = await dataSourcesUtils.getConfigMapInfo(configMapName, scope, scopeId);

            if (configMaps.length > 0) {
                throw new Error(
                    `ConfigMap ${configMapName} should not be visible in ${scope} scope but was found`
                );
            }

            logger.info(`Verified ConfigMap ${configMapName} is not visible in ${scope} scope`);
        }

        async function verifyConfigMapDataInWorkload(
            workloadName: string,
            projectName?: string
        ): Promise<void> {
            if (projectName) {
                runaiCliV2.projectSet(projectName);
            }

            const lsOutput = await waitUtils.waitForRunaiExecNonEmptyOutput(workloadName, `ls ${MOUNT_PATH}`);
            expect(lsOutput).toContain(CONFIG_KEY1);
            expect(lsOutput).toContain(CONFIG_KEY2);

            const value1 = await waitUtils.waitForRunaiExecNonEmptyOutput(
                workloadName,
                `cat ${MOUNT_PATH}/${CONFIG_KEY1}`
            );
            const value2 = await waitUtils.waitForRunaiExecNonEmptyOutput(
                workloadName,
                `cat ${MOUNT_PATH}/${CONFIG_KEY2}`
            );
            expect(value1.trim()).toBe(CONFIG_VALUE1);
            expect(value2.trim()).toBe(CONFIG_VALUE2);
        }

        test('T5247313: Create configmap data source in cluster scope', async () => {
            test.setTimeout(300 * 1000); // 5 minutes timeout
            const configMapName = objectGenerator.generateNameByPrefix('my-config-cluster');
            const dataSourceName = objectGenerator.generateNameByPrefix('e2e-ds-cm-cluster');
            const workloadName1 = objectGenerator.generateNameByPrefix('e2e-wl-cm-cluster-1');
            const workloadName2 = objectGenerator.generateNameByPrefix('e2e-wl-cm-cluster-2');

            await test.step('Create cluster-wide ConfigMap in runai namespace', async () => {
                await createConfigMap(configMapName, 'cluster');
            });

            await test.step('Verify cluster ConfigMap exists', async () => {
                await verifyConfigMapExists(configMapName);
            });

            await test.step('Verify ConfigMap is not visible in project scope', async () => {
                await verifyConfigMapNotVisibleInScope(configMapName, 'project', project.id.toString());
            });

            await test.step('Verify ConfigMap is not visible in department scope', async () => {
                await verifyConfigMapNotVisibleInScope(configMapName, 'department', 'default');
            });

            const dataSourceId = await dataSourcesUtils.createArbitraryConfigMap(
                dataSourceName,
                Scope.Cluster,
                configMapName,
                flowRootState.cluster.uuid
            );
            createdDataSourceIds.push(dataSourceId);

            await test.step('Wait for cluster data source to be ready', async () => {
                await dataSourcesUtils.validateExpectedStatus(dataSourceName, 'No issues found');
            });

            const workload1Id = await trainingUtils.createArbitraryTrainingWithConfigmap(
                workloadName1,
                project.id.toString(),
                dataSourceId
            );
            createdWorkloadIds.push(workload1Id);

            await test.step('Wait for first workload to be running', async () => {
                await workloadsUtilsV2.waitForWorkloadsPhase([workload1Id], 'Running');
            });

            await test.step('Verify ConfigMap data is mounted correctly in first workload', async () => {
                await verifyConfigMapDataInWorkload(workloadName1);
            });

            const workload2Id = await trainingUtils.createArbitraryTrainingWithConfigmap(
                workloadName2,
                projectInNewDepartment.id.toString(),
                dataSourceId
            );
            createdWorkloadIds.push(workload2Id);

            await test.step('Wait for second workload to be running', async () => {
                await workloadsUtilsV2.waitForWorkloadsPhase([workload2Id], 'Running');
            });

            await test.step('Verify cluster data source is accessible from different project', async () => {
                await verifyConfigMapDataInWorkload(workloadName2, projectInNewDepartment.name);
            });
        });

        test('T5247314: Create configmap data source in department scope', async () => {
            test.setTimeout(300 * 1000); // 5 minutes timeout
            const configMapName = objectGenerator.generateNameByPrefix('my-config-department');
            const dataSourceName = objectGenerator.generateNameByPrefix('e2e-ds-cm-dept');
            const workloadName1 = objectGenerator.generateNameByPrefix('e2e-wl-cm-dept-1');
            const workloadName2 = objectGenerator.generateNameByPrefix('e2e-wl-cm-dept-2');
            const workloadName3 = objectGenerator.generateNameByPrefix('e2e-wl-cm-dept-3');

            await test.step('Create department-scoped ConfigMap in runai namespace', async () => {
                await createConfigMap(configMapName, 'department', newDepartment.id.toString());
            });

            await test.step('Verify department ConfigMap exists', async () => {
                await verifyConfigMapExists(configMapName);
            });

            await test.step('Verify ConfigMap is not visible in project scope', async () => {
                await verifyConfigMapNotVisibleInScope(configMapName, 'project', project.id.toString());
            });

            await test.step('Verify ConfigMap is not visible in cluster scope', async () => {
                await verifyConfigMapNotVisibleInScope(configMapName, 'cluster', flowRootState.cluster.uuid);
            });

            const dataSourceId = await dataSourcesUtils.createArbitraryConfigMap(
                dataSourceName,
                Scope.Department,
                configMapName,
                newDepartment.id.toString(),
                flowRootState.users.l1_researcher.accessToken
            );
            createdDataSourceIds.push(dataSourceId);

            await test.step('Wait for department data source to be ready', async () => {
                await dataSourcesUtils.validateExpectedStatus(dataSourceName, 'No issues found');
            });

            const workload1Id = await trainingUtils.createArbitraryTrainingWithConfigmap(
                workloadName1,
                projectInNewDepartment.id.toString(),
                dataSourceId
            );
            createdWorkloadIds.push(workload1Id);

            await test.step('Wait for first workload to be running', async () => {
                await workloadsUtilsV2.waitForWorkloadsPhase([workload1Id], 'Running');
            });

            await test.step('Verify ConfigMap data is mounted correctly in first workload', async () => {
                await verifyConfigMapDataInWorkload(workloadName1, projectInNewDepartment.name);
            });

            await test.step('Create training workload in another project within same department', async () => {
                // Create another project in the same department
                const anotherProjectInSameDeptName =
                    objectGenerator.generateNameByPrefix('project-cm-same-dept-2');
                const anotherProjectInSameDept = await projectResourcesUtils.createProjectAndNamespace(
                    anotherProjectInSameDeptName,
                    newDepartment.id
                );

                const workload2Id = await trainingUtils.createArbitraryTrainingWithConfigmap(
                    workloadName2,
                    anotherProjectInSameDept.id.toString(),
                    dataSourceId
                );

                await workloadsUtilsV2.waitForWorkloadsPhase([workload2Id], 'Running');

                // Verify data source is accessible
                await verifyConfigMapDataInWorkload(workloadName2, anotherProjectInSameDept.name);

                // Clean up this temporary project
                await workloadsUtilsV2.deleteWorkload(workload2Id, WorkloadType.TRAINING);
                await operations.deleteProjectAndNamespace(anotherProjectInSameDept.id);
            });

            await test.step('Verify ConfigMap data source is not accessible from different department project', async () => {
                try {
                    await trainingUtils.createArbitraryTrainingWithConfigmap(
                        workloadName3,
                        anotherProjectInDefaultDept.id.toString(),
                        dataSourceId
                    );

                    throw new Error(
                        'Expected workload creation to fail when accessing department-scoped data source from different department'
                    );
                } catch (error) {
                    logger.info(
                        `Expected failure when accessing department-scoped data source from different department: ${error.message}`
                    );
                    const errorMessage = error.message.toLowerCase();
                    expect(
                        errorMessage.includes('not associated with the requested project or its ancestors')
                    ).toBe(true);
                    expect(errorMessage.includes('404')).toBe(true);
                }
            });
        });
    });
});
