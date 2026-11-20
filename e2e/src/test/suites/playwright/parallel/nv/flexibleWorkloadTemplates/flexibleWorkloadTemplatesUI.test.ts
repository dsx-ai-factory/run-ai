/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    departmentUtils,
    tags,
    Utils,
    templateUtils,
    config,
    waitUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';
import { WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    DISTRIBUTED_TOOLS_LABELS,
    TrainingPage
} from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { GeneralSettings } from '../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';

test.describe(`Flexible Workload Template Tests, tags: ${tags.tool.ui} ${tags.resource.templateV2} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let templatePageV2: TemplatePageV2;
        let hybridForm: HybridForm;
        let trainingPage: TrainingPage;
        let workspacePage: WorkspacePage;
        let generalSettinsPage: GeneralSettings;
        let loginUIPage: LoginUi;
        let videoName: string;
        let departmentIds: Array<number>;
        let departmentNames: Array<string>;
        let projectNameList: Array<string>;
        let namespaceList: Array<string>;
        let projectIdList: Array<number>;
        let templateName: string;
        let workloadName: string;

        const alpineImage = 'alpine:latest';
        const sleep10s = 'sleep 10s';
        const sleep20s = 'sleep 20s';
        const cpuOnly = 'cpu-only';
        const oneGpu = 'one-gpu';

        const workloadConfig = {
            image: alpineImage,
            runtimeOpts: {
                cmdArgsOpts: {
                    cmd: sleep10s
                }
            },
            existingComputeResName: cpuOnly
        };

        const masterConfig = {
            image: config.quickStartImage.ubuntu,
            runtimeOpts: {
                cmdArgsOpts: {
                    cmd: sleep20s
                }
            },
            existingComputeResName: cpuOnly
        };

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            const [dept1, dept2] = await Promise.all([
                departmentUtils.createBackendDepartmentWithName(
                    objectGenerator.generateShortNameByPrefix('d1'),
                    2,
                    2
                ),
                departmentUtils.createBackendDepartmentWithName(
                    objectGenerator.generateShortNameByPrefix('d2'),
                    2,
                    2
                )
            ]);
            departmentIds = [Number(dept1.body.id), Number(dept2.body.id)];
            departmentNames = [dept1.body.name, dept2.body.name];

            projectNameList = Array.from({ length: 2 }, () =>
                objectGenerator.generateShortNameByPrefix('project-for-flex-workload-template')
            );
            namespaceList = projectNameList.map((name) => Utils.namespaceFromProjectName(name));

            const projectsToCreate = [
                { namespace: namespaceList[0], name: projectNameList[0], departmentId: departmentIds[0] },
                { namespace: namespaceList[1], name: projectNameList[1], departmentId: departmentIds[1] }
            ];
            const projResponse = await Promise.all(
                projectsToCreate.map((config) =>
                    projectResourcesUtils.createProject(config.namespace, {
                        name: config.name,
                        deservedGpus: 2,
                        maxAllowedGpus: 2,
                        ...(config.departmentId && { departmentId: config.departmentId })
                    })
                )
            );
            projectIdList = projResponse.map((proj) => proj.id);
        });

        test.afterAll('deleting project and namespace', async () => {
            // delete projects and namespaces
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
            await Promise.all(departmentIds.map((deptId) => departmentUtils.deleteDepartment(deptId)));
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            templatePageV2 = new TemplatePageV2(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            workspacePage = new WorkspacePage(browserEngine.currentPage);
            generalSettinsPage = new GeneralSettings(browserEngine.currentPage);
            loginUIPage = new LoginUi(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );

            templateName = objectGenerator.generateShortNameByPrefix('alpine-template', 5);
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test.describe('Test create, update, make a copy and delete workload templates', function () {
            let updatedTemplateName: string;
            let copiedTemplateName: string;
            let templateCreated: string[];
            let testTemplateType: WorkloadType;

            test.beforeEach(function () {
                templateCreated = [];
            });

            test.afterEach(async ({}, testInfo) => {
                await Promise.all(
                    templateCreated.map(async (name) => {
                        const templateId = await templateUtils.getIdByName(name);
                        await templateUtils.deleteWorkloadTemplateV2(templateId, testTemplateType);
                    })
                );
            });

            const dataSets = [
                {
                    testCase:
                        'T5193628: test create/edit/make a copy/delete standard training template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Training,
                    multiNodeFramework: undefined,
                    workersOnly: undefined,
                    workloadConfig: workloadConfig,
                    masterConfig: undefined,
                    testEnable: true
                },
                {
                    testCase: 'T5193629: test create/edit/make a copy/delete distributed MPI template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Distributed,
                    multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.MPI,
                    workersOnly: undefined,
                    workloadConfig: workloadConfig,
                    masterConfig: undefined,
                    testEnable: true
                },
                {
                    testCase: 'T5193630: test create/edit/make a copy/delete workspace template on UI',
                    workloadType: WorkloadType.Workspace,
                    templateType: WorkloadType.Workspace,
                    multiNodeFramework: undefined,
                    workersOnly: undefined,
                    workloadConfig: workloadConfig,
                    masterConfig: undefined,
                    testEnable: true
                },
                {
                    testCase:
                        'T5198376: test create/edit/make a copy/delete distributed PyTorch Workers & masters template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Distributed,
                    multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                    workersOnly: undefined,
                    workloadConfig: workloadConfig,
                    masterConfig: masterConfig,
                    testEnable: true
                },
                {
                    testCase:
                        'T5198377: test create/edit/make a copy/delete distributed PyTorch Workers only template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Distributed,
                    multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                    workersOnly: true,
                    workloadConfig: workloadConfig,
                    masterConfig: undefined,
                    testEnable: true
                },
                {
                    testCase:
                        'T5198378: test create/edit/make a copy/delete distributed Tensorflow Workers & masters template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Distributed,
                    multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                    workersOnly: undefined,
                    workloadConfig: workloadConfig,
                    masterConfig: masterConfig,
                    testEnable: true
                },
                {
                    testCase:
                        'T5198379: test create/edit/make a copy/delete distributed Tensorflow Workers only template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Distributed,
                    multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                    workersOnly: true,
                    workloadConfig: workloadConfig,
                    masterConfig: undefined,
                    testEnable: true
                },
                {
                    testCase:
                        'T5198380: test create/edit/make a copy/delete distributed XGBoost template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Distributed,
                    multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.XG_BOOST,
                    workersOnly: undefined,
                    workloadConfig: workloadConfig,
                    masterConfig: undefined,
                    testEnable: true
                },
                {
                    testCase: 'T5198381: test create/edit/make a copy/delete distributed JAX template on UI',
                    workloadType: WorkloadType.Training,
                    templateType: WorkloadType.Distributed,
                    multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.JAX,
                    workersOnly: undefined,
                    workloadConfig: workloadConfig,
                    masterConfig: undefined,
                    testEnable: true
                }
            ];

            dataSets.forEach((caseData) => {
                if (!caseData.testEnable) return;
                testTemplateType = caseData.templateType;

                test(`${caseData.testCase}`, async function () {
                    videoName = 'create-update-copy-delete-template';
                    updatedTemplateName = `updated-${templateName}`;
                    copiedTemplateName = `copied-${templateName}`;

                    await templatePageV2.goToPage();

                    await test.step(`create a ${caseData.workloadType} template`, async function () {
                        if (caseData.templateType === WorkloadType.Distributed) {
                            await templatePageV2.createDistributedWorkloadTemplate(
                                caseData.workloadType,
                                templateName,
                                {
                                    scope: Scope.Cluster,
                                    scopeEntityName: flowRootState.cluster.name
                                },
                                {
                                    multiNodeFramework: caseData.multiNodeFramework,
                                    workerConfigOpts: caseData.workloadConfig,
                                    workersOnly: caseData.workersOnly,
                                    masterConfigOpts: caseData.masterConfig
                                }
                            );
                        } else {
                            await templatePageV2.createSingleNodeWorkloadTemplate(
                                caseData.workloadType,
                                templateName,
                                {
                                    scope: Scope.Cluster,
                                    scopeEntityName: flowRootState.cluster.name
                                },
                                caseData.workloadConfig
                            );
                        }

                        await templatePageV2.waitForTemplateExist(templateName);
                    });

                    await test.step('verify the created template', async function () {
                        const data = await templatePageV2.getTemplateData(templateName);
                        logger.info(`template data: ${JSON.stringify(data)}`);

                        expect(data['Template']).toEqual(templateName);
                        expect(data['Scope']).toEqual(flowRootState.cluster.name);
                        expect(data['Image']).toEqual(alpineImage);
                        expect(parseInt(data['GPU compute request'])).toEqual(0);
                        expect(data['Workload type']).toEqual(caseData.workloadType);
                    });

                    await test.step('update the template', async function () {
                        await templatePageV2.toggleTemplate(templateName);
                        await templatePageV2.clickUpadteTemplateBtn();
                        await templatePageV2.updateName(updatedTemplateName);
                        await templatePageV2.setComputeResources(oneGpu);
                        await templatePageV2.clickSubmitBtn();
                        await templatePageV2.page.waitForLoadState('load');
                        await templatePageV2.page.waitForTimeout(1000);

                        if (
                            templatePageV2.masterSettingsNeededForDistributedTemplate(
                                caseData.multiNodeFramework,
                                caseData.workersOnly
                            )
                        ) {
                            await templatePageV2.setComputeResources(oneGpu);
                            await hybridForm.waitForDistributedMasterSubmitButton();
                            await hybridForm.clickOnSubmitMasterForm();
                        }

                        if (caseData.templateType === WorkloadType.Distributed) {
                            await Utils.retry(async () => {
                                await templatePageV2.clickSubmitBtn();
                                await expect(
                                    templatePageV2.locator(templatePageV2.newTemplateBtn)
                                ).toBeVisible({ timeout: 3000 });
                            });
                        }

                        await templatePageV2.waitForTemplateExist(updatedTemplateName);
                        templateCreated.push(updatedTemplateName);
                    });

                    await test.step('verify updated template', async function () {
                        const data = await templatePageV2.getTemplateData(updatedTemplateName);
                        logger.info(`template data: ${JSON.stringify(data)}`);

                        expect(data['Template']).toEqual(updatedTemplateName);
                        expect(data['Scope']).toEqual(flowRootState.cluster.name);
                        expect(data['Image']).toEqual(caseData.workloadConfig.image);

                        if (
                            templatePageV2.masterSettingsNeededForDistributedTemplate(
                                caseData.multiNodeFramework,
                                caseData.workersOnly
                            )
                        ) {
                            expect(parseInt(data['GPU compute request'])).toEqual(2);
                        } else {
                            expect(parseInt(data['GPU compute request'])).toEqual(1);
                        }
                        expect(data['Workload type']).toEqual(caseData.workloadType);
                    });

                    await test.step('make a copy of a template', async function () {
                        await templatePageV2.toggleTemplate(updatedTemplateName);
                        await templatePageV2.clickCopyTemplateBtn();
                        await templatePageV2.fillName(copiedTemplateName);
                        await templatePageV2.clickSubmitBtn();

                        if (
                            templatePageV2.masterSettingsNeededForDistributedTemplate(
                                caseData.multiNodeFramework,
                                caseData.workersOnly
                            )
                        ) {
                            await Utils.retry(async () => {
                                await hybridForm.waitForDistributedMasterSubmitButton();
                                await hybridForm.clickOnSubmitMasterForm();
                            });
                        }

                        if (caseData.templateType === WorkloadType.Distributed) {
                            await Utils.retry(async () => {
                                await templatePageV2.clickSubmitBtn();
                                await expect(
                                    templatePageV2.locator(templatePageV2.newTemplateBtn)
                                ).toBeVisible({ timeout: 3000 });
                            });
                        }

                        await templatePageV2.waitForTemplateExist(copiedTemplateName);
                        templateCreated.push(copiedTemplateName);
                    });

                    await test.step('verify a copied template', async function () {
                        const data = await templatePageV2.getTemplateData(copiedTemplateName);
                        logger.info(`template data: ${JSON.stringify(data)}`);

                        expect(data['Template']).toEqual(copiedTemplateName);
                        expect(data['Scope']).toEqual(flowRootState.cluster.name);
                        expect(data['Image']).toEqual(caseData.workloadConfig.image);

                        if (
                            templatePageV2.masterSettingsNeededForDistributedTemplate(
                                caseData.multiNodeFramework,
                                caseData.workersOnly
                            )
                        ) {
                            expect(parseInt(data['GPU compute request'])).toEqual(2);
                        } else {
                            expect(parseInt(data['GPU compute request'])).toEqual(1);
                        }
                        expect(data['Workload type']).toEqual(caseData.workloadType);
                    });

                    await test.step('delete the templates', async function () {
                        await Promise.all(
                            templateCreated.map(async (name) => {
                                const templateId = await templateUtils.getIdByName(name);
                                await templateUtils.deleteWorkloadTemplateV2(
                                    templateId,
                                    caseData.templateType
                                );
                            })
                        );

                        templateCreated = [];
                    });
                });
            });
        });

        test.describe('Test using different scoped templates to create workload', function () {
            let templateExist = false;
            let templateType: WorkloadType;

            test.beforeEach(function () {
                workloadName = objectGenerator.generateShortNameByPrefix('workload', 5);
                templateType = undefined;
                templateExist = false;
            });

            test.afterEach(async ({}, testInfo) => {
                if (templateExist) {
                    const templateId = await templateUtils.getIdByName(templateName);
                    await templateUtils.deleteWorkloadTemplateV2(templateId, templateType);
                }
            });

            test('T5193631: create a project scoped training template and use it to create a training workload', async function () {
                videoName = 'test-project-scoped-training-template';
                templateType = WorkloadType.Training;

                await test.step('create a training template', async function () {
                    await templatePageV2.goToPage();

                    await templatePageV2.createSingleNodeWorkloadTemplate(
                        WorkloadType.Training,
                        templateName,
                        {
                            scope: Scope.Project,
                            scopeEntityName: projectNameList[0]
                        },
                        {
                            image: alpineImage,
                            runtimeOpts: {
                                cmdArgsOpts: {
                                    cmd: sleep10s
                                }
                            },
                            existingComputeResName: oneGpu
                        }
                    );

                    await templatePageV2.waitForTemplateExist(templateName);
                    templateExist = true;
                });

                await test.step('verify created template can not be seen in a different project', async function () {
                    await trainingPage.goToPage();
                    await trainingPage.openTrainingForm();
                    await trainingPage.workloadFormSelectProject(projectNameList[1]);
                    const templateExists = await trainingPage.templateV2Exists(templateName);
                    await trainingPage.closeWorkloadForm();

                    expect(templateExists).toEqual(false);
                });

                await test.step('create a training with the template', async function () {
                    await trainingPage.fillFlexTrainingFirstPage(
                        projectNameList[0],
                        workloadName,
                        false,
                        undefined,
                        false,
                        templateName
                    );

                    const image = await trainingPage.getEnvImage();
                    expect(image).toEqual(alpineImage);

                    const command = await trainingPage.getRuntimeCommand();
                    expect(command).toEqual(sleep10s);

                    const gpuDevice = await trainingPage.getGpuDeivce();
                    expect(gpuDevice).toEqual('1');

                    await trainingPage.clickOnSubmitBtn();
                    await trainingPage.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
                });

                await test.step('delete the workload', async function () {
                    await trainingPage.deleteWorkload(workloadName);
                });

                await test.step('delete the template', async function () {
                    const templateId = await templateUtils.getIdByName(templateName);
                    await templateUtils.deleteWorkloadTemplateV2(templateId, WorkloadType.Training);
                    templateExist = false;
                });
            });

            test('T5193632: create a department scoped workspace template and use it to create a workspace workload', async function () {
                videoName = 'test-depmt-scoped-workspace-template';
                templateType = WorkloadType.Workspace;

                await test.step('create a workspace template', async function () {
                    await templatePageV2.goToPage();

                    await templatePageV2.createSingleNodeWorkloadTemplate(
                        WorkloadType.Workspace,
                        templateName,
                        {
                            scope: Scope.Department,
                            scopeEntityName: departmentNames[0]
                        },
                        {
                            image: alpineImage,
                            runtimeOpts: {
                                cmdArgsOpts: {
                                    cmd: sleep10s
                                }
                            },
                            existingComputeResName: oneGpu
                        }
                    );

                    await templatePageV2.waitForTemplateExist(templateName);
                    templateExist = true;
                });

                await test.step('verify created template can not be seen in a different project', async function () {
                    await workspacePage.goToPage();
                    await workspacePage.openWorkspaceForm();
                    await workspacePage.workloadFormSelectProject(projectNameList[1]);
                    const templateExists = await trainingPage.templateV2Exists(templateName);
                    await workspacePage.closeWorkloadForm();
                    expect(templateExists).toEqual(false);
                });

                await test.step('create a workspace with the template', async function () {
                    await workspacePage.fillFlexWorkspaceFirstPage(
                        projectNameList[0],
                        workloadName,
                        templateName
                    );

                    const image = await workspacePage.getEnvImage();
                    expect(image).toEqual(alpineImage);

                    const command = await workspacePage.getRuntimeCommand();
                    expect(command).toEqual(sleep10s);

                    const gpuDevice = await workspacePage.getGpuDeivce();
                    expect(gpuDevice).toEqual('1');

                    await workspacePage.clickOnCreateWorkspaceBtn();
                    await workspacePage.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
                });

                await test.step('delete the workload', async function () {
                    await workspacePage.deleteWorkload(workloadName);
                });

                await test.step('delete the template', async function () {
                    const templateId = await templateUtils.getIdByName(templateName);
                    await templateUtils.deleteWorkloadTemplateV2(templateId, WorkloadType.Workspace);
                    templateExist = false;
                });
            });

            test('T5193633: create a cluster scoped distributed Tensorflow template (same setup for master & worker) and use it to create a distributed workload', async function () {
                videoName = 'test-dist-tf-same-workers-master-setup-template';
                templateType = WorkloadType.Distributed;

                await test.step('create a distributed Tensorflow template', async function () {
                    await templatePageV2.goToPage();

                    await templatePageV2.createDistributedWorkloadTemplate(
                        WorkloadType.Training,
                        templateName,
                        {
                            scope: Scope.Cluster,
                            scopeEntityName: flowRootState.cluster.name
                        },
                        {
                            multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                            workerConfigOpts: {
                                image: alpineImage,
                                runtimeOpts: {
                                    cmdArgsOpts: {
                                        cmd: sleep10s
                                    }
                                },
                                existingComputeResName: oneGpu
                            }
                        }
                    );

                    await templatePageV2.waitForTemplateExist(templateName);
                    templateExist = true;
                });

                await test.step('verify created template can be seen in different projects', async function () {
                    await trainingPage.goToPage();
                    await trainingPage.openTrainingForm();
                    await trainingPage.workloadFormSelectProject(projectNameList[1]);
                    await trainingPage.selectedMultiNodeOption();
                    const templateExists = await trainingPage.templateV2Exists(templateName);
                    await trainingPage.closeWorkloadForm();

                    expect(templateExists).toEqual(true);
                });

                await test.step('create a distribtued workload with the template', async function () {
                    await trainingPage.fillFlexTrainingFirstPage(
                        projectNameList[0],
                        workloadName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                        true,
                        templateName
                    );

                    const workerImage = await trainingPage.getEnvImage();
                    expect(workerImage).toEqual(alpineImage);

                    const workerCommand = await trainingPage.getRuntimeCommand();
                    expect(workerCommand).toEqual(sleep10s);

                    const workerGpuDevice = await trainingPage.getGpuDeivce();
                    expect(workerGpuDevice).toEqual('1');

                    await trainingPage.clickOnSubmitBtn();
                    await templatePageV2.page.waitForLoadState('load');
                    await templatePageV2.page.waitForTimeout(1000);

                    const masterImage = await trainingPage.getEnvImage();
                    expect(masterImage).toEqual(alpineImage);

                    const masterCommand = await trainingPage.getRuntimeCommand();
                    expect(masterCommand).toEqual(sleep10s);

                    const masterGpuDevice = await trainingPage.getGpuDeivce();
                    expect(masterGpuDevice).toEqual('1');

                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();

                    await Utils.retry(async () => {
                        await trainingPage.clickOnSubmitBtn();
                        await expect(trainingPage.locator(trainingPage.addNewWorkloadBtn)).toBeVisible({
                            timeout: 3000
                        });
                    });
                    await trainingPage.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
                });

                await test.step('delete the workload', async function () {
                    await trainingPage.deleteWorkload(workloadName);
                });

                await test.step('delete the template', async function () {
                    const templateId = await templateUtils.getIdByName(templateName);
                    await templateUtils.deleteWorkloadTemplateV2(templateId, templateType);
                    templateExist = false;
                });
            });

            test('T5193634: create a cluster scoped distributed Tensorflow template (diff setup for master & worker) and use it to create a distributed workload', async function () {
                videoName = 'test-dist-tf-diff-workers-master-setup-template';
                templateType = WorkloadType.Distributed;

                await test.step('create a distributed tensorflow template', async function () {
                    await templatePageV2.goToPage();

                    await templatePageV2.createDistributedWorkloadTemplate(
                        WorkloadType.Training,
                        templateName,
                        {
                            scope: Scope.Cluster,
                            scopeEntityName: flowRootState.cluster.name
                        },
                        {
                            multiNodeFramework: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                            workerConfigOpts: {
                                image: alpineImage,
                                runtimeOpts: {
                                    cmdArgsOpts: {
                                        cmd: sleep10s
                                    }
                                },
                                existingComputeResName: oneGpu
                            },
                            masterConfigOpts: {
                                image: config.quickStartImage.ubuntu,
                                runtimeOpts: {
                                    cmdArgsOpts: {
                                        cmd: sleep20s
                                    }
                                },
                                existingComputeResName: oneGpu
                            }
                        }
                    );

                    await templatePageV2.waitForTemplateExist(templateName);
                    templateExist = true;
                });

                await test.step('verify created template can be seen in different projects', async function () {
                    await trainingPage.goToPage();
                    await trainingPage.openTrainingForm();
                    await trainingPage.workloadFormSelectProject(projectNameList[1]);
                    await trainingPage.selectedMultiNodeOption();
                    const templateExists = await trainingPage.templateV2Exists(templateName);
                    await trainingPage.closeWorkloadForm();

                    expect(templateExists).toEqual(true);
                });

                await test.step('create a distributed workload with the template', async function () {
                    await trainingPage.fillFlexTrainingFirstPage(
                        projectNameList[0],
                        workloadName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                        true,
                        templateName
                    );

                    const workerImage = await trainingPage.getEnvImage();
                    expect(workerImage).toEqual(alpineImage);

                    const workerCommand = await trainingPage.getRuntimeCommand();
                    expect(workerCommand).toEqual(sleep10s);

                    const workerGpuDevice = await trainingPage.getGpuDeivce();
                    expect(workerGpuDevice).toEqual('1');

                    await trainingPage.clickOnSubmitBtn();
                    await templatePageV2.page.waitForLoadState('load');
                    await templatePageV2.page.waitForTimeout(1000);

                    const masterImage = await trainingPage.getEnvImage();
                    expect(masterImage).toEqual(config.quickStartImage.ubuntu);

                    const masterCommand = await trainingPage.getRuntimeCommand();
                    expect(masterCommand).toEqual(sleep20s);

                    const masterGpuDevice = await trainingPage.getGpuDeivce();
                    expect(masterGpuDevice).toEqual('1');

                    await Utils.retry(async () => {
                        await trainingPage.clickOnSubmitBtn();
                        await expect(trainingPage.locator(trainingPage.addNewWorkloadBtn)).toBeVisible({
                            timeout: 3000
                        });
                    });

                    await trainingPage.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
                });

                await test.step('delete the workload', async function () {
                    await trainingPage.deleteWorkload(workloadName);
                });

                await test.step('delete the template', async function () {
                    const templateId = await templateUtils.getIdByName(templateName);
                    await templateUtils.deleteWorkloadTemplateV2(templateId, templateType);
                    templateExist = false;
                });
            });

            test('T5193635: test flexible workload template feature toggle', async function () {
                videoName = 'test-flex-template-feature-toggle';

                await test.step('logout and login with system admin user', async function () {
                    await loginUIPage.logoutToLoginPage();
                    await authPage.signIn(
                        flowRootState.users.system_administrator.email,
                        flowRootState.users.system_administrator.password
                    );
                    await generalSettinsPage.goToPage();
                });

                await test.step('disable flex workload submission, then flex workload template will be disabled and disappear', async function () {
                    await generalSettinsPage.disableFlexWorkloadSubmission();
                    await waitUtils.waitForElementToNotExist(
                        generalSettinsPage,
                        generalSettinsPage.flexWorkloadTemplateToggle
                    );
                });

                await test.step('enable flex workload submission, then flex workload template will be shown with disabled status', async function () {
                    await generalSettinsPage.enableFlexWorkloadSubmission();
                    await waitUtils.waitForElementToExist(
                        generalSettinsPage,
                        generalSettinsPage.flexWorkloadTemplateToggle
                    );
                    const isEnabled = await generalSettinsPage.isFlexWorkloadTemplateEnabled();
                    expect(isEnabled).toEqual(false);
                });

                await test.step('enable flex workload template', async function () {
                    await generalSettinsPage.enableFlexWorkloadTemplate();
                    const isEnabled = await generalSettinsPage.isFlexWorkloadTemplateEnabled();
                    expect(isEnabled).toEqual(true);
                });
            });
        });
    });
});
