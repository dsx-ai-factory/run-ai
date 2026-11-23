/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    Utils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    environmentUtils,
    projectUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { EnvironmentPage } from '../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { AssetsScopeSection } from '../../../../../../modules/ui/runaiV2/components/AssetsScopeSection';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';

test.describe(`Environment UI, tags: ${tags.tool.ui} ${tags.resource.environment} ${tags.teams.viper} `, () => {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let environment: EnvironmentPage;
    let assetsScopeSection: AssetsScopeSection;

    let videoName: string;
    let projectName: string;
    let namespace: string;
    let project: Project;
    let environmentName: string;

    async function fillEnvironmentDetails(envDetails): Promise<void> {
        await environment.selectCurrentCluster();
        await assetsScopeSection.selectScope({ scope: Scope.Project, scopeEntityName: project.name });
        await environment.fillEnvironmentName(envDetails.name);
        await environment.fillEnvironmentImage(envDetails.image);
        await environment.openToolsSectionAndAddTool();
        await environment.selectVsCodeAToolType();
        await environment.clickOnRunTimeSettingsSection();
        await environment.openCommandAndArgumentSection();
        await environment.setCommand(envDetails.command);
        await environment.setArguments(envDetails.arguments);
        await environment.setEnvironmentVariable(envDetails.varName, envDetails.varValue);
        await environment.clickOnSecuritySection();
        await environment.clickCustomRadioOption();
        await environment.setUid(envDetails.uid);
        await environment.setGid(envDetails.gid);
        await environment.setOverideToggle(true);
    }

    test.beforeAll(async () => {
        browserEngine = new BrowserEngine();
        projectName = objectGenerator.generateNameByPrefix('project-e2e-for-environment');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 1,
            maxAllowedGpus: 1
        });

        const runaiResearchManagerRoleBindingName = projectUtils.getRunaiResearchManagerRoleBindingName();

        await projectUtils.waitForRoleBindingForUserInProject(
            namespace,
            flowRootState.users.research_manager.email,
            runaiResearchManagerRoleBindingName,
            true
        );
    });

    test.afterAll('deleting project and namespace', async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            logger.info(`deleted project ${projectName} an namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
            throw error;
        }
    });

    test.beforeEach(async () => {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        environment = new EnvironmentPage(browserEngine.currentPage);
        assetsScopeSection = new AssetsScopeSection(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.research_manager.email,
            flowRootState.users.research_manager.password
        );
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'environment-' + videoName);
    });

    test('Should create an environment, find it in the list and then delete it', async () => {
        videoName = 'create-and-delete-environment';
        environmentName = 'environment-e2e';
        const envDetails = {
            name: environmentName,
            image: 'ubuntu',
            command: 'sleep 100',
            arguments: '--arg1',
            varName: 'var1',
            varValue: 'value1',
            uid: '11111',
            gid: '22222',
            supplementalGrp: ''
        };
        await environment.goToPage();
        await environment.clickOnAddNewEnvironmentBtn();
        await fillEnvironmentDetails(envDetails);
        await environment.clickOnCreateEnvironmentBtn();
        const createdEnvNameCell = environment.findByText('td', environmentName);
        const createdEnvironmentName: string = (await createdEnvNameCell.textContent()).trim();
        expect(createdEnvironmentName).eql(environmentName);
        logger.info(`created environment ${environmentName} successfully`);
        await environment.removeEnvironment(environmentName);
        await environmentUtils.waitForEnvironmentToBeDeleted(environmentName);
        //validate list is empty
        logger.info(`removed environment ${environmentName} successfully`);
    });

    test('Should copy and edit an existing environment', async () => {
        videoName = 'copy-and-edit-environment';
        const environmentName = 'env-e2e';
        const envDetails = {
            name: environmentName,
            image: 'ubuntu',
            command: 'sleep 100',
            arguments: '--arg1',
            varName: 'var1',
            varValue: 'value1',
            uid: '11111',
            gid: '22222',
            supplementalGrp: ''
        };

        await environment.goToPage();
        await environment.clickOnAddNewEnvironmentBtn();
        await fillEnvironmentDetails(envDetails);
        await environment.clickOnCreateEnvironmentBtn();
        logger.info(`created environment ${environmentName} successfully`);

        await environment.copyEnvironment(environmentName);
        const copiedEnviromentName = 'copy-e2e';
        await environment.fillEnvironmentName(copiedEnviromentName);
        await environment.clickOnCreateEnvironmentBtn();

        const copiedEnvironmentNameCell = environment.findByText('td', copiedEnviromentName);
        const copiedEnvironmentNameCellText: string = (await copiedEnvironmentNameCell.textContent()).trim();
        expect(copiedEnvironmentNameCellText).eql(copiedEnviromentName);
        logger.info(
            `copied environment ${environmentName} successfully & created environment ${copiedEnviromentName} successfully`
        );
        await environment.removeEnvironment(environmentName);
        logger.info(`removed environment ${environmentName} successfully`);
        await environment.removeEnvironment(copiedEnviromentName);
        await environmentUtils.waitForEnvironmentToBeDeleted(environmentName);
        logger.info(`removed environment ${copiedEnviromentName} successfully`);
    });
});
