/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { TokenResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    TenantCreationRequest,
    TenantCreationResponse
} from '../../../../../../../modules/api/apis/tenantsManagerApi/generated/api';
import {
    ScopeType,
    TrainingTemplatesListResponseItem,
    TrainingTemplateUpdateRequest,
    WorkloadTemplatesApi
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    authUtils,
    config,
    logger,
    objectGenerator,
    settings,
    tags,
    tenantUtils
} from '../../../../../../../modules/helpers';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

shouldRunTests(settings.isMultiTenancyCluster)(() => {
    test.describe(`System template tenant isolation API, tags: ${tags.tool.api} ${tags.resource.templateV2} ${tags.teams.viper}`, () => {
        interface TestTenant {
            response: TenantCreationResponse;
            url: string;
            token: TokenResponse;
        }

        const systemTemplateName = 'pytorch-1-gpu';
        const tenants: TestTenant[] = [];
        let consoleToken: TokenResponse;
        let consoleTenantUrl: string;

        test.beforeAll(async function () {
            if (settings.isSelfHosted) {
                settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient(
                    'console',
                    'default'
                );
            }

            const baseUrl = config.tenantUrl.replace('https://', '');
            consoleTenantUrl = `https://console.${baseUrl}`;
            consoleToken = await authUtils.authenticateSuperAdminAppAndGetToken(
                'default',
                settings.tenantClientSecret,
                consoleTenantUrl
            );

            for (let index = 0; index < 2; index++) {
                const tenantName = objectGenerator.generateShortNameByPrefix(`template-tenant-${index + 1}`);
                const response = await tenantUtils.createTenant(
                    consoleToken.accessToken,
                    consoleTenantUrl,
                    tenantName,
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword,
                    TenantCreationRequest.ContractTypeEnum.Normal,
                    null,
                    null
                );
                const tenantUrl = response.additionalData['tenant_domain'];
                const token = await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
                    consoleToken.accessToken,
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword,
                    tenantName,
                    tenantUrl
                );

                tenants.push({ response, url: tenantUrl, token });
            }
        });

        test.afterAll(async function () {
            for (const tenant of tenants) {
                try {
                    await tenantUtils.deleteTenant(
                        consoleToken.accessToken,
                        consoleTenantUrl,
                        tenant.response.tenant.id,
                        true
                    );
                } catch (error) {
                    logger.warn(`Failed to clean up tenant ${tenant.response.tenant.name}: ${error.message}`);
                }
            }
        });

        test('update is rejected and deletion is isolated to the requesting tenant', async function () {
            const tenantA = tenants[0];
            const tenantB = tenants[1];
            const tenantATemplatesApi = new WorkloadTemplatesApi(tenantA.url);
            const tenantBTemplatesApi = new WorkloadTemplatesApi(tenantB.url);
            tenantATemplatesApi.accessToken = tenantA.token.accessToken;
            tenantBTemplatesApi.accessToken = tenantB.token.accessToken;

            const findSystemTemplate = async (
                api: WorkloadTemplatesApi
            ): Promise<TrainingTemplatesListResponseItem | undefined> => {
                const response = await api.listTrainingTemplates(
                    undefined,
                    undefined,
                    undefined,
                    undefined,
                    undefined,
                    systemTemplateName
                );
                return response.body.entries?.find(
                    (template) =>
                        template.meta.name === systemTemplateName &&
                        template.meta.scopeType === ScopeType.System
                );
            };

            const tenantATemplate = await findSystemTemplate(tenantATemplatesApi);
            const tenantBTemplate = await findSystemTemplate(tenantBTemplatesApi);

            expect(tenantATemplate, 'system template should be visible in tenant A').to.not.be.undefined;
            expect(tenantBTemplate, 'system template should be visible in tenant B').to.not.be.undefined;
            expect(tenantBTemplate.meta.id).to.equal(tenantATemplate.meta.id);

            const updateRequest: TrainingTemplateUpdateRequest = {
                meta: {
                    name: `${systemTemplateName}-updated`
                }
            };

            try {
                await tenantATemplatesApi.updateTrainingTemplate(tenantATemplate.meta.id, updateRequest);
                expect.fail('updating a system template through the public API should fail');
            } catch (error) {
                expect(error.statusCode).to.equal(StatusCodes.FORBIDDEN);
                expect(error.body.message).to.equal('system templates cannot be updated');
            }

            expect((await findSystemTemplate(tenantATemplatesApi)).meta.name).to.equal(systemTemplateName);
            expect((await findSystemTemplate(tenantBTemplatesApi)).meta.name).to.equal(systemTemplateName);

            const deleteResponse = await tenantATemplatesApi.deleteTrainingTemplate(tenantATemplate.meta.id);
            expect(deleteResponse.response.statusCode).to.equal(StatusCodes.NO_CONTENT);
            expect(await findSystemTemplate(tenantATemplatesApi)).to.be.undefined;

            const tenantBTemplateAfterDelete = await findSystemTemplate(tenantBTemplatesApi);
            expect(tenantBTemplateAfterDelete, 'system template should remain visible in tenant B').to.not.be
                .undefined;
            expect(tenantBTemplateAfterDelete.meta.id).to.equal(tenantBTemplate.meta.id);
            expect(tenantBTemplateAfterDelete.meta.name).to.equal(systemTemplateName);
        });
    });
});
