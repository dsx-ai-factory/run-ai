/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { TokenResponse } from '../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    TenantCreationRequest,
    TenantCreationResponse
} from '../../../../../modules/api/apis/tenantsManagerApi/generated/api';
import { ProjectPhase } from '../../../../../modules/api/apis/orgUnitApi';
import {
    authUtils,
    clusterUtils,
    config,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    settings,
    tags,
    tenantUtils,
    usersUtils,
    Utils,
    waitUtils,
    workloadUtils
} from '../../../../../modules/helpers';
import { WorkloadType } from '../../../../../modules/helpers/utils/WorkloadUtils';
import { BaseWorkload } from '../../../../../modules/helpers';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';
import { DatabaseConnection, db } from '../../../../../modules/db';
import { kubeCli } from '../../../../../modules/cli';
import { clusterApi } from '../../../../../modules/api';
import { ClusterState } from '../../../../../modules/api/apis/clusterApi';

test.describe(`Phoenix tests Delete Tenant SH Multi-Tenancy: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.tenant}`, () => {
    let tenantName: string;
    let token: TokenResponse;
    let tenant: TenantCreationResponse;
    let tenantUrl: string;

    test.beforeEach(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        const baseUrl = config.tenantUrl.replace('https://', '');
        tenantUrl = `https://console.${baseUrl}`;

        tenantName = `e2e${Date.now()}`;
        token = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            tenantUrl
        );
        logger.info(`Creating tenant: ${tenantName} with token: ${token.accessToken}`);
        tenant = await tenantUtils.createTenant(
            token.accessToken,
            tenantUrl,
            tenantName,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            TenantCreationRequest.ContractTypeEnum.Normal,
            null,
            null
        );
        logger.info(
            `Tenant created: status:${tenant.tenant.status} and name:${tenant.tenant.name} and id:${tenant.tenant.id}`
        );
        logger.debug('Before all Phoenix hook successful ended');
        await new Promise((resolve) => setTimeout(resolve, 60000));
    });

    test('should soft delete tenant', async function () {
        const softDeleted = await tenantUtils.deleteTenant(
            token.accessToken,
            tenantUrl,
            tenant.tenant.id,
            false
        );
        expect(softDeleted.uid, 'soft-deleted tenant uid should match the created tenant uuid').to.equal(tenant.tenant.uuid);
        const tenantSoftDeleted = await tenantUtils.getTenants(token.accessToken, tenantUrl);
        expect(tenantSoftDeleted.find((t) => t.uuid === tenant.tenant.uuid), 'soft-deleted tenant should not appear in the tenant list').to.be.undefined;
        logger.info(`Tenant soft deleted: ${softDeleted.uid}`);
        const keycloakDataValidation = await Keycloak.getRealmByName(tenant.tenant.name);
        expect(keycloakDataValidation, 'Keycloak realm should still exist after soft delete').to.be.exist;
        logger.info(`Keycloak data validation: ${keycloakDataValidation} exist in keycloak`);
    });

    test('should force delete tenant', async function () {
        const forceDeleted = await tenantUtils.deleteTenant(
            token.accessToken,
            tenantUrl,
            tenant.tenant.id,
            true
        );
        expect(forceDeleted.uid, 'force-deleted tenant uid should match the created tenant uuid').to.equal(tenant.tenant.uuid);
        const tenantSoftDeleted = await tenantUtils.getTenants(token.accessToken, tenantUrl);
        expect(tenantSoftDeleted.find((t) => t.uuid === tenant.tenant.uuid), 'force-deleted tenant should not appear in the tenant list').to.be.undefined;
        logger.info(`Tenant force deleted: ${forceDeleted.uid}`);
        const keycloakDataValidation = await Keycloak.getRealmByName(tenant.tenant.name);
        expect(keycloakDataValidation, 'Keycloak realm should be removed after force delete').to.be.null;
        logger.info(`Keycloak data validation: ${keycloakDataValidation} does not exist in keycloak`);
    });

    test('should verify tenant cleanup', async function () {
        test.setTimeout(20 * 60 * 1000);
        const tenantSpecificUrl = tenant.additionalData['tenant_domain'];
        logger.info(`Authenticating to tenant-specific URL: ${tenantSpecificUrl}`);
        const tokenResponse = await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
            token.accessToken,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            tenant.tenant.name,
            tenantSpecificUrl
        );
        const userToken = tokenResponse.idToken;
        const user = await usersUtils.createUserInSpecificTenant(userToken, tenantSpecificUrl, {
            email: 'test-e2e2@example.com',
            resetPassword: false
        });
        expect(user, 'test user should be created successfully in the new tenant').to.be.exist;
        await new Promise((resolve) => setTimeout(resolve, 10000));
        const tenantData = await tenantUtils.getTenant(userToken, tenantName);
        const cleanTenantUrl = config.tenantUrl.replace(/^https?:\/\//, '').replace(/\/$/, '');

        const controlPlaneUrl = `https://${tenantData.name}.${cleanTenantUrl}`;
        const clusterVersion = await clusterUtils.getLatestClusterVersion(userToken);
        const cluster = await clusterUtils.installCluster(
            userToken,
            `e2e-cluster-${Date.now()}`,
            config.tenantUrl,
            clusterVersion,
            controlPlaneUrl
        );
        await waitUtils.waitForCondition(
            async () => {
                const status = await clusterUtils.getClusterStatus(userToken, cluster.uuid);
                logger.info(`Cluster ${cluster.name} status: ${status}`);
                return status === ClusterState.Connected;
            },
            `Wait for cluster ${cluster.name} to be Connected`,
            true,
            300
        );

        logger.info('Creating project and job for tenant cleanup validation');
        const projectName = objectGenerator.generateNameByPrefix('test-project');
        const namespace = Utils.namespaceFromProjectName(projectName);
        flowRootState.cluster = {
            uuid: cluster.uuid,
            version: clusterVersion,
            tenantId: tenant.tenant.id
        };
        const departments = await orgUnitUtils.getDepartments(userToken);
        const defaultDepartment = departments.find((dep) => dep.name === 'default');
        if (!defaultDepartment) {
            throw new Error('Default department not found');
        }
        clusterApi.nodepoolsApi.accessToken = userToken;
        const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromCluster();
        logger.info(`Default node pool found: ${defaultNodePool.name} (ID: ${defaultNodePool.id})`);
        let project = await orgUnitUtils.createProject(
            {
                name: projectName,
                clusterId: cluster.uuid,
                parentId: defaultDepartment.id,
                resources: [
                    {
                        gpu: {
                            deserved: 1,
                            limit: 1
                        },
                        memory: null,
                        cpu: null,
                        nodePool: { id: defaultNodePool.id, name: defaultNodePool.name }
                    }
                ]
            },
            userToken
        );

        await waitUtils.waitForCondition(
            async () => {
                try {
                    const proj = await orgUnitUtils.getProject(project.id, userToken);
                    project = proj;
                    return proj.status.phase === ProjectPhase.Ready;
                } catch (e) {
                    logger.error(
                        `Failed getting project ${project.name}/${project.id} from backend; reason: ${JSON.stringify(
                            e
                        )}`
                    );
                    return false;
                }
            },
            `waitForProjectFromBackendToBeReady v2 project name: ${project.name}`,
            true,
            120
        );

        expect(project, 'test project should be created successfully in the new tenant').to.be.exist;
        logger.info(`Project created: ${project.name} (ID: ${project.id})`);

        const jobName = objectGenerator.generateNameByPrefix('test-job');
        let job: BaseWorkload;
        await waitUtils.waitForCondition(
            async () => {
                try {
                    job = await workloadUtils.createWorkloadAndWaitUntilCreated(
                        WorkloadType.Train,
                        namespace,
                        jobName,
                        config.quickStartImage.training
                    );
                    return true;
                } catch (e) {
                    const msg = e instanceof Error ? e.message : JSON.stringify(e);
                    if (msg.includes('migration in progress') || msg.includes('connection refused')) {
                        logger.warn('Transient error, retrying workload creation...');
                        return false;
                    }
                    throw e;
                }
            },
            'Wait for migration to complete before creating training workload',
            true,
            120
        );
        expect(job, 'training job should be created successfully in the new tenant').to.be.exist;
        logger.info(`Job created: ${job.name} in namespace: ${namespace}`);
        const interactiveJob = await workloadUtils.createWorkloadAndWaitUntilCreated(
            WorkloadType.Interactive,
            namespace,
            objectGenerator.generateNameByPrefix('test-interactive-job'),
            config.quickStartImage.interactive
        );
        expect(interactiveJob, 'interactive job should be created successfully in the new tenant').to.be.exist;
        logger.info(`Interactive job created: ${interactiveJob.name} in namespace: ${namespace}`);
        await new Promise((resolve) => setTimeout(resolve, 120000));
        const forceDeleted = await tenantUtils.deleteTenant(
            token.accessToken,
            tenantUrl,
            tenant.tenant.id,
            true
        );
        expect(forceDeleted.uid, 'force-deleted tenant uid should match the created tenant uuid for cleanup validation').to.equal(tenant.tenant.uuid);
        const portForwardProcess = kubeCli.portForwardPostgres();
        await new Promise((resolve) => setTimeout(resolve, 15000));

        let dbConnection: DatabaseConnection;
        try {
            await db.connect();
            dbConnection = db;

            const dbDataValidation = await db.checkTenantExists(tenant.tenant.id);
            expect(dbDataValidation, 'tenant record should not exist in the database after force delete').to.be.false;
            logger.info(`DB data validation: ${dbDataValidation} does not exist in db`);

            const tenantDataCount = await db.getTenantDataCountAcrossAllTables(tenant.tenant.id);
            logger.info(`=== COMPREHENSIVE TENANT ${tenant.tenant.id} CLEANUP ANALYSIS ===`);
            logger.info(`Total records across all tables: ${tenantDataCount.totalCount}`);
            logger.info(`Tables with remaining data: ${tenantDataCount.tableBreakdown.length}`);

            if (tenantDataCount.tableBreakdown.length > 0) {
                logger.warn('=== TABLES WITH REMAINING DATA ===');
                tenantDataCount.tableBreakdown.forEach((table) => {
                    logger.warn(`${table.fullTableName}: ${table.count} records`);
                });
            }

            expect(tenantDataCount.totalCount).to.equal(
                0,
                `Tenant cleanup incomplete. Found ${tenantDataCount.totalCount} records in ${tenantDataCount.tableBreakdown.length} tables: ${tenantDataCount.tableBreakdown.map((t) => `${t.fullTableName}: ${t.count} records`).join(', ')}`
            );

            logger.info(`Tenant cleanup: ${tenantDataCount.totalCount === 0}`);
        } catch (error) {
            logger.error('Database operation failed:', error);
            throw error;
        } finally {
            if (dbConnection) {
                try {
                    await dbConnection.disconnect();
                    logger.info('DB connection closed');
                } catch (disconnectError) {
                    logger.warn('Error disconnecting from database:', disconnectError);
                }
            }

            await new Promise((resolve) => setTimeout(resolve, 1000));

            if (portForwardProcess && !portForwardProcess.killed) {
                logger.info('Killing port forward process');
                portForwardProcess.kill('SIGTERM');

                await new Promise((resolve) => setTimeout(resolve, 500));

                if (!portForwardProcess.killed) {
                    portForwardProcess.kill('SIGKILL');
                }
            }
        }
    });
});
