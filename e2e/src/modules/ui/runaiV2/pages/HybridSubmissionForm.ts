/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { WorkloadPage } from './WorkloadPage';
import { WorkloadType } from '../../../api/apis/assetsApi';
import { expect } from 'allure-playwright';
import {
    ECredentialsSource,
    ECredentialsSourceType,
    IServingPortAccessOptions
} from '../../../helpers/models/inferenceBaseFlatUI.models';
import {
    ECredentialSource,
    ECredentialType,
    CREDENTIAL_SOURCE_LABELS,
    CREDENTIAL_TYPE_LABELS
} from '../../../helpers/models/credential.models';
import { logger, waitUtils } from '../../../helpers';

interface IServingPortAccess {
    servingPortAccessToggle: string;
    servingPortAccessEditBtn: string;
    servingPortAccessPublic: string;
    servingPortAccessEveryone: string;
    servingPortAccessGroup: string;
    servingPortAccessUsers: string;
    servingPortAccessSaveBtn: string;
    servingPortAccessCancelBtn: string;
    servingPortAccessModalOptions: string;
    servingPortAccessQuickView: string;
    servingPortAccessModalSaveBtn: string;
    servingPortAccessModalAddItemBtn: string;
    servingPortAccessModalRemoveItemBtn: string;
    servingPortAccessModalUserChips: string;
}

interface IAutoScalingSectionElements {
    sectionSelector: string;
    minReplicasInput: string;
    maxReplicasInput: string;
    metricSelect: string;
    metricThresholdInput: string;
    scaleDownSelect: string;
    scaleToZeroRetentionSecondsInput: string;
}

interface IHybridEnvironmentSectionElements {
    sectionSelector: string;
    hybridImageInput: string;
    addToolButton: string;
    toolTypeSelect: string;
    cmdAndArgsBtn: string;
    cmdInput: string;
    argsInput: string;
    toolSelect: string;
    toolCustomInput: string;
    toolWandbInput: string;
    toolCustomUrlInput: string;
    toolCustomPortInput: string;
    toolAutoGenSelect: string;
    toolConnectionTypeSelect: string;
    addEnvVarBtn: string;
    envVarNameInput: string;
    envVarSourceSelect: string;
    envVarValue: string;
    envVarPersonalSecretSelect: string;
    envVarCredentialTypeSelect: string;
    envVarSecretSelect: string;
    envVarSecretKey: string;
    envVarPersonalSecretKeyInput: string;
    envVarCustomKey: string;
    envVarConfigMapSelect: string;
    envVarConfigMapKey: string;
    envVarRemoveBtn: string;
    imagePullSecretsSelect: string;
    imagePullSecretsSourceSelect: string;
    servingPortProtocol: string;
    servingPortPort: string;
    servingPortAccess: IServingPortAccess;
    supplementalGroupsInput: string;
    uidInput: string;
    gidInput: string;
    securityToggle: string;
    runtimeToggle: string;
    imagePullPolicySubSection: string;
}
interface IHybridComputeSectionElements {
    sectionSelector: string;
    gpuDevicesInput: string;
    gpuRequestTypeSelect: string;
    gpuMigProfileSelect: string;
    gpuRequestValueInput: string;
    gpuLimitValueInput: string;
    cpuCoresRequestInput: string;
    cpuCoresLimitInput: string;
    cpuMemoryRequestInput: string;
    increaseSharedMemory: string;
    extendedResourceToggleBtn: string;
    extendedResourceBtn: string;
    nodesToggleBtn: string;
    nodeAffinityInput: string;
    autoScaling: IAutoScalingSectionElements;
    workersInput: string;
    topologyBtn: string;
    topologyKeyInput: string;
    topologyOperatorSelect: string;
}

export enum EDataSourceOption {
    Nfs,
    Pvc,
    S3,
    Git,
    HostPath,
    Volume
}

interface INfsFields {
    server: string;
    path: string;
    mountPath: string;
}

interface IPvcFields {
    storageClass: string;
    accessMode: string;
    claimSize: string;
    unit: string;
    volumeMode: string;
    path: string;
}
interface IHybridDataSourceSectionElements {
    sectionSelector: string;
    newBtn: string;
    newDataSourceOptions: {
        nfs: string;
        volume: string;
    };
    drawer: {
        options: {
            nfs: string;
            pvc: string;
            s3: string;
            git: string;
            hostPath: string;
            configMap: string;
        };
    };
    configMapInstance: {
        subPathInput: string;
    };
    containerPath: string;
    nfs: INfsFields;
    pvc: IPvcFields;
}

interface IHybridSectionsElements {
    loadExistingTemplatesButton: string;
    clearAssetTemplateButton: string;
    resetAssetTemplateButton: string;
}
interface IHybridDrawerElements {
    drawerItem: string;
    drawerItemLabel: string;
    selectDrawerItemsBtn: string;
    viewPolicyBtn: string;
    viewDetailsBtn: string;
    closeDrawerBtn: string;
}

interface INodePoolSectionElements {
    nodePoolListItem: string;
    addNodePoolBtn: string;
    nodePoolDropdownItem: (name: string) => string;
}

interface IHybridGeneralSectionElements {
    sectionSelector: string;
    servingConfigurationSection: {
        initializationTimeoutSeconds: {
            minutesInput: string;
            secondsInput: string;
        };
        requestTimeoutSeconds: {
            minutesInput: string;
            secondsInput: string;
        };
    };
    preemptibilitySelect: string;
    startupPolicy: string;
    prioritySelect: string;
}

interface IHybridHuggingFaceModelSectionElements {
    modelNameSelectInput: string;
    hfToken: string;
    provideToken: string;
    selectCredential: string;
    plaintextTokenWarning: string;
    credentials: {
        sourceSelect: string;
        secretNameSelect: string;
    };
    serverTypeSubSectionBtn: string;
    serverTypeTgi: string;
}

interface IHybridNimModelSectionElements {
    modelSourceSelect: string;
    modelNameSelectInput: string;
    modelNameSelectInputHuggingFace: string;
    apiKey: string;
    profileManually: string;
    profileInput: string;
    fromStorageOnly: string;
    accessFromCredentials: string;
    provideAToken: string;
    credentials: {
        sourceSelect: string;
        secretNameSelect: string;
        sourceTypeSelect: string;
    };
    imagePullCredentials: {
        sourceTypeSelect: string;
        secretNameSelect: string;
    };
    dockerRegistry: string;
}

export class HybridForm extends WorkloadPage {
    nodePoolSection: INodePoolSectionElements;

    hybridEnvironmentSection: IHybridEnvironmentSectionElements;

    hybridComputeSection: IHybridComputeSectionElements;

    hybridDataSourceSection: IHybridDataSourceSectionElements;

    hybridDrawer: IHybridDrawerElements;

    hybridSectionElement: IHybridSectionsElements;

    hybridGeneralSection: IHybridGeneralSectionElements;

    hybridNimModelSection: IHybridNimModelSectionElements;

    hybridHuggingFaceModelSection: IHybridHuggingFaceModelSectionElements;

    createWorkspaceNameInput: string;

    createTrainingNameInput: string;

    createWorkspaceContinueBtn: string;

    quickSubmitDropdownBtnComponent: string;

    quickSubmitDropdownBtn: string;

    quickSubmitDropdownSubmitOption: string;

    quickSubmitDropdownContinueOption: string;

    createTrainingContinueBtn: string;

    submitWorkloadBtn: string;

    addNewAssetBtn: string;

    addNewDropdownBtn: string;

    searchInputDrawer: string;

    imagePullSection: string;

    distributedMasterFormWrapper: string;

    constructor(page: Page) {
        super(page);
        this.createWorkspaceContinueBtn = '[aid="create-workspace-continue-btn"]';
        this.createTrainingContinueBtn = '[aid="create-training-continue-btn"]';
        this.createWorkspaceNameInput = '[aid="workspace-name-input"]';
        this.createTrainingNameInput = '[aid="training-name-input"]';
        this.submitWorkloadBtn = '[aid="submit-btn"]';
        this.addNewAssetBtn = '[aid="add-new-btn"]';
        this.addNewDropdownBtn = '[aid="add-new-dropdown-btn"]';
        this.searchInputDrawer = '[aid="search-input-drawer"]';
        this.imagePullSection = 'Image pull';
        this.quickSubmitDropdownBtnComponent = '.quick-submission-button';
        this.quickSubmitDropdownBtn = '[aid="quick-submission-dropdown-button"]';
        this.quickSubmitDropdownSubmitOption = '[aid="quick-submit-option"]';
        this.quickSubmitDropdownContinueOption = '[aid="advanced-setup-option"]';
        this.distributedMasterFormWrapper = '.distributed-hybrid-master';

        this.hybridSectionElement = {
            loadExistingTemplatesButton: '[aid="load-existing-templates"]',
            clearAssetTemplateButton: '[aid="clear-btn"]',
            resetAssetTemplateButton: '[aid="reset-btn"]'
        };
        this.hybridDrawer = {
            drawerItem: '.workload-drawer-item',
            drawerItemLabel: '.item-label',
            closeDrawerBtn: '.btn-icon-lg',
            selectDrawerItemsBtn: '[aid="multi-select-btn"]',
            viewPolicyBtn: '[aid="workload-drawer-item-policy-button"]',
            viewDetailsBtn: '[aid="workload-drawer-item-details-button"]'
        };
        this.hybridEnvironmentSection = {
            sectionSelector: '[aid="hybrid-environment-section"]',
            hybridImageInput: '[aid="environment-image-input"]',
            addToolButton: '[aid="add-tool-button"]',
            toolTypeSelect: '[aid="tool-type-select"]',
            cmdAndArgsBtn: '[aid="add-command-and-args-button"]',
            cmdInput: '[aid="command-input"]',
            argsInput: '[aid="args-input"]',
            toolSelect: '[aid="tool-type-select"]',
            toolCustomInput: '[aid="tool-type-custom-input"]',
            toolWandbInput: '[aid="tool-type-wandb-input"]',
            toolCustomUrlInput: '[aid="custom-url-input"]',
            toolCustomPortInput: '[aid="custom-port-input"]',
            toolAutoGenSelect: '[aid="auto-gen-select"]',
            toolConnectionTypeSelect: '[aid="connection-type-select"]',
            addEnvVarBtn: '[aid="add-environment-variable-button"]',
            envVarNameInput: '[aid="environment-variable-name"]',
            envVarSourceSelect: '[aid="environment-variable-source"]',
            envVarValue: '[aid="environment-variable-value"]',
            envVarSecretSelect: '[aid="environment-variable-secret"]',
            envVarSecretKey: '[aid="environment-variable-key-select"]',
            envVarPersonalSecretSelect: '.environment-variable-box [aid="credential-list-select"]',
            // The type control renders with class (image-pull section) or aid (NIM section) — accept both.
            envVarCredentialTypeSelect:
                '.environment-variable-box .credential-type-select, .environment-variable-box [aid="credential-type-select"]',
            envVarPersonalSecretKeyInput: '[aid="environment-variable-user-credentials-key-input"]',
            envVarCustomKey: '[aid="environment-variable-key-input"]',
            envVarConfigMapSelect: '[aid="environment-variable-config-map-name-select"]',
            envVarConfigMapKey: '[aid="environment-variable-config-map-key-input"]',
            envVarRemoveBtn: '[aid="remove-environment-variable-button"]',
            imagePullSecretsSelect: '.form-image-pull-policy [aid="credential-list-select"]',
            imagePullSecretsSourceSelect: '.form-image-pull-policy .credential-source-select',
            servingPortProtocol: '[aid="connection-protocol"]',
            servingPortPort: '[aid="container-port"]',
            servingPortAccess: {
                servingPortAccessToggle: '[aid="serving-endpoint-access-btn"]',
                servingPortAccessEditBtn: '[aid="edit-access-button"]',
                servingPortAccessPublic: '[aid="public"]',
                servingPortAccessEveryone: '[aid="everyone"]',
                servingPortAccessGroup: '[aid="group"]',
                servingPortAccessUsers: '[aid="specific-users"]',
                servingPortAccessSaveBtn: '[aid="save-connection-access-role"]',
                servingPortAccessCancelBtn: '[aid="cancel-connection-access-role"]',
                servingPortAccessQuickView: '.connection-access input',
                servingPortAccessModalOptions: '.connection-access-modal-body',
                servingPortAccessModalSaveBtn: '[aid="save-connection-access-role"]',
                servingPortAccessModalAddItemBtn: '.connection-access-modal-body [aid="add-item-button"]',
                servingPortAccessModalRemoveItemBtn: '[aid="remove-item-button"]',
                servingPortAccessModalUserChips:
                    '.connection-access-modal-body .subject-addable-multi-select .q-chip'
            },
            supplementalGroupsInput: '[aid="supplemental-groups-input"]',
            uidInput: '[aid="uid-input"]',
            gidInput: '[aid="gid-input"]',
            securityToggle: '[aid="security-sub-section"]',
            runtimeToggle: '[aid="runtime-settings-sub-section"]',
            imagePullPolicySubSection: '[aid="image-pull-policy-sub-section"]'
        };
        this.hybridComputeSection = {
            sectionSelector: '[aid="hybrid-compute-resource-section"]',
            gpuDevicesInput: '[aid="gpu-devices-input"]',
            gpuRequestTypeSelect: '[aid="gpu-type-select"]',
            gpuMigProfileSelect: '[aid="mig-profile-select"]',
            gpuRequestValueInput: '[aid="gpu-request-input"]',
            gpuLimitValueInput: '[aid="gpu-limit-input"]',
            cpuMemoryRequestInput: '.cpu-memory-request-input',
            cpuCoresRequestInput: '.cpu-request-input',
            cpuCoresLimitInput: '[aid="cpu-limit-input"]',
            increaseSharedMemory: '[aid="shm-toggle"]',
            extendedResourceToggleBtn: '[aid="extended-resources-sub-section"]',
            extendedResourceBtn: '.add-button',
            nodesToggleBtn: '[aid="nodes-sub-section"]',
            nodeAffinityInput: '[aid="node-affinity-input"]',
            autoScaling: {
                sectionSelector: '[aid="compute-section-auto-scale-btn"]',
                minReplicasInput: '[aid="min-rep-input"]',
                maxReplicasInput: '[aid="max-rep-input"]',
                metricSelect: '[aid="threshold-metric-options"]',
                metricThresholdInput: '[aid="threshold-value-input"]',
                scaleDownSelect: '[aid="scale-down-select"]',
                scaleToZeroRetentionSecondsInput: '[aid="scale-to-zero-retention-seconds-input"]'
            },
            workersInput: '[aid="workers-input"]',
            topologyBtn: '[aid="add-topology-btn"]',
            topologyKeyInput: '[aid="topology-key-input"]',
            topologyOperatorSelect: '[aid="topology-operator-select"]'
        };
        this.hybridDataSourceSection = {
            sectionSelector: '[aid="hybrid-data-source-section"]',
            newBtn: '[aid="add-one-off-tab"]',
            newDataSourceOptions: {
                nfs: '[aid="NFS-btn"]',
                volume: '[aid="Volume-btn"]'
            },
            drawer: {
                options: {
                    nfs: '[aid="ds-nfs-option"]',
                    pvc: '[aid="ds-pvc-option"]',
                    s3: '[aid="ds-s3-option"]',
                    git: '[aid="ds-git-option"]',
                    hostPath: '[aid="ds-hostPath-option"]',
                    configMap: '[aid="ds-configMap-option"]'
                }
            },
            configMapInstance: {
                subPathInput: '[aid="mount-subpath-input"]'
            },
            containerPath: '[aid="mount-path-input"]',
            nfs: {
                server: '[aid="server-input"]',
                path: '[aid="path-input"]',
                mountPath: '[aid="mount-path-input"]'
            },
            pvc: {
                storageClass: '[aid="storage-class-select"]',
                accessMode: '[aid="access-mode-select"]',
                claimSize: '[aid="claim-size-input"]',
                unit: '[aid="unit-select"]',
                volumeMode: '[aid="volume-mode-select"]',
                path: '[aid="path-input"]'
            }
        };
        this.nodePoolSection = {
            nodePoolListItem: '.workload-node-pool-list-item',
            addNodePoolBtn: 'button:has-text("NODE POOL")',
            nodePoolDropdownItem: (name: string): string => `.q-menu .q-item:has-text("${name}")`
        };

        this.hybridGeneralSection = {
            sectionSelector: '[aid="general-section"]',
            servingConfigurationSection: {
                initializationTimeoutSeconds: {
                    minutesInput:
                        '.duration-input-section.initialization-timeout-container [aid="minutes-input"]',
                    secondsInput:
                        '.duration-input-section.initialization-timeout-container [aid="seconds-input"]'
                },
                requestTimeoutSeconds: {
                    minutesInput: '.duration-input-section.request-timeout-container [aid="minutes-input"]',
                    secondsInput: '.duration-input-section.request-timeout-container [aid="seconds-input"]'
                }
            },
            preemptibilitySelect: '[aid="preemptibility-select"]',
            startupPolicy: '[aid="workers-startup-policy-select"]',
            prioritySelect: '[aid="priority-select"]'
        };

        this.hybridNimModelSection = {
            modelSourceSelect: '[aid="model-source-selector-select"]',
            modelNameSelectInput: '[aid="nim-autocomplete-select"]',
            modelNameSelectInputHuggingFace: '[aid="hf-autocomplete-select"]',
            apiKey: '[aid="ngc-api-key-input"]',
            profileManually: '[aid="radio-option-Manually"]',
            profileInput: '[aid="profile-input-input"]',
            fromStorageOnly: '[aid="access-from-storage"]',
            accessFromCredentials: '[aid="radio-option-Select credential"]',
            provideAToken: '[aid="radio-option-Provide a key"]',
            credentials: {
                sourceSelect: '.nim-access-model [aid="credential-source-select"]',
                secretNameSelect: '.nim-access-model [aid="credential-list-select"]',
                sourceTypeSelect: '.nim-access-model [aid="credential-type-select"]'
            },
            imagePullCredentials: {
                sourceTypeSelect: '.nim-image-pull-credential-selector [aid="credential-type-select"]',
                secretNameSelect: '.nim-image-pull-credential-selector [aid="credential-list-select"]'
            },
            dockerRegistry: '.nim-access-model [aid="docker-registry-selector"]'
        };

        this.hybridHuggingFaceModelSection = {
            modelNameSelectInput: '[aid="hf-autocomplete-select"]',
            hfToken: '[aid="credentials-token-input"]',
            provideToken: '[aid="radio-option-Provide a token"]',
            selectCredential: '[aid="radio-option-Select credential"]',
            plaintextTokenWarning: '[aid="hugging-face-plaintext-token-warning"]',
            credentials: {
                sourceSelect: '.hugging-face-access-section [aid="credential-source-select"]',
                secretNameSelect: '.hugging-face-access-section [aid="credential-list-select"]'
            },
            serverTypeSubSectionBtn: '[aid="inference-server-btn"]',
            serverTypeTgi: '[aid="radio-option-TGI"]'
        };
    }

    async fillWorkloadName(type: WorkloadType, name: string): Promise<void> {
        switch (type) {
            case WorkloadType.Workspace:
                return this.fillWorkspaceName(name);
            case WorkloadType.Training:
                return this.fillTrainingName(name);
            default: {
                throw new Error('Unsupported workload type');
            }
        }
    }

    async fillWorkspaceName(name: string): Promise<void> {
        await this.page.waitForTimeout(500); // wait some time for the page to load
        return this.page.locator(this.createWorkspaceNameInput).fill(name);
    }

    async fillTrainingName(name: string): Promise<void> {
        await this.page.waitForTimeout(500); // wait some time for the page to load
        return this.page.locator(this.createTrainingNameInput).fill(name);
    }

    async waitForWorkloadNameCount(name: string): Promise<void> {
        await this.waitForWorkloadsCountResponse({ nameInUrl: name });
    }

    async clickOnContinue(type: WorkloadType): Promise<void> {
        // if the quick submit dropdown is visible, we need to continue with it instead of the continue button
        if (await this.isQuickSubmitDropdownVisible()) {
            return this.continueWithQuickSubmit();
        }

        switch (type) {
            case WorkloadType.Workspace:
                return this.page.locator(this.createWorkspaceContinueBtn).click();
            case WorkloadType.Training:
                return this.page.locator(this.createTrainingContinueBtn).click();
            default: {
                throw new Error('Unsupported workload type');
            }
        }
    }

    async isContinueButtonVisible(): Promise<boolean> {
        return this.locator(this.createTrainingContinueBtn).isVisible();
    }

    async isCreateTrainingButtonVisible(): Promise<boolean> {
        return await this.v2frame
            .locator(`${this.quickSubmitDropdownBtnComponent}:has-text("Create training")`)
            .isVisible();
    }

    async isAdvancedSetupButtonVisible(): Promise<boolean> {
        return await this.locator(
            `${this.quickSubmitDropdownBtnComponent}:has-text("Advanced setup")`
        ).isVisible();
    }

    async clickOnAddNewAssetBtn(): Promise<void> {
        return this.page.locator(this.addNewAssetBtn).click();
    }

    async clickOnAddNewDropdownBtnInDrawer(): Promise<void> {
        return this.page.locator(this.addNewDropdownBtn).click();
    }

    async clickOnSubmit(): Promise<void> {
        return this.page.locator(this.submitWorkloadBtn).click();
    }

    async waitForDistributedMasterSubmitButton(): Promise<void> {
        await this.page.waitForSelector(`${this.submitWorkloadBtn}:has-text("continue")`);
        await this.page.waitForTimeout(1000);
    }

    async clickOnSubmitMasterForm(): Promise<void> {
        return this.page.locator(`${this.submitWorkloadBtn}:has-text("continue")`).click();
    }

    async fillImageInput(val: string): Promise<void> {
        return this.fillInputBySelector(this.hybridEnvironmentSection.hybridImageInput, val);
    }

    async isSubSectionOpenedByLabel(label: string): Promise<boolean> {
        const subSection = this.findByText('.runai-folding-sub-section-header', label);
        const classAttribute = await subSection.getAttribute('class');
        return classAttribute?.includes('is-open') ?? false;
    }

    async clickOnSubSectionByLabel(label: string): Promise<void> {
        return this.clickOnElementByText('.runai-folding-sub-section-header', label);
    }

    async clickOnRadioOptionByLabel(label: string): Promise<void> {
        const selector = `div[role="radio"]:has-text("${label}")`;
        logger.info(`Looking for radio option with label: "${label}" (selector: ${selector})`);
        const radio = this.locator(selector);
        await radio.waitFor({ state: 'visible' });
        logger.info(`Found radio option with label "${label}", clicking`);
        await radio.click();
        logger.info(`Clicked on radio option with label "${label}"`);
    }

    async clickOnAddNewToolButton(): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.addToolButton).click();
    }

    async addToolTypeByLabel(label: string, index = 0): Promise<void> {
        await this.clickOnAddNewToolButton();
        await this.page.locator(this.hybridEnvironmentSection.toolTypeSelect).nth(index).click();
        return this.page.getByRole('option', { name: label }).click();
    }

    async clickOnCommandAndArguments(): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.cmdAndArgsBtn).click();
    }

    async fillArgsInput(val: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.argsInput).fill(val);
    }

    async fillCmdInput(val: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.cmdInput).fill(val);
    }

    async selectServingPortProtocol(label: string): Promise<void> {
        return this.selectOptionByValue(this.hybridEnvironmentSection.servingPortProtocol, label);
    }

    async fillServingPortInput(val: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.servingPortPort).fill(val);
    }

    async selectServingPortAccess(servingPortAccessOption: IServingPortAccessOptions): Promise<void> {
        await this.locator(this.hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn).click();
        switch (servingPortAccessOption) {
            case IServingPortAccessOptions.Public:
                await this.locator(
                    this.hybridEnvironmentSection.servingPortAccess.servingPortAccessPublic
                ).click();
                break;
            case IServingPortAccessOptions.Everyone:
                await this.locator(
                    this.hybridEnvironmentSection.servingPortAccess.servingPortAccessEveryone
                ).click();
                break;
            case IServingPortAccessOptions.Group:
                await this.locator(
                    this.hybridEnvironmentSection.servingPortAccess.servingPortAccessGroup
                ).click();
                break;
            case IServingPortAccessOptions.Users:
                await this.locator(
                    this.hybridEnvironmentSection.servingPortAccess.servingPortAccessUsers
                ).click();
                break;
        }

        await this.locator(this.hybridEnvironmentSection.servingPortAccess.servingPortAccessSaveBtn).click();
    }

    // Asserts the connection-access modal's user multi-select is pre-populated with the given users (modal must be open).
    async expectServingPortAuthorizedUsersChips(expectedUsers: string[]): Promise<void> {
        const userChips = this.page.locator(
            this.hybridEnvironmentSection.servingPortAccess.servingPortAccessModalUserChips
        );
        await expect(userChips).toHaveCount(expectedUsers.length);
        for (const expectedUser of expectedUsers) {
            await expect(userChips.filter({ hasText: expectedUser })).toHaveCount(1);
        }
    }

    async fillGpuDevicesInput(val: string): Promise<void> {
        return this.fillInputBySelector(this.hybridComputeSection.gpuDevicesInput, val);
    }

    async focusGpuDevicesInput(): Promise<void> {
        return this.locator(this.hybridComputeSection.gpuDevicesInput).focus();
    }

    async pressEnterOnGpuDevicesInput(): Promise<void> {
        return this.pressKeyOnElement(this.hybridComputeSection.gpuDevicesInput, 'Enter');
    }

    async fillGpuRequestInput(val: string): Promise<void> {
        await this.switchToggleBySelector('[aid="gpu-fraction-toggle"]', true);
        return this.fillInputBySelector(this.hybridComputeSection.gpuRequestValueInput, val);
    }

    async fillCpuMemoryInput(val: string): Promise<void> {
        await this.toggleFoldingSubSection('.cpu-resources-section', true);
        return this.fillInputBySelector(this.hybridComputeSection.cpuMemoryRequestInput, val);
    }

    async fillCpuCoresInput(val: string): Promise<void> {
        await this.toggleFoldingSubSection('.cpu-resources-section', true);
        return this.fillInputBySelector(this.hybridComputeSection.cpuCoresRequestInput, val);
    }

    async openAssetTemplateDrawer(sectionSelector: string): Promise<void> {
        await this.page
            .locator(sectionSelector)
            .locator(this.hybridSectionElement.loadExistingTemplatesButton)
            .click();
        return this.page.waitForTimeout(500);
    }

    async resetSectionTemplate(sectionSelector: string): Promise<void> {
        await this.page
            .locator(sectionSelector)
            .locator(this.hybridSectionElement.resetAssetTemplateButton)
            .click();
        return this.page.waitForTimeout(500);
    }

    async clearSectionTemplate(sectionSelector: string): Promise<void> {
        await this.page
            .locator(sectionSelector)
            .locator(this.hybridSectionElement.clearAssetTemplateButton)
            .click();
        return this.page.waitForTimeout(500);
    }

    async closeDrawer(): Promise<void> {
        return this.page.locator(this.hybridDrawer.closeDrawerBtn).click();
    }

    async openEnvironmentsDrawer(): Promise<void> {
        return this.openAssetTemplateDrawer(this.hybridEnvironmentSection.sectionSelector);
    }

    async openComputesDrawer(): Promise<void> {
        return this.openAssetTemplateDrawer(this.hybridComputeSection.sectionSelector);
    }

    async openDataSourcesDrawer(): Promise<void> {
        return this.openAssetTemplateDrawer(this.hybridDataSourceSection.sectionSelector);
    }

    async openDataSourcesSection(): Promise<void> {
        // Check if already expanded
        const elements = await this.page.$$(this.hybridDataSourceSection.sectionSelector);
        if (elements.length === 1) {
            const isExpanded = await elements[0].evaluate((el) =>
                el.classList.contains('q-expansion-item--expanded')
            );
            if (isExpanded) {
                logger.debug('Data sources section already expanded');
                return;
            }
        }

        // Wait for animation to complete - ensure only one stable element exists
        logger.debug('Waiting for data sources section animation to complete');
        await this.page.waitForFunction(
            (selector) => {
                const elements = document.querySelectorAll(selector);
                return (
                    elements.length === 1 &&
                    (elements[0].classList.contains('q-expansion-item--expanded') ||
                        elements[0].classList.contains('q-expansion-item--collapsed'))
                );
            },
            this.hybridDataSourceSection.sectionSelector,
            { timeout: 5000 }
        );

        // Click only if collapsed
        const element = await this.page.$(this.hybridDataSourceSection.sectionSelector);
        const isCollapsed = await element.evaluate((el) =>
            el.classList.contains('q-expansion-item--collapsed')
        );
        if (isCollapsed) {
            logger.debug('Data sources section is collapsed, clicking to expand it');
            await this.locator(this.hybridDataSourceSection.sectionSelector).click();
        } else {
            logger.debug('Data sources section is already expanded, no need to click');
        }
    }

    async selectDataSourceInDrawer(ds: EDataSourceOption): Promise<void> {
        let aid = '';
        switch (ds) {
            case EDataSourceOption.Nfs:
                aid = this.hybridDataSourceSection.drawer.options.nfs;
                break;
            case EDataSourceOption.Git:
                aid = this.hybridDataSourceSection.drawer.options.git;
                break;
            case EDataSourceOption.Pvc:
                aid = this.hybridDataSourceSection.drawer.options.pvc;
                break;
            default:
                throw new Error(`Could not find the aid for ${ds}`);
        }
        await this.locator(aid).click();
    }

    async addNewDataSource(ds: EDataSourceOption): Promise<void> {
        let aid = '';
        switch (ds) {
            case EDataSourceOption.Nfs:
                aid = this.hybridDataSourceSection.newDataSourceOptions.nfs;
                break;
            case EDataSourceOption.Volume:
                aid = this.hybridDataSourceSection.newDataSourceOptions.volume;
                break;
            default:
                throw new Error(`Could not find the aid for ${ds}`);
        }

        await this.locator(this.hybridDataSourceSection.newBtn).click();
        await this.locator(aid).click();
    }

    async resetEnvironmentSection(): Promise<void> {
        return this.resetSectionTemplate(this.hybridEnvironmentSection.sectionSelector);
    }

    async resetComputesSection(): Promise<void> {
        return this.resetSectionTemplate(this.hybridComputeSection.sectionSelector);
    }

    async resetDataSourcesSection(): Promise<void> {
        return this.resetSectionTemplate(this.hybridDataSourceSection.sectionSelector);
    }

    async clearEnvironmentsSection(): Promise<void> {
        return this.clearSectionTemplate(this.hybridEnvironmentSection.sectionSelector);
    }

    async clearComputesSection(): Promise<void> {
        return this.clearSectionTemplate(this.hybridComputeSection.sectionSelector);
    }

    async clearDataSourcesSection(): Promise<void> {
        return this.clearSectionTemplate(this.hybridDataSourceSection.sectionSelector);
    }

    async clickOnAddEnvVarBtn(): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.addEnvVarBtn).click();
    }

    async clickLastEnvVarSource(): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.envVarSourceSelect).last().click();
    }

    async selectEnvVarSource(label: string): Promise<void> {
        return this.selectOptionByValue(this.hybridEnvironmentSection.envVarSourceSelect, label);
    }

    async clicLastEnvVarSource(index: number = 0): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.envVarSourceSelect).last().click();
    }

    async selectOptionFromDropdown(optionText: string): Promise<void> {
        const optionLocator = this.page.locator(`div[role="option"]:has-text("${optionText}")`);
        await optionLocator.scrollIntoViewIfNeeded();
        await optionLocator.click();
        await this.page.waitForTimeout(500);
    }

    async fillEnvVarName(val: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.envVarNameInput).fill(val);
    }

    async fillLastEnvVarName(val: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.envVarNameInput).last().fill(val);
    }

    async fillLastEnvVarSourceValue(label: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.envVarValue).last().fill(label);
    }

    async selectEnvVarConfigMap(val: string): Promise<void> {
        return this.selectOptionByValue(this.hybridEnvironmentSection.envVarConfigMapSelect, val);
    }

    async fillEnvVarConfigMapKeyInput(val: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.envVarConfigMapKey).fill(val);
    }

    async selectEnvVarSecret(val: string): Promise<void> {
        return this.selectOptionByValue(this.hybridEnvironmentSection.envVarSecretSelect, val);
    }

    async selectEnvVarPersonalSecret(val: string): Promise<void> {
        return this.selectOptionByValue(this.hybridEnvironmentSection.envVarPersonalSecretSelect, val);
    }

    async selectImagePullSecret(val: string): Promise<void> {
        return this.selectOptionByValue(this.hybridEnvironmentSection.imagePullSecretsSelect, val);
    }

    async selectImagePullSecretWithSource(
        source: ECredentialsSource | ECredentialSource,
        credentialName: string,
        credentialType: ECredentialType = ECredentialType.DockerRegistry
    ): Promise<void> {
        // Map source to label
        let sourceLabel: string;
        if (source === ECredentialsSource.Private || source === ECredentialSource.Private) {
            sourceLabel = CREDENTIAL_SOURCE_LABELS[ECredentialSource.Private];
        } else {
            sourceLabel = CREDENTIAL_SOURCE_LABELS[ECredentialSource.Shared];
        }

        await this.selectFromDropDown(
            this.hybridEnvironmentSection.imagePullSecretsSourceSelect,
            sourceLabel
        );

        // Select type (Docker registry / NGC API key / Generic secret)
        const typeSelector = '.form-image-pull-policy .credential-type-select';
        const typeLabel = CREDENTIAL_TYPE_LABELS[credentialType];
        await this.selectFromDropDown(typeSelector, typeLabel);

        // Wait for credential list to refresh after source/type selection
        await this.page.waitForTimeout(500);

        // Select the credential by name
        await this.selectFromDropDown(this.hybridEnvironmentSection.imagePullSecretsSelect, credentialName);
    }

    async selectLastEnvVarSecret(val: string): Promise<void> {
        await this.page.locator(this.hybridEnvironmentSection.envVarSecretSelect).last().click();
        await this.selectOptionFromDropdown(val);
    }

    async selectEnvVarSecretKey(val: string): Promise<void> {
        return this.selectOptionByValue(this.hybridEnvironmentSection.envVarSecretKey, val);
    }

    async fillEnvVarPersonalSecretKey(val: string): Promise<void> {
        return this.page.locator(this.hybridEnvironmentSection.envVarPersonalSecretKeyInput).fill(val);
    }

    // Appends one env var sourced from a personal ("My credentials") credential; .last()-scoped, so
    // repeated calls each target the box they just added. The box's Type filter (default Generic
    // secret) scopes the credential list; only Generic secret renders a key input — omit key otherwise.
    async addEnvVarFromPersonalCredential(
        name: string,
        credentialName: string,
        options: { key?: string; credentialType?: ECredentialType } = {}
    ): Promise<void> {
        const credentialType = options.credentialType ?? ECredentialType.GenericSecret;
        await this.clickOnAddEnvVarBtn();
        await this.fillLastEnvVarName(name);
        await this.clickLastEnvVarSource();
        await this.selectOptionFromDropdown(CREDENTIAL_SOURCE_LABELS[ECredentialSource.Private]);
        if (credentialType !== ECredentialType.GenericSecret) {
            await this.page
                .locator(this.hybridEnvironmentSection.envVarCredentialTypeSelect)
                .last()
                .click();
            await this.selectOptionFromDropdown(CREDENTIAL_TYPE_LABELS[credentialType]);
            // credential list refreshes after the type switch
            await this.page.waitForTimeout(500);
        }
        await this.page
            .locator(this.hybridEnvironmentSection.envVarPersonalSecretSelect)
            .last()
            .click();
        await this.selectOptionFromDropdown(credentialName);
        if (options.key !== undefined) {
            await this.page
                .locator(this.hybridEnvironmentSection.envVarPersonalSecretKeyInput)
                .last()
                .fill(options.key);
        }
    }

    async selectLastEnvVarSecretKey(val: string): Promise<void> {
        // Access key / password secrets expose their keys as a select, any other kind lets you type the key.
        const { envVarSecretKey, envVarCustomKey } = this.hybridEnvironmentSection;
        await this.page.waitForSelector(`${envVarSecretKey}, ${envVarCustomKey}`, { state: 'visible' });

        const keySelect = this.page.locator(envVarSecretKey).last();
        if ((await keySelect.count()) > 0) {
            await keySelect.click();
            await this.selectOptionFromDropdown(val);
            return;
        }
        await this.page.locator(envVarCustomKey).last().fill(val);
    }

    async getSelectValueBySelector(selector: string): Promise<string> {
        return this.page.locator(selector).locator('span').innerText();
    }

    getRadioOptionLocatorByLabel(label: string): string {
        return `[aid="radio-option-${label}"]`;
    }

    async fillLastEnvironmentVariablesCustomSource(sourcevalue: string): Promise<void> {
        await this.selectOptionFromDropdown('Custom');
        await this.fillLastEnvVarSourceValue(sourcevalue);
    }

    async fillLastEnvironmentVariablesSecretSource(sourcename: string, sourcevalue: string): Promise<void> {
        await this.selectOptionFromDropdown('Secret');
        await this.selectLastEnvVarSecret(sourcename);
        await this.selectLastEnvVarSecretKey(sourcevalue);
    }

    async fillLastEnvironmentVariablesConfigMapSource(
        sourcename: string,
        sourcevalue: string
    ): Promise<void> {
        // need to implement
    }

    async fillEnvironmentVariables(
        envVars: Array<{ name: string; sourcetype: string; sourcename?: string; sourcevalue: string }>
    ): Promise<void> {
        const envVarBtnVisible = await this.page
            .locator(this.hybridEnvironmentSection.addEnvVarBtn)
            .isVisible();
        if (!envVarBtnVisible) {
            await this.clickOnSubSectionByLabel('Runtime settings');
        }
        for (const envVar of envVars) {
            await this.clickOnAddEnvVarBtn();
            await this.fillLastEnvVarName(envVar.name);
            await this.clickLastEnvVarSource();
            if (envVar.sourcetype === 'Custom') {
                await this.fillLastEnvironmentVariablesCustomSource(envVar.sourcevalue);
            } else if (envVar.sourcetype === 'Secret') {
                await this.fillLastEnvironmentVariablesSecretSource(envVar.sourcename, envVar.sourcevalue);
            } else if (envVar.sourcetype === 'ConfigMap') {
                await this.fillLastEnvironmentVariablesConfigMapSource(envVar.sourcename, envVar.sourcevalue);
            } else {
                throw new Error(`Unsupported environment variable source: ${envVar.sourcetype}`);
            }
        }
    }

    async clickOnDrawerItemByLabel(label: string, waitForUnifiedRequest = true): Promise<void> {
        try {
            const unifiedResponse = this.page.waitForResponse((response) =>
                response.url().includes('/api/v1/workloads/unified')
            );
            await this.page
                .locator(this.hybridDrawer.drawerItemLabel)
                .getByText(label, { exact: true })
                .click();

            if (waitForUnifiedRequest) {
                const finishedResponse = await unifiedResponse;
                expect(finishedResponse.status()).toEqual(200);
            }
            await this.page.waitForTimeout(1000);
        } catch (error) {
            console.error(`Error clicking on drawer item by label: ${label}`, error);
            throw error;
        }
    }

    async clickOnLoadFromDrawer(): Promise<void> {
        await this.locator(this.hybridDrawer.selectDrawerItemsBtn).click();
    }

    async toggleIncreaseSharedMemory(): Promise<void> {
        await this.locator(this.hybridComputeSection.increaseSharedMemory).click();
    }

    async openExtendedResourceSection(): Promise<void> {
        await this.locator(this.hybridComputeSection.extendedResourceToggleBtn).click();
    }

    async fillExtendedResourcesValues(
        resources: Array<{ resource: string; quantity: string }>
    ): Promise<void> {
        for (const el of resources.entries()) {
            await this.page
                .locator(this.hybridComputeSection.extendedResourceBtn, {
                    hasText: 'EXTENDED RESOURCE'
                })
                .click();
        }

        const nameElements = await this.locator(
            '.extended-resources input[type="text"][aria-label="Resource"]'
        ).all();
        const quantityElements = await this.locator(
            '.extended-resources input[type="text"][aria-label="Quantity"]'
        ).all();

        if (nameElements.length !== quantityElements.length) {
            throw new Error(
                `Mismatch in name and values: found ${nameElements.length} name inputs, but found ${quantityElements.length} quantities.`
            );
        }
        if (resources.length !== quantityElements.length) {
            throw new Error(
                `Mismatch in name and quantities and resources: Expected ${resources.length} elements, but found ${quantityElements.length} names/quantities.`
            );
        }
        for (const [index, element] of nameElements.entries()) {
            await element.fill(resources[index].resource);
        }
        for (const [index, element] of quantityElements.entries()) {
            await element.fill(resources[index].quantity);
        }
    }

    async openNodesSection(): Promise<void> {
        await this.locator(this.hybridComputeSection.nodesToggleBtn).click();
        await this.page.waitForSelector(this.nodePoolSection.nodePoolListItem, {
            state: 'visible',
            timeout: 5000
        });
    }

    async fillNodeAffinityInput(val: string): Promise<void> {
        return this.locator(this.hybridComputeSection.nodeAffinityInput).fill(val);
    }

    async fillConfigMapSubPathInput(val: string): Promise<void> {
        return this.page.locator(this.hybridDataSourceSection.configMapInstance.subPathInput).fill(val);
    }

    async fillSeachInputInDrawer(val: string): Promise<void> {
        await this.page.locator(this.searchInputDrawer).fill(val);
    }

    async fillHybridFormByLoadingExistingConfig(
        existingEnvName: string,
        existingComputeResName: string,
        existingDsOption?: { dsName: string; containerPath?: string },
        existingEnvVars?: Array<{
            name: string;
            sourcetype: string;
            sourcename?: string;
            sourcevalue: string;
        }>
    ): Promise<void> {
        await this.page.waitForLoadState('load');
        await this.selectExistingEnv(existingEnvName);
        await this.selectExistingComputeResource(existingComputeResName);

        if (existingDsOption) {
            await this.openDataSourcesSection();
            await this.openDataSourcesDrawer();
            await this.fillSeachInputInDrawer(existingDsOption.dsName);
            await this.page.locator(`text="${existingDsOption.dsName}"`).click();
            await this.page.locator(this.hybridDrawer.selectDrawerItemsBtn).click();

            if (existingDsOption.containerPath) {
                await this.page
                    .locator(this.hybridDataSourceSection.containerPath)
                    .fill(existingDsOption.containerPath);
            }
        }

        if (existingEnvVars?.length) {
            await this.fillEnvironmentVariables(existingEnvVars);
        }
    }

    async selectExistingEnv(existingEnvName: string): Promise<void> {
        await this.openEnvironmentsDrawer();
        await this.fillSeachInputInDrawer(existingEnvName);
        await this.page.locator(`text="${existingEnvName}"`).click();
    }

    async selectExistingComputeResource(existingComputeResName: string): Promise<void> {
        await this.openComputesDrawer();
        await this.fillSeachInputInDrawer(existingComputeResName);
        await this.page.locator(`text="${existingComputeResName}"`).click();
    }

    async isNodePoolSelected(nodePoolName: string): Promise<boolean> {
        const selectedNodePools = await this.page
            .locator(this.nodePoolSection.nodePoolListItem)
            .allTextContents();
        return selectedNodePools.some((name) => name.includes(nodePoolName));
    }

    async addNodePoolIfNotSelected(nodePoolName: string): Promise<void> {
        if (!(await this.isNodePoolSelected(nodePoolName))) {
            await this.page.locator(this.nodePoolSection.addNodePoolBtn).click();
            await this.page.locator(this.nodePoolSection.nodePoolDropdownItem(nodePoolName)).click();
            await this.page.waitForTimeout(500);
            if (!(await this.isNodePoolSelected(nodePoolName))) {
                throw new Error(`Failed to select node pool: ${nodePoolName}`);
            }
        }
    }

    async dragNodePoolToTop(nodePoolName: string): Promise<void> {
        const allItems = this.page.locator(this.nodePoolSection.nodePoolListItem);

        await waitUtils.waitForCondition(
            async () => {
                const names = await allItems.allTextContents();
                const fromIndex = names.findIndex((name) => name.includes(nodePoolName));

                if (fromIndex === -1) {
                    throw new Error(`Node pool not found in selected list: ${nodePoolName}`);
                }
                if (fromIndex === 0) {
                    return true;
                }

                await this.page.waitForSelector(this.nodePoolSection.nodePoolListItem, { state: 'visible' });
                await allItems.nth(fromIndex).dragTo(allItems.nth(0));
                logger.warn(`dragNodePoolToTop: ${nodePoolName} is not at top yet, retrying`);

                return false;
            },
            `wait for node pool ${nodePoolName} to be dragged to the top`,
            true,
            15,
            1
        );
    }

    async ensureNodePoolSelectedAndOnTop(nodePoolName: string): Promise<void> {
        await this.addNodePoolIfNotSelected(nodePoolName);
        if (!(await this.isNodePoolSelected(nodePoolName))) {
            throw new Error(`Node Pool not selected: ${nodePoolName}`);
        }
        await this.dragNodePoolToTop(nodePoolName);
    }

    async fillAndSubmitTrainingWorkloadForm(
        workloadName: string,
        projectName: string,
        clusterName: string,
        options: {
            nodePoolList?: string[];
            nodeAffinity?: string;
            envTemplate?: string;
            computeTemplate?: string;
            image?: string;
            cmdInput?: string;
            argsInput?: string;
        } = {}
    ): Promise<void> {
        await this.openWorkloadForm(WorkloadType.Training);
        await this.selectClusterFromDropDown(clusterName);
        await this.workloadFormSelectProject(projectName);
        await this.workloadFormSelectTemplate();
        await this.fillWorkloadName(WorkloadType.Training, workloadName);
        await this.waitForWorkloadNameCount(workloadName);
        await this.clickOnContinue(WorkloadType.Training);
        if (options.image) {
            await this.fillImageSection(options.image);
        } else {
            if (!options.envTemplate) {
                throw new Error(
                    'envTemplate is required when image is not provided — use environmentUtils.presetNames() to resolve the correct name for the cluster routing'
                );
            }
            await this.openEnvironmentsDrawer();
            await this.clickOnDrawerItemByLabel(options.envTemplate);
        }
        options.computeTemplate = options.computeTemplate || 'one-gpu';
        await this.openComputesDrawer();
        await this.clickOnDrawerItemByLabel(options.computeTemplate);
        await this.openNodesSection();
        if (options.nodePoolList) {
            for (const nodePoolName of options.nodePoolList) {
                await this.addNodePoolIfNotSelected(nodePoolName);
            }
            await this.ensureNodePoolSelectedAndOnTop(options.nodePoolList[0]);
        }
        if (options.nodeAffinity) {
            await this.fillNodeAffinityInput(options.nodeAffinity);
        }
        if (options.cmdInput || options.argsInput) {
            await this.fillRuntimeSettings(options.cmdInput, options.argsInput);
        }
        await this.clickOnSubmit();
    }

    async fillImageSection(image: string, imagePullPolicy?: string): Promise<void> {
        await this.fillImageInput(image);
        await this.clickOnSubSectionByLabel(this.imagePullSection);
        await this.clickOnRadioOptionByLabel(imagePullPolicy || 'Always pull the image from the registry');
    }

    async fillToolsSection(toolName: string): Promise<void> {
        await this.clickOnSubSectionByLabel('Tools');
        await this.addToolTypeByLabel(toolName);
    }

    async fillRuntimeSettings(command?: string, args?: string): Promise<void> {
        await this.clickOnSubSectionByLabel('Runtime settings');
        const cmdInputExists = (await this.locator(this.hybridEnvironmentSection.cmdInput).count()) > 0;
        if (!cmdInputExists) {
            await this.clickOnCommandAndArguments();
        }
        if (command) {
            await this.fillCmdInput(command);
        }
        if (args) {
            await this.page.waitForTimeout(1000);
            await this.fillArgsInput(args);
        }
    }

    async clearRuntimeSettings(): Promise<void> {
        await this.clickOnSubSectionByLabel('Runtime settings');
        const cmdInputExists = (await this.locator(this.hybridEnvironmentSection.cmdInput).count()) > 0;
        if (cmdInputExists) {
            await this.fillCmdInput('');
            await this.fillArgsInput('');
        }
    }

    async fillPvcSection(pvcName: string): Promise<void> {
        await this.openDataSourcesSection();
        await this.openDataSourcesDrawer();
        await this.clickOnDrawerItemByLabel(pvcName, false);
        await this.clickOnLoadFromDrawer();
    }

    async fillContainerPathOverrideInput(containerPath: string): Promise<void> {
        await this.page.locator(this.hybridDataSourceSection.containerPath).fill(containerPath);
    }

    async fillComputeSection(computeLabel?: string, workersCount?: number): Promise<void> {
        if (computeLabel) {
            await this.openComputesDrawer();
            await this.clickOnDrawerItemByLabel(computeLabel);
        }
        if (workersCount) {
            await this.fillInputBySelector(this.hybridComputeSection.workersInput, workersCount.toString());
        }
    }

    async selectFromDropDownWithInput(selector: string, value: string): Promise<void> {
        await this.locator(selector).fill(value);
        await this.page.click(`css=.q-item[role="option"]:has-text("${value}")`);
    }

    async selectFromDropDown(selector: string, value: string): Promise<void> {
        await this.locator(selector).click();
        await this.page.click(`css=.q-item[role="option"]:has-text("${value}")`);
    }

    async fillHuggingFaceToken(hfToken: string): Promise<void> {
        const input = this.locator(this.hybridHuggingFaceModelSection.hfToken);
        await input.fill(hfToken);
        // Blur to flush the q-input 500ms debounce so modelValue.token commits before any re-render clears it.
        await input.blur();
    }

    async fillNimApiKey(apiKey: string): Promise<void> {
        await this.locator(this.hybridNimModelSection.apiKey).fill(apiKey);
    }

    async selectProfileManuallyAndFillTheProfile(profileOrHash: string): Promise<void> {
        await this.locator(this.hybridNimModelSection.profileManually).click();
        await this.locator(this.hybridNimModelSection.profileInput).fill(profileOrHash);
    }

    async selectNimAccessModelFromStorageOnly(): Promise<void> {
        await this.locator(this.hybridNimModelSection.fromStorageOnly).click();
    }

    async selectHuggingFaceTokenFromCredentials(): Promise<void> {
        await this.locator(this.hybridHuggingFaceModelSection.selectCredential).click();
    }

    async selectHuggingFaceTokenManually(): Promise<void> {
        await this.locator(this.hybridHuggingFaceModelSection.provideToken).click();
    }

    async selectNimAccessFromCredentials(): Promise<void> {
        await this.locator(this.hybridNimModelSection.accessFromCredentials).click();
    }

    async selectNimAccessFromProvideAToken(): Promise<void> {
        await this.locator(this.hybridNimModelSection.provideAToken).click();
    }

    async selectHuggingFaceCredentials(
        source: ECredentialsSource,
        secretName: string,
        sourceDisabled?: boolean
    ): Promise<void> {
        await this.selectCredentials(
            this.hybridHuggingFaceModelSection.credentials.sourceSelect,
            this.hybridHuggingFaceModelSection.credentials.secretNameSelect,
            secretName,
            source,
            undefined,
            undefined,
            sourceDisabled
        );
    }

    async selectNimCredentials(
        source: ECredentialsSource,
        secretName: string,
        sourceType: ECredentialsSourceType,
        sourceDisabled?: boolean
    ): Promise<void> {
        await this.selectCredentials(
            this.hybridNimModelSection.credentials.sourceSelect,
            this.hybridNimModelSection.credentials.secretNameSelect,
            secretName,
            source,
            this.hybridNimModelSection.credentials.sourceTypeSelect,
            sourceType,
            sourceDisabled
        );
    }

    async selectCredentials(
        sourceSelect: string,
        secretNameSelect: string,
        secretName: string,
        source?: ECredentialsSource,
        sourceTypeSelect?: string,
        sourceTypeValue?: string,
        sourceDisabled?: boolean
    ): Promise<void> {
        if (sourceDisabled) {
            const sourceValue = await this.locator(`${sourceSelect} .selected-value-container`).textContent();
            if (source === ECredentialsSource.Private) {
                if (sourceValue !== 'My credentials') {
                    throw new Error(`Invalid source: ${sourceValue} for my credentials`);
                } else {
                    logger.info('sourceValue is My credentials');
                }
            } else if (source === ECredentialsSource.Shared) {
                if (sourceValue !== 'Shared secret') {
                    throw new Error(`Invalid source: ${sourceValue} for shared credentials`);
                } else {
                    logger.info('sourceValue is Shared secret');
                }
            } else {
                throw new Error(`Invalid source: ${sourceValue}`);
            }
        } else {
            if (source === ECredentialsSource.Private) {
                await this.selectFromDropDown(sourceSelect, 'My credentials');
            } else if (source === ECredentialsSource.Shared) {
                await this.selectFromDropDown(sourceSelect, 'Shared secret');
            } else {
                throw new Error(`Invalid source: ${source}`);
            }
            // Wait for credential list to refresh after source selection
            await this.page.waitForTimeout(500);
        }

        if (sourceTypeValue === ECredentialsSourceType.GenericSecret) {
            if (!sourceTypeSelect) {
                throw new Error('sourceTypeSelect is required when sourceTypeValue is provided');
            }
            await this.selectFromDropDown(sourceTypeSelect, 'Generic secret');
        }

        await this.selectFromDropDown(secretNameSelect, secretName);
    }

    async fillWorkloadCreationFirstPage(
        clusterName: string,
        projectName: string,
        workloadName: string,
        workloadType: WorkloadType = WorkloadType.Workspace
    ): Promise<void> {
        if (workloadType === WorkloadType.Training) {
            await this.openTrainingForm();
        } else {
            await this.openWorkspaceForm();
        }

        await this.selectClusterFromDropDown(clusterName);
        await this.workloadFormSelectProject(projectName);
        await this.workloadFormSelectTemplate();

        if (workloadType === WorkloadType.Training) {
            await this.fillTrainingName(workloadName);
        } else {
            await this.fillWorkspaceName(workloadName);
        }

        await this.waitForWorkloadNameCount(workloadName);
        await this.clickOnContinue(workloadType);
    }

    async isQuickSubmitDropdownVisible(): Promise<boolean> {
        return await this.page.locator(this.quickSubmitDropdownBtnComponent).isVisible();
    }

    async openQuickSubmitDropdown(): Promise<void> {
        await this.page.locator(this.quickSubmitDropdownBtn).click();
    }

    async clickOnQuickSubmitDropdownSubmitOption(): Promise<void> {
        await this.page.locator(this.quickSubmitDropdownSubmitOption).click();
    }

    async isQuickSubmitDropdownSubmitOptionDisabled(): Promise<boolean> {
        return (
            (await this.page
                .locator(this.quickSubmitDropdownSubmitOption)
                .locator('..')
                .getAttribute('aria-disabled')) === 'true'
        );
    }

    async isQuickSubmitDropdownContinueOptionDisabled(): Promise<boolean> {
        return (
            (await this.page
                .locator(this.quickSubmitDropdownContinueOption)
                .locator('..')
                .getAttribute('aria-disabled')) === 'true'
        );
    }

    async clickOnQuickSubmitDropdownContinueOption(): Promise<void> {
        await this.page.locator(this.quickSubmitDropdownContinueOption).click();
    }

    async continueWithQuickSubmit(): Promise<void> {
        await this.openQuickSubmitDropdown();
        await this.clickOnQuickSubmitDropdownContinueOption();
    }

    async submitWithQuickSubmit(): Promise<void> {
        await this.openQuickSubmitDropdown();
        await this.clickOnQuickSubmitDropdownSubmitOption();
    }

    async quickSubmitMainButton(): Promise<void> {
        await this.page.locator(`${this.quickSubmitDropdownBtnComponent} button`).first().click();
    }

    async openServingPortSection(): Promise<void> {
        await this.locator(this.hybridEnvironmentSection.servingPortAccess.servingPortAccessToggle).click();
    }

    async openSecuritySection(): Promise<void> {
        await this.locator(this.hybridEnvironmentSection.securityToggle).click();
    }

    async openRuntimeSection(): Promise<void> {
        await this.locator(this.hybridEnvironmentSection.runtimeToggle).click();
    }

    async openAutoscalingSection(): Promise<void> {
        await this.locator(this.hybridComputeSection.autoScaling.sectionSelector).click();
    }

    async setTopology(topoKey: string, operator: string): Promise<void> {
        await this.locator(this.hybridComputeSection.topologyBtn).click();
        await this.locator(this.hybridComputeSection.topologyKeyInput).fill(topoKey);
        await this.locator(this.hybridComputeSection.topologyOperatorSelect).click();
        await this.page.getByRole('option', { name: operator }).click();
    }

    async openImagePullPolicySection(): Promise<void> {
        await this.locator(this.hybridEnvironmentSection.imagePullPolicySubSection).click();
    }

    async selectPreemptibility(preemptibility: string): Promise<void> {
        await this.selectOptionByExactValue(this.hybridGeneralSection.preemptibilitySelect, preemptibility);
    }

    async selectStartupPolicy(startupPolicy: string): Promise<void> {
        await this.selectOptionByValue(this.hybridGeneralSection.startupPolicy, startupPolicy);
    }

    async selectPriority(priority: string): Promise<void> {
        await this.selectOptionByExactValue(this.hybridGeneralSection.prioritySelect, priority);
    }
}
