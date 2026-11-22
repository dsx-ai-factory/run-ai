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
    tags,
    Utils,
    apiUtils,
    templateUtils,
    environmentUtils,
    computeResourceUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    ComputeAsset,
    EnvironmentAsset,
    EnvironmentCreationRequest,
    Scope
} from '../../../../../../modules/api/apis/assetsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';
import { WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';

test.describe(`Flexible Workload Template Assets Link Tests, tags: ${tags.tool.ui} ${tags.resource.templateV2} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let templatePageV2: TemplatePageV2;
        let hybridForm: HybridForm;
        let envAsset: EnvironmentAsset;
        let computeResource: ComputeAsset;
        let projectName: string;
        let environmentName: string;
        let computeResourceName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;
        let templateName: string;
        let templateCreated = [];
        let envDeleted = false;

        const tensorboardEnvName = 'tensorboard-tensorflow';
        const tensorboardUrl = 'tensorflow/tensorflow:2.20.0';
        const alpineUrl = 'alpine:latest';
        const ubuntuUrl = 'ubuntu:latest';
        const cmd = 'sleep infinity';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-templ-env-link', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            // create an env for propagation testing
            environmentName = objectGenerator.generateShortNameByPrefix('alpine-env', 5);
            const environmentCreationRequest: EnvironmentCreationRequest = {
                meta: {
                    name: environmentName,
                    scope: Scope.Project,
                    projectId: project.id,
                    workloadSupportedTypes: {
                        workspace: true,
                        training: true
                    }
                },
                spec: {
                    image: alpineUrl,
                    command: cmd
                }
            };
            envAsset = await environmentUtils.create(environmentCreationRequest);

            // create a compute resource
            computeResourceName = objectGenerator.generateShortNameByPrefix('cr-for-template', 5);
            const computeCreationRequest = {
                meta: {
                    name: computeResourceName,
                    scope: Scope.Project,
                    projectId: project.id,
                    clusterId: flowRootState.cluster.uuid
                },
                spec: {
                    cpuCoreRequest: 0.05,
                    gpuDevicesRequest: 0,
                    cpuMemoryRequest: '10M'
                }
            };
            computeResource = await computeResourceUtils.create(computeCreationRequest);
        });

        test.afterAll('deleting project and namespace', async () => {
            if (!envDeleted) {
                try {
                    await environmentUtils.delete(envAsset.meta.id);
                } catch (error) {
                    logger.error(`failed to delete env: ${environmentName}`);
                }
            }

            try {
                await computeResourceUtils.delete(computeResource.meta.id);
            } catch (error) {
                logger.error(`failed to delete compute resource: ${computeResourceName}`);
            }

            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`, error);
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();

            templateCreated = [];
            authPage = new AuthPage(browserEngine.currentPage);
            templatePageV2 = new TemplatePageV2(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );

            templateName = objectGenerator.generateShortNameByPrefix('flex-template-env-link', 5);
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await Promise.all(
                templateCreated.map(async (template) => {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(template.id, template.type);
                    } catch (error) {
                        logger.error(`failed to delete template with ID: ${template.id}`);
                    }
                })
            );

            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T5331903: link env with template and verify', async function () {
            videoName = 'link-env-with-templ';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('link env with template on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingEnv(tensorboardEnvName);
                await templatePageV2.toggleEnvLinkButton();
            });

            await test.step('verify image input is disabled', async () => {
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toBeDisabled();
            });

            await test.step('verify image input tooltip appears when being hovered on', async () => {
                await templatePageV2.hoverOnImageInput();
                await expect(templatePageV2.page.locator(templatePageV2.cantEditTooltip)).toBeVisible();
            });

            await test.step('verify asset reset button is disabled', async () => {
                await expect(templatePageV2.page.locator(templatePageV2.envAssetResetBtn)).toBeDisabled();
            });

            await test.step('verify asset reset button tooltip appears when being hovered on', async () => {
                await templatePageV2.hoverOnEnvAssetResetBtn();
                await expect(templatePageV2.page.locator(templatePageV2.assetResetBtnTooltip)).toBeVisible();
            });
        });

        test('T5331904: unlink env with template and verify', async function () {
            videoName = 'unlink-env-with-templ';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('link env with template on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingEnv(tensorboardEnvName);
                await templatePageV2.toggleEnvLinkButton();
            });

            await test.step('unlink env with template', async () => {
                await templatePageV2.toggleEnvLinkButton();
            });

            await test.step('verify image input is enabled', async () => {
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toBeEnabled();
            });

            await test.step('verify image input tooltip disappears when being hovered on', async () => {
                await templatePageV2.hoverOnImageInput();
                await expect(templatePageV2.page.locator(templatePageV2.cantEditTooltip)).toBeHidden();
            });

            await test.step('verify asset reset button is enabled', async () => {
                await expect(templatePageV2.page.locator(templatePageV2.envAssetResetBtn)).toBeEnabled();
            });
        });

        test('T5331905: reset env after unlinking env', async function () {
            videoName = 'reset_env_after_unlinking';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('select an env and leave it unlinked on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingEnv(tensorboardEnvName);
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toHaveValue(tensorboardUrl);
            });

            await test.step('modify image url info', async () => {
                await hybridForm.fillImageInput(alpineUrl);
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toHaveValue(alpineUrl);
            });

            await test.step('click the reset asset button and verify', async () => {
                await templatePageV2.clickEnvAssetResetBtn();
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toHaveValue(tensorboardUrl);
            });
        });

        test('T5331906: reset compute resource after unlinking it', async function () {
            videoName = 'reset_cr_after_unlnking';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('select cpu-only compute resource and link it on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingComputeResource('cpu-only');
                await templatePageV2.toggleComputeResourceLinkButton();
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeDisabled();
            });

            await test.step('unlink compute resource asset with the template', async () => {
                await templatePageV2.toggleComputeResourceLinkButton();
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeEnabled();
            });

            await test.step('modify the gpu device', async () => {
                await hybridForm.fillGpuDevicesInput('3');
                await expect(hybridForm.locator(hybridForm.hybridComputeSection.gpuDevicesInput)).toHaveValue(
                    '3'
                );
            });

            await test.step('click asset reset button and verify', async () => {
                await templatePageV2.clickComputeResourceAssetResetBtn();
                await expect(hybridForm.locator(hybridForm.hybridComputeSection.gpuDevicesInput)).toHaveValue(
                    '0'
                );
            });
        });

        test('T5331907: link compute resource with template', async function () {
            videoName = 'link-cr-with-templ';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('select cpu-only compute resource and link it on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingComputeResource('cpu-only');
                await templatePageV2.toggleComputeResourceLinkButton();
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeDisabled();
            });

            await test.step('verify gpu/cpu device fields are disabled', async () => {
                await templatePageV2.toggleFoldingSubSection('.cpu-resources-section', true);
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.gpuDevicesInput)
                ).toBeDisabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuCoresRequestInput)
                ).toBeDisabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuCoresLimitInput)
                ).toBeDisabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuMemoryRequestInput)
                ).toBeDisabled();
            });

            await test.step('verify tooltip is shown for the disabled GPU field', async () => {
                await templatePageV2.hoverOnGpuDeviceInput();
                await expect(templatePageV2.page.locator(templatePageV2.cantEditTooltip)).toBeVisible();
            });

            await test.step('verify asset reset button is disabled', async () => {
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeDisabled();
            });

            await test.step('verify asset reset button tooltip appears when being hovered on', async () => {
                await templatePageV2.hoverOnComputeResourceAssetResetBtn();
                await expect(templatePageV2.page.locator(templatePageV2.assetResetBtnTooltip)).toBeVisible();
            });
        });

        test('T5331908: unlink compute resource with template', async function () {
            videoName = 'unlink-cr-with-templ';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('select cpu-only compute resource and link it on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingComputeResource('cpu-only');
                await templatePageV2.toggleComputeResourceLinkButton();
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeDisabled();
            });

            await test.step('unlink the asset link and verify', async () => {
                await templatePageV2.toggleComputeResourceLinkButton();
                await templatePageV2.toggleFoldingSubSection('.cpu-resources-section', true);
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.gpuDevicesInput)
                ).toBeEnabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuCoresRequestInput)
                ).toBeEnabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuCoresLimitInput)
                ).toBeEnabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuMemoryRequestInput)
                ).toBeEnabled();
            });

            await test.step('verify tooltip is not shown for the disabled GPU field', async () => {
                await templatePageV2.hoverOnGpuDeviceInput();
                await expect(templatePageV2.page.locator(templatePageV2.cantEditTooltip)).toBeHidden();
            });

            await test.step('verify asset reset button is enabled', async () => {
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeEnabled();
            });
        });

        test('T5331909: link both env and compute resource with template', async function () {
            videoName = 'link-env-cr-with-templ';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('link env with template on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingEnv(tensorboardEnvName);
                await templatePageV2.toggleEnvLinkButton();
            });

            await test.step('select cpu-only compute resource and link it on the 2nd page', async () => {
                await templatePageV2.hybridForm.selectExistingComputeResource('cpu-only');
                await templatePageV2.toggleComputeResourceLinkButton();
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeDisabled();
            });

            await test.step('verify env and compute resource fields are disabled', async () => {
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toBeDisabled();
                await templatePageV2.toggleFoldingSubSection('.cpu-resources-section', true);
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.gpuDevicesInput)
                ).toBeDisabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuCoresRequestInput)
                ).toBeDisabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuCoresLimitInput)
                ).toBeDisabled();
                await expect(
                    hybridForm.locator(hybridForm.hybridComputeSection.cpuMemoryRequestInput)
                ).toBeDisabled();
            });

            await test.step('verify env and compute resource reset buttons are disabled', async () => {
                await expect(templatePageV2.page.locator(templatePageV2.envAssetResetBtn)).toBeDisabled();
                await expect(
                    templatePageV2.page.locator(templatePageV2.computeResourceAssetResetBtn)
                ).toBeDisabled();
            });
        });

        test('T5331910: verify updated env info is propagated to linked template', async function () {
            videoName = 'updated-env-propagated-to-template';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('link env with template on the 2nd page and submit', async () => {
                await templatePageV2.hybridForm.selectExistingEnv(environmentName);
                await templatePageV2.toggleEnvLinkButton();
                await templatePageV2.clickSubmitBtn();
                await templatePageV2.waitForTemplateExist(templateName);

                const templateId = await templateUtils.getIdByName(templateName);
                templateCreated.push({ id: templateId, type: WorkloadType.Workspace });
            });

            await test.step('update env image url', async () => {
                await environmentUtils.update(envAsset.meta.id, {
                    meta: {
                        name: environmentName
                    },
                    spec: {
                        image: ubuntuUrl
                    }
                });
            });

            await test.step('verify template env automatically updated', async () => {
                await templatePageV2.toggleTemplate(templateName);
                await templatePageV2.clickUpadteTemplateBtn();
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toHaveValue(ubuntuUrl);
            });
        });

        test('T5331911: verify updated compute resource info is propagated to linked template', async function () {
            videoName = 'updated-cr-propagated-to-template';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('link compute resource with template on the 2nd page and submit', async () => {
                await templatePageV2.hybridForm.selectExistingComputeResource(computeResourceName);
                await templatePageV2.toggleComputeResourceLinkButton();
                await templatePageV2.clickSubmitBtn();
                await templatePageV2.waitForTemplateExist(templateName);

                const templateId = await templateUtils.getIdByName(templateName);
                templateCreated.push({ id: templateId, type: WorkloadType.Workspace });
            });

            await test.step('update compute resource', async () => {
                await computeResourceUtils.update(computeResource.meta.id, {
                    meta: {
                        name: computeResourceName
                    },
                    spec: {
                        gpuDevicesRequest: 1
                    }
                });
            });

            await test.step('verify template compute resource automatically updated', async () => {
                await templatePageV2.toggleTemplate(templateName);
                await templatePageV2.clickUpadteTemplateBtn();
                await expect(hybridForm.locator(hybridForm.hybridComputeSection.gpuDevicesInput)).toHaveValue(
                    '1'
                );
            });
        });

        test('T5331912: verify deleting env is reflected in linked template', async function () {
            videoName = 'delete-env-reflect-in-template';

            await test.step('navigate to the template page', async () => {
                await templatePageV2.goToPage();
            });

            await test.step('open workspace template creation form and fill the 1st page', async () => {
                await templatePageV2.fillSingleNodeWorkloadTemplateFirstPage(
                    WorkloadType.Workspace,
                    templateName,
                    {
                        scope: Scope.Project,
                        scopeEntityName: projectName
                    }
                );
            });

            await test.step('link compute resource with template on the 2nd page and submit', async () => {
                await templatePageV2.hybridForm.selectExistingEnv(environmentName);
                await templatePageV2.toggleEnvLinkButton();
                await templatePageV2.clickSubmitBtn();
                await templatePageV2.waitForTemplateExist(templateName);

                const templateId = await templateUtils.getIdByName(templateName);
                templateCreated.push({ id: templateId, type: WorkloadType.Workspace });
            });

            await test.step('delete the env', async () => {
                await environmentUtils.delete(envAsset.meta.id);
                envDeleted = true;
            });

            await test.step('verify it is reflected in linked template', async () => {
                await templatePageV2.toggleTemplate(templateName);
                await templatePageV2.clickUpadteTemplateBtn();
                await expect(
                    hybridForm.locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                ).toHaveValue('');
            });
        });
    });
});
