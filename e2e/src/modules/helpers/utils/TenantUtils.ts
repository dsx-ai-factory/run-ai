/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { identityManagerApi, logoApi, tenantApi, tempBackendSettingsApi } from '../../api';
import { Setting, HttpBearerAuth } from '../../api/apis/tempBackendSettingsApi/generated/model/models';
import { AuthConfiguration, TokenResponse } from '../../api/apis/identityManagerApi/generated/api';
import {
    DeleteTenantById200Response,
    GetTenantLogo200Response,
    LogoCreationRequest,
    Roles,
    Tenant,
    TenantApi,
    TenantCreationRequest,
    TenantCreationRequestMetricStore,
    TenantCreationResponse
} from '../../api/apis/tenantsManagerApi/generated/api';
import { expect } from 'chai';
import { Keycloak } from './Keycloak';
import { authUtils, crds, logger, TokenObject, waitUtils } from '../index';
import { StatusCodes } from 'http-status-codes';

export class TenantUtils {
    async getTenantSettings(accessToken: string): Promise<Setting[]> {
        const bearerAuth = new HttpBearerAuth();
        bearerAuth.accessToken = accessToken;
        tempBackendSettingsApi.tenantApi.setDefaultAuthentication(bearerAuth);
        const tenantConfig = await tempBackendSettingsApi.tenantApi.getTenantSettings();
        return tenantConfig.body;
    }

    async getTenant(accessToken: string, tenantName?: string): Promise<Tenant> {
        tenantApi.accessToken = accessToken;
        const tenants = await tenantApi.getTenants();
        if (tenantName) {
            return tenants.body.find((t) => t.name === tenantName);
        }
        return tenants.body[0];
    }

    async getTenantLogo(accessToken: string): Promise<GetTenantLogo200Response> {
        logoApi.accessToken = accessToken;
        const tenantLogo = await logoApi.getTenantLogo();
        return tenantLogo.body;
    }

    async uploadTenantLogo(accessToken: string, logo: LogoCreationRequest, tenantId?: string): Promise<any> {
        logoApi.accessToken = accessToken;
        const uploaded = await logoApi.uploadTenantLogo(tenantId, logo);
        return uploaded.body;
    }

    async getTenantConfiguration(accessToken: string): Promise<AuthConfiguration> {
        identityManagerApi.configApi.accessToken = accessToken;
        const tenantConfig = await identityManagerApi.configApi.getTenantConfiguration();
        return tenantConfig.body;
    }

    async getTenants(accessToken: string, url: string): Promise<Array<Tenant>> {
        const tenantApi = new TenantApi(url);
        tenantApi.accessToken = accessToken;
        const tenants = await tenantApi.getTenants();
        return tenants.body;
    }

    async createTenant(
        accessToken: string,
        url: string,
        name: string,
        email: string,
        password: string,
        contractType?: TenantCreationRequest.ContractTypeEnum,
        metricStore?: TenantCreationRequestMetricStore | null,
        role?: Roles | null
    ): Promise<TenantCreationResponse> {
        const tenantApi = new TenantApi(url);
        tenantApi.accessToken = accessToken;
        const tenantCreationRequest: TenantCreationRequest = {
            email: email,
            name: name,
            password: password,
            contractType: contractType,
            metricStore: metricStore,
            role: role
        };
        const createdTenant = await tenantApi.createTenant(tenantCreationRequest);
        return createdTenant.body;
    }

    async updateTenantAlias(accessToken: string, url: string, tenantId: number, alias: string | null): Promise<Tenant> {
        const tenantApi = new TenantApi(url);
        tenantApi.accessToken = accessToken;
        const { body } = await tenantApi.updateTenantById(tenantId, { alias });
        return body;
    }

    async getTenantById(accessToken: string, url: string, tenantId: number): Promise<Tenant> {
        const tenantApi = new TenantApi(url);
        tenantApi.accessToken = accessToken;
        const tenant = await tenantApi.getTenantById(tenantId);
        return tenant.body;
    }

    async deleteTenant(
        accessToken: string,
        url: string,
        tenantId: number,
        forceDelete?: boolean
    ): Promise<DeleteTenantById200Response> {
        const tenantApi = new TenantApi(url);
        tenantApi.accessToken = accessToken;
        const deleteTenant = await tenantApi.deleteTenantById(tenantId, forceDelete);
        return deleteTenant.body;
    }

    async deleteTenantAndValidate(
        accessToken: string,
        url: string,
        tenantId: number,
        forceDelete?: boolean,
        realmName?: string
    ): Promise<void> {
        tenantApi.accessToken = accessToken;
        await this.deleteTenant(accessToken, url, tenantId, forceDelete);
        const tenants = await tenantApi.getTenants();
        expect(tenants.body.find((t) => t.id === tenantId)).to.be.undefined;
        if (forceDelete) {
            const keycloakDataValidation = await Keycloak.getRealmByName(realmName);
            expect(keycloakDataValidation).to.be.null;
        }
    }

    async createTenantAndRetrieveToken(
        consoleToken: string,
        consoleTenantUrl: string,
        appUrl: string,
        tenantName: string,
        adminEmail: string,
        adminPassword: string,
        contractType: TenantCreationRequest.ContractTypeEnum = TenantCreationRequest.ContractTypeEnum.Normal
    ): Promise<{
        tenant: TenantCreationResponse;
        adminToken: TokenResponse;
        tenantSpecificUrl: string;
    }> {
        const tenant = await this.createTenant(
            consoleToken,
            consoleTenantUrl,
            tenantName,
            adminEmail,
            adminPassword,
            contractType,
            null,
            null
        );
        logger.info(
            `Tenant created: status:${tenant.tenant.status} and name:${tenant.tenant.name} and id:${tenant.tenant.id}`
        );
        logger.info(`Tenant user: ${adminEmail}`);

        const tenantSpecificUrl = appUrl.replace('https://', `https://${tenant.tenant.name}.`);
        logger.info(`Tenant url: ${tenantSpecificUrl}`);

        const adminToken = await this.waitUntilNewTenantIsReadyAndRetrieveToken(
            consoleToken,
            adminEmail,
            adminPassword,
            tenant.tenant.name,
            tenantSpecificUrl
        );

        return { tenant, adminToken, tenantSpecificUrl };
    }

    async waitUntilNewTenantIsReadyAndRetrieveToken(
        accessToken: string,
        tenantAdminEmail: string,
        adminPassword: string,
        tenantName: string,
        tenantUrl: string
    ): Promise<TokenResponse> {
        let tokenObject: TokenResponse;
        const isNewTenantTokenReady = async (): Promise<boolean> => {
            try {
                logger.info('tenant url: ', tenantUrl);
                const tokenRes = await authUtils.authenticateTokenApi(
                    tenantAdminEmail,
                    adminPassword,
                    tenantUrl,
                    accessToken
                );
                expect(tokenRes.response.statusCode).to.equal(StatusCodes.OK);
                tokenObject = tokenRes.body;
                expect(tokenObject.accessToken).to.not.be.undefined;
                return true;
            } catch (error) {
                logger.info(`Tenant ${tenantName} still creating, retrying...`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isNewTenantTokenReady,
            `wait for new tenant ${tenantName} to be ready and retrieve admin token`,
            true,
            60,
            5
        );
        return tokenObject;
    }
}
