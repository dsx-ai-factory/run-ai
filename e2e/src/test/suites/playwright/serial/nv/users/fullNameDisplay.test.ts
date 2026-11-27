/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    serviceAccountUtils,
    settings,
    tags,
    usersManager,
    usersUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { AccessRulesPage } from '../../../../../../modules/ui/runaiV2/pages/AccessRulesPage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { ProjectPageV2 } from '../../../../../../modules/ui/runaiV2/pages/ProjectPageV2';
import { DepartmentPageV2 } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPageV2';
import { EventHistoryPage } from '../../../../../../modules/ui/runaiV2/pages/EventHistoryPage';
import { RolesPage } from '../../../../../../modules/ui/runaiV2/pages/RolesPage';
import { PolicyPage } from '../../../../../../modules/ui/runaiV2/pages/PolicyPage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';
import { QTable } from '../../../../../../modules/ui/runaiV2/components/QTable';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { authorizationApi } from '../../../../../../modules/api';
import { SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';

test.describe(`Full Name Display, tags: ${tags.teams.loong} ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_26)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let usersPage: UsersV2;
        let accessRulesPage: AccessRulesPage;
        let workloadPage: WorkloadPage;
        let hybridForm: HybridForm;
        let projectPage: ProjectPageV2;
        let departmentPage: DepartmentPageV2;
        let eventHistoryPage: EventHistoryPage;
        let rolesPage: RolesPage;
        let policyPage: PolicyPage;
        let templatePage: TemplatePageV2;
        let qTable: QTable;
        let videoName: string;
        let createdUserEmails: string[];
        let createdServiceAccountIds: string[];
        let createdAccessRuleSubjects: string[];
        let createdProjects: Project[];
        let workloadSubmissionProject: Project;

        test.beforeAll(function () {
            browserEngine = new BrowserEngine();
        });

        test.beforeEach(async function () {
            createdUserEmails = [];
            createdServiceAccountIds = [];
            createdAccessRuleSubjects = [];
            createdProjects = [];
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            usersPage = new UsersV2(browserEngine.currentPage);
            accessRulesPage = new AccessRulesPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            projectPage = new ProjectPageV2(browserEngine.currentPage);
            departmentPage = new DepartmentPageV2(browserEngine.currentPage);
            eventHistoryPage = new EventHistoryPage(browserEngine.currentPage);
            rolesPage = new RolesPage(browserEngine.currentPage);
            policyPage = new PolicyPage(browserEngine.currentPage);
            templatePage = new TemplatePageV2(browserEngine.currentPage);
            qTable = new QTable(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        const createEditorTenantAccessRuleForUser = async (subject: string): Promise<void> => {
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                subject,
                flowRootState.users.editor.role.id,
                SubjectType.User
            );
            createdAccessRuleSubjects.push(subject);
        };

        const createEditorTenantAccessRuleForServiceAccount = async (subject: string): Promise<void> => {
            await authorizationUtils.createAccessRuleForServiceAccount(
                subject,
                flowRootState.users.editor.role.name,
                ScopeType.Tenant,
                flowRootState.cluster.tenantId.toString(),
                flowRootState.users.system_administrator.accessToken
            );
            createdAccessRuleSubjects.push(subject);
        };

        test.afterEach('cleanup and stop browser', async ({}, testInfo) => {
            const adminToken = flowRootState.users.system_administrator.accessToken;
            const cancelConnectionAccessButton = browserEngine.currentPage.locator(
                hybridForm.hybridEnvironmentSection.servingPortAccess.servingPortAccessCancelBtn
            );
            if (await cancelConnectionAccessButton.isVisible({ timeout: 1000 }).catch(() => false)) {
                await cancelConnectionAccessButton.click();
            }

            const leavePageButton = browserEngine.currentPage.locator(workloadPage.headerLeavePageBtn);
            if (await leavePageButton.isVisible({ timeout: 1000 }).catch(() => false)) {
                await leavePageButton.click();
                const confirmLeaveButton = browserEngine.currentPage.locator(workloadPage.modalLeaveBtn);
                if (await confirmLeaveButton.isVisible({ timeout: 3000 }).catch(() => false)) {
                    await confirmLeaveButton.click();
                }
            }

            for (const subject of createdAccessRuleSubjects) {
                try {
                    const accessRulesResponse = await authorizationUtils.getAccessRules(
                        {
                            subjectType: SubjectType.User,
                            subjectIds: [subject]
                        },
                        adminToken
                    );
                    const accessRules = accessRulesResponse.accessRules.filter(
                        (accessRule) => accessRule.subjectId === subject
                    );
                    for (const accessRule of accessRules) {
                        await authorizationUtils.deleteAccessRule(accessRule, adminToken);
                    }
                } catch (err) {
                    logger.error(`Failed to delete access rule for ${subject}: ${err}`);
                }
            }

            for (const userEmail of createdUserEmails) {
                try {
                    const users = await usersUtils.searchUsersMinimal(adminToken, userEmail);
                    const user = users.find((minimalUser) => minimalUser.username === userEmail);
                    if (user) {
                        await usersManager.deleteLocalUser(adminToken, user.id);
                    }
                } catch (err) {
                    logger.error(`Failed to delete user ${userEmail}: ${err}`);
                }
            }
            for (const serviceAccountId of createdServiceAccountIds) {
                try {
                    await serviceAccountUtils.deleteServiceAccount(
                        flowRootState.users.system_administrator.accessToken,
                        serviceAccountId
                    );
                } catch (err) {
                    logger.error(`Failed to delete service account ${serviceAccountId}: ${err}`);
                }
            }
            for (const project of createdProjects) {
                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (err) {
                    logger.error(`Failed to delete project ${project.name}: ${err}`);
                }
            }
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test.describe('User full name display and fallback', function () {
            test('Display users as email Full Name when first and last name are available (T6037691)', async () => {
                videoName = 'full-name-display-when-available';

                const firstName = 'Alice';
                const lastName = 'Johnson';
                const fullName = `${firstName} ${lastName}`;
                const email = objectGenerator.generateEmailByPrefix('alice-johnson');
                const fullNameDisplay = `${email} ${fullName}`;

                await test.step('create local user with first and last name', async () => {
                    await usersPage.createLocalUser(email, firstName, lastName);
                    createdUserEmails.push(email);
                });

                await test.step('verify User and Full name columns in Users table', async () => {
                    await usersPage.verifyUserTableFullName(email, fullName);
                });

                await test.step('verify Full Name (email) in Access Rules subject column', async () => {
                    await createEditorTenantAccessRuleForUser(email);
                    await accessRulesPage.verifyAccessRuleSubject(email, fullNameDisplay);
                });
            });

            test('Fallback display handles partial or missing name attributes without layout issues (T6037692)', async () => {
                videoName = 'fallback-display-partial-missing-names';

                const users = [
                    {
                        email: objectGenerator.generateEmailByPrefix('bob-firstonly'),
                        firstName: 'Bob',
                        lastName: '',
                        expectedFullName: 'Bob'
                    },
                    {
                        email: objectGenerator.generateEmailByPrefix('smith-lastonly'),
                        firstName: '',
                        lastName: 'Smith',
                        expectedFullName: 'Smith'
                    },
                    {
                        email: objectGenerator.generateEmailByPrefix('noname'),
                        firstName: '',
                        lastName: '',
                        expectedFullName: '-'
                    }
                ];

                await test.step('create three users with partial or missing names', async () => {
                    for (const user of users) {
                        await usersPage.createLocalUser(user.email, user.firstName, user.lastName);
                        createdUserEmails.push(user.email);
                    }
                });

                await test.step('verify Full name fallback for each user in Users table', async () => {
                    for (const user of users) {
                        await usersPage.verifyUserTableFullName(user.email, user.expectedFullName);
                    }
                });

                await test.step('verify partial and missing name fallbacks in Access Rules subject column', async () => {
                    for (const user of users) {
                        await createEditorTenantAccessRuleForUser(user.email);
                        await accessRulesPage.verifyAccessRuleSubject(
                            user.email,
                            user.expectedFullName === '-'
                                ? user.email
                                : `${user.email} ${user.expectedFullName}`
                        );
                    }
                });
            });

            test('Non-user subjects do not receive user full-name decoration (T6037694)', async () => {
                videoName = 'non-user-subject-plain-text';

                const serviceAccountName = objectGenerator.generateShortNameByPrefix('fnd-app');
                const adminToken = flowRootState.users.system_administrator.accessToken;
                let nonUserProject: Project;

                await test.step('create service account via API', async () => {
                    const serviceAccount = await serviceAccountUtils.createServiceAccount(
                        flowRootState.users.system_administrator.accessToken,
                        { name: serviceAccountName }
                    );
                    createdServiceAccountIds.push(serviceAccount.body.id);
                    nonUserProject = await orgUnitUtils.createProjectAndValidate({
                        name: objectGenerator.generateShortNameByPrefix('fnd-non-user-project'),
                        clusterId: flowRootState.cluster.uuid,
                        resources: [{ gpu: { deserved: 0, limit: 0 } }]
                    });
                    createdProjects.push(nonUserProject);
                });

                await test.step('create access rules with service account as subject', async () => {
                    await createEditorTenantAccessRuleForServiceAccount(serviceAccountName);
                    await authorizationUtils.createAccessRuleForServiceAccount(
                        serviceAccountName,
                        flowRootState.users.editor.role.name,
                        ScopeType.Project,
                        nonUserProject.id,
                        adminToken
                    );
                    const defaultDepartment = await orgUnitUtils.getDefaultDepartment();
                    await authorizationUtils.createAccessRuleForServiceAccount(
                        serviceAccountName,
                        flowRootState.users.editor.role.name,
                        ScopeType.Department,
                        defaultDepartment.id,
                        adminToken
                    );
                });

                await test.step('verify subject column shows service account name as plain text', async () => {
                    await accessRulesPage.goToPage();
                    await accessRulesPage.searchByName(serviceAccountName);
                    const row = await qTable.getEntityData('Subject', serviceAccountName);
                    expect(row).not.toBeNull();
                    expect(row!['Subject']).toBe(serviceAccountName);
                    expect(row!['Subject']).not.toMatch(/@/);
                });

                await test.step('verify project subject details keep service account name as plain text', async () => {
                    await projectPage.goToPage();
                    const projectSubjects = await projectPage.getProjectSubjectInfo(nonUserProject.name);
                    const projectSubjectRow = projectSubjects.find((subjectRow) =>
                        Object.values(subjectRow).some((value) => value.includes(serviceAccountName))
                    );
                    expect(
                        projectSubjectRow,
                        `Expected project subject details to include "${serviceAccountName}". Actual rows: ${JSON.stringify(projectSubjects)}`
                    ).toBeTruthy();
                    expect(projectSubjectRow!['Subject']).toBe(serviceAccountName);
                    expect(projectSubjectRow!['Subject']).not.toMatch(/@/);
                });

                await test.step('verify department subject details keep service account name as plain text', async () => {
                    await departmentPage.goToPage();
                    const departmentSubjects = await departmentPage.getDepartmentSubjectInfo('default');
                    const departmentSubjectRow = departmentSubjects.find((subjectRow) =>
                        Object.values(subjectRow).some((value) => value.includes(serviceAccountName))
                    );
                    expect(
                        departmentSubjectRow,
                        `Expected department subject details to include "${serviceAccountName}". Actual rows: ${JSON.stringify(departmentSubjects)}`
                    ).toBeTruthy();
                    expect(departmentSubjectRow!['Subject']).toBe(serviceAccountName);
                    expect(departmentSubjectRow!['Subject']).not.toMatch(/@/);
                });
            });

            test('Create local user without First Name or Last Name falls back to email only (T6037695)', async () => {
                videoName = 'local-user-email-only-fallback';

                const noNameEmail = objectGenerator.generateEmailByPrefix('runai-noname');

                await test.step('create local user without first or last name', async () => {
                    await usersPage.createLocalUser(noNameEmail, '', '');
                    createdUserEmails.push(noNameEmail);
                });

                await test.step('verify Users table displays email-only fallback', async () => {
                    await usersPage.verifyUserTableFullName(noNameEmail, '-');
                });

                await test.step('verify Access Rules displays email-only fallback', async () => {
                    await createEditorTenantAccessRuleForUser(noNameEmail);
                    await accessRulesPage.verifyAccessRuleSubject(noNameEmail, noNameEmail);
                });
            });

            test('Edit first and last name for an existing local user (T6037698)', async () => {
                videoName = 'edit-local-user-full-name';

                const editEmail = objectGenerator.generateEmailByPrefix('runai-edit-local');
                const localFirstName = 'Runai';
                const localLastName = 'Tester';
                const updatedLocalFirstName = 'Updated';
                const updatedLocalLastName = 'Local';
                const updatedLocalFullName = `${updatedLocalFirstName} ${updatedLocalLastName}`;
                const expectedDisplay = `${editEmail} ${updatedLocalFullName}`;

                await test.step('create local user and access rule referencing them', async () => {
                    await usersPage.createLocalUser(editEmail, localFirstName, localLastName);
                    createdUserEmails.push(editEmail);
                    await createEditorTenantAccessRuleForUser(editEmail);
                });

                await test.step('edit first and last name for the local user', async () => {
                    await usersPage.updateLocalUserFullName(
                        editEmail,
                        updatedLocalFirstName,
                        updatedLocalLastName
                    );
                });

                await test.step('verify Users table and user detail do not show invalid placeholders', async () => {
                    await usersPage.verifyUserTableFullName(editEmail, updatedLocalFullName);
                    await usersPage.verifyUserDetailDoesNotContainInvalidName(
                        editEmail,
                        updatedLocalFullName
                    );
                });

                await test.step('verify Access Rules reflects the updated full name', async () => {
                    await accessRulesPage.verifyAccessRuleSubject(editEmail, expectedDisplay);
                });
            });

            test('Edit local user fallback behavior for partial and missing name fields (T6037699)', async () => {
                videoName = 'edit-local-user-fallbacks';

                const editEmail = objectGenerator.generateEmailByPrefix('runai-edit-fallback');

                await test.step('create local user for fallback edit validation', async () => {
                    await usersPage.createLocalUser(editEmail, 'Original', 'User');
                    createdUserEmails.push(editEmail);
                    await createEditorTenantAccessRuleForUser(editEmail);
                });

                await test.step('verify both names display after edit', async () => {
                    await usersPage.updateLocalUserFullName(editEmail, 'Edited', 'Fallback');
                    await usersPage.verifyUserTableFullName(editEmail, 'Edited Fallback');
                });

                await test.step('verify first-name-only fallback after clearing last name', async () => {
                    await usersPage.updateLocalUserFullName(editEmail, 'Edited', '');
                    await usersPage.verifyUserTableFullName(editEmail, 'Edited');
                });

                await test.step('verify last-name-only fallback after clearing first name', async () => {
                    await usersPage.updateLocalUserFullName(editEmail, '', 'Fallback');
                    await usersPage.verifyUserTableFullName(editEmail, 'Fallback');
                });

                await test.step('verify email-only fallback after clearing both names', async () => {
                    await usersPage.updateLocalUserFullName(editEmail, '', '');
                    await usersPage.verifyUserTableFullName(editEmail, '-');
                    await usersPage.verifyUserDetailDoesNotContainInvalidName(editEmail);
                });

                await test.step('verify Access Rules reflects email-only fallback after clearing names', async () => {
                    await accessRulesPage.verifyAccessRuleSubject(editEmail, editEmail);
                });
            });

            test('Create local user with both First Name and Last Name (T6037703)', async () => {
                videoName = 'create-local-user-full-name';

                const fullNameEmail = objectGenerator.generateEmailByPrefix('runai-fullname');
                const localFirstName = 'Runai';
                const localLastName = 'Tester';
                const localFullName = `${localFirstName} ${localLastName}`;
                const expectedDisplay = `${fullNameEmail} ${localFullName}`;

                await test.step('create local user with visible optional name fields', async () => {
                    await usersPage.openCreateUserForm();
                    await usersPage.submitCreateUserForm(fullNameEmail, localFirstName, localLastName);
                    createdUserEmails.push(fullNameEmail);
                });

                await test.step('verify Users table displays the local full name', async () => {
                    await usersPage.verifyUserTableFullName(fullNameEmail, localFullName);
                });

                await test.step('verify Access Rules displays the local user full name', async () => {
                    await createEditorTenantAccessRuleForUser(fullNameEmail);
                    await accessRulesPage.verifyAccessRuleSubject(fullNameEmail, expectedDisplay);
                });
            });

            test('Create local users with only First Name or only Last Name (T6037704)', async () => {
                videoName = 'create-local-user-partial-names';

                const partialUsers = [
                    {
                        email: objectGenerator.generateEmailByPrefix('runai-firstonly'),
                        firstName: 'Firstonly',
                        lastName: '',
                        expectedFullName: 'Firstonly'
                    },
                    {
                        email: objectGenerator.generateEmailByPrefix('runai-lastonly'),
                        firstName: '',
                        lastName: 'Lastonly',
                        expectedFullName: 'Lastonly'
                    }
                ];

                await test.step('create local users with partial name fields', async () => {
                    for (const user of partialUsers) {
                        await usersPage.createLocalUser(user.email, user.firstName, user.lastName);
                        createdUserEmails.push(user.email);
                    }
                });

                await test.step('verify Users table partial-name fallback display', async () => {
                    for (const user of partialUsers) {
                        await usersPage.verifyUserTableFullName(user.email, user.expectedFullName);
                    }
                });

                await test.step('verify Access Rules partial-name fallback display', async () => {
                    for (const user of partialUsers) {
                        await createEditorTenantAccessRuleForUser(user.email);
                        await accessRulesPage.verifyAccessRuleSubject(
                            user.email,
                            user.expectedFullName === '-'
                                ? user.email
                                : `${user.email} ${user.expectedFullName}`
                        );
                    }
                });
            });

            test('Pages without user entities are unaffected by Full Name Display changes (T6037705)', async () => {
                videoName = 'pages-without-user-entities-unaffected';

                await test.step('open Nodes page under Resources', async () => {
                    await browserEngine.currentPage
                        .locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                        .click({ timeout: 10000 });
                    await browserEngine.currentPage
                        .locator(`[aid="${MENU_SUB_ITEMS.Nodes}"]`)
                        .click({ timeout: 10000 });
                    await browserEngine.currentPage.waitForLoadState('networkidle');
                    await expect(browserEngine.currentPage.getByText('Nodes').first()).toBeVisible();
                });

                await test.step('verify no unexpected user full-name decoration appears', async () => {
                    const tableHeader = browserEngine.currentPage.locator('thead');
                    await expect(tableHeader).not.toContainText(/undefined|null/);
                    await expect(tableHeader).not.toContainText('Full name');
                });

                await test.step('verify basic table action remains available', async () => {
                    const columnsButton = browserEngine.currentPage.getByRole('button', { name: 'Columns' });
                    await expect(columnsButton).toBeVisible();
                    await columnsButton.click();
                    await expect(browserEngine.currentPage.locator('.q-menu').first()).toBeVisible();
                    await browserEngine.currentPage.keyboard.press('Escape');
                });
            });
        });

        test.describe('Full name display across UI surfaces and pickers', function () {
            test('User picker searches across first name, last name, full name, and email (T6037696)', async () => {
                videoName = 'user-picker-unified-search';

                const searchUser = {
                    email: objectGenerator.generateEmailByPrefix('tnguyen-search'),
                    firstName: 'Taylor',
                    lastName: 'Nguyen',
                    fullName: 'Taylor Nguyen'
                };
                const expectedDisplay = `${searchUser.email} ${searchUser.fullName}`;

                await test.step('create local user with searchable first and last name', async () => {
                    await usersPage.createLocalUser(
                        searchUser.email,
                        searchUser.firstName,
                        searchUser.lastName
                    );
                    createdUserEmails.push(searchUser.email);
                });

                await test.step('verify user picker finds the user by first name, last name, full name, and email', async () => {
                    await accessRulesPage.goToPage();
                    await accessRulesPage.clickNewAccessRule();
                    await accessRulesPage.selectSubjectType('User');

                    const searchCases = [
                        { label: 'first name', term: searchUser.firstName },
                        { label: 'last name', term: searchUser.lastName },
                        { label: 'full name', term: searchUser.fullName },
                        { label: 'email', term: searchUser.email }
                    ];
                    for (const { label, term } of searchCases) {
                        await test.step(`search by ${label}`, async () => {
                            await accessRulesPage.verifyUserPickerResult(
                                term,
                                searchUser.email,
                                searchUser.firstName,
                                searchUser.lastName
                            );
                        });
                    }

                    await browserEngine.currentPage.locator(accessRulesPage.cancelBtn).click();
                });

                await test.step('verify saved access rule displays the selected user with full name', async () => {
                    await createEditorTenantAccessRuleForUser(searchUser.email);
                    await accessRulesPage.verifyAccessRuleSubject(searchUser.email, expectedDisplay);
                });
            });

            test('User picker allows manual email input when search result is unavailable (T6037697)', async () => {
                videoName = 'user-picker-manual-email';

                const externalEmail = objectGenerator.generateEmailByPrefix('external-pending');

                await test.step('create access rule by manually entering an unknown user email', async () => {
                    await createEditorTenantAccessRuleForUser(externalEmail);
                });

                await test.step('verify manual email renders email-only without invalid placeholders', async () => {
                    await accessRulesPage.verifyAccessRuleSubject(externalEmail, externalEmail);
                    await expect(browserEngine.currentPage.getByText(externalEmail).first()).not.toHaveText(
                        /undefined|null/
                    );
                });
            });

            test('Users page and user detail surfaces display email Full Name (T6037700)', async () => {
                videoName = 'users-page-and-detail-full-name';

                const detailsEmail = objectGenerator.generateEmailByPrefix('runai-detail-user');
                const detailsFirstName = 'Details';
                const detailsLastName = 'Surface';
                const detailsFullName = `${detailsFirstName} ${detailsLastName}`;
                const adminEmail = flowRootState.users.system_administrator.email;

                await test.step('create local user with first and last name', async () => {
                    await usersPage.createLocalUser(detailsEmail, detailsFirstName, detailsLastName);
                    createdUserEmails.push(detailsEmail);
                });

                await test.step('verify Users table displays the full name', async () => {
                    await usersPage.verifyUserTableFullName(detailsEmail, detailsFullName);
                });

                await test.step('verify Users table Created by column displays user identity safely', async () => {
                    await usersPage.navigateToUsers();
                    await usersPage.searchByName(detailsEmail);
                    const row = await qTable.getEntityData('User', detailsEmail);
                    expect(row).not.toBeNull();
                    const createdByDisplay = row!['Created by'];
                    expect(createdByDisplay).toBeTruthy();
                    expect(createdByDisplay).not.toMatch(/undefined|null/);
                    const normalizedCreatedBy = createdByDisplay.replace(/\s+/g, ' ').trim();
                    const normalizedAdminEmail = adminEmail.replace(/\s+/g, ' ').trim();
                    expect(
                        createdByDisplay === '-' || normalizedCreatedBy.includes(normalizedAdminEmail)
                    ).toBe(true);
                });

                await test.step('verify user detail surface displays the user identity without invalid placeholders', async () => {
                    await usersPage.verifyUserDetailDoesNotContainInvalidName(detailsEmail, detailsFullName);
                });
            });

            test('Long names render without layout regressions and expose full identity across surfaces (T6037701)', async () => {
                videoName = 'long-name-display-tooltip';

                const longNameEmail = objectGenerator.generateEmailByPrefix('alexandria-longname');
                const longFirstName = 'Alexandria-Cassandra';
                const longLastName = 'Montgomery-Worthington';
                const longFullName = `${longFirstName} ${longLastName}`;
                const expectedDisplay = `${longNameEmail} ${longFullName}`;

                await test.step('create user with long first and last name', async () => {
                    await usersPage.createLocalUser(longNameEmail, longFirstName, longLastName);
                    createdUserEmails.push(longNameEmail);
                });

                await test.step('verify long name appears in Users table without invalid placeholders', async () => {
                    await usersPage.verifyUserTableFullName(longNameEmail, longFullName);
                    const row = browserEngine.currentPage
                        .getByRole('row')
                        .filter({ hasText: longNameEmail })
                        .first();
                    await expect(row).toBeVisible();
                    const rowHeight = await row.evaluate((element) => element.getBoundingClientRect().height);
                    expect(rowHeight).toBeGreaterThan(0);
                    expect(rowHeight).toBeLessThan(120);
                });

                await test.step('verify Users table row keeps the full identity available', async () => {
                    const row = browserEngine.currentPage
                        .getByRole('row')
                        .filter({ hasText: longNameEmail })
                        .first();
                    await row.hover();
                    await expect(row).toContainText(longNameEmail);
                    await expect(row).toContainText(longFullName);
                });

                await test.step('verify long name display is consistent in Access Rules', async () => {
                    await createEditorTenantAccessRuleForUser(longNameEmail);
                    await accessRulesPage.verifyAccessRuleSubject(longNameEmail, expectedDisplay);
                });
            });

            test('System navigation and user settings display current user identity without invalid placeholders (T6037702)', async () => {
                videoName = 'system-navigation-user-identity';

                const adminEmail = flowRootState.users.system_administrator.email;

                await test.step('verify profile menu displays current user identity safely', async () => {
                    await browserEngine.currentPage.locator(usersPage.profileIcon).click();
                    const profileMenu = browserEngine.currentPage
                        .locator('.q-menu')
                        .filter({ hasText: adminEmail })
                        .first();
                    await expect(profileMenu).toBeVisible();
                    await expect(profileMenu).not.toContainText(/undefined|null/);
                });

                await test.step('verify user settings displays current user identity safely after refresh', async () => {
                    await browserEngine.currentPage.reload();
                    await browserEngine.currentPage.waitForLoadState('networkidle');
                    await usersPage.clickOnSettings();
                    const settingsIdentity = browserEngine.currentPage.getByText(adminEmail).first();
                    await expect(settingsIdentity).toBeVisible();
                    await expect(settingsIdentity).not.toHaveText(/undefined|null/);
                });
            });
        });

        test.describe('Additional user identity surfaces', function () {
            test.beforeAll(async function () {
                workloadSubmissionProject = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('fullname-workload-project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 0, limit: 0 } }]
                });
            });

            test.afterAll(async function () {
                if (!workloadSubmissionProject) return;

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(
                        workloadSubmissionProject.id,
                        workloadSubmissionProject.name
                    );
                } catch (err) {
                    logger.error(`Failed to delete project ${workloadSubmissionProject.name}: ${err}`);
                }
            });

            test('Workload submission Access List picker supports unified user search and manual email entry (T6037711)', async () => {
                videoName = 'workload-access-list-user-search';

                const pickerEmail = objectGenerator.generateEmailByPrefix('morgan-lee');
                const pickerFirstName = 'Morgan';
                const pickerLastName = 'Lee';
                const pickerFullName = `${pickerFirstName} ${pickerLastName}`;
                const manualEmail = objectGenerator.generateEmailByPrefix('manual-access-list');
                const workloadName = objectGenerator.generateNameByPrefix('fullname-training-access-list');

                await test.step('create local user for access-list picker search', async () => {
                    await usersPage.createLocalUser(pickerEmail, pickerFirstName, pickerLastName);
                    createdUserEmails.push(pickerEmail);
                });

                await test.step('open workload submission serving port access picker when available', async () => {
                    await workloadPage.navigateToWorkloads();
                    const createWorkloadButton = browserEngine.currentPage
                        .locator(workloadPage.addNewWorkloadBtn)
                        .first();
                    await expect(createWorkloadButton).toBeVisible({ timeout: 10000 });
                    await expect(createWorkloadButton).toBeEnabled({ timeout: 10000 });

                    await hybridForm.openTrainingForm();
                    await hybridForm.selectClusterFromDropDown(flowRootState.cluster.name);
                    await hybridForm.workloadFormSelectProject(workloadSubmissionProject.name);
                    await hybridForm.workloadFormSelectTemplate();
                    await hybridForm.fillTrainingName(workloadName);
                    const continueButton = browserEngine.currentPage.locator(
                        hybridForm.createTrainingContinueBtn
                    );
                    await expect(continueButton).toBeEnabled({ timeout: 10000 });
                    await continueButton.click();

                    await hybridForm.fillImageInput(config.quickStartImage.ubuntu);
                    await hybridForm.clickOnSubSectionByLabel('Tools');
                    await hybridForm.addToolTypeByLabel('Jupyter');

                    const toolAccessEditButton = browserEngine.currentPage
                        .locator(
                            hybridForm.hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn
                        )
                        .first();
                    await expect(toolAccessEditButton).toBeVisible({ timeout: 10000 });
                    await toolAccessEditButton.click();
                    await browserEngine.currentPage
                        .locator(hybridForm.hybridEnvironmentSection.servingPortAccess.servingPortAccessUsers)
                        .click();
                });

                await test.step('verify access-list picker finds the same user by name and email', async () => {
                    const modal = browserEngine.currentPage.locator(
                        hybridForm.hybridEnvironmentSection.servingPortAccess.servingPortAccessModalOptions
                    );
                    const subjectInput = modal.locator(accessRulesPage.addableMultiSelect);

                    for (const searchTerm of [pickerFirstName, pickerLastName, pickerFullName, pickerEmail]) {
                        await subjectInput.fill(searchTerm);
                        const option = browserEngine.currentPage
                            .locator(accessRulesPage.dropdownMenu)
                            .last()
                            .locator(accessRulesPage.dropdownOption)
                            .filter({ hasText: pickerEmail })
                            .filter({ hasText: pickerFullName })
                            .first();
                        await expect(option).toBeVisible({ timeout: 10000 });
                    }
                });

                await test.step('verify manual email input renders safely without creating a user', async () => {
                    const modal = browserEngine.currentPage.locator(
                        hybridForm.hybridEnvironmentSection.servingPortAccess.servingPortAccessModalOptions
                    );
                    const subjectInput = modal.locator(accessRulesPage.addableMultiSelect);
                    await subjectInput.fill(manualEmail);
                    await expect(
                        browserEngine.currentPage.getByText(`${manualEmail} (new user)`).first()
                    ).toBeVisible();
                    await expect(modal).not.toContainText(/undefined|null/);
                });

                await test.step('verify selected access-list users render with full-name and email-only display', async () => {
                    const modal = browserEngine.currentPage.locator(
                        hybridForm.hybridEnvironmentSection.servingPortAccess.servingPortAccessModalOptions
                    );
                    const subjectInput = modal.locator(accessRulesPage.addableMultiSelect);

                    await subjectInput.fill(pickerEmail);
                    await browserEngine.currentPage
                        .locator(accessRulesPage.dropdownMenu)
                        .last()
                        .locator(accessRulesPage.dropdownOption)
                        .filter({ hasText: pickerEmail })
                        .filter({ hasText: pickerFullName })
                        .first()
                        .click();
                    await expect(modal.getByText(pickerEmail).first()).toBeVisible();
                    await expect(modal.getByText(pickerFullName).first()).toBeVisible();

                    await subjectInput.fill(manualEmail);
                    await browserEngine.currentPage.getByText(`${manualEmail} (new user)`).first().click();
                    await expect(modal.getByText(manualEmail).first()).toBeVisible();
                    await expect(modal).not.toContainText(/undefined|null/);
                });

                await test.step('cancel workload submission without creating a workload', async () => {
                    await browserEngine.currentPage
                        .locator(
                            hybridForm.hybridEnvironmentSection.servingPortAccess.servingPortAccessCancelBtn
                        )
                        .click();
                    await browserEngine.currentPage.locator(workloadPage.headerLeavePageBtn).click();
                    await browserEngine.currentPage.locator(workloadPage.modalLeaveBtn).click();
                });
            });

            test('Access, project, department, workload, and event surfaces display user identities safely (T6037712)', async () => {
                videoName = 'identity-surfaces-full-name-regression';

                const surfaceEmail = objectGenerator.generateEmailByPrefix('surface-coverage');
                const surfaceFirstName = 'Surface';
                const surfaceLastName = 'Coverage';
                const surfaceFullName = `${surfaceFirstName} ${surfaceLastName}`;
                const surfaceDisplay = `${surfaceEmail} ${surfaceFullName}`;
                const adminToken = flowRootState.users.system_administrator.accessToken;
                const invalidDisplayPlaceholder = /^(undefined|null)$/i;

                await test.step('create user and access rule for identity surface coverage', async () => {
                    await usersPage.createLocalUser(surfaceEmail, surfaceFirstName, surfaceLastName);
                    createdUserEmails.push(surfaceEmail);
                    createdAccessRuleSubjects.push(surfaceEmail);

                    const defaultDepartment = await orgUnitUtils.getDefaultDepartment();

                    await authorizationUtils.createAccessRuleForUser(
                        adminToken,
                        surfaceEmail,
                        flowRootState.users.editor.role.id,
                        SubjectType.User
                    );
                    await authorizationUtils.createAccessRuleForUser(
                        adminToken,
                        surfaceEmail,
                        flowRootState.users.editor.role.id,
                        SubjectType.User,
                        workloadSubmissionProject.id,
                        ScopeType.Project
                    );
                    await authorizationUtils.createAccessRuleForUser(
                        adminToken,
                        surfaceEmail,
                        flowRootState.users.editor.role.id,
                        SubjectType.User,
                        defaultDepartment.id,
                        ScopeType.Department
                    );
                });

                await test.step('verify Access Rules subject surface uses email Full Name', async () => {
                    await accessRulesPage.verifyAccessRuleSubject(surfaceEmail, surfaceDisplay);
                });

                await test.step('verify Project subject details surface uses email Full Name', async () => {
                    await projectPage.goToPage();
                    await projectPage.waitForProjectToBeReady(workloadSubmissionProject.name);
                    await projectPage.openProjectSubjectDetails(workloadSubmissionProject.name);

                    const projectSubjectsDialog =
                        browserEngine.currentPage.locator('.dialog-content-container');
                    await expect(projectSubjectsDialog).toBeVisible({ timeout: 15000 });

                    const subjectCell = projectSubjectsDialog
                        .locator('table.q-table tbody tr')
                        .filter({ hasText: surfaceEmail.slice(0, 24) })
                        .locator('td')
                        .first();

                    await expect(subjectCell).toBeVisible({ timeout: 10000 });
                    await subjectCell.hover();

                    const tooltip = browserEngine.currentPage.locator('[role="tooltip"]').last();
                    await expect(tooltip).toBeVisible({ timeout: 10000 });
                    await expect(tooltip).toContainText(surfaceEmail);
                    await expect(tooltip).toContainText(surfaceFullName);
                    await expect(tooltip).not.toContainText(invalidDisplayPlaceholder);

                    await projectPage.orgUnitsForm.getCloseBtn().click();
                });

                await test.step('verify Department subject details surface uses email Full Name', async () => {
                    await departmentPage.goToPage();
                    const departmentSubjects = await departmentPage.getDepartmentSubjectInfo('default');
                    const normalizedSurfaceDisplay = surfaceDisplay.replace(/\s+/g, ' ').trim();
                    const surfaceRows = departmentSubjects.filter((row) =>
                        Object.values(row).some(
                            (value) => typeof value === 'string' && value.includes(surfaceEmail)
                        )
                    );
                    expect(
                        surfaceRows.some((row) =>
                            Object.values(row).some(
                                (value) =>
                                    typeof value === 'string' &&
                                    value.replace(/\s+/g, ' ').trim().includes(normalizedSurfaceDisplay)
                            )
                        ),
                        `Expected department subject details to include "${surfaceDisplay}". Actual rows: ${JSON.stringify(surfaceRows)}`
                    ).toBeTruthy();
                    expect(
                        surfaceRows.some((row) =>
                            Object.values(row).some(
                                (value) =>
                                    typeof value === 'string' && invalidDisplayPlaceholder.test(value.trim())
                            )
                        )
                    ).toBeFalsy();
                });

                await test.step('verify Workloads surface remains free of invalid placeholders', async () => {
                    await workloadPage.navigateToWorkloads();
                    const rows = await qTable.getTableData();
                    const surfaceRows = rows.filter((row) =>
                        Object.values(row).some(
                            (value) => typeof value === 'string' && value.includes(surfaceEmail)
                        )
                    );
                    expect(
                        surfaceRows.some((row) =>
                            Object.values(row).some(
                                (value) =>
                                    typeof value === 'string' && invalidDisplayPlaceholder.test(value.trim())
                            )
                        )
                    ).toBeFalsy();
                });

                await test.step('verify Event History surface renders user identities safely', async () => {
                    await eventHistoryPage.goToPage();
                    const rows = await qTable.getTableData();
                    const normalizedSurfaceDisplay = surfaceDisplay.replace(/\s+/g, ' ').trim();
                    const surfaceRows = rows.filter((row) =>
                        Object.values(row).some(
                            (value) => typeof value === 'string' && value.includes(surfaceEmail)
                        )
                    );
                    const surfaceRowValues = surfaceRows
                        .flatMap((row) => Object.values(row))
                        .filter((value): value is string => typeof value === 'string')
                        .map((value) => value.replace(/\s+/g, ' ').trim());
                    const surfaceDisplayPresent = surfaceRowValues.some((value) =>
                        value.includes(normalizedSurfaceDisplay)
                    );
                    expect(surfaceRows.length === 0 || surfaceDisplayPresent).toBe(true);
                    expect(
                        surfaceRowValues.some((value) => invalidDisplayPlaceholder.test(value))
                    ).toBeFalsy();
                });
            });

            test('Created By columns across resource tables use safe user identity rendering (T6037713)', async () => {
                videoName = 'created-by-columns-full-name-regression';

                const pagesToVerify = [
                    {
                        name: 'Workloads',
                        navigate: (): Promise<void> => workloadPage.navigateToWorkloads()
                    },
                    {
                        name: 'Templates',
                        navigate: (): Promise<void> => templatePage.goToPage()
                    },
                    {
                        name: 'Policies',
                        navigate: (): Promise<void> => policyPage.navigateToPolicyPage()
                    },
                    {
                        name: 'Roles',
                        navigate: (): Promise<void> => rolesPage.goToPage()
                    },
                    {
                        name: 'Users',
                        navigate: (): Promise<void> => usersPage.navigateToUsers()
                    }
                ];

                for (const pageToVerify of pagesToVerify) {
                    await test.step(`verify ${pageToVerify.name} identity columns render safely`, async () => {
                        await pageToVerify.navigate();
                        const createdByHeader = browserEngine.currentPage
                            .getByRole('columnheader')
                            .filter({ hasText: /^Created by$/i })
                            .first();
                        const headerVisible = await createdByHeader
                            .isVisible({ timeout: 3000 })
                            .catch(() => false);
                        if (!headerVisible) {
                            return;
                        }
                        await expect(createdByHeader).toBeVisible();
                        const rows = await qTable.getTableData();
                        const emailRowsToCheck = rows
                            .map((row) => {
                                const display = row['Created by']?.trim() ?? '';
                                const normalized = display.replace(/\s+/g, ' ').trim();
                                const email =
                                    normalized.match(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i)?.[0] ?? '';
                                return { row, display, normalized, email };
                            })
                            .filter(({ display, email }) => display && display !== '-' && email);

                        for (const { row, display, normalized, email } of emailRowsToCheck) {
                            expect(display).not.toMatch(/undefined|null/);
                            expect(
                                normalized,
                                `${pageToVerify.name} Created by should include the rendered email. Actual row: ${JSON.stringify(row)}`
                            ).toContain(email);
                        }
                    });
                }
            });
        });

        test.describe('Users minimal API and permission coverage', function () {
            const usersMinimalRequiredPermissionSetNames = [
                'Workspace edit',
                'Training edit',
                'Inference edit',
                'YAML workload edit'
            ];

            test('users-minimal API searches by first name, last name, full name, and email (T6037706)', async () => {
                videoName = 'users-minimal-api-search';

                const apiSearchEmail = objectGenerator.generateEmailByPrefix('jordan-parker');
                const apiSearchFirstName = 'Jordan';
                const apiSearchLastName = 'Parker';
                const apiSearchFullName = `${apiSearchFirstName} ${apiSearchLastName}`;
                const adminToken = flowRootState.users.system_administrator.accessToken;

                await test.step('create local user for users-minimal search validation', async () => {
                    await usersPage.createLocalUser(apiSearchEmail, apiSearchFirstName, apiSearchLastName);
                    createdUserEmails.push(apiSearchEmail);
                });

                await test.step('search users-minimal by first name, last name, full name, and email', async () => {
                    const searchTerms = [
                        apiSearchFirstName,
                        apiSearchLastName,
                        apiSearchFullName,
                        apiSearchEmail
                    ];

                    for (const searchTerm of searchTerms) {
                        const users = await usersUtils.searchUsersMinimal(adminToken, searchTerm);

                        expect(Array.isArray(users)).toBeTruthy();
                        const matchedUser = users.find((user) => user.username === apiSearchEmail);
                        expect(matchedUser).toBeTruthy();
                        expect(matchedUser.firstName).toBe(apiSearchFirstName);
                        expect(matchedUser.lastName).toBe(apiSearchLastName);
                    }
                });

                await test.step('search users-minimal with no-match term', async () => {
                    const users = await usersUtils.searchUsersMinimal(
                        adminToken,
                        objectGenerator.generateNameByPrefix('no-such-user-runai')
                    );

                    expect(Array.isArray(users)).toBeTruthy();
                    expect(users).toHaveLength(0);
                });
            });

            test('User Minimal Read Access is included in required permission sets (T6037708)', async () => {
                videoName = 'users-minimal-required-permission-sets';

                const adminToken = flowRootState.users.system_administrator.accessToken;

                await test.step('get permission set catalog', async () => {
                    authorizationApi.permissionsApi.accessToken = adminToken;
                    const response = await authorizationApi.permissionsApi.getPermissionSets();
                    expect(response.response.statusCode).toBe(200);
                    expect(response.body.permissionSets.length).toBeGreaterThan(0);

                    for (const permissionSetName of usersMinimalRequiredPermissionSetNames) {
                        const permissionSet = response.body.permissionSets.find(
                            (set) => set.name === permissionSetName
                        );
                        expect(
                            permissionSet,
                            `Permission set "${permissionSetName}" was not returned by the permission sets API`
                        ).toBeTruthy();

                        const usersMinimalPermission = permissionSet.permissions.find(
                            (permission) => `${permission.resourceType}` === 'users-minimal'
                        );
                        expect(usersMinimalPermission).toBeTruthy();
                        expect(usersMinimalPermission.actions).toContain('read');
                        expect(usersMinimalPermission.actions).not.toContain('create');
                        expect(usersMinimalPermission.actions).not.toContain('update');
                        expect(usersMinimalPermission.actions).not.toContain('delete');
                    }
                });
            });

            test('users-minimal permission is granted to predefined roles that use required permission sets (T6037709)', async () => {
                videoName = 'users-minimal-predefined-roles';

                const adminToken = flowRootState.users.system_administrator.accessToken;
                const rbacLookupFirstName = 'Rbac';
                const rbacLookupLastName = 'Lookup';
                const rbacLookupFullName = `${rbacLookupFirstName} ${rbacLookupLastName}`;
                const rbacLookupEmail = objectGenerator.generateEmailByPrefix('rbac-lookup');
                const editorUser = flowRootState.users.editor;

                await test.step('get predefined roles and verify users-minimal read permission', async () => {
                    const roles = await authorizationUtils.getRolesV2(
                        adminToken,
                        authorizationApi.rolesApi,
                        undefined,
                        true
                    );

                    const predefinedRoles = roles.filter((role) => role.custom === false);
                    const applicableRoles = predefinedRoles.filter((role) =>
                        role.permissionSets?.some((permissionSet) =>
                            usersMinimalRequiredPermissionSetNames.includes(permissionSet.name)
                        )
                    );

                    expect(applicableRoles.length).toBeGreaterThan(0);

                    for (const role of applicableRoles) {
                        const usersMinimalPermission = role.permissions.find(
                            (permission) => `${permission.resourceType}` === 'users-minimal'
                        );
                        expect(usersMinimalPermission).toBeTruthy();
                        expect(usersMinimalPermission.actions).toContain('read');
                    }
                });

                await test.step('verify Editor can see another user full-name display in UI lookup', async () => {
                    await usersPage.createLocalUser(rbacLookupEmail, rbacLookupFirstName, rbacLookupLastName);
                    createdUserEmails.push(rbacLookupEmail);

                    const rbacProject = await orgUnitUtils.createProjectAndValidate({
                        name: objectGenerator.generateShortNameByPrefix('fullname-rbac-project'),
                        clusterId: flowRootState.cluster.uuid,
                        resources: [{ gpu: { deserved: 0, limit: 0 } }]
                    });
                    createdProjects.push(rbacProject);

                    await authorizationUtils.createAccessRuleForUser(
                        adminToken,
                        editorUser.email,
                        flowRootState.users.editor.role.id,
                        SubjectType.User,
                        rbacProject.id,
                        ScopeType.Project
                    );

                    let editorBrowserEngine: BrowserEngine;
                    try {
                        editorBrowserEngine = new BrowserEngine();
                        await editorBrowserEngine.start();
                        const editorAuthPage = new AuthPage(editorBrowserEngine.currentPage);
                        const editorWorkloadPage = new WorkloadPage(editorBrowserEngine.currentPage);
                        const editorHybridForm = new HybridForm(editorBrowserEngine.currentPage);

                        await editorAuthPage.signIn(editorUser.email, editorUser.password);
                        await editorWorkloadPage.navigateToWorkloads();
                        const createWorkloadButton = editorBrowserEngine.currentPage
                            .locator(editorWorkloadPage.addNewWorkloadBtn)
                            .first();
                        await expect(createWorkloadButton).toBeVisible({ timeout: 10000 });
                        await expect(createWorkloadButton).toBeEnabled({ timeout: 10000 });

                        await editorHybridForm.openTrainingForm();
                        await editorHybridForm.selectClusterFromDropDown(flowRootState.cluster.name);
                        await editorHybridForm.workloadFormSelectProject(rbacProject.name);
                        await editorHybridForm.workloadFormSelectTemplate();
                        await editorHybridForm.fillTrainingName(
                            objectGenerator.generateNameByPrefix('fullname-rbac-training')
                        );
                        const continueButton = editorBrowserEngine.currentPage.locator(
                            editorHybridForm.createTrainingContinueBtn
                        );
                        await expect(continueButton).toBeEnabled({ timeout: 10000 });
                        await continueButton.click();

                        await editorHybridForm.fillImageInput(config.quickStartImage.ubuntu);
                        await editorHybridForm.clickOnSubSectionByLabel('Tools');
                        await editorHybridForm.addToolTypeByLabel('Jupyter');

                        const toolAccessEditButton = editorBrowserEngine.currentPage
                            .locator(
                                editorHybridForm.hybridEnvironmentSection.servingPortAccess
                                    .servingPortAccessEditBtn
                            )
                            .first();
                        await expect(toolAccessEditButton).toBeVisible({ timeout: 10000 });
                        await toolAccessEditButton.click();
                        await editorBrowserEngine.currentPage
                            .locator(
                                editorHybridForm.hybridEnvironmentSection.servingPortAccess
                                    .servingPortAccessUsers
                            )
                            .click();

                        const modal = editorBrowserEngine.currentPage.locator(
                            editorHybridForm.hybridEnvironmentSection.servingPortAccess
                                .servingPortAccessModalOptions
                        );
                        const subjectInput = modal.locator(accessRulesPage.addableMultiSelect);
                        await subjectInput.fill(rbacLookupEmail);
                        await expect(
                            editorBrowserEngine.currentPage
                                .locator(accessRulesPage.dropdownMenu)
                                .last()
                                .locator(accessRulesPage.dropdownOption)
                                .filter({ hasText: rbacLookupEmail })
                                .filter({ hasText: rbacLookupFullName })
                                .first()
                        ).toBeVisible({ timeout: 10000 });
                        await expect(modal).not.toContainText(/undefined|null/);
                    } finally {
                        if (editorBrowserEngine) {
                            await editorBrowserEngine.stopAndArchive(
                                test.info(),
                                `${videoName}-editor-session`
                            );
                        }
                    }
                });
            });
        });
    });
});
