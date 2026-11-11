/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { TemplatePage } from '../../../../../../../modules/ui/runaiV2/pages/TemplatePage';
import { EnvironmentPage } from '../../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { ComputeResourcePage } from '../../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    objectGenerator,
    logger,
    tags,
    waitUtils,
    projectResourcesUtils,
    Utils,
    operations,
    apiUtils
} from '../../../../../../../modules/helpers';

import { Scope } from '../../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect } from 'chai';
import { Locator } from 'playwright';
import { AssetsScopeSection } from '../../../../../../../modules/ui/runaiV2/components/AssetsScopeSection';
import { test } from '@playwright/test';
import { supportedFrom, SupportedVersions } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Workspace Templates UI, tags: ${tags.tool.ui} ${tags.resource.workspaceTemplate} ${tags.teams.viper}`, { tag: [tags.services.assets, tags.services.frontend, tags.runOn.serialCI] }, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
    let projectName: string;
    let namespace: string;
    let project: Project;
    let vidoeName: string;

    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let workspaceTemplate: TemplatePage;
    let environment: EnvironmentPage;
    let computeResource: ComputeResourcePage;

    // components
    let assetsScopeSection: AssetsScopeSection;

    test.beforeAll('creating project and namespace', async function () {
        browserEngine = new BrowserEngine();

        projectName = objectGenerator.generateNameByPrefix('project-e2e-for-templates');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 1,
            maxAllowedGpus: 1
        });

        await apiUtils.setFlexTemplateFeature(false);
        logger.info(`created project ${projectName}`);
    });

    test.afterAll('deleting project and namespace', async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            await apiUtils.setFlexTemplateFeature(true);
            logger.info(`deleted project ${projectName} an namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
            throw error;
        }
    });

    test.beforeEach('start browser', async function () {
        await browserEngine.start();
        // setting pages:
        auth = new AuthPage(browserEngine.currentPage);
        workspaceTemplate = new TemplatePage(browserEngine.currentPage);
        // setting components:
        assetsScopeSection = new AssetsScopeSection(browserEngine.currentPage);

        await auth.signIn(
            flowRootState.users.research_manager.email,
            flowRootState.users.research_manager.password
        );
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'workspace-templates-' + vidoeName);
    });

    test.describe('Create new legacy workload template', function () {
        async function deleteEnvironment(name: string): Promise<void> {
            logger.debug(`deleting environment: ${name}`);
            await environment.goToPage();
            await environment.removeEnvironment(name);
            logger.debug(`environment ${name} deleted successfully`);
        }

        async function deleteCompute(name: string): Promise<void> {
            logger.debug(`deleting compute: ${name}`);
            await computeResource.goToPage();
            await computeResource.removeCompute(name);
            logger.debug(`compute resource ${name} deleted successfully`);
        }

        test('should create new legacy workload template from scratch', async () => {
            vidoeName = 'create-workload-template-from-scratch';
            environment = new EnvironmentPage(browserEngine.currentPage);
            computeResource = new ComputeResourcePage(browserEngine.currentPage);

            const templateName = objectGenerator.generateShortNameByPrefix('template');
            const environmentName = objectGenerator.generateShortNameByPrefix('environment');
            const computeName = objectGenerator.generateShortNameByPrefix('compute');

            await workspaceTemplate.goToPage();
            await workspaceTemplate.openNewTemplatePage();

            //select cluster
            await workspaceTemplate.selectCurrentCluster();
            await assetsScopeSection.selectScope({ scope: Scope.Project, scopeEntityName: projectName });

            // fill the workspace form details
            await workspaceTemplate.fillTemplateName(templateName);

            // create environment
            await workspaceTemplate.clickOnNewEnvironmentBtn();
            await environment.createNewEnvironment(environmentName, projectName, {
                waitForProjectScope: true,
                command: 'sleep infinity'
            });

            // create compute
            await workspaceTemplate.clickOnNewComputeBtn();
            await computeResource.createNewCompute(computeName, projectName);

            await waitUtils.sleep(2000); //remove and call workspaceTemplate.waitForTenantScopeSection() after scope-section-subheader aid will be added

            // create workspace
            await workspaceTemplate.clickOnCreateTemplateBtn();
            const locator: Locator = workspaceTemplate.findByText('td', templateName);
            const name: string = await locator.textContent();
            expect(name.trim()).eql(templateName);
            logger.debug('workload template created successfully');

            // cleanup
            logger.debug('starting cleanup');

            // delete workspace
            logger.debug('deleting workload template: ' + templateName);
            // await waitUtils.waitForWorkspaceCreation(flowRootState.cluster.uuid, templateName);
            await workspaceTemplate.removeTemplate(templateName);
            await waitUtils.waitForWorkspaceTemplateDeletion(templateName);
            logger.debug('workload template deleted successfully');

            // cleanup assets
            logger.debug('starting cleanup');
            await deleteEnvironment(environmentName);
            await deleteCompute(computeName);
        });
    });
    });
});
