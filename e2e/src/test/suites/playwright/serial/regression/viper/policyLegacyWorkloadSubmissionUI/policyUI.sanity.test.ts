/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    Utils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    projectUtils,
    apiUtils
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { PolicyPage } from '../../../../../../../modules/ui/runaiV2/pages/PolicyPage';
import { AssetsScopeSection } from '../../../../../../../modules/ui/runaiV2/components/AssetsScopeSection';
import { Scope } from '../../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';
import { policyApi as policyApiHandler } from '../../../../../../../modules/api';

const consts = {
    uppWorkspace: 'Workspace',
    lowWorkspace: 'workspace'
};

test.describe(
    `Policy UI, tags: ${tags.tool.ui} ${tags.resource.policy} ${tags.teams.viper} `,
    { tag: [tags.services.frontend, tags.services.policy, tags.services.workloads, tags.runOn.serialCI] },
    () => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let policyPage: PolicyPage;
        // components
        let assetsScopeSection: AssetsScopeSection;

        let videoName: string;
        let projectName: string;
        let namespace: string;
        let project: Project;

        async function initProject(): Promise<void> {
            projectName = objectGenerator.generateNameByPrefix('project-e2e-for-policy');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            try {
                const runaiResearchManagerRoleBindingName =
                    projectUtils.getRunaiResearchManagerRoleBindingName();
                await projectUtils.waitForRoleBindingForUserInProject(
                    namespace,
                    flowRootState.users.research_manager.email,
                    runaiResearchManagerRoleBindingName,
                    true
                );
            } catch (e) {
                logger.error(e);
            }
        }

        test.beforeAll(async () => {
            browserEngine = new BrowserEngine();
            await apiUtils.enableNewPolicyFeature();
        });

        test.afterAll(async () => {
            await apiUtils.disableNewPolicyFeature();
        });

        test.beforeEach(async function () {
            await initProject();

            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            policyPage = new PolicyPage(browserEngine.currentPage);
            assetsScopeSection = new AssetsScopeSection(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });
        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'policy-' + videoName);
            if (testInfo.status === 'failed') {
                logger.info('trying to dele project scoped policy, in case it exists');
                try {
                    await policyApiHandler.policyApi.deleteWorkspacePolicy(
                        'project',
                        undefined,
                        project.id.toString()
                    );
                } catch (error) {
                    logger.error(`failed to delete policy, project id: ${project.id.toString()} ${error}`);
                }
            }

            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} an namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });

        test(`Should create a policy ${tags.runOn.parallelCI}`, async () => {
            videoName = 'policy-created-successfully';
            await policyPage.navigateToPolicyPage();
            await policyPage.clickOnAddNewPolicyBtn();

            await assetsScopeSection.selectScope({
                scope: Scope.Project,
                scopeEntityName: projectName
            });

            await policyPage.selectWorkloadByText(consts.uppWorkspace);
            await policyPage.clickOnPolicyYamlAddBtn();
            await policyPage.InjectPolicyToMonacoYamlEditor('simplePolicy.yaml');
            await policyPage.clickOnPolicyApplyYamlBtn();

            await policyPage.waitForOnPolicySubmitBtnToAppear();
            await policyPage.clickOnPolicySubmitBtn();

            const policyName = `${projectName}-${consts.lowWorkspace}`;
            await policyPage.searchByName(policyName);
            const policyFromTable = await policyPage.getPolicyFromTableByName(policyName);

            expect(policyFromTable).equal(policyName);

            await policyPage.selectPolicyInTableByName(policyName);
            await policyPage.clickOnPolicyDeleteBtn();
            await policyPage.clickOnPolicyModalDeleteBtn();

            await policyPage.verifyPolicyDeletedSuccessfully(policyName);
        });
    }
);
