/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import axios from 'axios';
import { logger } from '../../logger/logger';
import { UnleashConfig, Constrain, Strategy, FeatureToggle } from './models';
import { config, flowRootState, settings } from '../../config/config';

const tenantIdContext = 'tenantId';

export class UnleashClient {
    private client: UnleashConfig;

    constructor() {
        this.client = {
            url: config.unleash.url,
            apiToken: config.unleash.apiToken,
            env: 'development',
            projectName: 'runai',
            strategyTitle: 'e2e-tenants'
        };
    }

    async isFeatureEnabled(featureName: string, tenantId?: number): Promise<boolean> {
        if (!tenantId) {
            tenantId = flowRootState.currentTenant.id;
        }
        const tenantIdStr = String(tenantId);
        logger.info(`Fetching feature toggle by name: '${featureName}'`);
        if (settings.isSelfHosted) {
            return true;
        }

        try {
            const feature: FeatureToggle = await this.getFeatureFlagByName(featureName);
            if (!feature.enabled) return false;
            const strategies: Strategy[] = await this.getFeatureToggleStrategies(featureName);
            const tenantStrategy: Strategy | undefined = strategies.find(
                (s) => s.title === this.client.strategyTitle
            );
            if (!tenantStrategy) {
                logger.debug(
                    `No strategy found: '${this.client.strategyTitle}' for feature flag: '${featureName}'`
                );
                return false;
            }
            const tenantIdConstrain: Constrain | undefined = tenantStrategy.constraints.find(
                (c) => c.contextName === tenantIdContext
            );

            if (!tenantIdConstrain) {
                logger.debug(`TenantId context not found for strategy: '${this.client.strategyTitle}'`);
                return false;
            }

            return tenantIdConstrain.values.includes(tenantIdStr);
        } catch (error) {
            throw new Error(
                `Error validating feature toggle '${featureName}' for tenant: '${tenantIdStr}'`,
                error
            );
        }
    }

    async enableFeature(featureName: string, tenantId?: number): Promise<void | Error> {
        if (!tenantId) {
            tenantId = flowRootState.currentTenant.id;
        }
        const tenantIdStr = String(tenantId);
        logger.info(`Enabling feature toggle: '${featureName}' for tenant '${tenantIdStr}'`);
        const strategies: Strategy[] = await this.getFeatureToggleStrategies(featureName);
        const tenantStrategy: Strategy = strategies.find((s) => s.title === this.client.strategyTitle);
        try {
            if (tenantStrategy) {
                for (const c of tenantStrategy.constraints) {
                    if (c.contextName === tenantIdContext) {
                        const isTenantIdInValues = c.values.includes(tenantIdStr);
                        if (!isTenantIdInValues) {
                            c.values.push(tenantIdStr);
                            await this.updateStrategyConstrains(
                                featureName,
                                tenantStrategy.id,
                                tenantStrategy.constraints
                            );
                        }
                    } else {
                        logger.warn(`No tenantId context for strategy: '${this.client.strategyTitle}'`);
                    }
                }
            } else {
                await this.createNewStrategy(featureName, tenantIdStr);
                await this.enableFeatureToggleEnvironment(featureName);
            }
        } catch (error) {
            return new Error(
                `Error enabling feature toggle '${featureName}' for tenant: '${tenantIdStr}'`,
                error
            );
        }
    }

    async disableFeature(featureName: string, tenantId?: number): Promise<void | Error> {
        if (!tenantId) {
            tenantId = flowRootState.currentTenant.id;
        }
        const tenantIdStr = String(tenantId);
        logger.info(`Disabling feature toggle: '${featureName}' for tenant '${tenantIdStr}'`);
        const strategies: Strategy[] = await this.getFeatureToggleStrategies(featureName);
        const tenantStrategy: Strategy = strategies.find((s) => s.title === this.client.strategyTitle);
        if (tenantStrategy) {
            for (const c of tenantStrategy.constraints) {
                if (c.contextName === tenantIdContext) {
                    const isTenantIdInValues = c.values.includes(tenantIdStr);
                    if (isTenantIdInValues) {
                        const newValues = c.values.filter((element) => element !== tenantIdStr);
                        c.values = newValues;
                        try {
                            if (newValues.length === 0) {
                                await this.deleteStrategy(featureName, tenantStrategy.id);
                            } else {
                                await this.updateStrategyConstrains(
                                    featureName,
                                    tenantStrategy.id,
                                    tenantStrategy.constraints
                                );
                            }
                        } catch (error) {
                            return new Error(
                                `Error deleting strategy: '${this.client.strategyTitle}' for feature toggle: '${featureName}'`,
                                error
                            );
                        }
                    } else {
                        logger.warn(
                            `Tenant id: '${tenantIdStr}' is missing from: '${this.client.strategyTitle}'`
                        );
                    }
                }
            }
        } else {
            logger.warn(`No strategy: '${this.client.strategyTitle}' for feature flag: '${featureName}'`);
        }
    }

    private async getFeatureFlagByName(featureName: string): Promise<FeatureToggle> {
        logger.info(`Fetching feature toggle by name: '${featureName}'`);
        try {
            const response = await axios.get(
                `${this.client.url}/api/admin/projects/${this.client.projectName}/features/${featureName}/environments/${this.client.env}`,
                {
                    headers: {
                        Accept: 'application/json',
                        Authorization: this.client.apiToken
                    }
                }
            );
            return response.data;
        } catch (error) {
            const errorMessage = `Error fetching feature toggle by name: '${featureName}'`;
            if (axios.isAxiosError(error)) {
                logger.error(
                    `${errorMessage}, status: ${error.response?.status}, data: ${error.response?.data}`
                );
            } else {
                logger.error(`${errorMessage}`, error);
            }
            throw error;
        }
    }

    private async enableFeatureToggleEnvironment(featureName: string): Promise<unknown> {
        logger.info(`Enabling environment for feature toggle: '${featureName}'`);

        const config = {
            method: 'post',
            url: `${this.client.url}/api/admin/projects/${this.client.projectName}/features/${featureName}/environments/${this.client.env}/on`,
            headers: {
                Accept: 'application/json',
                Authorization: this.client.apiToken
            },
            data: {}
        };
        try {
            const res = await axios.request(config);
            return res.data;
        } catch (error) {
            const errorMessage = `Error enabling feature toggle '${featureName}' enviroment`;
            if (axios.isAxiosError(error)) {
                logger.error(
                    `${errorMessage}, status: ${error.response?.status}, data: ${error.response?.data}`
                );
            } else {
                logger.error(`${errorMessage}`, error);
            }
            throw error;
        }
    }

    private async getFeatureToggleStrategies(featureName: string): Promise<Strategy[]> {
        logger.info(`Getting strategies for feature toggle: '${featureName}'`);
        try {
            const response = await axios.get(
                `${this.client.url}/api/admin/projects/${this.client.projectName}/features/${featureName}/environments/${this.client.env}/strategies`,
                {
                    headers: {
                        Accept: 'application/json',
                        Authorization: this.client.apiToken
                    }
                }
            );
            return response.data;
        } catch (error) {
            const errorMessage = `Error fetching strategies for feature toggle: '${featureName}'`;
            if (axios.isAxiosError(error)) {
                logger.error(
                    `${errorMessage}, status: ${error.response?.status}, data: ${error.response?.data}`
                );
            } else {
                logger.error(`${errorMessage}`, error);
            }
            throw error;
        }
    }

    private async createNewStrategy(featureName: string, tenantId: string): Promise<unknown> {
        logger.info(
            `Creating new '${this.client.strategyTitle}' strategy for feature toggle: '${featureName}'`
        );
        const data = {
            name: 'default',
            sortOrder: 0,
            title: this.client.strategyTitle,
            disabled: false,
            parameters: {},
            constraints: [
                {
                    contextName: tenantIdContext,
                    operator: 'IN',
                    caseInsensitive: false,
                    inverted: false,
                    values: [tenantId]
                }
            ]
        };

        const config = {
            method: 'post',
            url: `${this.client.url}/api/admin/projects/${this.client.projectName}/features/${featureName}/environments/${this.client.env}/strategies`,
            headers: {
                'Content-Type': 'application/json',
                Accept: 'application/json',
                Authorization: this.client.apiToken
            },
            data: data
        };
        try {
            const res = await axios.request(config);
            return res.data;
        } catch (error) {
            const errorMessage = `Error creating new '${this.client.strategyTitle}' strategy for feature toggle '${featureName}'`;
            if (axios.isAxiosError(error)) {
                logger.error(
                    `${errorMessage}, status: ${error.response?.status}, data: ${error.response?.data}`
                );
            } else {
                logger.error(`${errorMessage}`, error);
            }
            throw error;
        }
    }

    private async updateStrategyConstrains(
        featureName: string,
        strategyId: string,
        newConstrains: Constrain[]
    ): Promise<unknown> {
        logger.info(`Update strategy: '${this.client.strategyTitle}' for feature toggle: '${featureName}'`);
        const data = [
            {
                path: '/constraints',
                op: 'replace',
                from: '/constraints',
                value: newConstrains
            }
        ];

        const config = {
            method: 'patch',
            url: `${this.client.url}/api/admin/projects/${this.client.projectName}/features/${featureName}/environments/${this.client.env}/strategies/${strategyId}`,
            headers: {
                'Content-Type': 'application/json',
                Accept: 'application/json',
                Authorization: this.client.apiToken
            },
            data: data
        };
        try {
            const res = await axios.request(config);
            return res.data;
        } catch (error) {
            const errorMessage = `Error updating strategy: '${this.client.strategyTitle}' constrains array for feature toggle '${featureName}'`;
            if (axios.isAxiosError(error)) {
                logger.error(
                    `${errorMessage}, status: ${error.response?.status}, data: ${error.response?.data}`
                );
            } else {
                logger.error(`${errorMessage}`, error);
            }
            throw error;
        }
    }

    private async deleteStrategy(featureName: string, strategyId: string): Promise<void> {
        logger.info(`Deleting strategy: '${this.client.strategyTitle}' for feature toggle: '${featureName}'`);

        const config = {
            method: 'delete',
            url: `${this.client.url}/api/admin/projects/${this.client.projectName}/features/${featureName}/environments/${this.client.env}/strategies/${strategyId}`,
            headers: {
                Authorization: this.client.apiToken
            },
            data: {}
        };
        try {
            await axios.request(config);
        } catch (error) {
            const errorMessage = `Error deleting strategy: '${this.client.strategyTitle}' for feature toggle '${featureName}'`;
            if (axios.isAxiosError(error)) {
                logger.error(
                    `${errorMessage}, status: ${error.response?.status}, data: ${error.response?.data}`
                );
            } else {
                logger.error(`${errorMessage}`, error);
            }
            throw error;
        }
    }

    public setToken(token: string): void {
        this.client.apiToken = token;
    }
}
