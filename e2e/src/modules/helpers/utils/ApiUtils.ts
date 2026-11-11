/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import axios from 'axios';
import { logger, flowRootState, UtilsError, Utils, versionUtils, apiUtils, config } from '..';
import { backendClusterApi, tempBackendSettingsApi, clusterApi, orgUnitService } from '../../api';
import { Cluster } from '../../api/apis/backendClusterApi';
import { Setting } from '../../api/apis/tempBackendSettingsApi';
import { HttpBearerAuth } from '../../api/apis/tempBackendSettingsApi/generated/model/models';
import {
    NodePoolSettingsKeyEnum,
    HttpBearerAuth as ClusterHttpBearerAuth
} from '../../api/apis/clusterApi';
import {
    OrgUnitSettingsKeyEnum,
    HttpBearerAuth as OrgUnitHttpBearerAuth
} from '../../api/apis/orgUnitApi';

import { EFeatureFlagsKeys, EFeatureFlagsNames } from '../models/featureFlags';
import { SupportedVersions } from './SupporetedEnvsUtils';
import { isEmpty } from 'lodash-es';

interface WorkloadsBooleanSetting {
    value?: {
        enabled?: boolean;
    };
}

interface WorkloadsSettingsResponse {
    enableTemplatesV2?: WorkloadsBooleanSetting;
}

type WorkloadsSettingsKey =
    | 'enablePolicies'
    | 'workloadSubmissionV2Enabled'
    | 'enableTemplatesV2'
    | 'enableHybridSubmission'
    | 'gpuOverProvisioning';

type DatavolumesSettingsKey = 'enableDataVolumes';

type CatalogSettingsKey = 'enableNgcIntegration';

interface UpdateWorkloadsSettingRequest {
    enabled: boolean;
}

export class ApiUtils {
    private readonly serviceSettingsRequestTimeoutMs = 15000;

    private async getWorkloadsSettings(): Promise<WorkloadsSettingsResponse> {
        const url = `${config.tenantUrl}/api/v1/workloads/settings`;
        const accessToken = flowRootState.users.system_administrator.accessToken;
        const maxRetries = 3;
        let lastError: unknown;

        for (let attempt = 1; attempt <= maxRetries; attempt++) {
            try {
                const response = await axios.get<WorkloadsSettingsResponse>(url, {
                    timeout: this.serviceSettingsRequestTimeoutMs,
                    headers: {
                        Authorization: `Bearer ${accessToken}`,
                        Accept: 'application/json'
                    }
                });
                return response.data;
            } catch (error) {
                lastError = error;
                logger.error(
                    `failed getting workloads settings from workloads service, attempt ${attempt}/${maxRetries}, error: ${error}`
                );
                if (attempt < maxRetries) {
                    await Utils.delay(1000);
                }
            }
        }

        throw new UtilsError(
            new Error(`Failed getting workloads settings from workloads service, error: ${lastError}`)
        );
    }

    private async getEnableTemplatesV2FromWorkloadsService(): Promise<boolean> {
        const settings = await this.getWorkloadsSettings();
        const enabled = settings.enableTemplatesV2?.value?.enabled;

        if (enabled === undefined) {
            throw new UtilsError(
                new Error('Failed getting flexible workload templates feature from workloads service settings')
            );
        }

        return enabled;
    }

    private async updateBooleanServiceSetting(
        servicePath: string,
        key: WorkloadsSettingsKey | DatavolumesSettingsKey | CatalogSettingsKey,
        request: UpdateWorkloadsSettingRequest
    ): Promise<void> {
        const url = `${config.tenantUrl}${servicePath}/${key}`;
        const accessToken = flowRootState.users.system_administrator.accessToken;

        try {
            await axios.put<UpdateWorkloadsSettingRequest>(url, request, {
                timeout: this.serviceSettingsRequestTimeoutMs,
                headers: {
                    Authorization: `Bearer ${accessToken}`,
                    Accept: 'application/json'
                }
            });
            logger.info(`${servicePath} setting ${key} set to ${request.enabled} successfully`);
        } catch (error) {
            logger.error(
                `${servicePath} setting ${key} failed to be set to ${request.enabled}, error: ${error}, ` +
                    `response body: ${JSON.stringify(error?.response?.data)}`
            );
            throw new UtilsError(
                new Error(`Failed to set ${servicePath} setting ${key} to ${request.enabled}, error: ${error}`)
            );
        }
    }

    private async updateWorkloadsSetting(
        key: WorkloadsSettingsKey,
        request: UpdateWorkloadsSettingRequest
    ): Promise<void> {
        await this.updateBooleanServiceSetting('/api/v1/workloads/settings', key, request);
    }

    private async updateDatavolumesSetting(
        key: DatavolumesSettingsKey,
        request: UpdateWorkloadsSettingRequest
    ): Promise<void> {
        await this.updateBooleanServiceSetting('/api/v1/datavolumes/settings', key, request);
    }

    private async updateCatalogSetting(
        key: CatalogSettingsKey,
        request: UpdateWorkloadsSettingRequest
    ): Promise<void> {
        await this.updateBooleanServiceSetting('/api/v1/catalog/settings', key, request);
    }

    private async setHybridSubmissionFeature(enabled: boolean): Promise<void> {
        await this.updateWorkloadsSetting('enableHybridSubmission', { enabled });
    }

    private async setTemplatesV2Feature(enabled: boolean): Promise<void> {
        await this.updateWorkloadsSetting('enableTemplatesV2', { enabled });
    }

    private async setPoliciesFeature(enabled: boolean): Promise<void> {
        await this.updateWorkloadsSetting('enablePolicies', { enabled });
    }

    private async setWorkloadV2SubmissionFeature(enabled: boolean): Promise<void> {
        await this.updateWorkloadsSetting('workloadSubmissionV2Enabled', { enabled });
    }

    private async setResourceOptimizationFeature(enabled: boolean): Promise<void> {
        await this.updateWorkloadsSetting('gpuOverProvisioning', { enabled });
    }

    private async setDataVolumesFeature(enabled: boolean): Promise<void> {
        await this.updateDatavolumesSetting('enableDataVolumes', { enabled });
    }

    private async setNgcIntegrationFeature(enabled: boolean): Promise<void> {
        await this.updateCatalogSetting('enableNgcIntegration', { enabled });
    }

    extractTokenFromHeaders(headers: object): string {
        logger.debug(`extracting token from header ${JSON.stringify(headers)}`);
        try {
            const authHeader = headers['authorization'];
            return authHeader.replace('Bearer ', '');
        } catch (error) {
            logger.error(`unable to get token from header, error: ${error}`);
            throw new UtilsError(error);
        }
    }

    async getClusterByPrefix(prefix: string, accessToken: string): Promise<Cluster> {
        logger.debug(`Get cluster by prefix ${prefix}`);
        backendClusterApi.clusterApi.accessToken = accessToken;
        try {
            const clusters = (await backendClusterApi.clusterApi.v1K8sClustersGet()).body;
            logger.debug(`Found clusters: ${JSON.stringify(clusters)}`);
            const filteredClusters = clusters.filter((cls) => !isEmpty(cls.version));
            const cluster = filteredClusters.find((cls) => Utils.strStartsWith(cls.name, prefix));
            if (cluster) return cluster;
            throw new Error(`Cluster with prefix ${prefix} not found or version not set`);
        } catch (error) {
            logger.error(`Cannot get cluster by prefix, error: ${error}`);
            throw new UtilsError(error);
        }
    }

    async getTenantSettings(): Promise<Array<Setting>> {
        try {
            const bearerAuth = new HttpBearerAuth();
            bearerAuth.accessToken = flowRootState.users.system_administrator.accessToken;
            tempBackendSettingsApi.tenantApi.setDefaultAuthentication(bearerAuth);
            const tenantSettings = (await tempBackendSettingsApi.tenantApi.getTenantSettings()).body;
            logger.debug('Get tenant settings from backend - success');
            return tenantSettings;
        } catch (error) {
            logger.error(`failed getting tenant settings from backend, error: ${error}`);
            throw new UtilsError(new Error(`Failed getting tenant settings from backend, error: ${error}`));
        }
    }

    async toggleFeature(featureKey: string, featureName: string, enable: boolean, accessToken?: string): Promise<void> {
        return this.setFeatureValue(featureKey, featureName, enable, accessToken);
    }

    async setFeatureValue(
        featureKey: string,
        featureName: string,
        value: number | string | boolean,
        accessToken?: string
    ): Promise<void> {
        logger.debug(`set ${featureName} feature to ${value} `);
        const booleVal = value as boolean;
        try {
            const bearerAuth = new HttpBearerAuth();
            bearerAuth.accessToken = accessToken ?? flowRootState.users.system_administrator.accessToken;
            tempBackendSettingsApi.tenantApi.setDefaultAuthentication(bearerAuth);
            await tempBackendSettingsApi.tenantApi.updateTenantSetting(undefined, {
                key: featureKey,
                value: booleVal
            });
            logger.info(`${featureName} feature set to ${value} successfully`);
        } catch (error) {
            logger.error(`${featureName} feature failed to be set to ${value} , error: ${error}`);
            throw new UtilsError(new Error(`Failed to set ${featureName} feature, error: ${error}`));
        }
    }

    async getFeatureValue(featureKey: string, featureName: string): Promise<string | number | boolean> {
        logger.debug(`get ${featureName} feature, key: ${featureKey} `);
        try {
            const settings = await this.getTenantSettings();
            const ts = settings.find((setting) => setting.key === featureKey);
            if (ts === undefined) {
                logger.debug(
                    `Feature ${featureName} (key: ${featureKey}) not found in settings, returning false`
                );
                return false;
            }
            return ts.value;
        } catch (error) {
            logger.error(`failed to get ${featureName} feature , error: ${error}`);
            throw new UtilsError(new Error(`Failed to set ${featureName} feature, error: ${error}`));
        }
    }


    async enableHybridSubmission(): Promise<void> {
        await this.setHybridSubmissionFeature(true);
    }

    /**
     * @deprecated **Do NOT disable this feature** in the parallel suite as it may cause test instability
     * and unexpected behavior in parallel test environments.
     */
    async disableHybridSubmission(): Promise<void> {
        await this.setHybridSubmissionFeature(false);
    }

    async enableTemplatesV2(): Promise<void> {
        await this.setFlexTemplateFeature(true);
    }

    // setNodePoolSetting toggles a cluster-service node-pool setting via its public API
    // (PUT /api/v1/node-pools/settings/{key}), the owner of these settings after the migration.
    private async setNodePoolSetting(key: NodePoolSettingsKeyEnum, enabled: boolean): Promise<void> {
        logger.debug(`set node-pool setting ${key} to ${enabled} `);
        try {
            const bearerAuth = new ClusterHttpBearerAuth();
            bearerAuth.accessToken = flowRootState.users.system_administrator.accessToken;
            clusterApi.settingsApi.setDefaultAuthentication(bearerAuth);
            await clusterApi.settingsApi.updateNodePoolSettingByKey(key, { enabled });
            logger.info(`node-pool setting ${key} set to ${enabled} successfully`);
        } catch (error) {
            logger.error(`node-pool setting ${key} failed to be set to ${enabled} , error: ${error}`);
            throw new UtilsError(new Error(`Failed to set node-pool setting ${key}, error: ${error}`));
        }
    }

    // setOrgUnitSetting toggles an org-unit-service setting via its public API
    // (PUT /api/v1/org-unit/settings/{key}), the owner of these settings after the migration.
    private async setOrgUnitSetting(
        key: OrgUnitSettingsKeyEnum,
        enabled: boolean,
        accessToken?: string
    ): Promise<void> {
        logger.debug(`set org-unit setting ${key} to ${enabled} `);
        try {
            const bearerAuth = new OrgUnitHttpBearerAuth();
            bearerAuth.accessToken = accessToken ?? flowRootState.users.system_administrator.accessToken;
            orgUnitService.settingsApi.setDefaultAuthentication(bearerAuth);
            await orgUnitService.settingsApi.updateOrgUnitSettingByKey(key, { enabled });
            logger.info(`org-unit setting ${key} set to ${enabled} successfully`);
        } catch (error) {
            logger.error(`org-unit setting ${key} failed to be set to ${enabled} , error: ${error}`);
            throw new UtilsError(new Error(`Failed to set org-unit setting ${key}, error: ${error}`));
        }
    }

    // org-unit-service only owns this setting from 3.0; older control planes still serve it as a
    // tenant feature flag and return 404 on the org-unit route.
    private async setLimitOverQuotaSubscription(enabled: boolean, accessToken?: string): Promise<void> {
        if (versionUtils.isVersionSupported(flowRootState.controlPlaneVersion, SupportedVersions.V3_0)) {
            await this.setOrgUnitSetting(
                OrgUnitSettingsKeyEnum.LimitQuotaOverSubscription,
                enabled,
                accessToken
            );
            return;
        }
        await this.toggleFeature(
            EFeatureFlagsKeys.ENABLE_LIMIT_OQ_SUBSCRIPTION,
            EFeatureFlagsNames.ENABLE_LIMIT_OQ_SUBSCRIPTION,
            enabled,
            accessToken
        );
    }

    async enableLimitOverQuotaSubscription(accessToken?: string): Promise<void> {
        await this.setLimitOverQuotaSubscription(true, accessToken);
    }

    async disableLimitOverQuotaSubscription(accessToken?: string): Promise<void> {
        await this.setLimitOverQuotaSubscription(false, accessToken);
    }

    async enableCpuQuotaFeature(): Promise<void> {
        await this.setNodePoolSetting(NodePoolSettingsKeyEnum.CpuResourcesQuota, true);
    }

    async disableCpuQuotaFeature(): Promise<void> {
        await this.setNodePoolSetting(NodePoolSettingsKeyEnum.CpuResourcesQuota, false);
    }

    async enableProjectOverQuotaFeature(): Promise<void> {
        await this.setNodePoolSetting(NodePoolSettingsKeyEnum.ProjectManualOverQuota, true);
    }

    async disableProjectOverQuotaFeature(): Promise<void> {
        await this.setNodePoolSetting(NodePoolSettingsKeyEnum.ProjectManualOverQuota, false);
    }

    async enableNewPolicyFeature(): Promise<void> {
        await this.setPoliciesFeature(true);
    }

    async disableNewPolicyFeature(): Promise<void> {
        await this.setPoliciesFeature(false);
    }

    async resetLogoutUri(): Promise<void> {
        await this.setFeatureValue(EFeatureFlagsKeys.LOGOUT_URI, EFeatureFlagsNames.LOGOUT_URI, '');
    }

    async setLogoutUri(url: string): Promise<void> {
        await this.setFeatureValue(EFeatureFlagsKeys.LOGOUT_URI, EFeatureFlagsNames.LOGOUT_URI, url);
    }

    async getLogoutUri(): Promise<string> {
        return (await this.getFeatureValue(
            EFeatureFlagsKeys.LOGOUT_URI,
            EFeatureFlagsNames.LOGOUT_URI
        )) as string;
    }

    async setInactivityLogoutTimeout(millisecond: number): Promise<void> {
        await this.setFeatureValue(
            EFeatureFlagsKeys.INACTIVITY_LOGOUT_TIMEOUT,
            EFeatureFlagsNames.INACTIVITY_LOGOUT_TIMEOUT,
            millisecond
        );
    }

    async getLogoutTimeout(): Promise<number> {
        return (await this.getFeatureValue(
            EFeatureFlagsKeys.INACTIVITY_LOGOUT_TIMEOUT,
            EFeatureFlagsNames.INACTIVITY_LOGOUT_TIMEOUT
        )) as number;
    }

    async setUIWorkloadV2SubmissionFeature(value: boolean): Promise<void> {
        await this.setWorkloadV2SubmissionFeature(value);
    }

    async enableDataVolumes(): Promise<void> {
        await this.setDataVolumesFeature(true);
    }

    async disableDataVolumes(): Promise<void> {
        await this.setDataVolumesFeature(false);
    }

    async enableUiCliV2(): Promise<void> {
        await this.toggleFeature(EFeatureFlagsKeys.CLI_V2, EFeatureFlagsNames.CLI_V2, true);
    }

    async disableUiCliV2(): Promise<void> {
        await this.toggleFeature(EFeatureFlagsKeys.CLI_V2, EFeatureFlagsNames.CLI_V2, false);
    }

    async enableModelCatalog(): Promise<void> {
        await this.toggleFeature(EFeatureFlagsKeys.MODEL_CATALOG, EFeatureFlagsNames.MODEL_CATALOG, true);
    }

    async enableNIM(): Promise<void> {
        await this.toggleFeature(EFeatureFlagsKeys.NIM, EFeatureFlagsNames.NIM, true);
    }

    async disableModelCatalog(): Promise<void> {
        await this.toggleFeature(EFeatureFlagsKeys.MODEL_CATALOG, EFeatureFlagsNames.MODEL_CATALOG, false);
    }

    async disableNIM(): Promise<void> {
        await this.toggleFeature(EFeatureFlagsKeys.NIM, EFeatureFlagsNames.NIM, false);
    }

    async enableGPT2Model(): Promise<void> {
        await this.toggleFeature(
            EFeatureFlagsKeys.ALLOW_GPT2_MODEL,
            EFeatureFlagsNames.ALLOW_GPT2_MODEL,
            true
        );
    }

    async disableGPT2Model(): Promise<void> {
        await this.toggleFeature(
            EFeatureFlagsKeys.ALLOW_GPT2_MODEL,
            EFeatureFlagsNames.ALLOW_GPT2_MODEL,
            false
        );
    }

    async enableNodePool(): Promise<void> {
        await this.toggleFeature(
            EFeatureFlagsKeys.ENABLE_NODE_POOLS,
            EFeatureFlagsNames.ENABLE_NODE_POOLS,
            true
        );
    }

    async disableNodePool(): Promise<void> {
        await this.toggleFeature(
            EFeatureFlagsKeys.ENABLE_NODE_POOLS,
            EFeatureFlagsNames.ENABLE_NODE_POOLS,
            false
        );
    }

    async enableNgcCatalog(): Promise<void> {
        await this.setNgcIntegrationFeature(true);
    }

    async disableNgcCatalog(): Promise<void> {
        await this.setNgcIntegrationFeature(false);
    }

    async setFeatureFlags(valueToSet: boolean): Promise<void> {
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_16)) {
            await apiUtils.setFeatureValue(
                EFeatureFlagsKeys.MODEL_CATALOG,
                EFeatureFlagsNames.MODEL_CATALOG,
                valueToSet
            );
        }

        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_18)) {
            await this.setPoliciesFeature(valueToSet);
        }
    }

    async setFlexTemplateFeature(value: boolean): Promise<void> {
        await this.setTemplatesV2Feature(value);
    }

    async setResourceOptimization(value: boolean): Promise<void> {
        await this.setResourceOptimizationFeature(value);
    }

    async isTemplateV2Enabled(): Promise<boolean> {
        return await this.getEnableTemplatesV2FromWorkloadsService();
    }
}
