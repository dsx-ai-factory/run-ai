/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { TokenResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    Tenant,
    TenantCreationRequest,
    TenantCreationResponse
} from '../../../../../../../modules/api/apis/tenantsManagerApi/generated/api';
import {
    authUtils,
    clusterService,
    clusterUtils,
    config,
    flowRootState,
    logger,
    nodePoolUtils,
    nodeUtils,
    objectGenerator,
    settings,
    tags,
    tenantUtils,
    Utils,
    waitUtils
} from '../../../../../../../modules/helpers';
import { clusterApi, orgUnitService } from '../../../../../../../modules/api';
import { ProjectCreationRequest } from '../../../../../../../modules/api/apis/orgUnitApi/generated/model/projectCreationRequest';
import { Project as OrgUnitProject } from '../../../../../../../modules/api/apis/orgUnitApi/generated/model/project';
import { Units } from '../../../../../../../modules/api/apis/orgUnitApi/generated/model/units';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { Browser } from 'playwright';
import {
    DisplayedCluster,
    ClusterDisplayedStatus,
    MetricsType
} from '../../../../../../../modules/api/apis/clusterApi';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Cluster } from '../../../../../../../modules/api/apis/backendClusterApi';

shouldRunTests(settings.isMultiTenancyCluster)(() => {
    test.describe(`Multi-Tenancy Control Plane Installation E2E: ${tags.teams.phoenix} ${tags.tool.api} ${tags.resource.tenant} ${tags.runOn.multiTenancyUpgrade}`, () => {
        let superAdminToken: TokenResponse;
        let tenantUrl: string;
        let browser: Browser | null = null;

        // Tenant data structure
        interface TenantTestData {
            name: string;
            response: TenantCreationResponse;
            specificUrl: string;
            userToken: TokenResponse;
            project: OrgUnitProject;
            workloadId: string;
            userName: string;
            cluster: DisplayedCluster;
        }

        const tenants: TenantTestData[] = [];
        const numberOfTenants = 3;

        test.beforeAll(async function () {
            if (!settings.isSelfHosted) {
                logger.debug('Retrieving Keycloak client secret from env');
            } else {
                settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient(
                    'console',
                    'default'
                );
            }

            const baseUrl = config.tenantUrl.replace('https://', '');
            tenantUrl = `https://console.${baseUrl}`;
            // tenantUrl = `https://${baseUrl}`;
        });

        test.afterAll(async function () {
            // Cleanup all tenants
            for (const tenant of tenants) {
                try {
                    if (tenant.response?.tenant?.id) {
                        await tenantUtils.deleteTenant(
                            superAdminToken.accessToken,
                            tenantUrl,
                            tenant.response.tenant.id,
                            true
                        );
                        logger.debug(`Tenant force deleted: ${tenant.response.tenant.id}`);
                    }
                } catch (error) {
                    logger.warn(`Failed to cleanup tenant ${tenant.name}: ${error.message}`);
                }
            }

            if (browser) {
                await browser.close();
                browser = null;
            }
        });

        test('should authenticate super admin and get token', async function () {
            superAdminToken = await authUtils.authenticateSuperAdminAppAndGetToken(
                'default',
                settings.tenantClientSecret,
                tenantUrl
            );
            logger.info(superAdminToken.accessToken);

            expect(superAdminToken.accessToken).to.not.be.undefined;
            expect(superAdminToken.accessToken).to.not.be.empty;
        });

        test('should create multiple tenants for isolation testing', async function () {
            logger.debug('Creating tenants');
            const timestamp = Date.now();
            for (let i = 0; i < numberOfTenants; i++) {
                const tenantName = `e2e-tenant-${i + 1}-${timestamp}`;
                const tenantResponse = await tenantUtils.createTenant(
                    superAdminToken.accessToken,
                    tenantUrl,
                    tenantName,
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword,
                    TenantCreationRequest.ContractTypeEnum.Normal,
                    null,
                    null
                );

                expect(tenantResponse.tenant.name).to.equal(tenantName);
                expect(tenantResponse.tenant.status).to.equal('Creating');
                expect(tenantResponse.tenant.id).to.not.be.null;
                expect(tenantResponse.additionalData['tenant_domain']).to.not.be.undefined;

                const tenantData: TenantTestData = {
                    name: tenantName,
                    response: tenantResponse,
                    specificUrl: tenantResponse.additionalData['tenant_domain'],
                    userToken: null,
                    project: null,
                    workloadId: null,
                    userName: `user-${tenantName}`,
                    cluster: null
                };

                // Print the tenant info
                logger.info(`Tenant ${i + 1} created successfully:`);
                logger.info(`  - Name: ${tenantData.name}`);
                logger.info(`  - ID: ${tenantData.response.tenant.id}`);
                logger.info(`  - UUID: ${tenantData.response.tenant.uuid}`);
                logger.info(`  - Status: ${tenantData.response.tenant.status}`);
                logger.info(`  - Domain: ${tenantData.specificUrl}`);
                logger.info(`  - User Name: ${tenantData.userName}`);

                tenants.push(tenantData);
            }
        });

        test('should wait for all tenants to be ready', async function () {
            logger.debug('Waiting for all tenants to be ready');
            for (const tenant of tenants) {
                const isTenantReady = async (): Promise<boolean> => {
                    try {
                        const getTenantStatus = await tenantUtils.getTenantById(
                            superAdminToken.accessToken,
                            tenantUrl,
                            tenant.response.tenant.id
                        );
                        logger.debug(`Tenant ${tenant.name} status: ${getTenantStatus.status}`);
                        return getTenantStatus.status === Tenant.StatusEnum.Ready;
                    } catch (error) {
                        logger.debug(`Failed to get tenant ${tenant.name} status: ${error.message}`);
                        return false;
                    }
                };

                await waitUtils.waitForCondition(
                    isTenantReady,
                    `tenant ${tenant.name} to become Ready`,
                    true,
                    120,
                    5
                );

                const readyTenant = await tenantUtils.getTenantById(
                    superAdminToken.accessToken,
                    tenantUrl,
                    tenant.response.tenant.id
                );
                expect(readyTenant.status).to.equal('Ready');
            }
        });

        test('should authenticate to all tenants and get user tokens', async function () {
            for (const tenant of tenants) {
                tenant.userToken = await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
                    superAdminToken.accessToken,
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword,
                    tenant.response.tenant.name,
                    tenant.specificUrl
                );

                expect(tenant.userToken.accessToken).to.not.be.undefined;
                expect(tenant.userToken.idToken).to.not.be.undefined;
            }
            logger.debug('Successfully authenticated to tenants');
        });

        test('should verify tenant uniqueness properties', async function () {
            logger.debug('Verifying tenant uniqueness properties...');
            const allTenants = await tenantUtils.getTenants(superAdminToken.accessToken, tenantUrl);

            // Filter to only our created tenants
            const testTenants = allTenants.filter((tenant) =>
                tenants.some((createdTenant) => createdTenant.response.tenant.id === tenant.id)
            );

            expect(testTenants.length).to.equal(numberOfTenants);

            // Verify all tenant IDs are unique
            const tenantIds = testTenants.map((tenant) => tenant.id);
            const uniqueTenantIds = [...new Set(tenantIds)];
            expect(uniqueTenantIds.length).to.equal(numberOfTenants);

            // Verify all tenant UUIDs are unique
            const tenantUuids = testTenants.map((tenant) => tenant.uuid);
            const uniqueTenantUuids = [...new Set(tenantUuids)];
            expect(uniqueTenantUuids.length).to.equal(numberOfTenants);

            // Verify all tenant names are unique
            const tenantNames = testTenants.map((tenant) => tenant.name);
            const uniqueTenantNames = [...new Set(tenantNames)];
            expect(uniqueTenantNames.length).to.equal(numberOfTenants);

            // Verify all console URLs are unique
            const consoleUrls = tenants.map((tenant) => tenant.specificUrl);
            const uniqueConsoleUrls = [...new Set(consoleUrls)];
            expect(uniqueConsoleUrls.length).to.equal(numberOfTenants);

            // Verify all mimirIds are unique (if they exist)
            const mimirIds = testTenants.filter((tenant) => tenant.mimirId).map((tenant) => tenant.mimirId);
            if (mimirIds.length > 0) {
                const uniqueMimirIds = [...new Set(mimirIds)];
                expect(uniqueMimirIds.length).to.equal(mimirIds.length);
            } else {
                logger.info('ⓘ No mimirIds found for verification');
            }
            logger.debug('Tenant uniqueness verification completed successfully');
        });

        test('create and validate clusters in tenants', async function () {
            const bootstrapTokens: string[] = [];
            for (const tenant of tenants) {
                const clusterName = `e2e-cluster-${tenant.name}-${Date.now()}`;
                const clusterDomain = `${config.tenantUrl}:6443`;
                const clusterVersion = await clusterUtils.getLatestClusterVersion(tenant.userToken.idToken);

                tenant.cluster = await clusterUtils.createCluster(
                    tenant.userToken.idToken,
                    clusterName,
                    clusterDomain,
                    clusterVersion
                );

                expect(tenant.cluster).to.not.be.undefined;
                expect(tenant.cluster.uuid).to.not.be.undefined;

                // The shared installer client no longer exists in new realms; the installer authenticates
                // with a per-cluster bootstrap token.
                const installationInfo = await clusterService.getClusterInfo(
                    tenant.userToken.idToken,
                    tenant.cluster.uuid,
                    tenant.cluster.version
                );
                expect(installationInfo.bootstrapToken, `bootstrap token for ${tenant.name}`).to.not.be.empty;
                bootstrapTokens.push(installationInfo.bootstrapToken);
            }

            expect(new Set(bootstrapTokens).size, 'bootstrap tokens should be unique per tenant').to.equal(
                numberOfTenants
            );

            // Verify all cluster UIDs are unique
            const clusterUids = tenants.map((tenant) => tenant.cluster.uuid);
            const uniqueClusterUids = [...new Set(clusterUids)];
            expect(uniqueClusterUids.length).to.equal(numberOfTenants);

            // Verify all cluster names are unique
            const clusterNames = tenants.map((tenant) => tenant.cluster.name);
            const uniqueClusterNames = [...new Set(clusterNames)];
            expect(uniqueClusterNames.length).to.equal(numberOfTenants);
        });

        test('create project per tenant', async function () {
            for (const tenant of tenants) {
                const projectName = objectGenerator.generateNameByPrefix(`project-${tenant.name}`);
                const namespace = Utils.namespaceFromProjectName(projectName);
                orgUnitService.projectApi.accessToken = tenant.userToken.accessToken;
                clusterApi.nodepoolsApi.accessToken = tenant.userToken.accessToken;
                flowRootState.cluster = new Cluster();
                flowRootState.cluster.uuid = tenant.cluster.uuid;
                const np = await nodePoolUtils.getDefaultNodePoolFromCluster();
                const projectCreationRequest: ProjectCreationRequest = {
                    name: projectName,
                    clusterId: tenant.cluster.uuid,
                    requestedNamespace: namespace,
                    resources: [
                        {
                            nodePool: { id: np.id.toString(), name: np.name },
                            gpu: { deserved: 1, limit: 1, overQuotaWeight: 1 },
                            cpu: { deserved: null, limit: null, overQuotaWeight: null },
                            memory: { deserved: null, limit: null, overQuotaWeight: null, units: Units.Mib }
                        }
                    ]
                };
                const createResponse = await orgUnitService.projectApi.createProject(projectCreationRequest);
                tenant.project = createResponse.body;

                expect(tenant.project).to.be.exist;
            }
        });

        test('should validate tenant isolation - each tenant can only see their own resources', async function () {
            for (let i = 0; i < tenants.length; i++) {
                logger.debug('Validating resource isolation for tenants');
                const currentTenant = tenants[i];
                orgUnitService.projectApi.accessToken = currentTenant.userToken.accessToken;
                const projectsResponse = await orgUnitService.projectApi.getProjects();
                const projects = projectsResponse.body.projects;

                // Current tenant should see their own project
                const ownProject = projects.find((p) => p.name === currentTenant.project.name);
                expect(ownProject).to.not.be.undefined;
                // Current tenant should NOT see other tenants' projects in their context
                for (let j = 0; j < tenants.length; j++) {
                    if (i !== j) {
                        const otherTenant = tenants[j];
                        const otherProject = projects.find((p) => p.name === otherTenant.project.name);
                        // In a properly isolated system, other tenant projects should NOT be visible
                        expect(otherProject).to.be.undefined;
                    }
                }
            }
        });

        test('should validate tenant deletion and cleanup', async function () {
            logger.info('Performing soft delete on tenants');
            // Test soft delete on first tenant
            const firstTenant = tenants[0];
            const softDeleted = await tenantUtils.deleteTenant(
                superAdminToken.accessToken,
                tenantUrl,
                firstTenant.response.tenant.id,
                false
            );

            expect(softDeleted.uid).to.equal(firstTenant.response.tenant.uuid);
            // Verify tenant is not visible in list
            const tenantList = await tenantUtils.getTenants(superAdminToken.accessToken, tenantUrl);
            const deletedTenant = tenantList.find((t) => t.uuid === firstTenant.response.tenant.uuid);
            expect(deletedTenant).to.be.undefined;

            // Verify Keycloak realm still exists for soft delete
            const keycloakRealm = await Keycloak.getRealmByName(firstTenant.response.tenant.name);
            expect(keycloakRealm).to.not.be.null;
            // Test hard delete on second tenant
            const secondTenant = tenants[1];
            const forceDeleted = await tenantUtils.deleteTenant(
                superAdminToken.accessToken,
                tenantUrl,
                secondTenant.response.tenant.id,
                true // force delete
            );

            expect(forceDeleted.uid).to.equal(secondTenant.response.tenant.uuid);

            // Verify Keycloak realm is completely removed for hard delete
            const hardDeletedRealm = await Keycloak.getRealmByName(secondTenant.response.tenant.name);
            expect(hardDeletedRealm).to.be.null;
        });
    });
});
