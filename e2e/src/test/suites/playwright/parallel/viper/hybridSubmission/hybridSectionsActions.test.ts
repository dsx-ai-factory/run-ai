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
    environmentUtils,
    unleashClient
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect, test } from '@playwright/test';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import {
    EnvironmentAsset,
    EnvironmentCreationRequest,
    ImagePullPolicy,
    InternalConnectionType,
    InternalToolType,
    Scope,
    UidGidSource,
    WorkloadType
} from '../../../../../../modules/api/apis/assetsApi';
import { EFeatureToggleKeys } from '../../../../../../modules/helpers/utils/Unleash/featureToggle';

type TElementDetails = {
    name: string;
    selector: string;
    setValue: CallableFunction;
    defaultValue: string;
    overrideWith: string;
    getCurrentValue: CallableFunction;
    templateValue?: string;
    triggerToShow?: CallableFunction;
};

type TSectionsMeta = {
    drawerItemToSelect: string;
    selectItemInDrawer: CallableFunction;
    openDrawer: CallableFunction;
    subSections: string[];
    clearSection: CallableFunction;
    resetSection: CallableFunction;
    elements: TElementDetails[];
};

type THybridElementTests = {
    environment: TSectionsMeta;
    compute: TSectionsMeta;
    // dataSource: TSectionsMeta;
};

test.describe(
    `Hybrid sections actions, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.viper}`,
    { tag: [tags.services.policy, tags.services.workloads, tags.services.frontend, tags.runOn.parallelCI] },
    function () {
        let project: Project;
        let videoName: string;
        let createdEnv: EnvironmentAsset;

        let hybridForm: HybridForm;
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workload: WorkloadPage;
        let workloadName: string;

        const createHybridEnvironment = async (projectId: number): Promise<void> => {
            const environmentName = objectGenerator.generateNameByPrefix('e2e-hybrid-env');
            const environmentCreationRequest: EnvironmentCreationRequest = {
                meta: {
                    name: environmentName,
                    scope: Scope.Project,
                    projectId,
                    workloadSupportedTypes: {
                        training: true,
                        workspace: true
                    }
                },
                spec: {
                    command: 'das ist der befehl',
                    args: 'etwas argument fur dich',
                    environmentVariables: [
                        {
                            name: 'umgebungsvariable',
                            value: 'etwas wert'
                        }
                    ],
                    image: 'alpine',
                    imagePullPolicy: ImagePullPolicy.IfNotPresent,
                    workingDir: null,
                    connections: [
                        {
                            name: 'Jupyter',
                            isExternal: false,
                            internalToolInfo: {
                                toolType: InternalToolType.JupyterNotebook,
                                connectionType: InternalConnectionType.ExternalUrl,
                                containerPort: 8888,
                                externalUrlInfo: {
                                    isCustomUrl: false,
                                    externalUrl: null
                                }
                            }
                        }
                    ],
                    uidGidSource: UidGidSource.FromTheImage,
                    overrideUidGidInWorkspace: false,
                    capabilities: []
                }
            };
            createdEnv = await environmentUtils.create(environmentCreationRequest);
        };

        const deleteEnvironment = async (): Promise<void> => {
            try {
                await environmentUtils.delete(createdEnv.meta.id);
            } catch {
                logger.error(`failed to delete environment with id: ${createdEnv.meta.id}`);
            }
        };

        test.beforeAll('creating project and namespace', async function () {
            browserEngine = new BrowserEngine();
            project = await projectResourcesUtils.createProjectAndNamespace('h-p');
            logger.info(`created project ${project.name} ,namespace: "${project.namespace}"`);
            logger.info(`created project ${project.name}`);
        });

        test.afterAll('deleting project and namespace', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${project.name} an namespace ${project.namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${project.name} an namespace ${project.namespace}`);
                throw error;
            }
        });

        test.beforeEach('start browser', async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.l2_researcher.email,
                flowRootState.users.l2_researcher.password
            );
            workloadName = objectGenerator.generateNameByPrefix('hybrid-sections');
            await workload.navigateToWorkloads();
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, 'hybrid-sections' + videoName);
            } catch (e) {
                logger.error(e);
            }
        });

        async function fillWorkloadCreationFirstPage(): Promise<void> {
            await workload.openWorkspaceForm();

            await workload.selectClusterFromDropDown(flowRootState.cluster.name);
            await workload.workloadFormSelectProject(project.name);
            await workload.workloadFormSelectTemplate();
            await hybridForm.fillWorkspaceName(workloadName);
            await hybridForm.waitForWorkloadNameCount(workloadName);
            await hybridForm.clickOnContinue(WorkloadType.Workspace);
        }

        const getHybridSectionsTestConfig = (envName: string, computeName: string): THybridElementTests => {
            return {
                environment: {
                    drawerItemToSelect: envName,
                    selectItemInDrawer: async (val: string) => hybridForm.clickOnDrawerItemByLabel(val),
                    openDrawer: async () => hybridForm.openEnvironmentsDrawer(),
                    resetSection: async () => hybridForm.resetEnvironmentSection(),
                    clearSection: async () => hybridForm.clearEnvironmentsSection(),
                    subSections: [hybridForm.imagePullSection, 'Tools', 'Runtime settings', 'Security'],
                    elements: [
                        {
                            name: 'image input',
                            selector: hybridForm.hybridEnvironmentSection.hybridImageInput,
                            setValue: async (val: string) => hybridForm.fillImageInput(val),
                            overrideWith: 'ubuntu',
                            defaultValue: '',
                            getCurrentValue: async () =>
                                hybridForm.page
                                    .locator(hybridForm.hybridEnvironmentSection.hybridImageInput)
                                    .inputValue()
                        },
                        {
                            name: 'args input',
                            selector: hybridForm.hybridEnvironmentSection.argsInput,
                            setValue: async (val: string) => hybridForm.fillArgsInput(val),
                            overrideWith: 'x=1',
                            defaultValue: '',
                            triggerToShow: async () => hybridForm.clickOnCommandAndArguments(),
                            getCurrentValue: async () =>
                                hybridForm.page
                                    .locator(hybridForm.hybridEnvironmentSection.argsInput)
                                    .inputValue()
                        },
                        {
                            name: 'command input',
                            selector: hybridForm.hybridEnvironmentSection.cmdInput,
                            setValue: async (val: string) => hybridForm.fillCmdInput(val),
                            overrideWith: 'sleep 10000',
                            defaultValue: '',
                            triggerToShow: async () => hybridForm.clickOnCommandAndArguments(),
                            getCurrentValue: async () =>
                                hybridForm.page
                                    .locator(hybridForm.hybridEnvironmentSection.cmdInput)
                                    .inputValue()
                        },
                        {
                            name: 'tool select',
                            selector: hybridForm.hybridEnvironmentSection.toolSelect,
                            setValue: async (val: string) =>
                                hybridForm.selectOptionByValue(
                                    hybridForm.hybridEnvironmentSection.toolSelect,
                                    val
                                ),
                            overrideWith: 'VSCode',
                            defaultValue: 'Select tool',
                            triggerToShow: async () => hybridForm.clickOnAddNewToolButton(),
                            getCurrentValue: async () =>
                                hybridForm.getSelectValueBySelector(
                                    hybridForm.hybridEnvironmentSection.toolSelect
                                )
                        },
                        {
                            name: 'env var custom name input',
                            selector: hybridForm.hybridEnvironmentSection.envVarNameInput,
                            setValue: async (val: string) =>
                                hybridForm.fillInputBySelector(
                                    hybridForm.hybridEnvironmentSection.envVarNameInput,
                                    val
                                ),
                            overrideWith: 'some-env-var-name',
                            defaultValue: '',
                            triggerToShow: async () => hybridForm.clickOnAddEnvVarBtn(),
                            getCurrentValue: async () =>
                                hybridForm.page
                                    .locator(hybridForm.hybridEnvironmentSection.envVarNameInput)
                                    .inputValue()
                        }
                    ]
                },
                compute: {
                    drawerItemToSelect: computeName,
                    selectItemInDrawer: async (val: string) => hybridForm.clickOnDrawerItemByLabel(val),
                    openDrawer: async () => hybridForm.openComputesDrawer(),
                    resetSection: async () => hybridForm.resetComputesSection(),
                    clearSection: async () => hybridForm.clearComputesSection(),
                    subSections: ['Extended resources', 'Nodes'],
                    elements: [
                        {
                            name: 'gpu devices request input',
                            selector: hybridForm.hybridComputeSection.gpuDevicesInput,
                            setValue: async (val: string) => hybridForm.fillGpuDevicesInput(val),
                            overrideWith: '2',
                            defaultValue: '0',
                            getCurrentValue: async () =>
                                hybridForm.page
                                    .locator(hybridForm.hybridComputeSection.gpuDevicesInput)
                                    .inputValue()
                        },
                        {
                            name: 'cpu core request input',
                            selector: hybridForm.hybridComputeSection.cpuCoresRequestInput,
                            setValue: async (val: string) => hybridForm.fillCpuCoresInput(val),
                            overrideWith: '5',
                            defaultValue: '0',
                            getCurrentValue: async () =>
                                hybridForm.page
                                    .locator(hybridForm.hybridComputeSection.cpuCoresRequestInput)
                                    .inputValue()
                        },
                        {
                            name: 'cpu memory request input',
                            selector: hybridForm.hybridComputeSection.cpuMemoryRequestInput,
                            setValue: async (val: string) => hybridForm.fillCpuMemoryInput(val),
                            overrideWith: '600',
                            defaultValue: '0',
                            getCurrentValue: async () =>
                                hybridForm.page
                                    .locator(hybridForm.hybridComputeSection.cpuMemoryRequestInput)
                                    .inputValue()
                        },
                        {
                            name: 'gpu request input',
                            selector: hybridForm.hybridComputeSection.gpuRequestValueInput,
                            setValue: async (val: string) => hybridForm.fillGpuRequestInput(val),
                            overrideWith: '15',
                            defaultValue: '50',
                            getCurrentValue: async (): Promise<string> => {
                                await hybridForm.switchToggleBySelector('[aid="gpu-fraction-toggle"]', true);
                                return hybridForm.page
                                    .locator(hybridForm.hybridComputeSection.gpuRequestValueInput)
                                    .inputValue();
                            },
                            triggerToShow: async (): Promise<void> => {
                                await hybridForm.fillGpuDevicesInput('1');
                                await hybridForm.toggleFoldingSubSection('[aid="gpu-fraction-toggle"]', true);
                            }
                        }
                    ]
                }
                // to be added
                // dataSource: {}
            };
        };

        test('select an asset template, go over each relevant field change its value and check section actions  -> reset -> clear ', async function () {
            videoName = 'hybrid-sections-reset-clear-actions';
            await createHybridEnvironment(project.id);
            const testConfig = getHybridSectionsTestConfig(createdEnv.meta.name, 'one-gpu');

            await fillWorkloadCreationFirstPage();

            for (const section of Object.values(testConfig)) {
                // opening all sub-sections to make most elements visible
                for (const subSection of section.subSections) {
                    await hybridForm.clickOnSubSectionByLabel(subSection);
                }

                // selecting a template
                await section.openDrawer();
                await section.selectItemInDrawer(section.drawerItemToSelect);

                // overriding the template value
                for (const el of section.elements) {
                    el.templateValue = await el.getCurrentValue();
                    await el.setValue(el.overrideWith);

                    const currentValue = await el.getCurrentValue();
                    expect(currentValue).toEqual(el.overrideWith);
                }

                // reset to template defaults
                await section.resetSection();

                // checking that original templates values are set
                for (const el of section.elements) {
                    const isVisible = await hybridForm.locator(el.selector).isVisible();
                    if (!isVisible && !!el.triggerToShow) {
                        logger.debug(`element ${el.name} is not visible - attempting to trigger visiblity`);
                        await el.triggerToShow();
                    }
                    const currentValue = await el.getCurrentValue();
                    expect(currentValue).toEqual(el.templateValue);
                }

                // clearing the section entirely to the original state
                await section.clearSection();

                for (const subSection of section.subSections) {
                    await hybridForm.clickOnSubSectionByLabel(subSection);
                }

                for (const el of section.elements) {
                    const isVisible = await hybridForm.locator(el.selector).isVisible();
                    if (!isVisible && !!el.triggerToShow) {
                        logger.debug(`element ${el.name} is not visible - attempting to trigger visiblity`);
                        await el.triggerToShow();
                    }
                    const currentValue = await el.getCurrentValue();
                    expect(currentValue).toEqual(el.defaultValue);
                }
                await deleteEnvironment();
            }
        });
    }
);
