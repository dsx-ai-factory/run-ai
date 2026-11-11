/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    objectGenerator,
    orgUnitUtils,
    tags,
    apiUtils,
    crds,
    crdUtils,
    logger,
    projectUtils,
    nodeUtils,
    nodePoolUtils,
    departmentUtils,
    operations,
    workloadsCrudUtils,
    waitUtils
} from '../../../../../../modules/helpers';
import { StatusCodes } from 'http-status-codes';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect, TestInfo } from '@playwright/test';
import { ProjectPageV2 } from '../../../../../../modules/ui/runaiV2/pages/ProjectPageV2';
import { kubeCli } from '../../../../../../modules/cli';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';
import {
    AccessRulesSubjectType,
    OrgUnitsFormName,
    SchedulingRulesIdleGpuTimeoutWorkloadType
} from '../../../../../../modules/ui/runaiV2/pages/OrgUnitsForm';
import { NodepoolCreateFields, NodepoolPhase } from '../../../../../../modules/api/apis/clusterApi';
import { DepartmentPageV2 } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPageV2';
import { Phase, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { Preemptibility } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { workloadsCrudService } from '../../../../../../modules/api';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`Project UI E2E Tests, tags: ${tags.tool.ui} ${tags.resource.projects} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPageV2;
        let departmentPage: DepartmentPageV2;

        let projectName: string;
        let projectDescription: string;
        let videoName: string;
        let gpuWorkerNodeNames: string[];
        let totalGpuNodeCnt: number;
        let nodeToLabel: string;
        let npId1: string;

        const createdNodePools: { id: string; name: string }[] = [];
        const nodeLabelKey = 'test';
        const nodeLabelVal1 = 'proj-test-group1';
        const projectOverDeptQuotaWarning = /The project.s quota exceeds the available department quota/;
        const projectQuotaLtNonPreempUsageError =
            /The project.s quota is less than the usage by non-preemptive workloads/;
        const npName1 = objectGenerator.generateShortNameByPrefix('np1', 5);
        let deptIds: number[] = [];
        let projectIds: string[] = [];

        test.beforeAll(async () => {
            browserEngine = new BrowserEngine();

            // get gpu node list
            gpuWorkerNodeNames = await nodeUtils.getGpuWorkerNodeNames();
            totalGpuNodeCnt = gpuWorkerNodeNames.length;

            expect(totalGpuNodeCnt).toBeGreaterThan(0);

            // label a gpu node for node pool
            nodeToLabel = gpuWorkerNodeNames[0];
            kubeCli.labelNode(nodeToLabel, nodeLabelKey, nodeLabelVal1);

            // create node pool 1 includes 1 gpu node
            const npReq1 = Object.assign(new NodepoolCreateFields(), {
                name: npName1,
                labelKey: nodeLabelKey,
                labelValue: nodeLabelVal1,
                clusterId: flowRootState.cluster.uuid
            });

            const resp = await nodePoolUtils.createClusterNodePoolWithRequest(npReq1).then((rsp) => {
                createdNodePools.push({ id: rsp.body.id, name: npName1 });
                npId1 = rsp.body.id;
                return rsp;
            });

            await nodePoolUtils.waitForClusterNodePoolPhase(npId1, [NodepoolPhase.Ready], 60);
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            projectPage = new ProjectPageV2(browserEngine.currentPage);
            departmentPage = new DepartmentPageV2(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );

            projectName = objectGenerator.generateShortNameByPrefix('proj-ui', 5);
            projectDescription = objectGenerator.generateShortNameByPrefix('desc-ui', 5);

            deptIds = [];
            projectIds = [];
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            await apiUtils.disableCpuQuotaFeature();
            await apiUtils.disableLimitOverQuotaSubscription();
            await apiUtils.disableProjectOverQuotaFeature();

            // delete projects
            await Promise.all(
                projectIds.map(async (projectId) => {
                    try {
                        await operations.deleteProjectAndNamespace(Number(projectId));
                    } catch (e) {
                        logger.error(`Failed to delete project ${projectId}`, e);
                    }
                })
            );

            // delete departments
            if (deptIds.length > 0) {
                await Promise.all(
                    deptIds.map(async (deptId) => {
                        try {
                            await departmentUtils.deleteDepartment(deptId);
                        } catch (e) {
                            logger.error(`Failed to delete department ${deptId}`, e);
                        }
                    })
                );
            }
        });

        test.afterAll('clean up', async () => {
            // Unlabel nodes
            try {
                if (gpuWorkerNodeNames && gpuWorkerNodeNames.length > 0) {
                    kubeCli.unLabelNode(nodeToLabel, nodeLabelKey);
                }
            } catch (e) {
                logger.error(`Failed to unlabel node ${nodeToLabel}`, e);
            }

            // Delete node pools
            await Promise.all(
                createdNodePools.map(async (np) => {
                    try {
                        await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                    } catch (e) {
                        logger.error(`Failed to delete node pool ${np.id}`, e);
                    }
                })
            );
        });

        test.describe('Project CRUD tests', () => {
            test.beforeEach(() => {
                projectIds = [];
            });

            test.afterEach('delete project', async () => {
                if (projectIds.length > 0) {
                    await Promise.all(
                        projectIds.map(async (projectId) => {
                            try {
                                await orgUnitUtils.deleteProject(projectId);
                            } catch (e) {
                                logger.error(`Failed to delete project ${projectId}`, e);
                            }
                        })
                    );
                }
            });

            test('T5689776: test create a project by "Create Department & Continue" button', async () => {
                await apiUtils.enableCpuQuotaFeature();

                videoName = 'create-project-continue-button';
                const gpuQuota = 1;
                const maxGpuAlloc = 2;
                const cpuQuota = 2;
                const cpuMemory = 4096;
                const userName = objectGenerator.generateShortNameByPrefix('usr', 5) + '@example.com';

                await test.step('Go to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('Create project using the create & continue workflow', async () => {
                    await projectPage.createProject(
                        projectName,
                        projectDescription,
                        [{ default: gpuQuota }],
                        [{ default: { allowExceed: true, maxGpuAlloc: maxGpuAlloc } }],
                        [{ default: { cpus: cpuQuota, cpuMemory: cpuMemory } }],
                        {
                            idleGpuTimeout: {
                                workloadType: SchedulingRulesIdleGpuTimeoutWorkloadType.Training,
                                days: 1,
                                hours: 10,
                                minutes: 10
                            }
                        },
                        {
                            subjectName: AccessRulesSubjectType.User,
                            userName: userName,
                            roleName: PROPERTY_TO_ROLE_MAPPING.system_administrator
                        }
                    );
                });

                await test.step('Verify the project is created successfully and verify it is shown at the top row', async () => {
                    await projectPage.waitForProjectToBeReady(projectName);
                    const projects = await projectPage.getProjectsFromGrid();

                    // Verify the project exists in the top row of the grid
                    expect(projects.length).toBeGreaterThan(0);
                    expect(projects[0]['Project']).toBe(projectName);
                });

                await test.step('Check the "projects" and "queues" CRD specs and verify the settings', async () => {
                    // Get project ID first for later use
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());

                    // Validate Projects CRD
                    const projectCrd = await kubeCli.getCrd(crds.runaiProjectCrd, projectName);
                    logger.info(`Project CRD spec: ${JSON.stringify(projectCrd['spec'])}`);

                    expect(projectCrd).toBeDefined();
                    expect(projectCrd['metadata']['name']).toBe(projectName);
                    expect(projectCrd['status']['phase']).toBe('Ready');

                    // Validate Queues CRD
                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: maxGpuAlloc },
                        cpu: { quota: cpuQuota * 1000, overQuotaWeight: cpuQuota * 1000, limit: -1 },
                        memory: { quota: cpuMemory, overQuotaWeight: cpuMemory, limit: -1 }
                    });
                });
            });

            test('T5689777: test create a project by "Create Project & Close" button', async () => {
                videoName = 'create-project-create-close-button';

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('Create project using the create project and close button shortcut', async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.fillProjectNameInput(projectName);
                    await projectPage.orgUnitsForm.clickSecondaryBtn();
                    await projectPage.orgUnitsForm.clickModalCreateAndCloseBtn();
                });

                await test.step('Wait for project to be ready', async () => {
                    await projectPage.waitForProjectToBeReady(projectName);
                });

                await test.step('validate the gpu quota', async () => {
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());

                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: 0, overQuotaWeight: 0.01, limit: -1 }
                    });
                });
            });

            test('T5689778: test create a project by "Create Project & Continue Editing" button', async () => {
                videoName = 'create-project-continue-editing-button';

                const gpuQuota = 1;

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('Click "New Project" button and fill name and description', async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.fillProjectNameInput(projectName);
                    await projectPage.fillProjectDescriptionInput(projectDescription);
                });

                await test.step('Click "CREATE PROJECT & CLOSE" button', async () => {
                    await projectPage.orgUnitsForm.clickSecondaryBtn();
                });

                await test.step('Click "CONTINUE EDITING" button in the popup dialog', async () => {
                    await projectPage.orgUnitsForm.clickModalContinueEditingBtn();
                });

                await test.step('Click "Create Project & Continue" button', async () => {
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                });

                await test.step('Fill GPU quota settings', async () => {
                    await projectPage.orgUnitsForm.configGpuQuota([{ default: gpuQuota }]);
                    await projectPage.orgUnitsForm.getSaveAndContinueBtn().click();
                });

                await test.step('Fill GPU over quota settings', async () => {
                    await projectPage.orgUnitsForm.configGpuOverQuota([{ default: { allowExceed: true } }]);
                    await projectPage.orgUnitsForm.getSaveAndContinueBtn().click();
                });

                await test.step('Close the form', async () => {
                    await projectPage.orgUnitsForm.getCloseBtn().click();
                });

                await test.step('Verify the project is created successfully', async () => {
                    await projectPage.waitForProjectToBeReady(projectName);
                });

                await test.step('Validate the config in the queues CRD', async () => {
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());

                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: -1 }
                    });
                });
            });

            test('T5689779: test free a project gpu quota during a 2nd project creation', async () => {
                videoName = 'free-project-quota-during-2nd-project-creation';
                const project2Name = `${projectName}-2`;
                const project2Description = `${projectDescription}-2`;
                const gpuQuota = 4;
                const newGpuQuota = 2;

                await test.step(`Create a project ${projectName}`, async () => {
                    const project = await orgUnitUtils.createProjectAndValidate({
                        name: projectName,
                        clusterId: flowRootState.cluster.uuid,
                        resources: [
                            {
                                gpu: { deserved: gpuQuota, limit: -1, overQuotaWeight: gpuQuota }
                            }
                        ]
                    });
                    projectIds.push(project.id);
                });

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step(`Free ${projectName} gpu quota during a 2nd project creation`, async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.fillProjectNameInput(project2Name);
                    await projectPage.fillProjectDescriptionInput(project2Description);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.waitForNotificationDisappear();
                    await projectPage.orgUnitsForm.updateOtherOrgUnitGpuQuota([
                        { default: { orgUnitName: projectName, newGpuQuota: newGpuQuota } }
                    ]);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step(`Validate ${projectName} gpu quota is updated from ${gpuQuota} to ${newGpuQuota}`, async () => {
                    await projectPage.waitForProjectToBeReady(project2Name);
                    const project2Id = await projectUtils.getProjectIdByProjectName(project2Name);
                    projectIds.push(project2Id.toString());

                    const projectData = await projectPage.getProjectByNameFromGrid(projectName);

                    expect(projectData).toBeDefined();
                    expect(Number(projectData['GPU quota'])).toBe(newGpuQuota);
                });
            });

            test('T5689780: test edit project', async () => {
                await apiUtils.enableCpuQuotaFeature();

                videoName = 'edit-project';
                const gpuQuota = 2;
                const newDescription = objectGenerator.generateShortNameByPrefix('desc', 3);
                const newGpuQuota = 1;
                const maxGpuAlloc = 2;
                const newCpuQuota = 6;
                const newCpuMemory = 8096;
                const userName = objectGenerator.generateShortNameByPrefix('usr-', 5) + '@example.com';
                const newSchedulingRules = {
                    workloadType: SchedulingRulesIdleGpuTimeoutWorkloadType.Training,
                    days: 1,
                    hours: 10,
                    minutes: 10
                };
                const newAccessRules = {
                    subjectName: AccessRulesSubjectType.User,
                    userName: userName,
                    roleName: PROPERTY_TO_ROLE_MAPPING.system_administrator
                };

                await test.step(`Create a project ${projectName}`, async () => {
                    const project = await orgUnitUtils.createProjectAndValidate({
                        name: projectName,
                        clusterId: flowRootState.cluster.uuid,
                        resources: [
                            {
                                gpu: { deserved: gpuQuota, limit: -1 }
                            }
                        ]
                    });
                    projectIds.push(project.id);
                });

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('Test edit description', async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.Description);
                    await projectPage.fillProjectDescriptionInput(newDescription);
                    await projectPage.orgUnitsForm.clickSecondaryBtn();
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.Description);

                    const desc = await projectPage.page
                        .locator(projectPage.projectDescriptionInput)
                        .inputValue();
                    expect(desc).toBe(newDescription);
                    await projectPage.orgUnitsForm.clickSecondaryBtn();
                });

                await test.step('Test edit GPU quota', async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuQuota);
                    await projectPage.orgUnitsForm.configGpuQuota([{ default: newGpuQuota }]);
                    await projectPage.orgUnitsForm.clickCloseFormBtn();

                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: newGpuQuota, overQuotaWeight: newGpuQuota, limit: -1 }
                    });
                });

                await test.step('Test edit GPU over quota', async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuOverQuota);
                    await projectPage.orgUnitsForm.configGpuOverQuota([
                        { default: { allowExceed: true, maxGpuAlloc: maxGpuAlloc } }
                    ]);
                    await projectPage.orgUnitsForm.clickCloseFormBtn();

                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: newGpuQuota, overQuotaWeight: newGpuQuota, limit: maxGpuAlloc }
                    });
                });

                await test.step('Test edit CPU quota', async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.CpuQuota);
                    await projectPage.orgUnitsForm.configCpuQuota([
                        { default: { cpus: newCpuQuota, cpuMemory: newCpuMemory } }
                    ]);
                    await projectPage.orgUnitsForm.clickCloseFormBtn();

                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: newGpuQuota, overQuotaWeight: newGpuQuota, limit: maxGpuAlloc },
                        cpu: { quota: newCpuQuota * 1000, overQuotaWeight: newCpuQuota * 1000, limit: -1 },
                        memory: { quota: newCpuMemory, overQuotaWeight: newCpuMemory, limit: -1 }
                    });
                });

                await test.step('Test edit Scheduling rules', async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.SchedulingRules);
                    await projectPage.orgUnitsForm.setSchedulingRulesIdleGpuTimeout(
                        newSchedulingRules.workloadType,
                        newSchedulingRules.days,
                        newSchedulingRules.hours,
                        newSchedulingRules.minutes
                    );
                    await projectPage.orgUnitsForm.clickCloseFormBtn();

                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.SchedulingRules);

                    const workloadType =
                        await projectPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutWorkloadType();
                    const days = await projectPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutDays();
                    const hours = await projectPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutHours();
                    const minutes = await projectPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutMinutes();
                    await projectPage.orgUnitsForm.clickCloseFormBtn();

                    expect(workloadType).toBe(newSchedulingRules.workloadType);
                    expect(days).toBe(newSchedulingRules.days);
                    expect(hours).toBe(newSchedulingRules.hours);
                    expect(minutes).toBe(newSchedulingRules.minutes);
                });

                await test.step('Test edit Access rules', async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.AccessRules);
                    await projectPage.orgUnitsForm.setAccessRules(
                        newAccessRules.subjectName,
                        newAccessRules.userName,
                        newAccessRules.roleName
                    );
                    await projectPage.orgUnitsForm.clickCloseFormBtn();

                    const projectData = await projectPage.getProjectByNameFromGrid(projectName);
                    expect(projectData['Subject(s)']).toBeDefined();

                    const subject = await projectPage.getProjectSubjectInfo(projectName);
                    expect(subject[0]['Type']).toBe(newAccessRules.subjectName);
                    expect(subject[0]['Subject']).toBe(newAccessRules.userName);
                    expect(subject[0]['Role']).toBe(newAccessRules.roleName);
                });
            });
        });

        test.describe('Resource toggle tests', () => {
            test('T5689785: test the "Limit projects from exceeding department quota" toggle for node pool/project', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'can not create project when the toggle is enabled and dept has limited CPU quota'
                );

                await apiUtils.enableLimitOverQuotaSubscription();
                await apiUtils.enableCpuQuotaFeature();

                videoName = 'test-limit-proj-exceeding-dept-quota-toggle-proj-np';
                const deptName = objectGenerator.generateShortNameByPrefix('dept', 5);
                const deptGpuQuota = 2;
                const deptCpuQuota = 2;
                const deptCpuMemory = 4096;
                const projectGpuQuota = 4;
                const projectCpuQuota = 4;
                const projectCpuMemory = 8192;

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Create a department', async () => {
                    await departmentPage.createDepartment(
                        deptName,
                        undefined,
                        [{ default: deptGpuQuota }],
                        undefined,
                        [{ default: { cpus: deptCpuQuota, cpuMemory: deptCpuMemory } }]
                    );

                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);
                });

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step(`Create a project under the department ${deptName} and verify its quota can not exceed department quotas`, async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.selectScope({
                        scope: Scope.Department,
                        scopeEntityName: deptName
                    });
                    await projectPage.fillProjectNameInput(projectName);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();

                    // verify it can not set the project gpu quota to exceed department quota
                    await projectPage.orgUnitsForm.configGpuQuota([{ default: projectGpuQuota }]);
                    await expect(
                        projectPage.page.locator(projectPage.orgUnitsForm.quotaMessageDiv)
                    ).toContainText(projectOverDeptQuotaWarning, { timeout: 3000 });
                    await projectPage.orgUnitsForm.clearGpuQuotaInput();

                    // verify it can not set the project cpu quota to exceed department quota
                    await projectPage.orgUnitsForm.navigateToCpuQuotaStep();
                    await projectPage.orgUnitsForm.fillCpusInput(projectCpuQuota);
                    await expect(
                        projectPage.page.locator(projectPage.orgUnitsForm.quotaMessageDiv)
                    ).toContainText(projectOverDeptQuotaWarning, { timeout: 3000 });
                    await projectPage.orgUnitsForm.clearCpusInput();

                    await projectPage.orgUnitsForm.fillCpuMemoryInput(projectCpuMemory);
                    await expect(
                        projectPage.page.locator(projectPage.orgUnitsForm.quotaMessageDiv)
                    ).toContainText(projectOverDeptQuotaWarning, { timeout: 3000 });
                    await projectPage.orgUnitsForm.clearCpuMemoryInput();
                });

                await test.step('Submit and the project should be created successfully', async () => {
                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                    await projectPage.waitForProjectToBeReady(projectName);
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());
                });
            });

            test('T5689788: test the "cpu quota" toggle for project', async () => {
                await apiUtils.enableCpuQuotaFeature();

                videoName = 'test-cpu-quota-project';

                const gpuQuota = 2;
                const cpuQuota = 4;
                const cpuMemory = 8192;

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('Create a project', async () => {
                    await projectPage.createProject(
                        projectName,
                        projectDescription,
                        [{ default: gpuQuota }],
                        undefined,
                        [{ default: { cpus: cpuQuota, cpuMemory: cpuMemory } }]
                    );
                });

                await test.step('Verify the project is created and cpu quota is correct', async () => {
                    await projectPage.waitForProjectToBeReady(projectName);
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());

                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        cpu: { quota: cpuQuota * 1000, overQuotaWeight: cpuQuota * 1000, limit: -1 },
                        memory: { quota: cpuMemory, overQuotaWeight: cpuMemory, limit: -1 }
                    });
                });

                await test.step('Disable cpu quota toggle', async () => {
                    await apiUtils.disableCpuQuotaFeature();
                });

                await test.step('Verify that cpu quota disappears', async () => {
                    await projectPage.goToPage();
                    await projectPage.toggleProject(projectName);
                    await projectPage.page.locator(projectPage.editProjectBtn).click();
                    await expect(async () => {
                        await projectPage.page
                            .getByRole('listitem')
                            .filter({ hasText: OrgUnitsFormName.CpuQuota })
                            .waitFor({ state: 'visible', timeout: 2000 });
                    }).rejects.toThrow(/timeout/i);
                });

                await test.step('Validate cpu and memory quota is reset to -1, overQuotaWeight to 1 for the project', async () => {
                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        cpu: { quota: -1, overQuotaWeight: 1, limit: -1 },
                        memory: { quota: -1, overQuotaWeight: 1, limit: -1 }
                    });
                });
            });

            test('T5689790: test the "over quota weight" toggle for project', async () => {
                videoName = 'test-over-quota-weight-project';
                const gpuQuota = 2;
                const overQuotaWeight = 4;

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('Create a project', async () => {
                    await projectPage.createProject(projectName, projectDescription, [{ default: gpuQuota }]);
                });

                await test.step('Wait the project to be created', async () => {
                    await projectPage.waitForProjectToBeReady(projectName);
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());
                });

                await test.step('Verify gpu over weight can not be edited', async () => {
                    await projectPage.goToPage();
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuOverQuota);
                    await projectPage.orgUnitsForm.switchToNodepoolTab('default');
                    // click the weight cell
                    await projectPage.page
                        .locator('.runai-table-row')
                        .filter({
                            has: projectPage.page.locator(projectPage.orgUnitsForm.gpuOverQuotaOrgNameTd, {
                                hasText: projectName
                            })
                        })
                        .locator(projectPage.orgUnitsForm.gpuOverQuotaGpuWeightTd)
                        .click();

                    // verify the pop edit form is not displayed (meaning weight cell is not editable)
                    await expect(async () => {
                        await projectPage.page
                            .locator('.popup-card')
                            .waitFor({ state: 'visible', timeout: 2000 });
                    }).rejects.toThrow(/timeout/i);

                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('enable the "over quota weight" toggle', async () => {
                    await apiUtils.enableProjectOverQuotaFeature();
                });

                await test.step('Edit project gpu over quota weight to a new value', async () => {
                    await projectPage.goToPage();
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuOverQuota);
                    await projectPage.orgUnitsForm.waitUntilGpuOverWeightCellClickable(
                        'default',
                        projectName
                    );
                    await projectPage.orgUnitsForm.updateOrgUnitGpuWeight([
                        { default: { orgUnitName: projectName, newGpuWeight: overQuotaWeight } }
                    ]);
                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('Validate gpu over quota weight is updated successfully', async () => {
                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: gpuQuota, overQuotaWeight: overQuotaWeight, limit: -1 }
                    });
                });
            });
        });

        test.describe('Allocated Non-Preemtible resource tests', () => {
            const deptName = objectGenerator.generateShortNameByPrefix('non-preemp-dept', 3);
            const projectName = objectGenerator.generateShortNameByPrefix('non-preemp-proj', 3);
            const trainingName = objectGenerator.generateShortNameByPrefix('non-preemp-training', 3);
            const deptGpuQuota = 2;
            const projGpuQuota = 1;
            const projMaxGpuLimit = 4;
            let deptId: string;
            let trainingId: string;
            const nonPreempInfoMsg = `Non-preemptive workload usage: ${projGpuQuota} GPU`;

            test.beforeAll('prepare the test setup', async () => {
                deptIds = [];
                projectIds = [];
                workloadsCrudService.trainingsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;

                // create a department with gpu quota = 2
                const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
                const request = orgUnitUtils.getDepartmentCreateRequestFromV1(
                    {
                        name: deptName,
                        deservedGpus: deptGpuQuota,
                        maxAllowedGpus: deptGpuQuota
                    },
                    flowRootState.cluster.uuid,
                    defaultNodePool
                );

                const resp = await orgUnitUtils.createDepartmentApi(request);
                expect(resp.response.statusCode).toBe(StatusCodes.CREATED);
                deptId = resp.body.id;
                deptIds.push(Number(deptId));

                // create a project under the department above with gpu quota = 1
                const projectResourceConfig = [
                    {
                        nodePool: {
                            name: 'default',
                            id: defaultNodePool.id.toString()
                        },
                        gpu: {
                            deserved: projGpuQuota,
                            overQuotaWeight: projGpuQuota
                        },
                        priority: 'MediumLow',
                        rank: 'MediumLow'
                    }
                ];

                const project = await orgUnitUtils.createProjectAndValidate({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    parentId: deptId,
                    resources: projectResourceConfig
                });
                projectIds.push(project.id);

                // create a non-preemptible training under the project with requested gpu = 1
                const { body: workload } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id,
                    {
                        image: 'alpine:latest',
                        command: 'sleep infinity',
                        compute: {
                            gpuDevicesRequest: projGpuQuota
                        },
                        preemptibility: Preemptibility.NonPreemptible
                    }
                );

                trainingId = workload.workloadId;

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(workload.workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workload.name} to be Running`,
                    true,
                    90
                );
            });

            test.afterAll('delete the workload', async () => {
                try {
                    await workloadsCrudUtils.deleteTraining(trainingId);
                } catch (error) {
                    logger.warn(`failed to delete training ${trainingName}`);
                }
            });

            test('T5689793: test non-preemptible messages displayed when editing project gpu quota form', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(test, testInfo, 'non-preemptible messages are not displayed');

                videoName = 'test-non-preemptible-workload-msg';

                await test.step('Navigate to the project page', async () => {
                    await projectPage.goToPage();
                });

                await test.step(`Open project ${projectName} GPU quota edit form`, async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuQuota);
                });

                await test.step('Verify non-preemptive workload usage tooltip is displayed', async () => {
                    const msg = await projectPage.orgUnitsForm.getQuotaMessage();
                    expect(msg).toContain(nonPreempInfoMsg);
                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('Enable the Limit projects from exceeding department quota toggle', async () => {
                    await apiUtils.enableLimitOverQuotaSubscription();
                });

                await test.step('Verify non-preemptive workload error tooltip is displayed when set GPU quota to 0', async () => {
                    await projectPage.goToPage();
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuQuota);
                    await projectPage.orgUnitsForm.fillGpuQuotaInput(0);
                    await projectPage.page.waitForTimeout(200);
                    const msg = await projectPage.orgUnitsForm.getQuotaMessage();
                    expect(msg).toMatch(projectQuotaLtNonPreempUsageError);
                    await projectPage.orgUnitsForm.fillGpuQuotaInput(projGpuQuota);
                    await projectPage.page.waitForTimeout(200);
                });

                await test.step('Update GPU over quota', async () => {
                    await projectPage.orgUnitsForm.navigateToGpuOverQuotaStep();
                    await projectPage.orgUnitsForm.fillGpuOverQuotaMaxGpuAllocationInput(projMaxGpuLimit);
                    await projectPage.orgUnitsForm.clickSaveBtn();
                    await projectPage.page.waitForTimeout(500);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.page
                        .locator(projectPage.orgUnitsForm.schedulingRulesAddRuleBtn)
                        .waitFor({ state: 'visible', timeout: 2000 });
                });

                await test.step('Go back to the GPU quota form and verify the non-preemptible tooltips again', async () => {
                    await projectPage.orgUnitsForm.navigateToGpuQuotaStep();
                    await projectPage.page.waitForTimeout(500);
                    await projectPage.orgUnitsForm.fillGpuQuotaInput(0);
                    await projectPage.page.waitForTimeout(200);
                    const msg = await projectPage.orgUnitsForm.getQuotaMessage();
                    expect(msg).toMatch(projectQuotaLtNonPreempUsageError);

                    await projectPage.orgUnitsForm.fillGpuQuotaInput(projGpuQuota);
                    await projectPage.page.waitForTimeout(200);
                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('Disbale the Limit projects from exceeding department quota toggle', async () => {
                    await apiUtils.disableLimitOverQuotaSubscription();
                });

                await test.step('Verify non-preemptive workload warning tooltip is displayed when set GPU quota to 0', async () => {
                    await projectPage.goToPage();
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuQuota);
                    await projectPage.orgUnitsForm.fillGpuQuotaInput(0);
                    await projectPage.page.waitForTimeout(200);
                    const msg = await projectPage.orgUnitsForm.getQuotaMessage();
                    expect(msg).toMatch(projectQuotaLtNonPreempUsageError);

                    await projectPage.orgUnitsForm.clickSaveBtn();
                    await projectPage.page.waitForTimeout(500);
                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                    await projectPage.page
                        .locator(projectPage.newProjectBtn)
                        .waitFor({ state: 'visible', timeout: 3000 });
                });
            });
        });

        test.describe('Miscellaneous tests', () => {
            test.afterEach('restore project creation modal settings', async () => {
                await projectPage.page.evaluate(() => {
                    localStorage.removeItem('project-creation-warning-dialog-dont-show');
                });
            });

            test('T5712872: test do not show project creation warning dialog', async () => {
                videoName = 'test-dont-show-proj-creation-warning-dialog';
                const projectName2 = `${projectName}-2`;

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step(`Create project ${projectName} with create & close button and with do not show this again option ticked`, async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.fillProjectNameInput(projectName);
                    await projectPage.orgUnitsForm.clickSecondaryBtn();
                    await projectPage.orgUnitsForm.clickDontShowThisAgainCheckbox();
                    await projectPage.orgUnitsForm.clickModalCreateAndCloseBtn();
                });

                await test.step(`Verify that ${projectName} is created`, async () => {
                    await projectPage.waitForProjectToBeReady(projectName);
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());
                });

                await test.step(`Create project ${projectName2} with create & close button and creation warning dialog not shown`, async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.fillProjectNameInput(projectName2);
                    await projectPage.orgUnitsForm.clickSecondaryBtn();
                });

                await test.step(`Verify that ${projectName2} is created`, async () => {
                    await projectPage.waitForProjectToBeReady(projectName2);
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName2);
                    projectIds.push(projectId.toString());
                });
            });

            test('T5689792: project step navigation test', async () => {
                await apiUtils.enableCpuQuotaFeature();
                videoName = 'test-proj-step-navigation';

                const gpuQuota = 2;
                const maxGpuAlloc = 4;
                const cpuQuota = 4;
                const cpuMemory = 8192;
                const schedulingRulesIdleGpuTimeout = {
                    workloadType: SchedulingRulesIdleGpuTimeoutWorkloadType.Training,
                    days: 1,
                    hours: 10,
                    minutes: 10
                };
                const userName = objectGenerator.generateShortNameByPrefix('usr', 5) + '@example.com';

                await test.step('Navigate to the Projects page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('fill all the form fields for a project', async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.fillProjectNameInput(projectName);
                    await projectPage.fillProjectDescriptionInput(projectDescription);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.orgUnitsForm.configGpuQuota([{ default: gpuQuota }]);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.orgUnitsForm.configGpuOverQuota([
                        { default: { allowExceed: true, maxGpuAlloc: maxGpuAlloc } }
                    ]);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.orgUnitsForm.configCpuQuota([
                        { default: { cpus: cpuQuota, cpuMemory: cpuMemory } }
                    ]);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.orgUnitsForm.setSchedulingRulesIdleGpuTimeout(
                        schedulingRulesIdleGpuTimeout.workloadType,
                        schedulingRulesIdleGpuTimeout.days,
                        schedulingRulesIdleGpuTimeout.hours,
                        schedulingRulesIdleGpuTimeout.minutes
                    );
                    await projectPage.orgUnitsForm.clickSaveBtn();
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                    await projectPage.orgUnitsForm.setAccessRules(
                        AccessRulesSubjectType.User,
                        userName,
                        PROPERTY_TO_ROLE_MAPPING.system_administrator
                    );
                });

                await test.step('Navigate to the cluster & name step', async () => {
                    await projectPage.orgUnitsForm.navigateToClusterAndNameStep();
                    await expect(projectPage.page.locator(projectPage.projectNameInput)).toHaveValue(
                        projectName,
                        { timeout: 3000 }
                    );
                    await expect(projectPage.page.locator(projectPage.projectDescriptionInput)).toHaveValue(
                        projectDescription,
                        { timeout: 3000 }
                    );
                });

                await test.step('Navigate to the gpu quota step', async () => {
                    await projectPage.orgUnitsForm.navigateToGpuQuotaStep();
                    await projectPage.orgUnitsForm.switchToNodepoolTab('default');
                    await expect(projectPage.page.locator('.gpu-quota-card').getByLabel('GPUs')).toHaveValue(
                        gpuQuota.toString(),
                        { timeout: 3000 }
                    );
                });

                await test.step('Navigate to the gpu over quota step', async () => {
                    await projectPage.orgUnitsForm.navigateToGpuOverQuotaStep();
                    await projectPage.orgUnitsForm.switchToNodepoolTab('default');
                    await expect(projectPage.page.getByLabel('Max GPU allocation')).toHaveValue(
                        maxGpuAlloc.toString(),
                        { timeout: 3000 }
                    );
                });

                await test.step('Navigate to the cpu quota step', async () => {
                    await projectPage.orgUnitsForm.navigateToCpuQuotaStep();
                    await projectPage.orgUnitsForm.switchToNodepoolTab('default');
                    await expect(projectPage.page.getByLabel('CPUs (Cores)')).toHaveValue(
                        cpuQuota.toString(),
                        { timeout: 3000 }
                    );
                    await expect(projectPage.page.getByLabel('CPU memory')).toHaveValue(
                        cpuMemory.toString(),
                        { timeout: 3000 }
                    );
                });

                await test.step('Navigate to the scheduling rules step', async () => {
                    await projectPage.orgUnitsForm.navigateToSchedulingRulesStep();
                    await expect(
                        projectPage.page.locator(projectPage.orgUnitsForm.schedulingRulesWorkloadTypeSelect)
                    ).toHaveValue(schedulingRulesIdleGpuTimeout.workloadType, { timeout: 3000 });
                    await expect(projectPage.page.locator(projectPage.orgUnitsForm.daysInput)).toHaveValue(
                        schedulingRulesIdleGpuTimeout.days.toString(),
                        { timeout: 3000 }
                    );
                    await expect(projectPage.page.locator(projectPage.orgUnitsForm.hoursInput)).toHaveValue(
                        schedulingRulesIdleGpuTimeout.hours.toString(),
                        { timeout: 3000 }
                    );
                    await expect(projectPage.page.locator(projectPage.orgUnitsForm.minutesInput)).toHaveValue(
                        schedulingRulesIdleGpuTimeout.minutes.toString(),
                        { timeout: 3000 }
                    );
                });

                await test.step('Submit the form and verify it is created successfully', async () => {
                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                    await projectPage.waitForProjectToBeReady(projectName);
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());

                    await crdUtils.validateProjectQueueCrdResources(projectName, {
                        gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: maxGpuAlloc },
                        cpu: { quota: cpuQuota * 1000, overQuotaWeight: cpuQuota * 1000, limit: -1 },
                        memory: { quota: cpuMemory, overQuotaWeight: cpuMemory, limit: -1 }
                    });
                });
            });

            test('T5689795: quota info display validation for projects', async () => {
                videoName = 'test-project-quota-info';
                const deptName = objectGenerator.generateShortNameByPrefix('quota-info-dept', 3);
                const proj2Name = `${projectName}-2`;
                const deptGpuQuota = 4;
                const proj1GpuQuota = 1;

                await test.step('create a department', async () => {
                    await departmentPage.goToPage();
                    await departmentPage.createDepartment(deptName, 'test-desc', [
                        { [npName1]: deptGpuQuota }
                    ]);
                });

                await test.step('wait for department appear', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);
                });

                await test.step(`create project ${projectName} under dept ${deptName}`, async () => {
                    const proj1 = await orgUnitUtils.createProjectAndValidate({
                        name: projectName,
                        clusterId: flowRootState.cluster.uuid,
                        parentId: deptIds[0].toString(),
                        resources: [
                            {
                                nodePool: {
                                    name: npName1,
                                    id: npId1
                                },
                                gpu: { deserved: proj1GpuQuota, limit: proj1GpuQuota }
                            }
                        ]
                    });

                    projectIds.push(proj1.id);
                });

                await test.step('navigate to the project page', async () => {
                    await projectPage.goToPage();
                });

                await test.step(`create a 2nd project ${proj2Name} and validate the displayed quota info`, async () => {
                    await projectPage.openNewProjectForm();
                    await projectPage.selectScope({
                        scope: Scope.Department,
                        scopeEntityName: deptName
                    });
                    await projectPage.fillProjectNameInput(proj2Name);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();

                    const gpuInfo = await projectPage.orgUnitsForm.getGpuQuotaCardValue(npName1);
                    expect(gpuInfo.unassigned).toEqual(deptGpuQuota - proj1GpuQuota);
                    expect(gpuInfo.total).toEqual(deptGpuQuota);

                    const projNpDataList = await projectPage.orgUnitsForm.getOtherOrgUnitGpuQuotaData();
                    const otherProjNames = projNpDataList.map((item: any) => item['Project']);
                    expect(otherProjNames).toContain(projectName);

                    for (const projNpData of projNpDataList) {
                        if (projNpData['Project'] === projectName) {
                            expect(Number(projNpData['GPUs'])).toEqual(proj1GpuQuota);
                        } else {
                            expect(Number(projNpData['GPUs'])).toEqual(0);
                        }
                    }

                    await projectPage.orgUnitsForm.clickCloseFormBtn();
                    await projectPage.waitForProjectToBeReady(proj2Name);
                    const proj2Id = await projectUtils.getProjectIdByProjectName(proj2Name);
                    deptIds.push(proj2Id);
                });
            });

            test('T5743762: test the save and discard workflows in the "unsaved changes modal" for project', async () => {
                videoName = 'test-save-discard-workflows-for-project';
                const gpuQuota = 1;
                const maxGpuAlloc = 4;
                const newMaxGpuAlloc = 2;

                await test.step('navigate to the project page', async () => {
                    await projectPage.goToPage();
                });

                await test.step('open project creation form', async () => {
                    await projectPage.openNewProjectForm();
                });

                await test.step('fill project name and continue', async () => {
                    await projectPage.fillProjectNameInput(projectName);
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                });

                await test.step(`switch to the ${npName1} nodepool from default nodepool without saving and then save`, async () => {
                    // a new project has the 'default' node pool only, add the one to switch to
                    await projectPage.orgUnitsForm.addNodePool(npName1);
                    await projectPage.orgUnitsForm.switchToNodepoolTab('default');
                    await projectPage.orgUnitsForm.fillGpuQuotaInput(gpuQuota);
                    // switch nodepool without saving, then unsaved changes modal will appear
                    await projectPage.orgUnitsForm.switchToNodepoolTab(npName1);
                    await projectPage.orgUnitsForm.clickSaveChangesBtn();
                });

                await test.step('click continue to go to the GPU over quota form', async () => {
                    await projectPage.orgUnitsForm.clickPrimaryBtn();
                });

                await test.step(`set default nodepool max gpu allocation to ${maxGpuAlloc} and close and save changes`, async () => {
                    await projectPage.orgUnitsForm.switchToNodepoolTab('default');
                    await projectPage.orgUnitsForm.fillGpuOverQuotaMaxGpuAllocationInput(maxGpuAlloc);
                    await projectPage.orgUnitsForm.clickCloseFormBtn(true);
                });

                await test.step('verify project is created', async () => {
                    await projectPage.waitForProjectToBeReady(projectName);
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    projectIds.push(projectId.toString());
                });

                await test.step('verify node pools gpu quotas', async () => {
                    await crdUtils.validateProjectQueueCrdResources(
                        projectName,
                        {
                            gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: maxGpuAlloc }
                        },
                        'default'
                    );

                    await crdUtils.validateProjectQueueCrdResources(
                        projectName,
                        {
                            gpu: { quota: 0, overQuotaWeight: 0.01, limit: -1 }
                        },
                        npName1
                    );
                });

                await test.step('open project edit form', async () => {
                    await projectPage.openEditProjectForm(projectName, OrgUnitsFormName.GpuQuota);
                });

                await test.step('set default nodepool gpu quota to 0 and click the gpu over quota sidemenu and save', async () => {
                    await projectPage.orgUnitsForm.switchToNodepoolTab('default');
                    await projectPage.orgUnitsForm.fillGpuQuotaInput(0);
                    // click the sidemenu and the unsaved changes modal will appear
                    await projectPage.orgUnitsForm.navigateToGpuOverQuotaStep();
                    await projectPage.orgUnitsForm.clickSaveChangesBtn();
                });

                await test.step(`set default nodepool max gpu allocation to ${newMaxGpuAlloc} and close and discard`, async () => {
                    await projectPage.orgUnitsForm.fillGpuOverQuotaMaxGpuAllocationInput(newMaxGpuAlloc);
                    await projectPage.orgUnitsForm.clickCloseFormBtn(false);
                });

                await test.step('verify node pools gpu quotas', async () => {
                    await crdUtils.validateProjectQueueCrdResources(
                        projectName,
                        {
                            gpu: { quota: 0, overQuotaWeight: 0.01, limit: maxGpuAlloc }
                        },
                        'default'
                    );
                });
            });
        });
    });
});
