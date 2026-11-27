/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    authUtils,
    config,
    flowRootState,
    idpUtils,
    logger,
    objectGenerator,
    operations,
    PodStatus,
    projectResourcesUtils,
    projectUtils,
    settings,
    tags,
    Utils,
    usersManager,
    versionUtils,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../../modules/helpers';
import {
    idpNameOidc,
    keycloakClientName,
    oidcSsoUser,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import {
    shouldRunTests,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

test.describe(
    `Submit Job with SSO OIDC, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`,
    { tag: [tags.services.identityManager, tags.services.keycloak] },
    function () {
        shouldRunTests(settings.isSelfHosted && !isOCP())(
            function () {
                let projectName: string;
                let namespace: string;

                test.beforeAll(async function () {
                    projectName = objectGenerator.generateNameByPrefix('project');
                    namespace = Utils.namespaceFromProjectName(projectName);

                    try {
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                    } catch (error) {
                        if (!error.message?.includes('404')) {
                            logger.debug(`IDP cleanup: ${error.message}`);
                        }
                    }

                    try {
                        await Keycloak.deleteRealm(realmName);
                        await new Promise((resolve) => setTimeout(resolve, 1000));
                    } catch (error) {
                        logger.debug(`Realm cleanup: ${error.message}`);
                    }

                    await Keycloak.setupSSORealmOIDC(realmName, keycloakClientName);
                    await authorizationUtils.createAccessRuleForUser(
                        flowRootState.users.system_administrator.accessToken,
                        oidcSsoUser,
                        flowRootState.users.system_administrator.role.id
                    );
                });
                test.afterAll(async function () {
                    try {
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                        await Keycloak.deleteRealm(realmName);
                        await operations.forceCleanup();
                    } catch (error) {
                        logger.error(error);
                    } finally {
                        await usersManager.restorePreSeededOidcIdp();
                    }
                });

                test('should define OIDC SSO configuration on test tenant and submit new job', async function () {
                    const token = await authUtils.authenticateUser(
                        oidcSsoUser,
                        config.controlPlaneAdminPassword,
                        true
                    );
                    const user = await authUtils.getUserV2(oidcSsoUser, token.idToken);
                    logger.debug(`user: ${JSON.stringify(user)}`);
                    const project = await projectResourcesUtils.createProject(
                        namespace,
                        {
                            name: projectName,
                            deservedGpus: 2,
                            maxAllowedGpus: 2,
                            permissions: { users: [oidcSsoUser] }
                        },
                        token.idToken
                    );
                    if (
                        versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_14
                        )
                    ) {
                        const runaiResearcherRoleBindingName =
                            projectUtils.getRunaiResearcherRoleBindingName();
                        await projectUtils.waitForRoleBindingForUserInProject(
                            namespace,
                            oidcSsoUser,
                            runaiResearcherRoleBindingName,
                            true
                        );
                    }
                    const workloadName = objectGenerator.generateNameByPrefix('w');
                    const submitRes = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        project.clusterUuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity'
                        },
                        token.idToken
                    );
                    const workload = submitRes.body;
                    expect(workload.name, 'submitted OIDC SSO workload should have the correct name').equal(workloadName);
                    expect(workload.projectId, 'submitted workload should be associated with the correct project').equal(project.id.toString());
                    expect(workload.spec.image, 'submitted workload should use the configured ubuntu image').equal(config.quickStartImage.ubuntu);
                    await waitUtils.waitForPods(workload.name, namespace, 1, undefined, [PodStatus.RUNNING]);
                });
            }
        );
    }
);
