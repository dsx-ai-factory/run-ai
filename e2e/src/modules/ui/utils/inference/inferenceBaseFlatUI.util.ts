/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import {
    ECredentialsSourceType,
    EModelSourceForNim,
    IAutoScalingOptions,
    IComputeOptions,
    IDataSourceOptions,
    IEnvironmentOptions,
    IGeneralSectionOptions,
    IHuggingFaceModelOptions,
    INimModelOptions,
    IServingPortOptions
} from '../../../helpers/models/inferenceBaseFlatUI.models';
import { EDataSourceOption, HybridForm } from '../../runaiV2/pages/HybridSubmissionForm';
import { ComputeResourcePage } from '../../runaiV2/pages/ComputeResourcePage';
import { EnvironmentPage } from '../../runaiV2/pages/EnvironmentPage';
import { DataSourcePage } from '../../runaiV2/pages/DataSourcePage';
import { InferencePage } from '../../runaiV2/pages/InferencePage';
import { logger, waitUtils } from '../../../helpers';

export class InferenceBaseFlatUIUtil {
    inference: InferencePage;

    computeResource: ComputeResourcePage;

    environment: EnvironmentPage;

    dataSource: DataSourcePage;

    hybridForm: HybridForm;

    constructor(private readonly page: Page) {
        this.inference = new InferencePage(page);
        this.computeResource = new ComputeResourcePage(page);
        this.environment = new EnvironmentPage(page);
        this.dataSource = new DataSourcePage(page);
        this.hybridForm = new HybridForm(page);
    }

    async fillAutoScalingOptions(autoScaling: IAutoScalingOptions): Promise<void> {
        await this.inference.clickOnAutoScaleToggleBtn();

        if (autoScaling.minReplicas !== undefined) {
            await this.inference.fillMinReplicasInput(autoScaling.minReplicas);
        }

        if (autoScaling.maxReplicas) {
            await this.inference.fillMaxReplicasInput(autoScaling.maxReplicas);
        }

        if (autoScaling.thresholdValue) {
            await this.inference.fillThresholdValueInput(autoScaling.thresholdValue);
        }

        if (autoScaling.autoScalingPeriod && autoScaling.minReplicas !== 0) {
            throw new Error('autoScalingPeriod is not supported for minReplicas > 0');
        }

        if (autoScaling.minReplicas === 0 && autoScaling.autoScalingPeriod) {
            if (autoScaling.autoScalingPeriod === '1m') {
                await this.inference.selectAfter1Minute();
            }
        }
    }

    async fillServingPortOptions(options: IServingPortOptions): Promise<void> {
        if (options.protocol) {
            await this.hybridForm.selectServingPortProtocol(options.protocol);
        }

        if (options.port) {
            await this.hybridForm.fillServingPortInput(options.port.toString());
        }

        if (options.access) {
            await this.hybridForm.selectServingPortAccess(options.access.option);
        }
    }

    async fillEnvironmentData(options?: IEnvironmentOptions): Promise<void> {
        if (!options) return;

        if (options.assetName) {
            await this.hybridForm.openEnvironmentsDrawer();
            await this.hybridForm.clickOnDrawerItemByLabel(options.assetName);
        }

        if (options.newEnvironmentResource) {
            await this.hybridForm.openEnvironmentsDrawer();
            await this.hybridForm.clickOnAddNewAssetBtn();

            await this.environment.createNewEnvironment(
                options.newEnvironmentResource.assetName,
                options.newEnvironmentResource.projectName,
                {
                    image: options.newEnvironmentResource.image,
                    containerPort: options.newEnvironmentResource.servingPort.container,
                    command: options.newEnvironmentResource.runtimeSettings?.cmd,
                    args: options.newEnvironmentResource.runtimeSettings?.args
                }
            );
            await this.hybridForm.clickOnDrawerItemByLabel(options.newEnvironmentResource.assetName, false);
        }

        if (options.image) {
            await this.hybridForm.fillImageInput(options.image);
        }

        if (options.imagePullPolicy) {
            await this.hybridForm.clickOnSubSectionByLabel(this.hybridForm.imagePullSection);
            await this.page
                .locator(`div[role="radio"]:has-text("${options.imagePullPolicy}")`)
                .waitFor({ state: 'visible' });
            await this.hybridForm.clickOnRadioOptionByLabel(options.imagePullPolicy);
        }

        if (options.servingPort) {
            await this.fillServingPortOptions(options.servingPort);
        }

        if (options.tool) {
            await this.hybridForm.clickOnSubSectionByLabel('Tools');
            await this.hybridForm.addToolTypeByLabel(options.tool);
        }

        if (options.runtimeSettings) {
            if (!(await this.hybridForm.isSubSectionOpenedByLabel('Runtime settings'))) {
                logger.debug('Runtime settings section is not opened, opening it');
                await this.hybridForm.clickOnSubSectionByLabel('Runtime settings');
            } else {
                logger.debug('Runtime settings section is already opened');
            }
            if (options.runtimeSettings.cmd && options.runtimeSettings.args) {
                if (!options.runtimeSettings.isEditCmd) {
                    const addCmdArgsBtn = this.page.locator(
                        this.hybridForm.hybridEnvironmentSection.cmdAndArgsBtn
                    );
                    if (await addCmdArgsBtn.isVisible().catch(() => false)) {
                        await this.hybridForm.clickOnCommandAndArguments();
                    }
                }
                await this.hybridForm.fillCmdInput(options.runtimeSettings.cmd);
                await this.hybridForm.fillArgsInput(options.runtimeSettings.args);
            }

            if (options.runtimeSettings.environmentVariablesValues) {
                await this.inference.fillEnvironmentVariablesValues(
                    options.runtimeSettings.environmentVariablesValues
                );
            }

            if (options.runtimeSettings.environmentVariablesValuesByName) {
                await this.inference.fillEnvironmentVariablesValuesByName(
                    options.runtimeSettings.environmentVariablesValuesByName
                );
            }

            if (options.runtimeSettings.fullEnvironmentVariables) {
                await this.inference.addEnvironmentVariables(
                    options.runtimeSettings.fullEnvironmentVariables
                );
            }
        }
    }

    async fillComputeData(options?: IComputeOptions): Promise<void> {
        if (!options) return;

        if (options.assetName) {
            await this.hybridForm.page.waitForTimeout(300);
            await this.hybridForm.openComputesDrawer();
            await this.hybridForm.clickOnDrawerItemByLabel(
                options.assetName,
                options.waitForUnifiedRequest ?? true
            );
        }

        if (options.gpu) {
            if (options.gpu.devicesInput) {
                await this.hybridForm.fillGpuDevicesInput(options.gpu.devicesInput);
            }

            if (options.gpu.memoryRequest) {
                await this.hybridForm.fillGpuRequestInput(options.gpu.memoryRequest);
            }
        }

        if (options.newComputeResource) {
            await this.hybridForm.page.waitForTimeout(300);
            await this.hybridForm.openComputesDrawer();
            await this.hybridForm.clickOnAddNewAssetBtn();
            await this.computeResource.createNewCompute(
                options.newComputeResource.assetName,
                options.newComputeResource.projectName,
                options.newComputeResource.gpu?.devicesInput,
                options.newComputeResource.isProjectScope
            );
            await this.hybridForm.clickOnDrawerItemByLabel(options.newComputeResource.assetName, false);
        }

        if (options.cpu) {
            if (options.cpu.compute) {
                if (options.cpu.compute.request) {
                    await this.hybridForm.fillCpuCoresInput(options.cpu.compute.request);
                }
            }

            if (options.cpu.memory) {
                if (options.cpu.memory.request) {
                    await this.hybridForm.fillCpuMemoryInput(options.cpu.memory.request);
                }
            }
        }

        if (options.extendedResource) {
            await this.hybridForm.openExtendedResourceSection();

            if (options.extendedResource.increaseSharedMemory) {
                await this.hybridForm.toggleIncreaseSharedMemory();
            }

            if (options.extendedResource.resources) {
                await this.hybridForm.fillExtendedResourcesValues(options.extendedResource.resources);
            }
        }

        if (options.nodes) {
            await this.hybridForm.openNodesSection();
            if (options.nodes.nodeType) {
                await this.hybridForm.fillNodeAffinityInput(options.nodes.nodeType);
            }
        }

        // auto scale
        if (options.autoScaling) {
            await this.fillAutoScalingOptions(options.autoScaling);
        }
    }

    async fillDataSourceData(
        options: IDataSourceOptions,
        supportedMultiStorages: boolean = true
    ): Promise<void> {
        if (!options) return;

        await this.hybridForm.openDataSourcesSection();

        if (options.nfs) {
            await this.hybridForm.openDataSourcesDrawer();
            await this.hybridForm.clickOnAddNewDropdownBtnInDrawer();
            await this.hybridForm.selectDataSourceInDrawer(EDataSourceOption.Nfs);
            await this.dataSource.fillNFSForm(options.nfs);
            await this.hybridForm.clickOnDrawerItemByLabel(options.nfs, false);
            if (supportedMultiStorages) {
                await this.hybridForm.clickOnLoadFromDrawer();
            }
        }

        if (options.pvc) {
            await this.hybridForm.openDataSourcesDrawer();
            await this.hybridForm.clickOnAddNewDropdownBtnInDrawer();
            await this.hybridForm.selectDataSourceInDrawer(EDataSourceOption.Pvc);
            await this.dataSource.clickOnNewPvcOption();
            await waitUtils.waitForElementToExist(this.dataSource, this.dataSource.PvcStorageClassSelect);
            await this.dataSource.fillPVCForm(options.pvc);
            await waitUtils.waitForElementToBeEnabled(
                this.dataSource,
                this.hybridForm.hybridDrawer.drawerItemLabel,
                options.pvc.name
            );
            await this.hybridForm.clickOnDrawerItemByLabel(options.pvc.name, false);
            if (supportedMultiStorages) {
                await this.hybridForm.clickOnLoadFromDrawer();
            }
        }

        if (options.volume) {
            await this.hybridForm.addNewDataSource(EDataSourceOption.Volume);
            await this.inference.setWorkloadVolume(options.volume);
        }

        if (options.assetNames) {
            await this.hybridForm.openDataSourcesDrawer();
            await this.hybridForm.clickOnDrawerItemByLabel(
                options.assetNames[0],
                options.waitForUnifiedRequest ?? true
            );
            if (supportedMultiStorages) {
                await this.hybridForm.clickOnLoadFromDrawer();
            }
        }
    }

    async fillHuggingFaceModelData(options?: IHuggingFaceModelOptions): Promise<void> {
        if (!options) return;

        if (options.credentials && options.hfToken) {
            throw new Error('Cannot provide both credentials and an hfToken. Please provide only one.');
        }

        if (options.modelName) {
            await this.hybridForm.selectFromDropDownWithInput(
                this.hybridForm.hybridHuggingFaceModelSection.modelNameSelectInput,
                options.modelName
            );
        }

        if (options.hfToken) {
            await this.hybridForm.selectHuggingFaceTokenManually();
            await this.hybridForm.fillHuggingFaceToken(options.hfToken);
        }

        if (options.credentials) {
            await this.hybridForm.selectHuggingFaceTokenFromCredentials();
            await this.hybridForm.selectHuggingFaceCredentials(
                options.credentials.source,
                options.credentials.secretName,
                options.credentials.sourceDisabled
            );
        }

        if (options.servingPort) {
            if (options.shouldOpenServingPortSection) {
                await this.inference.openServingPortSection();
            }
            await this.fillServingPortOptions(options.servingPort);
        }
    }

    async fillNimModelData(options?: INimModelOptions): Promise<void> {
        if (!options) return;

        if (options.credentials && options.apiKey) {
            throw new Error('Cannot provide both credentials and an API key. Please provide only one.');
        }

        if (options.modelSource) {
            await this.hybridForm.selectFromDropDown(
                this.hybridForm.hybridNimModelSection.modelSourceSelect,
                options.modelSource
            );
        }

        if (options.modelSource === EModelSourceForNim.HuggingFace) {
            await this.hybridForm.selectFromDropDownWithInput(
                this.hybridForm.hybridNimModelSection.modelNameSelectInputHuggingFace,
                options.modelName
            );
        } else {
            await this.hybridForm.selectFromDropDownWithInput(
                this.hybridForm.hybridNimModelSection.modelNameSelectInput,
                options.modelName
            );
        }

        if (options.hfToken) {
            await this.hybridForm.selectHuggingFaceTokenManually();
            await this.hybridForm.fillHuggingFaceToken(options.hfToken);
        }

        if (options.profileOrHash) {
            await this.hybridForm.selectProfileManuallyAndFillTheProfile(options.profileOrHash);
        }

        if (options.fromStorageOnly) {
            await this.hybridForm.selectNimAccessModelFromStorageOnly();
        }

        if (options.apiKey) {
            await this.hybridForm.selectNimAccessFromProvideAToken();
            await this.hybridForm.fillNimApiKey(options.apiKey);
        }

        if (options.credentials) {
            if (options.modelSource === EModelSourceForNim.HuggingFace) {
                await this.hybridForm.selectHuggingFaceTokenFromCredentials();
                await this.hybridForm.selectHuggingFaceCredentials(
                    options.credentials.source,
                    options.credentials.secretName,
                    options.credentials.sourceDisabled
                );
            } else {
                await this.hybridForm.selectNimAccessFromCredentials();
                await this.hybridForm.selectNimCredentials(
                    options.credentials.source,
                    options.credentials.secretName,
                    options.credentials.sourceType,
                    options.credentials.sourceDisabled
                );
            }
        }

        if (options.servingPort) {
            await this.fillServingPortOptions(options.servingPort);
        }

        if (options.nimPullImageCredential) {
            const isHuggingFaceSource = options.modelSource === EModelSourceForNim.HuggingFace;
            if (isHuggingFaceSource) {
                const pull = this.hybridForm.hybridNimModelSection.imagePullCredentials;
                if (options.nimPullImageCredential.dockerRegistry) {
                    await this.hybridForm.selectFromDropDown(pull.sourceTypeSelect, 'Docker Registry');
                    await this.hybridForm.selectFromDropDown(
                        pull.secretNameSelect,
                        options.nimPullImageCredential.dockerRegistry
                    );
                } else if (options.nimPullImageCredential.ngcApiKey) {
                    await this.hybridForm.selectFromDropDown(
                        pull.secretNameSelect,
                        options.nimPullImageCredential.ngcApiKey
                    );
                } else {
                    throw new Error('No docker registry or NGC API key provided');
                }
            } else {
                // NVIDIA NGC catalog (default): image pull uses docker-registry-selector in nim-access-model
                if (options.nimPullImageCredential.dockerRegistry) {
                    const dockerRegistrySelect = this.hybridForm.hybridNimModelSection.dockerRegistry;
                    await this.hybridForm.locator(dockerRegistrySelect).waitFor({ state: 'visible', timeout: 15000 });
                    await this.hybridForm.selectFromDropDown(
                        dockerRegistrySelect,
                        options.nimPullImageCredential.dockerRegistry
                    );
                } else if (options.nimPullImageCredential.ngcApiKey) {
                    throw new Error(
                        'nimPullImageCredential.ngcApiKey is only supported with modelSource Hugging Face'
                    );
                } else {
                    throw new Error('No docker registry or NGC API key provided');
                }
            }
        }
    }

    async fillGeneralSection(generalSectionOptions?: IGeneralSectionOptions): Promise<void> {
        if (!generalSectionOptions) return;

        await this.inference.page.waitForTimeout(1000); // wait for 1 second to ensure the section is loaded
        const generalSectionSelector = this.hybridForm.hybridGeneralSection.sectionSelector;

        // After navigating to advanced setup, the general section may take a moment to render.
        // Wait for it (or for timeout inputs) so we don't proceed while the view is still loading.
        const sectionLocator = this.inference.page.locator(generalSectionSelector);
        const inferenceScopedSelector = `.inference-app ${generalSectionSelector}`;
        const inferenceSectionLocator = this.inference.page.locator(inferenceScopedSelector);
        try {
            await Promise.race([
                sectionLocator.waitFor({ state: 'visible', timeout: 10000 }),
                inferenceSectionLocator.waitFor({ state: 'visible', timeout: 10000 })
            ]);
            logger.debug('general section became visible');
        } catch {
            logger.debug('general section did not appear within timeout, continuing to fill fields');
        }

        // Check if the general section exists and needs to be opened
        const sectionExists = await this.inference.page
            .locator(generalSectionSelector)
            .isVisible()
            .catch(() => false);
        if (!sectionExists) {
            const sectionExistsInference = await inferenceSectionLocator.isVisible().catch(() => false);
            if (sectionExistsInference) {
                logger.debug('general section found inside .inference-app');
                const isCollapsed = await inferenceSectionLocator
                    .evaluate((el) => el.classList.contains('q-expansion-item--collapsed'))
                    .catch(() => false);
                if (isCollapsed) {
                    logger.debug('general section is collapsed, opening it');
                    await inferenceSectionLocator.click();
                }
            } else {
                logger.debug('general section does not exist');
            }
        } else {
            const isCollapsed = await this.inference.page
                .locator(generalSectionSelector)
                .evaluate((el) => el.classList.contains('q-expansion-item--collapsed'))
                .catch(() => false);
            if (isCollapsed) {
                logger.debug('general section is collapsed, opening it');
                await this.inference.openGeneralSection();
                await this.inference.page.waitForTimeout(300); // wait for 300 milliseconds to ensure the section is opened
            } else {
                logger.debug('general section is not collapsed');
            }
        }

        if (generalSectionOptions.initializationTimeout) {
            await this.inference.fillInitializationTimeoutSeconds(
                generalSectionOptions.initializationTimeout
            );
        }

        if (generalSectionOptions.requestTimeout) {
            await this.inference.fillRequestTimeoutSeconds(generalSectionOptions.requestTimeout);
        }

        if (generalSectionOptions.annotations) {
            await this.inference.fillAnnotationValues(generalSectionOptions.annotations);
        }

        if (generalSectionOptions.labels) {
            await this.inference.fillLabelsValues(generalSectionOptions.labels);
        }

        if (generalSectionOptions.preemptibility) {
            await this.hybridForm.selectPreemptibility(generalSectionOptions.preemptibility);
        }

        if (generalSectionOptions.startupPolicy) {
            await this.hybridForm.selectStartupPolicy(generalSectionOptions.startupPolicy);
        }

        if (generalSectionOptions.priority) {
            await this.hybridForm.selectPriority(generalSectionOptions.priority);
        }
    }
}
