/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    objectGenerator,
    logger,
    tags,
    projectResourcesUtils,
    operations,
    policyUtils,
    apiUtils,
    environmentUtils
} from '../../../../../../modules/helpers';
import { expect } from '@playwright/test';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { WorkloadType as PolicyTypeInCluster } from '../../../../../../modules/helpers/utils/policyUtils';
import { WorkloadType } from '../../../../../../modules/api/apis/assetsApi';
import { policyApi } from '../../../../../../modules/api';
import { ScopeType } from '../../../../../../modules/api/apis/backendApi';
import { TrainingPolicyDefaultsAndRulesV2 } from '../../../../../../modules/api/apis/policyApi';

const policyWithDefaultsAndRules: TrainingPolicyDefaultsAndRulesV2 = {
    defaults: {
        image: 'alpine',
        command: 'sleep 1000',
        compute: {
            gpuDevicesRequest: 1,
            cpuMemoryRequest: '400M'
        }
    },
    rules: {
        image: {
            options: [{ value: 'ubuntu' }, { value: 'alpine' }, { value: 'jupyter/scipy-notebook' }]
        },
        command: {
            options: [{ value: 'sleep infinite' }, { value: 'sleep 1000' }, { value: 'start-notebook.sh' }]
        },
        compute: {
            gpuDevicesRequest: {
                max: 1,
                min: 0
            }
        }
    }
};

const policyWithCanEditFalse: TrainingPolicyDefaultsAndRulesV2 = {
    defaults: {
        image: 'alpine',
        command: 'sleep 1000',
        compute: {
            gpuDevicesRequest: 1,
            cpuMemoryRequest: '400M'
        }
    },
    rules: {
        image: {
            options: [{ value: 'ubuntu' }, { value: 'alpine' }, { value: 'jupyter/scipy-notebook' }],
            canEdit: false
        },
        command: {
            options: [{ value: 'sleep infinite' }, { value: 'sleep 1000' }, { value: 'start-notebook.sh' }]
        },
        compute: {
            gpuDevicesRequest: {
                max: 1,
                min: 0
            }
        }
    }
};

const testParams = [
    {
        workloadType: WorkloadType.Training,
        kindInCluster: PolicyTypeInCluster.Training,
        policy: policyWithDefaultsAndRules
    },
    {
        workloadType: WorkloadType.Workspace,
        kindInCluster: PolicyTypeInCluster.Interactive,
        policy: policyWithDefaultsAndRules
    }
];
testParams.forEach((params) => {
    test.describe(
        `Hybrid ${params.workloadType} submission UI with policy, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.viper}`,
        {
            tag: [
                tags.services.policy,
                tags.services.workloads,
                tags.services.frontend,
                tags.runOn.parallelCI
            ]
        },
        () => {
            let project: Project;
            let videoName: string;
            let policyName: string;
            let workloadName: string;

            let browserEngine: BrowserEngine;
            let hybridForm: HybridForm;
            let auth: AuthPage;
            let workload: WorkloadPage;

            async function createProjectPolicy(
                policyName: string,
                policy: { defaults?: object; rules?: object }
            ): Promise<void> {
                const meta = { scope: ScopeType.Project, projectId: project.id, name: policyName };

                switch (params.workloadType) {
                    case WorkloadType.Training:
                        await policyApi.policyApi.overwriteTrainingPolicyV2(false, false, {
                            meta,
                            policy
                        });
                        break;
                    case WorkloadType.Workspace:
                        await policyApi.policyApi.overwriteWorkspacePolicyV2(false, false, {
                            meta,
                            policy
                        });
                        break;
                }
                await policyUtils.waitForPolicyToBeSynced(
                    params.kindInCluster,
                    project.namespace,
                    policyName
                );
            }

            const deletePolicy = async (): Promise<void> => {
                try {
                    switch (params.workloadType) {
                        case WorkloadType.Training:
                            await policyApi.policyApi.deleteTrainingPolicy(
                                'project',
                                undefined,
                                project.id.toString()
                            );
                            break;
                        case WorkloadType.Workspace:
                            await policyApi.policyApi.deleteWorkspacePolicy(
                                'project',
                                undefined,
                                project.id.toString()
                            );
                            break;
                        default:
                            throw 'Unsupported workload type';
                    }
                } catch (e) {
                    logger.error(`failed to delete policy, error: ${e}`);
                }
            };

            test.beforeAll('creating project and namespace', async () => {
                browserEngine = new BrowserEngine();
                try {
                    await apiUtils.enableHybridSubmission();
                    project = await projectResourcesUtils.createProjectAndNamespace('h-p');
                    logger.info(`created project ${project.name}`);
                    policyName = `project-${project.id}`;
                    policyApi.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
                    await createProjectPolicy(policyName, params.policy);
                } catch (e) {
                    logger.error(`failed to create project/policy, error: ${e}`);
                    throw e;
                }
            });

            test.afterAll('cleanup', async ({}, testInfo) => {
                try {
                    await deletePolicy();
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${project.name} an namespace ${project.namespace}`);
                } catch (error) {
                    logger.error(
                        `failed to delete project ${project.name} an namespace ${project.namespace}`
                    );
                    throw error;
                }
            });

            test.beforeEach('start browser', async () => {
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                hybridForm = new HybridForm(browserEngine.currentPage);
                workload = new WorkloadPage(browserEngine.currentPage);
                await auth.signIn(
                    flowRootState.users.l2_researcher.email,
                    flowRootState.users.l2_researcher.password
                );
                workloadName = objectGenerator.generateNameByPrefix('hybrid-w');
                await workload.navigateToWorkloads();
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                try {
                    await browserEngine.stopAndArchive(testInfo, 'hybrid-w' + videoName);
                } catch (e) {
                    logger.error(e);
                }
            });

            async function fillWorkloadCreationFirstPage(): Promise<void> {
                await hybridForm.openWorkloadForm(params.workloadType);
                await workload.selectClusterFromDropDown(flowRootState.cluster.name);
                await workload.workloadFormSelectProject(project.name);
                await workload.workloadFormSelectTemplate();
                await hybridForm.fillWorkloadName(params.workloadType, workloadName);
                await hybridForm.waitForWorkloadNameCount(workloadName);
                await hybridForm.clickOnContinue(params.workloadType);
            }

            test.describe(`${params.workloadType}: Policy - Positive Tests`, () => {
                test(`${params.workloadType}: check policy defaults on form`, async () => {
                    videoName = 'hybrid-policy-default-on-form';

                    await fillWorkloadCreationFirstPage();

                    // override the form defaults with an asset template
                    await hybridForm.openEnvironmentsDrawer();
                    await hybridForm.clickOnDrawerItemByLabel((await environmentUtils.presetNames()).jupyter);

                    await hybridForm.openComputesDrawer();
                    await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                    await hybridForm.clearEnvironmentsSection();
                    await hybridForm.clearComputesSection();

                    // verify and defaults from the policy are presented in the form after the asset's templates were cleared
                    const imageFieldValue = await hybridForm.getSelectValueBySelector(
                        hybridForm.hybridEnvironmentSection.hybridImageInput
                    );
                    expect(imageFieldValue).toEqual(params.policy.defaults.image);

                    const argsFieldValue = await hybridForm
                        .locator(hybridForm.hybridEnvironmentSection.argsInput)
                        .inputValue();
                    // Args field should be empty when not specified in policy defaults
                    expect(argsFieldValue).toEqual('');

                    const cmdFieldValue = await hybridForm.getSelectValueBySelector(
                        hybridForm.hybridEnvironmentSection.cmdInput
                    );
                    expect(cmdFieldValue).toEqual(params.policy.defaults.command);

                    const gpuDevicesRequestValue = await hybridForm
                        .locator(hybridForm.hybridComputeSection.gpuDevicesInput)
                        .inputValue();
                    expect(+gpuDevicesRequestValue).toEqual(params.policy.defaults.compute.gpuDevicesRequest);

                    const cpuMemoryRequestValue = await hybridForm
                        .locator(hybridForm.hybridComputeSection.cpuMemoryRequestInput)
                        .inputValue();
                    expect(cpuMemoryRequestValue).toEqual(
                        // splitting size string to numbers and letters
                        params.policy.defaults.compute.cpuMemoryRequest.match(/[a-zA-Z]+|[0-9]+/g)[0]
                    );

                    await hybridForm.clickOnSubmit();
                    await workload.waitForWorkloadUIStatus(workloadName, 'Running');
                    await workload.deleteWorkload(workloadName);
                });

                test(`${params.workloadType}: ensure assets with canEdit false are disabled in the drawer`, async () => {
                    videoName = 'hybrid-policy-can-edit-false-disabled';

                    try {
                        // Update policy to include canEdit: false for this specific test
                        await createProjectPolicy(policyName, policyWithCanEditFalse);

                        await fillWorkloadCreationFirstPage();

                        await hybridForm.openEnvironmentsDrawer();

                        const tensorboardEnvName = (await environmentUtils.presetNames()).tensorboard;
                        const tensorboardTemplate = hybridForm
                            .locator(hybridForm.hybridDrawer.drawerItem)
                            .filter({
                                has: hybridForm.page.locator(
                                    `${hybridForm.hybridDrawer.drawerItemLabel}:text-is("${tensorboardEnvName}")`
                                )
                            });
                        const envClasses = await tensorboardTemplate
                            .locator('.text-section')
                            .getAttribute('class');
                        expect(envClasses).toContain('disabled');

                        await expect(
                            tensorboardTemplate.locator(hybridForm.hybridDrawer.viewPolicyBtn)
                        ).toBeVisible();
                    } finally {
                        // Restore original policy for other tests
                        await createProjectPolicy(policyName, params.policy);
                    }
                });
            });
        }
    );
});
