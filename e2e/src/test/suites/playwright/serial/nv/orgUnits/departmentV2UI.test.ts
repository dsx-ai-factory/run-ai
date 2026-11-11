/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    objectGenerator,
    departmentUtils,
    tags,
    apiUtils,
    crds,
    crdUtils,
    logger,
    nodeUtils,
    nodePoolUtils,
    orgUnitUtils,
    operations
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect, TestInfo } from '@playwright/test';
import { DepartmentPageV2 } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPageV2';
import { kubeCli } from '../../../../../../modules/cli';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';
import {
    AccessRulesSubjectType,
    OrgUnitsFormName,
    SchedulingRulesIdleGpuTimeoutWorkloadType
} from '../../../../../../modules/ui/runaiV2/pages/OrgUnitsForm';
import { NodepoolCreateFields, NodepoolPhase } from '../../../../../../modules/api/apis/clusterApi';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`Department UI E2E Tests, tags: ${tags.tool.ui} ${tags.resource.department} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let departmentPage: DepartmentPageV2;

        let gpuWorkerNodeNames: string[];
        let totalGpuNodeCnt: number;
        let nodeGpuCnt: number;
        let nodeToLabel: string;
        let npId1: string;
        let deptName: string;
        let deptDescription: string;
        let videoName: string;

        const createdNodePools: { id: string; name: string }[] = [];
        const nodeLabelKey = 'test';
        const nodeLabelVal1 = 'dept-test-group1';
        const deptOverNodePoolWarning = /The department.s quota exceeds the node pool.s total/;
        const deptQuotaLtProjQuotaMsg =
            /The department.s quota is less than the quota assigned to its subordinate projects/;
        const npName1 = objectGenerator.generateShortNameByPrefix('np1', 5);
        let deptIds: number[];
        let projectIds: string[];

        test.beforeAll(async () => {
            browserEngine = new BrowserEngine();

            // get gpu node list
            gpuWorkerNodeNames = await nodeUtils.getGpuWorkerNodeNames();
            totalGpuNodeCnt = gpuWorkerNodeNames.length;

            expect(totalGpuNodeCnt).toBeGreaterThan(0);

            // label a gpu node for node pool
            nodeToLabel = gpuWorkerNodeNames[0];
            kubeCli.labelNode(nodeToLabel, nodeLabelKey, nodeLabelVal1);
            nodeGpuCnt = await nodeUtils.getNodeGpuDeviceCnt(gpuWorkerNodeNames[0]);

            // create node pool 1 includes 1 gpu node
            const npReq1 = Object.assign(new NodepoolCreateFields(), {
                name: npName1,
                labelKey: nodeLabelKey,
                labelValue: nodeLabelVal1,
                clusterId: flowRootState.cluster.uuid
            });

            await nodePoolUtils.createClusterNodePoolWithRequest(npReq1).then((rsp) => {
                createdNodePools.push({ id: rsp.body.id, name: npName1 });
                npId1 = rsp.body.id;
                return rsp;
            });

            await nodePoolUtils.waitForClusterNodePoolPhase(npId1, [NodepoolPhase.Ready], 60);
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            departmentPage = new DepartmentPageV2(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );

            deptName = objectGenerator.generateShortNameByPrefix('dept-ui', 5);
            deptDescription = objectGenerator.generateShortNameByPrefix('desc-ui', 5);
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

        test.describe('Department CRUD tests', () => {
            test('T5689771: test create a department by "Create Department & Continue" button', async () => {
                await apiUtils.enableCpuQuotaFeature();

                videoName = 'create-dept-continue-button';
                const gpuQuota = 2;
                const maxGpuAlloc = 4;
                const cpuQuota = 4;
                const cpuMemory = 8192;
                const userName = objectGenerator.generateShortNameByPrefix('usr', 5) + '@example.com';

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Create department using the create & continue workflow', async () => {
                    await departmentPage.createDepartment(
                        deptName,
                        deptDescription,
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

                await test.step('Verify the created department is displayed at the top row of the grid', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const departments = await departmentPage.getDepartmentsFromGrid();

                    // verify the department exists in the top row of the grid
                    expect(departments.length).toBeGreaterThan(0);
                    expect(departments[0]['Department']).toBe(deptName);
                });

                await test.step('Validate the config in the queues CRD', async () => {
                    const deptCrd = await kubeCli.getCrd(crds.departmentCrd, deptName);
                    const deptId = Number(deptCrd['metadata']['labels']['runai/department-id']);
                    deptIds.push(deptId);

                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: maxGpuAlloc },
                        cpu: { quota: cpuQuota * 1000, overQuotaWeight: cpuQuota * 1000, limit: -1 },
                        memory: { quota: cpuMemory, overQuotaWeight: cpuMemory, limit: -1 }
                    });
                });
            });

            test('T5689772: test create a department by "Create Department & Close" button', async () => {
                videoName = 'create-dept-create-close-button';

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Create department using the create department and close button shortcut', async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.orgUnitsForm.clickSecondaryBtn();
                    await departmentPage.orgUnitsForm.clickModalCreateAndCloseBtn();
                });

                await test.step('Wait for department to appear', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                });

                await test.step('validate the gpu quota', async () => {
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);

                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        gpu: { quota: 0, overQuotaWeight: 0.01, limit: -1 }
                    });
                });
            });

            test('T5689773: test create a department by "Create Department & Continue Editing" button', async () => {
                videoName = 'create-dept-continue-editing-button';

                const gpuQuota = 1;

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Click "New Department" button and fill name and description', async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.fillDepartmentDescriptionInput(deptDescription);
                });

                await test.step('Click "CREATE DEPARTMENT & CLOSE" button', async () => {
                    await departmentPage.orgUnitsForm.clickSecondaryBtn();
                });

                await test.step('Click "CONTINUE EDITING" button in the popup dialog', async () => {
                    await departmentPage.orgUnitsForm.clickModalContinueEditingBtn();
                });

                await test.step('Click "Create Department & Continue" button', async () => {
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                });

                await test.step('Fill GPU quota settings', async () => {
                    await departmentPage.orgUnitsForm.configGpuQuota([{ default: gpuQuota }]);
                    await departmentPage.orgUnitsForm.getSaveAndContinueBtn().click();
                });

                await test.step('Fill GPU over quota settings', async () => {
                    await departmentPage.orgUnitsForm.configGpuOverQuota([
                        { default: { allowExceed: true } }
                    ]);
                    await departmentPage.orgUnitsForm.getSaveAndContinueBtn().click();
                });

                await test.step('Close the form', async () => {
                    await departmentPage.orgUnitsForm.getCloseBtn().click();
                });

                await test.step('Verify the department is created successfully', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                });

                await test.step('Validate the config in the queues CRD', async () => {
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);

                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: -1 }
                    });
                });
            });

            test('T5689774: test free a department gpu quota during a 2nd department creation', async () => {
                videoName = 'free-dept-quota-during-2nd-dept-creation';
                const dept2Name = `${deptName}-2`;
                const dept2Description = `${deptDescription}-2`;
                const gpuQuota = 8;
                const newGpuQuota = 2;

                await test.step(`Create a department ${deptName}`, async () => {
                    await departmentUtils
                        .createBackendDepartmentWithName(deptName, gpuQuota, gpuQuota)
                        .then((rsp) => {
                            deptIds.push(parseInt(rsp.body.id));
                        });
                });

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step(`Free ${deptName} gpu quota during a 2nd department creation`, async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(dept2Name);
                    await departmentPage.fillDepartmentDescriptionInput(dept2Description);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.updateOtherOrgUnitGpuQuota([
                        { default: { orgUnitName: deptName, newGpuQuota: newGpuQuota } }
                    ]);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step(`Validate ${dept2Name} gpu quota is updated from ${gpuQuota} to ${newGpuQuota}`, async () => {
                    await departmentPage.waitForDepartmentToAppear(dept2Name);
                    const dept2Id = await departmentUtils.getDepartmentId(dept2Name);
                    deptIds.push(dept2Id);

                    const deptData = await departmentPage.getDepartmentByNameFromGrid(deptName);

                    expect(deptData).toBeDefined();
                    expect(Number(deptData['GPU quota'])).toBe(newGpuQuota);
                });
            });

            test('T5689775: test edit department', async () => {
                await apiUtils.enableCpuQuotaFeature();

                videoName = 'edit-department';
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

                await test.step(`Create a department ${deptName}`, async () => {
                    await departmentUtils
                        .createBackendDepartmentWithName(deptName, gpuQuota, -1)
                        .then((rsp) => {
                            deptIds.push(parseInt(rsp.body.id));
                        });
                });

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Test edit description', async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.Description);
                    await departmentPage.fillDepartmentDescriptionInput(newDescription);
                    await departmentPage.orgUnitsForm.clickSecondaryBtn();
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.Description);

                    const desc = await departmentPage.page
                        .locator(departmentPage.departmentDescriptionInput)
                        .inputValue();
                    expect(desc).toBe(newDescription);
                    await departmentPage.orgUnitsForm.clickSecondaryBtn();
                });

                await test.step('Test edit GPU quota', async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.GpuQuota);
                    await departmentPage.orgUnitsForm.configGpuQuota([{ default: newGpuQuota }]);
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();

                    const deptId = deptIds[0];
                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        gpu: { quota: newGpuQuota, overQuotaWeight: newGpuQuota, limit: -1 }
                    });
                });

                await test.step('Test edit GPU over quota', async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.GpuOverQuota);
                    await departmentPage.orgUnitsForm.configGpuOverQuota([
                        { default: { allowExceed: true, maxGpuAlloc: maxGpuAlloc } }
                    ]);
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();

                    const deptId = deptIds[0];
                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        gpu: { quota: newGpuQuota, overQuotaWeight: newGpuQuota, limit: maxGpuAlloc }
                    });
                });

                await test.step('Test edit CPU quota', async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.CpuQuota);
                    await departmentPage.orgUnitsForm.configCpuQuota([
                        { default: { cpus: newCpuQuota, cpuMemory: newCpuMemory } }
                    ]);
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();

                    const deptId = deptIds[0];
                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        gpu: { quota: newGpuQuota, overQuotaWeight: newGpuQuota, limit: maxGpuAlloc },
                        cpu: { quota: newCpuQuota * 1000, overQuotaWeight: newCpuQuota * 1000, limit: -1 },
                        memory: { quota: newCpuMemory, overQuotaWeight: newCpuMemory, limit: -1 }
                    });
                });

                await test.step('Test edit Scheduling rules', async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.SchedulingRules);
                    await departmentPage.orgUnitsForm.setSchedulingRulesIdleGpuTimeout(
                        newSchedulingRules.workloadType,
                        newSchedulingRules.days,
                        newSchedulingRules.hours,
                        newSchedulingRules.minutes
                    );
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();

                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.SchedulingRules);

                    const workloadType =
                        await departmentPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutWorkloadType();
                    const days = await departmentPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutDays();
                    const hours = await departmentPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutHours();
                    const minutes =
                        await departmentPage.orgUnitsForm.getSchedulingRulesIdleGpuTimeoutMinutes();
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();

                    expect(workloadType).toBe(newSchedulingRules.workloadType);
                    expect(days).toBe(newSchedulingRules.days);
                    expect(hours).toBe(newSchedulingRules.hours);
                    expect(minutes).toBe(newSchedulingRules.minutes);
                });

                await test.step('Test edit Access rules', async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.AccessRules);
                    await departmentPage.orgUnitsForm.setAccessRules(
                        newAccessRules.subjectName,
                        newAccessRules.userName,
                        newAccessRules.roleName
                    );
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();

                    const departData = await departmentPage.getDepartmentByNameFromGrid(deptName);
                    expect(departData['Subject(s)']).toBeDefined();

                    const subject = await departmentPage.getDepartmentSubjectInfo(deptName);
                    expect(subject[0]['Type']).toBe(newAccessRules.subjectName);
                    expect(subject[0]['Subject']).toBe(newAccessRules.userName);
                    expect(subject[0]['Role']).toBe(newAccessRules.roleName);
                });
            });
        });

        test.describe('Resource toggle tests', () => {
            test('T5689786: test the "Limit projects from exceeding department quota" toggle for node pool/department', async () => {
                await apiUtils.enableLimitOverQuotaSubscription();

                videoName = 'test-limit-proj-exceeding-dept-quota-toggle-dept-np';
                const projectName = objectGenerator.generateShortNameByPrefix('proj', 5);
                const gpuQuota = 2;
                const newGpuQuota = 20;
                let deptId: number;

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step(`Create a department with gpu quota equal to ${gpuQuota}`, async () => {
                    await departmentPage.createDepartment(deptName, deptDescription, [
                        { [npName1]: gpuQuota }
                    ]);
                });

                await test.step('Verify department is created', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);
                });

                await test.step(`Create a project under department ${deptName} with gpu quota equal to 2`, async () => {
                    const project = await orgUnitUtils.createProjectAndValidate({
                        name: projectName,
                        clusterId: flowRootState.cluster.uuid,
                        parentId: deptId.toString(),
                        defaultNodePools: [npName1],
                        resources: [
                            {
                                nodePool: { name: npName1, id: npId1 },
                                gpu: { deserved: gpuQuota, limit: -1 }
                            }
                        ]
                    });
                    projectIds.push(project.id);
                });

                await test.step(`it should not allow to set the department ${deptName} gpu quota equal to 1`, async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.GpuQuota);
                    await departmentPage.orgUnitsForm.switchToNodepoolTab(npName1);
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(1);
                    const msg = await departmentPage.orgUnitsForm.getQuotaMessage();

                    expect(msg).toMatch(deptQuotaLtProjQuotaMsg);

                    // click the continue button, it should popup a modal saying changes not saved
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.page
                        .getByRole('button', { name: 'Save Changes' })
                        .waitFor({ state: 'visible', timeout: 3000 });
                    await departmentPage.clickBaseDialogCloseBtn();
                });

                await test.step(`it should allow to set the department ${deptName} gpu quota equal to 20 with warning`, async () => {
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(20);
                    await expect(
                        departmentPage.page.locator(departmentPage.orgUnitsForm.quotaMessageDiv)
                    ).toContainText(deptOverNodePoolWarning, { timeout: 3000 });

                    // click save and continue button, it should stay on the gpu quota form
                    await departmentPage.orgUnitsForm.clickSaveBtn();
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step(`Validate ${deptName} gpu quota is updated from ${gpuQuota} to ${newGpuQuota}`, async () => {
                    const deptData = await departmentPage.getDepartmentByNameFromGrid(deptName);

                    expect(deptData).toBeDefined();
                    expect(Number(deptData['GPU quota'])).toBe(newGpuQuota);
                });
            });

            test('T5689787: test the "cpu quota" toggle for department', async () => {
                await apiUtils.enableCpuQuotaFeature();

                videoName = 'test-cpu-quota-dept';

                const gpuQuota = 2;
                const cpuQuota = 4;
                const cpuMemory = 8192;

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Create a department', async () => {
                    await departmentPage.createDepartment(
                        deptName,
                        deptDescription,
                        [{ default: gpuQuota }],
                        undefined,
                        [{ default: { cpus: cpuQuota, cpuMemory: cpuMemory } }]
                    );
                });

                await test.step('Verify the department is created and cpu quota is correct', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);

                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        cpu: { quota: cpuQuota * 1000, overQuotaWeight: cpuQuota * 1000, limit: -1 },
                        memory: { quota: cpuMemory, overQuotaWeight: cpuMemory, limit: -1 }
                    });
                });

                await test.step('Disable cpu quota toggle', async () => {
                    await apiUtils.disableCpuQuotaFeature();
                });

                await test.step('Verify that cpu quota disappears', async () => {
                    await departmentPage.goToPage();
                    await departmentPage.toggleDepartment(deptName);
                    await departmentPage.page.locator(departmentPage.editDepartmentBtn).click();
                    await expect(async () => {
                        await departmentPage.page
                            .getByRole('listitem')
                            .filter({ hasText: OrgUnitsFormName.CpuQuota })
                            .waitFor({ state: 'visible', timeout: 2000 });
                    }).rejects.toThrow(/timeout/i);
                });

                await test.step('Validate cpu and memory quota is reset to -1, overQuotaWeight to 1 for the department', async () => {
                    await crdUtils.validateDepartmentQueueCrdResources(deptIds[0], {
                        cpu: { quota: -1, overQuotaWeight: 1, limit: -1 },
                        memory: { quota: -1, overQuotaWeight: 1, limit: -1 }
                    });
                });
            });

            test('T5689789: test the "over quota weight" toggle for department', async () => {
                videoName = 'test-over-quota-weight-dept';
                const gpuQuota = 2;
                const overQuotaWeight = 4;

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Create a department', async () => {
                    await departmentPage.createDepartment(deptName, deptDescription, [{ default: gpuQuota }]);
                });

                await test.step('Wait the department to be created', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);
                });

                await test.step('Verify gpu over weight can not be edited', async () => {
                    await departmentPage.goToPage();
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.GpuOverQuota);
                    await departmentPage.orgUnitsForm.switchToNodepoolTab('default');
                    // click the weight cell
                    await departmentPage.page
                        .locator('.runai-table-row')
                        .filter({
                            has: departmentPage.page.locator(
                                departmentPage.orgUnitsForm.gpuOverQuotaOrgNameTd,
                                { hasText: deptName }
                            )
                        })
                        .locator(departmentPage.orgUnitsForm.gpuOverQuotaGpuWeightTd)
                        .click();

                    // verify the pop edit form is not displayed (meaning weight cell is not editable)
                    await expect(async () => {
                        await departmentPage.page
                            .locator('.popup-card')
                            .waitFor({ state: 'visible', timeout: 2000 });
                    }).rejects.toThrow(/timeout/i);

                    await departmentPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('enable the "over quota weight" toggle', async () => {
                    await apiUtils.enableProjectOverQuotaFeature();
                });

                await test.step('Edit department gpu over quota weight to a new value', async () => {
                    await departmentPage.goToPage();
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.GpuOverQuota);
                    await departmentPage.orgUnitsForm.waitUntilGpuOverWeightCellClickable(
                        'default',
                        deptName
                    );
                    await departmentPage.orgUnitsForm.updateOrgUnitGpuWeight([
                        { default: { orgUnitName: deptName, newGpuWeight: overQuotaWeight } }
                    ]);
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('Validate gpu over quota weight is updated successfully', async () => {
                    await crdUtils.validateDepartmentQueueCrdResources(deptIds[0], {
                        gpu: { quota: gpuQuota, overQuotaWeight: overQuotaWeight, limit: -1 }
                    });
                });
            });
        });

        test.describe('Miscellaneous tests', () => {
            test.afterEach('restore department creation modal settings', async () => {
                await departmentPage.page.evaluate(() => {
                    localStorage.removeItem('department-creation-warning-dialog-dont-show');
                });
            });

            test('T5689781: gpu quota value validations on the gpu quota form', async () => {
                videoName = 'validate-gpu-quota-values';

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Open the default department gpu quota form', async () => {
                    await departmentPage.openEditDepartmentForm('default', OrgUnitsFormName.GpuQuota);
                });

                await test.step('Verify that -1 is invalid for gpu quota', async () => {
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(-1);
                    await expect(
                        departmentPage.page.locator(departmentPage.orgUnitsForm.quotaMessageDiv)
                    ).toContainText('Enter a value equal to or above 0', { timeout: 3000 });
                });

                await test.step('Verify that it shows warning when set gpu quota to 3000', async () => {
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(3000);
                    await expect(
                        departmentPage.page.locator(departmentPage.orgUnitsForm.quotaMessageDiv)
                    ).toContainText(deptOverNodePoolWarning, { timeout: 3000 });
                });
            });

            test('T5689782: create a department with gpu quota over total GPU device number', async () => {
                videoName = 'create-dept-with-gpu-quota-over-avail-gpus';
                const gpuQuota = 3000;

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step(`Create department with gpu quota ${gpuQuota}`, async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.switchToNodepoolTab(npName1);
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(3000);
                    await expect(
                        departmentPage.page.locator(departmentPage.orgUnitsForm.quotaMessageDiv)
                    ).toContainText(deptOverNodePoolWarning, { timeout: 3000 });
                    await departmentPage.orgUnitsForm.clickSaveBtn();
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('Wait for department to appear', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                });

                await test.step('validate the gpu quota', async () => {
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);

                    await crdUtils.validateDepartmentQueueCrdResources(
                        deptId,
                        {
                            gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: -1 }
                        },
                        npName1
                    );
                });
            });

            test('T5689783: create remove node pool during department creation', async () => {
                videoName = 'remove-np-during-dept-creation';
                const npToRemove = 'default';

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('Remove a node pool during department creation', async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.switchToNodepoolTab(npToRemove);
                    await departmentPage.orgUnitsForm.removeNodePool(npToRemove);
                    await departmentPage.orgUnitsForm.clickSaveBtn();
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();
                });

                await test.step('Verify that the default nodepool is not in the department', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);

                    const nodepools = await departmentPage.getDepartmentNodepoolInfo(deptName);
                    let defaultNp = undefined;
                    for (const nodepool of nodepools) {
                        if (nodepool['Node pool'] === npToRemove) {
                            defaultNp = nodepool;
                            break;
                        }
                    }

                    expect(defaultNp).toBeDefined();
                    expect(defaultNp['Order of priority']).toBe('empty');
                });
            });

            test('T5689784: test do not show department creation warning dialog', async () => {
                videoName = 'test-dont-show-dept-creation-warning-dialog';
                const deptName2 = `${deptName}-2`;

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step(`Create department ${deptName} with create & close button and with do not show this again option ticked`, async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.orgUnitsForm.clickSecondaryBtn();
                    await departmentPage.orgUnitsForm.clickDontShowThisAgainCheckbox();
                    await departmentPage.orgUnitsForm.clickModalCreateAndCloseBtn();
                });

                await test.step(`Verify that ${deptName} is created`, async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);
                });

                await test.step(`Create department ${deptName2} with create & close button and creation warning dialog not shown`, async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName2);
                    await departmentPage.orgUnitsForm.clickSecondaryBtn();
                });

                await test.step(`Verify that ${deptName2} is created`, async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName2);
                    const deptId = await departmentUtils.getDepartmentId(deptName2);
                    deptIds.push(deptId);
                });
            });

            test('T5689791: department step navigation test', async () => {
                await apiUtils.enableCpuQuotaFeature();
                videoName = 'test-dept-step-navigation';

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

                await test.step('Navigate to the Departments page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('fill all the form fields for a department', async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.fillDepartmentDescriptionInput(deptDescription);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.configGpuQuota([{ default: gpuQuota }]);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.configGpuOverQuota([
                        { default: { allowExceed: true, maxGpuAlloc: maxGpuAlloc } }
                    ]);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.configCpuQuota([
                        { default: { cpus: cpuQuota, cpuMemory: cpuMemory } }
                    ]);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.setSchedulingRulesIdleGpuTimeout(
                        schedulingRulesIdleGpuTimeout.workloadType,
                        schedulingRulesIdleGpuTimeout.days,
                        schedulingRulesIdleGpuTimeout.hours,
                        schedulingRulesIdleGpuTimeout.minutes
                    );
                    await departmentPage.orgUnitsForm.clickSaveBtn();
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.setAccessRules(
                        AccessRulesSubjectType.User,
                        userName,
                        PROPERTY_TO_ROLE_MAPPING.system_administrator
                    );
                });

                await test.step('Navigate to the cluster & name step', async () => {
                    await departmentPage.orgUnitsForm.navigateToClusterAndNameStep();
                    await expect(departmentPage.page.locator(departmentPage.departmentNameInput)).toHaveValue(
                        deptName,
                        { timeout: 3000 }
                    );
                    await expect(
                        departmentPage.page.locator(departmentPage.departmentDescriptionInput)
                    ).toHaveValue(deptDescription, { timeout: 3000 });
                });

                await test.step('Navigate to the gpu quota step', async () => {
                    await departmentPage.orgUnitsForm.navigateToGpuQuotaStep();
                    await departmentPage.orgUnitsForm.switchToNodepoolTab('default');
                    await expect(
                        departmentPage.page.locator('.gpu-quota-card').getByLabel('GPUs')
                    ).toHaveValue(gpuQuota.toString(), { timeout: 3000 });
                });

                await test.step('Navigate to the gpu over quota step', async () => {
                    await departmentPage.orgUnitsForm.navigateToGpuOverQuotaStep();
                    await departmentPage.orgUnitsForm.switchToNodepoolTab('default');
                    await expect(departmentPage.page.getByLabel('Max GPU allocation')).toHaveValue(
                        maxGpuAlloc.toString(),
                        { timeout: 3000 }
                    );
                });

                await test.step('Navigate to the cpu quota step', async () => {
                    await departmentPage.orgUnitsForm.navigateToCpuQuotaStep();
                    await departmentPage.orgUnitsForm.switchToNodepoolTab('default');
                    await expect(departmentPage.page.getByLabel('CPUs (Cores)')).toHaveValue(
                        cpuQuota.toString(),
                        { timeout: 3000 }
                    );
                    await expect(departmentPage.page.getByLabel('CPU memory')).toHaveValue(
                        cpuMemory.toString(),
                        { timeout: 3000 }
                    );
                });

                await test.step('Navigate to the scheduling rules step', async () => {
                    await departmentPage.orgUnitsForm.navigateToSchedulingRulesStep();
                    await expect(
                        departmentPage.page.locator(
                            departmentPage.orgUnitsForm.schedulingRulesWorkloadTypeSelect
                        )
                    ).toHaveValue(schedulingRulesIdleGpuTimeout.workloadType, { timeout: 3000 });
                    await expect(
                        departmentPage.page.locator(departmentPage.orgUnitsForm.daysInput)
                    ).toHaveValue(schedulingRulesIdleGpuTimeout.days.toString(), { timeout: 3000 });
                    await expect(
                        departmentPage.page.locator(departmentPage.orgUnitsForm.hoursInput)
                    ).toHaveValue(schedulingRulesIdleGpuTimeout.hours.toString(), { timeout: 3000 });
                    await expect(
                        departmentPage.page.locator(departmentPage.orgUnitsForm.minutesInput)
                    ).toHaveValue(schedulingRulesIdleGpuTimeout.minutes.toString(), { timeout: 3000 });
                });

                await test.step('Submit the form and verify it is created successfully', async () => {
                    await departmentPage.orgUnitsForm.clickCloseFormBtn();
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);

                    await crdUtils.validateDepartmentQueueCrdResources(deptId, {
                        gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: maxGpuAlloc },
                        cpu: { quota: cpuQuota * 1000, overQuotaWeight: cpuQuota * 1000, limit: -1 },
                        memory: { quota: cpuMemory, overQuotaWeight: cpuMemory, limit: -1 }
                    });
                });
            });

            test('T5689794: quota info display validation for departments', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(test, testInfo, 'Unassigned GPU quota value is wrong');

                videoName = 'test-dept-quota-info';

                await test.step('navigate to the department page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('create a department', async () => {
                    await departmentPage.openNewDepartmentForm();
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();

                    const gpuInfo = await departmentPage.orgUnitsForm.getGpuQuotaCardValue(npName1);
                    expect(gpuInfo.unassigned).toEqual(nodeGpuCnt);
                    expect(gpuInfo.total).toEqual(nodeGpuCnt);

                    const deptNpDataList = await departmentPage.orgUnitsForm.getOtherOrgUnitGpuQuotaData();
                    for (const deptNpData of deptNpDataList) {
                        expect(Number(deptNpData['GPUs'])).toEqual(0);
                    }
                });
            });

            test('T5743761: test the save and discard workflows in the "unsaved changes modal" for department', async () => {
                videoName = 'test-save-discard-workflows-for-dept';
                const gpuQuota = 1;
                const maxGpuAlloc = 4;

                await test.step('navigate to the department page', async () => {
                    await departmentPage.goToPage();
                });

                await test.step('open department creation form', async () => {
                    await departmentPage.openNewDepartmentForm();
                });

                await test.step('fill department name and continue', async () => {
                    await departmentPage.fillDepartmentNameInput(deptName);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                });

                await test.step(`switch to the default nodepool from ${npName1} nodepool without saving and discard settings`, async () => {
                    await departmentPage.orgUnitsForm.switchToNodepoolTab(npName1);
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(gpuQuota);
                    // switch nodepool without saving, then unsaved changes modal will appear
                    await departmentPage.orgUnitsForm.switchToNodepoolTab('default');
                    await departmentPage.orgUnitsForm.clickDiscardBtn();
                });

                await test.step('set default nodepool gpu quota and click continue without saving and then save changes', async () => {
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(gpuQuota);
                    await departmentPage.orgUnitsForm.clickPrimaryBtn();
                    await departmentPage.orgUnitsForm.clickSaveChangesBtn();
                });

                await test.step(`set default nodepool max gpu allocation to ${maxGpuAlloc} and close and discard settings`, async () => {
                    await departmentPage.orgUnitsForm.switchToNodepoolTab('default');
                    await departmentPage.orgUnitsForm.fillGpuOverQuotaMaxGpuAllocationInput(maxGpuAlloc);
                    await departmentPage.orgUnitsForm.clickCloseFormBtn(false);
                });

                await test.step('verify department is created', async () => {
                    await departmentPage.waitForDepartmentToAppear(deptName);
                    const deptId = await departmentUtils.getDepartmentId(deptName);
                    deptIds.push(deptId);
                });

                await test.step('verify node pools gpu quotas', async () => {
                    await crdUtils.validateDepartmentQueueCrdResources(
                        deptIds[0],
                        {
                            gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: -1 }
                        },
                        'default'
                    );

                    await crdUtils.validateDepartmentQueueCrdResources(
                        deptIds[0],
                        {
                            gpu: { quota: 0, overQuotaWeight: 0.01, limit: -1 }
                        },
                        npName1
                    );
                });

                await test.step('open department edit form', async () => {
                    await departmentPage.openEditDepartmentForm(deptName, OrgUnitsFormName.GpuQuota);
                });

                await test.step('set default nodepool gpu quota and click the gpu over quota sidemenu without saving', async () => {
                    await departmentPage.orgUnitsForm.switchToNodepoolTab('default');
                    await departmentPage.orgUnitsForm.fillGpuQuotaInput(0);
                    // click the sidemenu and the unsaved changes modal will appear
                    await departmentPage.orgUnitsForm.navigateToGpuOverQuotaStep();
                    await departmentPage.orgUnitsForm.clickDiscardBtn();
                });

                await test.step('set default nodepool max gpu allocation to 4 and close and save', async () => {
                    await departmentPage.orgUnitsForm.fillGpuOverQuotaMaxGpuAllocationInput(maxGpuAlloc);
                    await departmentPage.orgUnitsForm.clickCloseFormBtn(true);
                });

                await test.step('verify node pools gpu quotas', async () => {
                    await crdUtils.validateDepartmentQueueCrdResources(
                        deptIds[0],
                        {
                            gpu: { quota: gpuQuota, overQuotaWeight: gpuQuota, limit: maxGpuAlloc }
                        },
                        'default'
                    );
                });
            });
        });
    });
});
