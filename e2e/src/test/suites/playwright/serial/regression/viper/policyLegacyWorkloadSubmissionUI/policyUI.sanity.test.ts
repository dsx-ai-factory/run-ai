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
    apiUtils,
    waitUtils,
    unleashClient,
    settings
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { PolicyPage } from '../../../../../../../modules/ui/runaiV2/pages/PolicyPage';
import { AssetsScopeSection } from '../../../../../../../modules/ui/runaiV2/components/AssetsScopeSection';
import { AssetKind, Scope } from '../../../../../../../modules/api/apis/assetsApi';
import { WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { WorkspacePage } from '../../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { EnvironmentPage } from '../../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { test } from '@playwright/test';
import { EFeatureToggleKeys } from '../../../../../../../modules/helpers/utils/Unleash/featureToggle';
import { policyApi as policyApiHandler } from '../../../../../../../modules/api';

const consts = {
    lowPolicies: 'policies',
    uppPolicies: 'Workload policies',
    newPolicies: 'policies/new',
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
        let workloadPage: WorkloadPage;
        let workspacePage: WorkspacePage;
        let environment: EnvironmentPage;
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
            if (!settings.isSelfHosted) {
                await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            }
            await apiUtils.disableHybridSubmission();
        });

        test.afterAll(async () => {
            await apiUtils.disableNewPolicyFeature();
            if (!settings.isSelfHosted) {
                await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            }
            await apiUtils.enableHybridSubmission();
        });

        test.beforeEach(async function () {
            await initProject();

            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            policyPage = new PolicyPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            workspacePage = new WorkspacePage(browserEngine.currentPage);
            assetsScopeSection = new AssetsScopeSection(browserEngine.currentPage);
            environment = new EnvironmentPage(browserEngine.currentPage);
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

        test('Should create a policy, move to workload page and verify policy has an effect on assets', async function () {
            videoName = 'policy-has-an-effect';
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

            // go to workspace page and proceed through the inital steps
            await workloadPage.navigateToWorkloads();
            await workloadPage.openWorkspaceForm();
            await workloadPage.workloadFormSelectProject(projectName);

            await workspacePage.fillWorkspaceName('some-name');
            await workspacePage.waitForWorkspaceNameCount('some-name');
            await workspacePage.selectAssetsSubmissionOption();
            await workspacePage.clickOnContinue();

            // verify a compute asset does not comply with the policy we set during the test
            // open on compliance modal
            await workspacePage.fillGeneralSearchInput(workspacePage.computeResourceSection, 'two-gpus');
            await workspacePage.clickOnCardCompliaceButton('two-gpus', AssetKind.Compute);

            // verify compliance modal details are correct
            expect(
                await workspacePage.locator(workspacePage.complianceInfoModalSubtitle).textContent()
            ).to.include('This compute resource');

            expect(await workspacePage.locator('[aid=parameter]').first().textContent()).to.include(
                'gpuDevicesRequest'
            );
            expect(await workspacePage.locator('[aid=rule]').first().textContent()).to.include('• max = 1');
            expect(await workspacePage.locator('[aid=source]').first().textContent()).to.include(
                `Project = ${projectName}`
            );

            // checks done, go back to policy page and delete the policy
            await workspacePage.closeComplianceModal();
            await workspacePage.closeWorkspaceForm();

            // Deleting policy
            await policyPage.navigateToPolicyPage();
            await policyPage.searchByName(policyName);
            await policyPage.selectPolicyInTableByName(policyName);
            await policyPage.clickOnPolicyDeleteBtn();
            await policyPage.clickOnPolicyModalDeleteBtn();
        });

        test('Should create a policy, move to workload page and verify policy is enforced on form fields', async function () {
            videoName = 'policy-is-enforced-on-form-fields';
            await policyPage.navigateToPolicyPage();
            await policyPage.clickOnAddNewPolicyBtn();

            await assetsScopeSection.selectScope({
                scope: Scope.Project,
                scopeEntityName: projectName
            });

            await policyPage.selectWorkloadByText(consts.uppWorkspace);
            await policyPage.clickOnPolicyYamlAddBtn();

            await policyPage.InjectPolicyToMonacoYamlEditor('multiFieldsPolicy.yaml');
            await policyPage.clickOnPolicyApplyYamlBtn();

            await policyPage.waitForOnPolicySubmitBtnToAppear();
            await policyPage.clickOnPolicySubmitBtn();

            const policyName = `${projectName}-${consts.lowWorkspace}`;
            await policyPage.searchByName(policyName);
            const policyFromTable = await policyPage.getPolicyFromTableByName(policyName);

            expect(policyFromTable).equal(policyName);

            // go to workspace page and proceed through the inital steps
            await workloadPage.navigateToWorkloads();
            await workloadPage.openWorkspaceForm();
            await workloadPage.workloadFormSelectProject(projectName);

            await workspacePage.fillWorkspaceName('enforce-policy-on-form-fields');
            await workspacePage.waitForWorkspaceNameCount('enforce-policy-on-form-fields');
            await workspacePage.selectAssetsSubmissionOption();
            await workspacePage.clickOnContinue();

            // creating environment to reveal relevant fields
            await workspacePage.clickOnNewEnvironmentBtn();
            const envName = objectGenerator.generateShortNameByPrefix('enforce-policy-on-form-fields-env', 5);

            await environment.fillEnvironmentName(envName);
            await environment.fillEnvironmentImage('ubuntu');
            await environment.clickOnSecuritySection();
            await environment.clickCustomRadioOption();
            await environment.setOverideToggle(true);
            await environment.clickOnCreateEnvironmentBtn();

            // check relevant policy fields on the form
            await workspacePage.clickEnvSectionMoreSettings();
            await workspacePage.fillGidInput(501);

            const errorLocator = workspacePage
                .locator('div')
                .filter({ hasText: /^Enter a value that follows steps of 3, between 500-510$/ })
                .first();

            await waitUtils.waitForElementToExistWithLocator(errorLocator);
            await workspacePage.fillGidInput(503);
            await waitUtils.waitForElementToNotExistWithLocator(errorLocator);

            expect(
                await workspacePage.locator(workspacePage.environmentVarNameInput).first().isDisabled()
            ).to.eql(true);

            expect(
                await workspacePage.locator(workspacePage.environmentVarNameInput).first().inputValue()
            ).to.include('LOG_DIR');

            expect(
                await workspacePage
                    .locator(workspacePage.environmentVarValueInput)
                    .first()

                    .inputValue()
            ).to.include('/home/log');

            await workspacePage.openGeneralSection();

            expect(
                await workspacePage.locator(workspacePage.annotationNameInput).first().isDisabled()
            ).to.eql(true);

            expect(
                await workspacePage
                    .locator(workspacePage.annotationNameInput)
                    .first()

                    .inputValue()
            ).to.include('anno-1');
            expect(
                await workspacePage
                    .locator(workspacePage.annotationValueInput)
                    .first()

                    .inputValue()
            ).to.include('ann1');
            expect(
                await workspacePage
                    .locator(workspacePage.annotationNameInput)
                    .nth(1)

                    .inputValue()
            ).to.include('anno-2');
            expect(
                await workspacePage
                    .locator(workspacePage.annotationValueInput)
                    .nth(1)

                    .inputValue()
            ).to.equal('ann2');

            await workspacePage.closeWorkspaceForm();

            await environment.goToPage();
            await environment.removeEnvironment(envName);

            // Deleting policy
            await policyPage.navigateToPolicyPage();
            await policyPage.searchByName(policyName);
            await policyPage.selectPolicyInTableByName(policyName);
            await policyPage.clickOnPolicyDeleteBtn();
            await policyPage.clickOnPolicyModalDeleteBtn();
        });
    }
);
