/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { AssetKind, Scope } from '../../../api/apis/assetsApi';
import { dataSourcesUtils, logger, objectGenerator, waitUtils } from '../../../helpers';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';
import { Response } from 'playwright-core';

export class DataSourcePage extends BaseV2Page {
    // components
    assetsScopeSection: AssetsScopeSection;

    // Data source table elements
    addNewDataSourceBtn: string;

    addNewS3Btn: string;

    addNewHostPathBtn: string;

    addNewNFSBtn: string;

    addNewGitBtn: string;

    addNewPvcBtn: string;

    addNewConfigMapBtn: string;

    addNewSecretBtn: string;

    deleteDataSourceBtn: string;

    copyEditDataSourceBtn: string;

    renameDataSourceBtn: string;

    renameModalNameInput: string;

    renameModalSaveBtn: string;

    deleteModalDeleteBtn: string;

    // S3 creation elements
    s3NameInput: string;

    s3ServiceUrlInput: string;

    s3CredentialsSelect: string;

    s3BucketInput: string;

    s3ContainerPathInput: string;

    // Host path creation elements
    hostPathNameInput: string;

    hostPathHostInput: string;

    hostPathContainerInput: string;

    // NFS creation elements
    nfsNameInput: string;

    nfsServerInput: string;

    nfsPathInput: string;

    nfsMountPath: string;

    nfsContainerInput: string;

    // Git creation elements

    gitNameInput: string;

    gitRepositoryInput: string;

    gitCredentialsSelect: string;

    gitBranchInput: string;

    gitRevisionInput: string;

    gitContainerInput: string;

    gitPrivateRadioOption: string;

    gitPublicRadioOption: string;

    // PVC creation elements
    PvcNameInput: string;

    PvcRadioNewPvcOption: string;

    PvcStorageClassSelect: string;

    PvcAccessModeSelect: string;

    PvcClaimSizeInput: string;

    PvcClaimUnitSelect: string;

    PvcVolumeModeSelect: string;

    PvcContainerInput: string;

    // ConfigMap creation elements
    ConfigMapNameAssetInput: string;

    ConfigMapResourceSelect: string;

    ConfigMapNameMountPathInput: string;

    // Secret creation elements
    SecretNameAssetInput: string;

    SecretMountPathInput: string;

    CredentialResourceSelect: string;

    // Data source creation elements

    submitBtn: string;

    cancelBtn: string;

    leaveBtn: string;

    existingPvcSelect: string;

    nameAlreadyExistsError: string;

    dataSourceNameErrorDiv: string;

    dataSourceNameErrorMsg: string;

    constructor(page: Page) {
        super(page);

        this.assetsScopeSection = new AssetsScopeSection(page);

        this.mainPageTab = '[aid="menu-icon-data-sources"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.DataSources}"]`;

        // Data source table elements
        this.addNewDataSourceBtn = '.runai-page-actions button[aid="primary-cta-btn"]';
        this.addNewS3Btn = '[aid="S3 Bucket-btn"]';
        this.addNewHostPathBtn = '[aid="Host path-btn"]';
        this.addNewNFSBtn = '[aid="NFS-btn"]';
        this.addNewGitBtn = '[aid=Git-btn]';
        this.addNewPvcBtn = '[aid=PVC-btn]';
        this.addNewConfigMapBtn = '[aid=ConfigMap-btn]';
        this.addNewSecretBtn = '[aid=Secret-btn]';
        this.deleteDataSourceBtn = '[aid="delete-data-source-btn"]';
        this.copyEditDataSourceBtn = '[aid="copy-edit-data-source-btn"]';
        this.renameDataSourceBtn = '[aid="rename-data-source-btn"]';
        this.renameModalNameInput = '[aid="runai-name-validation"]';
        this.renameModalSaveBtn = '[aid="rename-modal-save-button"]';
        this.deleteModalDeleteBtn = '[aid="delete-modal-delete-button"]';

        // S3 creation elements
        this.s3NameInput = `[aid="${AssetKind.S3}-name-input"]`;
        this.s3ServiceUrlInput = '[aid="s3-service-url-input"]';
        this.s3CredentialsSelect = '[aid="s3-credentials-select"]';
        this.s3BucketInput = '[aid="s3-bucket-input"]';
        this.s3ContainerPathInput = '[aid="s3-container-path-input"]';

        // Host path creation elements
        this.hostPathNameInput = `[aid="${AssetKind.HostPath}-name-input"]`;
        this.hostPathHostInput = '[aid="host-path-input"]';
        this.hostPathContainerInput = '[aid="mount-path-input"]';

        // NFS creation elements
        this.nfsNameInput = `[aid="${AssetKind.Nfs}-name-input"]`;
        this.nfsServerInput = '[aid="server-input"]';
        this.nfsPathInput = '[aid="path-input"]';
        this.nfsContainerInput = '[aid="mount-path-input"]';

        // Git creation elements
        this.gitNameInput = `[aid="${AssetKind.Git}-name-input"]`;
        this.gitRepositoryInput = '[aid="repository-input"]';
        this.gitCredentialsSelect = '[aid="credentials-select"]';
        this.gitBranchInput = '[aid="git-branch-input"]';
        this.gitRevisionInput = '[aid="git-revision-input"]';
        this.gitContainerInput = '[aid="git-container-input"]';
        this.gitPrivateRadioOption = 'div[role="radio"]:has-text("Private")';
        this.gitPublicRadioOption = 'div[role="radio"]:has-text("Public")';

        // PVC creation elements
        this.PvcNameInput = `[aid="${AssetKind.Pvc}-name-input"]`;
        this.PvcRadioNewPvcOption = 'div[role="radio"]:has-text("New PVC")';
        this.PvcStorageClassSelect = '[aid=storage-class-select]';
        this.PvcAccessModeSelect = '[aid=access-mode-select]';
        this.PvcClaimSizeInput = '[aid=claim-size-input]';
        this.PvcClaimUnitSelect = '[aid=unit-select]';
        this.PvcVolumeModeSelect = '[aid=volume-mode-select]';
        this.PvcContainerInput = '[aid=container-path-input]';
        this.existingPvcSelect = '[class="pvc-data-mount-existing"]';
        this.dataSourceNameErrorDiv =
            'div[aid="data source name & description-expansion-toggle"] div[role="alert"]';
        this.dataSourceNameErrorMsg =
            'Enter a name starting with a lowercase letter and including only a-z, 0-9, -';

        // ConfigMap creation elements
        this.ConfigMapNameAssetInput = '[aid="config-map-name-input"]';
        this.ConfigMapResourceSelect = '.configmap-name-select';
        this.ConfigMapNameMountPathInput = '[aid="mount-path-input"]';

        // Secret creation elements
        this.SecretNameAssetInput = '[aid="secret-volume-name-input"]';
        // TODO replace the selecter name, Wait for FE to proivde a better name
        this.CredentialResourceSelect = '[aid="s3-credentials-select"]';
        this.SecretMountPathInput = '[aid="mount-path-input"]';

        // Data source creation elements
        this.submitBtn = '[aid="submit-btn"]';
        this.cancelBtn = '[aid="cancel-btn"]';
        this.leaveBtn = '[aid="leave-page-modal-cancel-btn"]';
        this.nameAlreadyExistsError = 'div[role="alert"]:has-text("This name already exists")';
    }

    async validateNewDataSourceButtonExist(): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const dataSourceBtnCount = await this.v2frame.locator(this.addNewDataSourceBtn).count();
                return dataSourceBtnCount !== 0;
            },
            'is new data source button exists',
            true,
            10
        );
    }

    async clickOnAddNewDataSourceBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewDataSourceBtn).click();
    }

    async clickOnAddS3Btn(): Promise<void> {
        return await this.v2frame.locator(this.addNewS3Btn).click();
    }

    async selectProjectByName(projectName: string): Promise<void> {
        return this.v2frame.locator(`div[role="option"]:has-text("${projectName}")`).click();
    }

    async fillS3NameInput(s3Name: string): Promise<void> {
        return await this.v2frame.locator(this.s3NameInput).fill(s3Name);
    }

    async fillS3ServiceUrlInput(s3ServiceUrl: string): Promise<void> {
        return await this.v2frame.locator(this.s3ServiceUrlInput).fill(s3ServiceUrl);
    }

    async selectS3CredentialsByName(credentialName: string): Promise<void> {
        await this.v2frame.locator(this.s3CredentialsSelect).click();
        return await this.v2frame.locator(`div[role="option"]:has-text("${credentialName}")`).click();
    }

    async fillS3BucketInput(s3Bucket: string): Promise<void> {
        return await this.v2frame.locator(this.s3BucketInput).fill(s3Bucket);
    }

    async fillS3ContainerPathInput(s3ContainerPath: string): Promise<void> {
        return await this.v2frame.locator(this.s3ContainerPathInput).fill(s3ContainerPath);
    }

    async clickOnSubmitBtn(): Promise<void> {
        return await this.v2frame.locator(this.submitBtn).click();
    }

    async clickOnCancelBtn(): Promise<void> {
        return await this.page.locator(this.cancelBtn).click();
    }

    async clickOnLeaveBtn(): Promise<void> {
        return await this.page.locator(this.leaveBtn).click();
    }

    async clickOnDeleteDataSourceBtn(): Promise<void> {
        return this.v2frame.locator(this.deleteDataSourceBtn).click();
    }

    async clickOnConfirmDeleteBtn(): Promise<void> {
        return this.v2frame.locator(this.deleteModalDeleteBtn).click();
    }

    async clickOnRenameDataSourceBtn(): Promise<void> {
        return this.v2frame.locator(this.renameDataSourceBtn).click();
    }

    async fillRenameModalNameInput(newName: string): Promise<void> {
        return this.v2frame.locator(this.renameModalNameInput).fill(newName);
    }

    async clickOnRenameModalSaveBtn(): Promise<void> {
        return this.v2frame.locator(this.renameModalSaveBtn).click();
    }

    async deleteDataSource(name: string): Promise<void> {
        await this.searchByName(name);
        await this.selectDataSourceInTableByName(name);
        await this.clickOnDeleteDataSourceBtn();
        await this.clickOnConfirmDeleteBtn();
        await dataSourcesUtils.waitForDataSourceToBeDeleted(name);
        await this.waitForAssetToBeDeletedFromUI(this, name);
    }

    async selectDataSourceInTableByName(dataSourceName?: string): Promise<void> {
        await this.searchByName(dataSourceName);
        await this.findByText('td[aid="data-source"]', dataSourceName).click();
    }

    async checkDataSourceExistsInTableByName(dataSourceName: string): Promise<void> {
        await this.searchByName(dataSourceName);
        await this.findByText('td', dataSourceName).click();
    }

    async selectFirstProject(): Promise<void> {
        await this.v2frame.locator('.q-tree__node-header >> nth=0').click();
        await this.v2frame.locator('.q-tree__children > div > div >> nth=0').click();
        await this.v2frame.locator('div[role="checkbox"] >> nth=2').click();
    }

    async findDataSourceStatus(dsName: string): Promise<string> {
        await this.searchByName(dsName);
        await waitUtils.waitForCondition(
            async () => {
                const rowsNum = await this.v2frame.locator('.runai-table-row').count();
                return rowsNum === 1;
            },
            'waiting for filter to apply',
            true,
            10
        );
        return (await this.v2frame.locator('.q-td[aid="status"]').textContent()).trim();
    }

    // Host path
    async clickOnAddNewHostPathBtn(): Promise<void> {
        return this.v2frame.locator(this.addNewHostPathBtn).click();
    }

    async fillHostPathNameInput(hostPathName: string): Promise<void> {
        return this.v2frame.locator(this.hostPathNameInput).fill(hostPathName);
    }

    async fillHostPathHostInput(hostPathHost: string): Promise<void> {
        return this.v2frame.locator(this.hostPathHostInput).fill(hostPathHost);
    }

    async fillHostPathContainerInput(hostPathContainer: string): Promise<void> {
        return this.v2frame.locator(this.hostPathContainerInput).fill(hostPathContainer);
    }

    async openCopyAndEditForm(dataSourceName: string): Promise<void> {
        await this.selectDataSourceInTableByName(dataSourceName);
        await this.clickOnCopyEditSourceBtn();
        await this.waitForNameAlreadyExistsErrorToAppear();
    }

    async clickOnCopyEditSourceBtn(): Promise<void> {
        return this.v2frame.locator(this.copyEditDataSourceBtn).click();
    }

    // NFS
    async clickOnAddNewNFSBtn(): Promise<void> {
        return this.locator(this.addNewNFSBtn).click();
    }

    async fillNFSNameInput(nfsName: string): Promise<void> {
        await this.locator(this.nfsNameInput).fill(nfsName);
        return await this.waitForNameCount(nfsName);
    }

    async waitForNameCount(name: string): Promise<void> {
        function isFinished(response: Response): boolean {
            return (
                response.url().includes('api/v1/asset/name-availability') &&
                response.url().includes(name) &&
                response.status() >= 200 &&
                response.status() < 300
            );
        }

        await this.waitForResponseByCallback(isFinished);
    }

    async fillNFSServerInput(nfsServer: string): Promise<void> {
        return this.v2frame.locator(this.nfsServerInput).fill(nfsServer);
    }

    async fillNFSPathInput(nfsPath: string): Promise<void> {
        return this.v2frame.locator(this.nfsPathInput).fill(nfsPath);
    }

    async fillNFSContainerInput(nfsContainer: string): Promise<void> {
        return this.v2frame.locator(this.nfsContainerInput).fill(nfsContainer);
    }

    // GIT
    async clickOnAddNewGitBtn(): Promise<void> {
        return this.v2frame.locator(this.addNewGitBtn).click();
    }

    async fillGitNameInput(gitName: string): Promise<void> {
        return this.v2frame.locator(this.gitNameInput).fill(gitName);
    }

    async fillGitRepositoryInput(gitRepository: string): Promise<void> {
        return this.v2frame.locator(this.gitRepositoryInput).fill(gitRepository);
    }

    async fillGitBranchInput(gitBranch: string): Promise<void> {
        return this.v2frame.locator(this.gitBranchInput).fill(gitBranch);
    }

    async fillGitRevisionInput(gitRevision: string): Promise<void> {
        return this.v2frame.locator(this.gitRevisionInput).fill(gitRevision);
    }

    async fillGitContainerInput(gitContainer: string): Promise<void> {
        return this.v2frame.locator(this.gitContainerInput).fill(gitContainer);
    }

    async selectGitCredentialByName(credentialName: string): Promise<void> {
        return this.selectOptionByValue(this.gitCredentialsSelect, credentialName);
    }

    async clickOnGitPrivateRadioOption(): Promise<void> {
        return this.page.locator(this.gitPrivateRadioOption).click();
    }

    async clickOnGitPublicRadioOption(): Promise<void> {
        return this.page.locator(this.gitPublicRadioOption).click();
    }

    // PVC
    async clickOnAddNewPvcBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewPvcBtn).click();
    }

    async clickOnNewPvcOption(): Promise<void> {
        return this.v2frame.locator(this.PvcRadioNewPvcOption).click();
    }

    async selectPvcVolumeSharingOption(sharingOption: string): Promise<void> {
        return this.clickOnElementByAidName(`radio-option-${sharingOption}`);
    }

    async fillPvcNameInput(pvcName: string): Promise<void> {
        return await this.v2frame.locator(this.PvcNameInput).fill(pvcName);
    }

    async selectFirstPvcStorageClass(): Promise<void> {
        return await this.selectFirstOption(this.PvcStorageClassSelect);
    }

    async selectPvcAccessMode(accessMode: string): Promise<void> {
        return this.selectOptionByValue(this.PvcAccessModeSelect, accessMode);
    }

    async selectPvcVolumeMode(volumeMode: string): Promise<void> {
        return this.selectOptionByValue(this.PvcVolumeModeSelect, volumeMode);
    }

    async fillPvcClaimSizeInput(claimSize: string): Promise<void> {
        return this.v2frame.locator(this.PvcClaimSizeInput).fill(claimSize);
    }

    async selectPvcClaimUnit(claimUnit: string): Promise<void> {
        return this.selectOptionByValue(this.PvcClaimUnitSelect, claimUnit);
    }

    async fillPvcContainerInput(containerPath = '/home/local/data'): Promise<void> {
        return this.v2frame.locator(this.PvcContainerInput).fill(containerPath);
    }

    async createPvc(
        name: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
            pvcAccessMode?: string;
            claimSizeInput?: string;
            claimUnit?: string;
            volumeMode?: string;
            containerInput?: string;
        }
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewPvcBtn();
        await this.fillNewPvcCreationForm(name, options);
    }

    async fillNewPvcCreationForm(
        pvcName: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
            pvcAccessMode?: string;
            claimSizeInput?: string;
            claimUnit?: string;
            volumeMode?: string;
            containerInput?: string;
        }
    ): Promise<void> {
        await this.selectCurrentCluster();
        if (options?.scope) {
            await this.assetsScopeSection.selectScope({
                scope: options.scope,
                scopeEntityName: options.scopeEntityName
            });
        }

        await this.fillPvcNameInput(pvcName);
        await this.clickOnNewPvcOption();
        await this.selectFirstPvcStorageClass();
        await this.selectPvcAccessMode(options?.pvcAccessMode || 'Read-write by one node');
        await this.fillPvcClaimSizeInput(options?.claimSizeInput || '2');
        await this.selectPvcClaimUnit(options?.claimUnit || 'GB');
        await this.selectPvcVolumeMode(options?.volumeMode || 'Filesystem');
        await this.fillPvcContainerInput(options?.containerInput || '/home/local/data');
        await this.clickOnSubmitBtn();
    }

    async fillHostPathCreationForm(
        hostpathName: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
            hostPath?: string;
            containerPath?: string;
        }
    ): Promise<void> {
        await this.selectCurrentCluster();
        if (options?.scope && options?.scopeEntityName) {
            await this.assetsScopeSection.selectScope({
                scope: options.scope,
                scopeEntityName: options.scopeEntityName
            });
        }
        await this.fillHostPathNameInput(hostpathName);
        await this.fillHostPathHostInput(
            options?.hostPath || objectGenerator.generateShortNameByPrefix('/tmp/test')
        );
        await this.fillHostPathContainerInput(
            options?.containerPath || objectGenerator.generateShortNameByPrefix('/home')
        );
        await this.clickOnSubmitBtn();
    }

    async createS3DataSource(
        name: string,
        options?: {
            credentialName?: string;
            serviceUrl?: string;
            bucket?: string;
            containerPath?: string;
            scope?: Scope;
            scopeEntityName?: string;
        }
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddS3Btn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope({
            scope: options.scope,
            scopeEntityName: options.scopeEntityName
        });
        await this.fillS3NameInput(name);
        await this.fillS3ServiceUrlInput(options?.serviceUrl || 'https://s3.amazonaws.com');
        await this.selectS3CredentialsByName(options?.credentialName || 'None (for public buckets)');
        await this.fillS3BucketInput(options?.bucket || 'test-bucket');
        await this.fillS3ContainerPathInput(options?.containerPath || '/test/path');
        await this.clickOnSubmitBtn();
        logger.debug('data source created successfully');
    }

    async createPvcFromExistingPvc(
        newPvcName: string,
        existingPvcName: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        }
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewPvcBtn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope({
            scope: options.scope,
            scopeEntityName: options.scopeEntityName
        });
        await this.fillPvcNameInput(newPvcName);
        await this.selectOptionByValue(this.existingPvcSelect, existingPvcName);
        await this.fillPvcContainerInput();

        await this.clickOnSubmitBtn();
    }

    async waitForNameAlreadyExistsErrorToAppear(): Promise<void> {
        return waitUtils.waitForElementToExist(this, this.nameAlreadyExistsError);
    }

    async waitForNameAlreadyExistsErrorToDisappear(): Promise<void> {
        return waitUtils.waitForElementToNotExist(this, this.nameAlreadyExistsError);
    }

    async clickOnAddNewConfigMapBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewConfigMapBtn).click();
    }

    async fillConfigMapAssetNameInput(value: string): Promise<void> {
        return await this.v2frame.locator(this.ConfigMapNameAssetInput).fill(value);
    }

    async fillConfigMapMountPathInput(value: string): Promise<void> {
        return await this.v2frame.locator(this.ConfigMapNameMountPathInput).fill(value);
    }

    async createConfigMapAsset(
        assetName: string,
        scope: Scope,
        scopeEntityName: string,
        configMapName: string,
        mountPath: string
    ): Promise<void> {
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewConfigMapBtn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope({
            scope: scope,
            scopeEntityName: scopeEntityName
        });
        await this.fillConfigMapAssetNameInput(assetName);
        await this.selectOptionByValue(this.ConfigMapResourceSelect, configMapName);
        await this.fillConfigMapMountPathInput(mountPath);
        await this.clickOnSubmitBtn();
    }

    async clickOnAddNewSecretBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewSecretBtn).click();
    }

    async fillSecretAssetNameInput(value: string): Promise<void> {
        return await this.v2frame.locator(this.SecretNameAssetInput).fill(value);
    }

    async clearSecretAssetNameInput(): Promise<void> {
        return await this.page.locator(this.SecretNameAssetInput).clear();
    }

    async fillSecretMountPathInput(value: string): Promise<void> {
        return await this.v2frame.locator(this.SecretMountPathInput).fill(value);
    }

    async createSecretDataSource(
        assetName: string,
        scope: Scope,
        scopeEntityName: string,
        credentialName: string,
        mountPath: string
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewSecretBtn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope({
            scope: scope,
            scopeEntityName: scopeEntityName
        });
        await this.fillSecretAssetNameInput(assetName);
        await this.selectOptionByValue(this.CredentialResourceSelect, credentialName);
        await this.fillSecretMountPathInput(mountPath);
        await this.clickOnSubmitBtn();
        logger.debug('secret data source created successfully');
    }

    async renameDataSource(oldName: string, newName: string): Promise<void> {
        // Select the data source
        await this.selectDataSourceInTableByName(oldName);
        // Click rename button
        await this.clickOnRenameDataSourceBtn();
        // Fill the new name
        await this.fillRenameModalNameInput(newName);
        // Click save and wait for the rename API to complete
        await this.clickOnRenameModalSaveBtn();
        await this.page.waitForResponse(
            (response) => response.url().includes('/api/v1/asset/datasource') && response.status() === 200
        );
        // Wait for the name to change in UI
        await this.selectDataSourceInTableByName(newName);
    }

    async createHostPath(
        name: string,
        options: {
            scope: Scope;
            scopeEntityName: string;
            hostPath?: string;
            containerPath?: string;
        }
    ): Promise<void> {
        // Moving to host path creation page
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewHostPathBtn();
        await this.selectCurrentCluster();

        // Filling the form fields
        await this.assetsScopeSection.selectScope({
            scope: options.scope,
            scopeEntityName: options.scopeEntityName
        });
        await this.fillHostPathNameInput(name);
        await this.fillHostPathHostInput(options.hostPath || '/user');
        await this.fillHostPathContainerInput(options.containerPath || '/home/local/data');
        await this.clickOnSubmitBtn();
        logger.info(`Host path - ${name} was created successfully`);
    }

    async createNFS(
        name: string,
        options: {
            scope: Scope;
            scopeEntityName: string;
            nfsServer?: string;
            nfsPath?: string;
            containerPath?: string;
        }
    ): Promise<void> {
        // Moving to NFS creation page
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewNFSBtn();
        await this.selectCurrentCluster();

        // Filling the form fields
        await this.fillNFSForm(name, options);
    }

    async fillNFSForm(
        name: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
            nfsServer?: string;
            nfsPath?: string;
            containerPath?: string;
        }
    ): Promise<void> {
        if (options?.scope && options?.scopeEntityName) {
            await this.assetsScopeSection.selectScope({
                scope: options.scope,
                scopeEntityName: options.scopeEntityName
            });
        }
        await this.fillNFSNameInput(name);
        await this.fillNFSServerInput(options?.nfsServer || 'test-server');
        await this.fillNFSPathInput(options?.nfsPath || '/test/sample');
        await this.fillNFSContainerInput(options?.containerPath || '/home/local/data');
        await this.clickOnSubmitBtn();
        logger.info(`NFS - ${name} was created successfully`);
    }

    async createGit(
        name: string,
        options: {
            scope: Scope;
            scopeEntityName: string;
            repository?: string;
            branch?: string;
            revision?: string;
            containerPath?: string;
            credentialName?: string;
        }
    ): Promise<void> {
        // Moving to Git creation page
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewGitBtn();
        await this.selectCurrentCluster();

        // Filling the form fields
        await this.assetsScopeSection.selectScope({
            scope: options.scope,
            scopeEntityName: options.scopeEntityName
        });
        await this.fillGitNameInput(name);
        await this.fillGitRepositoryInput(options.repository || 'https://test.com');
        if (options.credentialName) {
            await this.clickOnGitPrivateRadioOption();
            await this.selectGitCredentialByName(options.credentialName);
        }
        await this.fillGitContainerInput(options.containerPath || '/home/local/data');
        await this.clickOnSubmitBtn();
        logger.info(`Git - ${name} was created successfully`);
    }

    async createS3(options: {
        name: string;
        scope: Scope;
        scopeEntityName: string;
        credentialName?: string;
        serviceUrl?: string;
        bucket?: string;
        containerPath?: string;
    }): Promise<void> {
        // Moving to S3 creation page
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddS3Btn();
        await this.selectCurrentCluster();

        // Filling the form fields
        await this.assetsScopeSection.selectScope({
            scope: options.scope,
            scopeEntityName: options.scopeEntityName
        });
        await this.fillS3NameInput(options.name);
        await this.fillS3ServiceUrlInput(options.serviceUrl || 'https://s3.amazonaws.com');
        await this.fillS3BucketInput(options.bucket || 'test-bucket');
        await this.fillS3ContainerPathInput(options.containerPath || '/test/path');
        await this.selectS3CredentialsByName(options.credentialName || 'None (for public buckets)');
        await this.clickOnSubmitBtn();
        logger.info(`S3 - ${options.name} was created successfully`);
    }

    async createPVC(options: {
        name: string;
        scope: Scope;
        scopeEntityName: string;
        storageClass?: string;
        accessMode?: string;
        claimSize?: string;
        claimUnit?: string;
        containerPath?: string;
        volumeMode?: string;
    }): Promise<void> {
        // Moving to PVC creation page
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewPvcBtn();
        await this.selectCurrentCluster();

        await this.fillPVCForm(options);
    }

    async fillPVCForm(options: {
        name: string;
        scope?: Scope;
        scopeEntityName?: string;
        storageClass?: string;
        accessMode?: string;
        claimSize?: string;
        claimUnit?: string;
        containerPath?: string;
        volumeMode?: string;
    }): Promise<void> {
        // Filling the form fields
        if (options?.scope && options?.scopeEntityName) {
            await this.assetsScopeSection.selectScope({
                scope: options.scope,
                scopeEntityName: options.scopeEntityName
            });
        }

        await this.fillPvcNameInput(options.name);
        await this.selectOptionByValue(this.PvcStorageClassSelect, options.storageClass || 'standard');
        await this.selectPvcAccessMode(options.accessMode || 'Read-write by one node');
        await this.fillPvcClaimSizeInput(options.claimSize || '2');
        await this.selectPvcClaimUnit(options.claimUnit || 'GB');
        await this.selectPvcVolumeMode(options?.volumeMode || 'Filesystem');
        await this.fillPvcContainerInput(options.containerPath || '/home/local/data');
        await this.clickOnSubmitBtn();
        logger.info(`PVC - ${options.name} was created successfully`);
    }

    async createConfigMap(options: {
        name: string;
        scope: Scope;
        scopeEntityName: string;
        configMapName: string;
        mountPath: string;
    }): Promise<void> {
        // Moving to ConfigMap creation page
        await this.goToPage();
        await this.clickOnAddNewDataSourceBtn();
        await this.clickOnAddNewConfigMapBtn();
        await this.selectCurrentCluster();

        // Filling the form fields
        await this.assetsScopeSection.selectScope({
            scope: options.scope,
            scopeEntityName: options.scopeEntityName
        });
        await this.fillConfigMapAssetNameInput(options.name);
        await this.selectOptionByValue(this.ConfigMapResourceSelect, options.configMapName);
        await this.fillConfigMapMountPathInput(options.mountPath);
        await this.clickOnSubmitBtn();
        logger.info(`ConfigMap - ${options.name} was created successfully`);
    }

    async isDataSourceNameErrorShown(): Promise<boolean> {
        const alertMsg = await this.page.locator(this.dataSourceNameErrorDiv).innerText();
        return alertMsg === this.dataSourceNameErrorMsg;
    }

    async clearPvcNameInput(): Promise<void> {
        await this.page.locator(this.PvcNameInput).clear();
    }
}
