/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

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
import { HybridForm } from '../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { expect, test } from '@playwright/test';
import { policyApi as policyApiHandler } from '../../../../../../../modules/api';

test.describe(
    `Policy UI, tags: ${tags.tool.ui} ${tags.resource.policy} ${tags.teams.viper} `,
    { tag: [tags.services.frontend, tags.services.policy, tags.services.workloads, tags.runOn.serialCI] },
    () => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let policyPage: PolicyPage;
        let hybridForm: HybridForm;
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
            hybridForm = new HybridForm(browserEngine.currentPage);
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

        async function createWorkspacePolicy(fixture: string): Promise<string> {
            await policyPage.navigateToPolicyPage();
            await policyPage.clickOnAddNewPolicyBtn();

            await assetsScopeSection.selectScope({
                scope: Scope.Project,
                scopeEntityName: projectName
            });

            await policyPage.selectWorkloadByText('Workspace');
            await policyPage.clickOnPolicyYamlAddBtn();
            await policyPage.InjectPolicyToMonacoYamlEditor(fixture);
            await policyPage.clickOnPolicyApplyYamlBtn();

            await policyPage.waitForOnPolicySubmitBtnToAppear();
            await policyPage.clickOnPolicySubmitBtn();

            const policyName = `${projectName}-workspace`;
            await policyPage.searchByName(policyName);
            const policyFromTable = await policyPage.getPolicyFromTableByName(policyName);

            expect(policyFromTable).toBe(policyName);
            return policyName;
        }

        async function deleteWorkspacePolicy(policyName: string): Promise<void> {
            await policyPage.searchByName(policyName);
            await policyPage.selectPolicyInTableByName(policyName);
            await policyPage.clickOnPolicyDeleteBtn();
            await policyPage.clickOnPolicyModalDeleteBtn();
            await policyPage.verifyPolicyDeletedSuccessfully(policyName);
        }

        async function openWorkspace(name: string): Promise<void> {
            await hybridForm.navigateToWorkloads();
            await hybridForm.fillWorkloadCreationFirstPage(flowRootState.cluster.name, projectName, name);
        }

        test(`Should create a policy ${tags.runOn.parallelCI}`, async () => {
            videoName = 'policy-created-successfully';
            const policyName = await createWorkspacePolicy('simplePolicy.yaml');
            await deleteWorkspacePolicy(policyName);
        });

        test('Shows the policy rule for a noncompliant compute resource', async () => {
            videoName = 'policy-has-an-effect';
            const policyName = await createWorkspacePolicy('simplePolicy.yaml');
            await openWorkspace('check-compute-policy');
            await hybridForm.openComputesDrawer();
            await hybridForm.fillSeachInputInDrawer('two-gpus');
            const drawer = hybridForm.hybridDrawer;
            const asset = hybridForm.locator(drawer.drawerItem).filter({ hasText: 'two-gpus' });
            await asset.locator(drawer.viewPolicyBtn).click();

            const policy = hybridForm.locator('.policy-info');
            const rule = policy.getByRole('row').filter({ hasText: 'gpuDevicesRequest' });
            await expect(rule).toContainText(/max\s*=\s*1/);
            await expect(rule).toContainText(projectName);
            await hybridForm
                .locator('.primary-section')
                .getByRole('button', { name: 'Close', exact: true })
                .click();
            await hybridForm.closeWorkloadForm();
            await policyPage.navigateToPolicyPage();
            await deleteWorkspacePolicy(policyName);
        });

        test('Enforces GID restrictions and preserves locked defaults in the workspace form', async () => {
            videoName = 'policy-is-enforced-on-form-fields';
            const policyName = await createWorkspacePolicy('multiFieldsPolicy.yaml');
            await openWorkspace('enforce-policy-on-form-fields');
            await hybridForm.openSecuritySection();
            await hybridForm.page.getByRole('radio', { name: /Custom/ }).click();
            const environment = hybridForm.hybridEnvironmentSection;
            const gid = hybridForm.locator(environment.gidInput);
            const error = hybridForm.page.getByText('Enter a value that follows steps of 3, between 500-510');
            await gid.fill('501');
            await gid.blur();
            await expect(error).toBeVisible();
            await gid.fill('503');
            await gid.blur();
            await expect(error).toBeHidden();

            await hybridForm.openRuntimeSection();
            const variable = hybridForm.locator(environment.envVarNameInput).first();
            await expect(variable).toHaveValue('LOG_DIR');
            await expect(variable).toBeDisabled();
            await expect(hybridForm.locator(environment.envVarValue).first()).toHaveValue('/home/log');

            await hybridForm.locator(hybridForm.hybridGeneralSection.sectionSelector).click();
            const annotations = hybridForm.locator('.annotations-section');
            const names = annotations.locator('[aid="itemized-name-input"]');
            const values = annotations.locator('[aid="itemized-value-input"]');
            await expect(names.nth(0)).toHaveValue('anno-1');
            await expect(names.nth(0)).toBeDisabled();
            await expect(values.nth(0)).toHaveValue('ann1');
            await expect(names.nth(1)).toHaveValue('anno-2');
            await expect(values.nth(1)).toHaveValue('ann2');

            await hybridForm.closeWorkloadForm();
            await policyPage.navigateToPolicyPage();
            await deleteWorkspacePolicy(policyName);
        });
    }
);
