/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../api';
import { EnvironmentAsset, EnvironmentUpdateRequest, WorkloadSupportedTypes } from '../../api/apis/assetsApi';
import { EnvironmentCreationRequest, Scope } from '../../api/apis/assetsApi';
import { flowRootState } from '../config/config';
import { objectGenerator, settings, supportedPlatforms, waitUtils } from '../index';
import { logger } from '..';

// Environment presets have different names for path-based vs host-based routing.
// The list API filters by the cluster's routing method, so only one variant is visible at a time.
// See migration: assets-service/pkg/database/migrations/up/20210008_new_environments_preset.up.sql
const ENV_ROUTING_PAIRS: [string, string][] = [
    ['tensorboard', 'tensorboard-tensorflow'],
    ['jupyter-lab', 'jupyter-scipy'],
    // jupyter-tensorboard has no direct host-based equivalent; jupyter-scipy is the closest match
    ['jupyter-tensorboard', 'jupyter-scipy'],
];

// Template presets similarly have routing-specific names.
// See: control-plane/workloads/pkg/services_crud/template_presets.go
const TEMPLATE_ROUTING_PAIRS: [string, string][] = [
    ['jupyter-lab-one-gpu', 'jupyter-lab-1-gpu'],
];

function getAlternateEnvName(name: string): string | undefined {
    for (const [pathName, hostName] of ENV_ROUTING_PAIRS) {
        if (name === pathName) return hostName;
        if (name === hostName) return pathName;
    }
    return undefined;
}

export interface EnvPresetNames {
    jupyter: string;
    tensorboard: string;
    jupyterTensorboard: string;
}

export interface TemplatePresetNames {
    jupyterLabOneGpu: string;
}

export class EnvironmentUtils {
    private _isHostBased: boolean | null = null;

    private _presets: EnvPresetNames | null = null;

    private _templatePresets: TemplatePresetNames | null = null;

    set accessToken(accessToken: string) {
        assetsApi.environmentApi.accessToken = accessToken;
    }

    async list(): Promise<Array<EnvironmentAsset>> {
        assetsApi.environmentApi.accessToken = flowRootState.users.system_administrator.accessToken;
        return (await assetsApi.environmentApi.listEnvironmentAssets()).body.entries;
    }

    async getIdOfSomeEnv(): Promise<string> {
        return (await this.list())[0].meta.id;
    }

    async getId(environmentName: string): Promise<string> {
        return (await this.get(environmentName)).meta.id;
    }

    async get(environmentName: string): Promise<EnvironmentAsset> {
        const environments = await this.list();
        const env = environments.find((e) => e.meta.name === environmentName);
        if (env) return env;
        const alt = getAlternateEnvName(environmentName);
        const resolved = alt ? environments.find((e) => e.meta.name === alt) : undefined;
        if (resolved) return resolved;
        throw new Error(`Environment "${environmentName}" was not found`);
    }

    async isHostBasedRouting(): Promise<boolean> {
        if (this._isHostBased !== null) return this._isHostBased;
        const environments = await this.list();
        const names = new Set(environments.map((e) => e.meta.name));
        const hasHost = ENV_ROUTING_PAIRS.some(([, hostName]) => names.has(hostName));
        const hasPath = ENV_ROUTING_PAIRS.some(([pathName]) => names.has(pathName));

        if (hasHost) this._isHostBased = true;
        else if (hasPath) this._isHostBased = false;

        return this._isHostBased ?? false;
    }

    async presetNames(): Promise<EnvPresetNames> {
        if (this._presets) return this._presets;
        const isHost = await this.isHostBasedRouting();
        this._presets = {
            jupyter: isHost ? 'jupyter-scipy' : 'jupyter-lab',
            tensorboard: isHost ? 'tensorboard-tensorflow' : 'tensorboard',
            jupyterTensorboard: isHost ? 'jupyter-scipy' : 'jupyter-tensorboard'
        };
        return this._presets;
    }

    async templatePresetNames(): Promise<TemplatePresetNames> {
        if (this._templatePresets) return this._templatePresets;
        const isHost = await this.isHostBasedRouting();
        this._templatePresets = {
            jupyterLabOneGpu: isHost ? 'jupyter-lab-1-gpu' : 'jupyter-lab-one-gpu'
        };
        return this._templatePresets;
    }

    async resolveEnvPresetName(logicalName: string): Promise<string> {
        const hostBased = await this.isHostBasedRouting();
        for (const [pathName, hostName] of ENV_ROUTING_PAIRS) {
            if (logicalName === pathName || logicalName === hostName) {
                return hostBased ? hostName : pathName;
            }
        }
        return logicalName;
    }

    async resolveTemplatePresetName(logicalName: string): Promise<string> {
        const hostBased = await this.isHostBasedRouting();
        for (const [pathName, hostName] of TEMPLATE_ROUTING_PAIRS) {
            if (logicalName === pathName || logicalName === hostName) {
                return hostBased ? hostName : pathName;
            }
        }
        return logicalName;
    }

    async waitBeforeAbleToRemoveEnv(environmentName: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const environments = await this.list();
                return environments.find((env) => env.meta.name === environmentName).usedBy === undefined;
            },
            'waiting for usedBy field to be undefined',
            true
        );
        await waitUtils.sleep(40000); // TODO: remove it after the above func is fixed
    }

    async create(environment: EnvironmentCreationRequest): Promise<EnvironmentAsset> {
        try {
            logger.info('creating environment', environment.meta.name);
            assetsApi.environmentApi.accessToken = flowRootState.users.system_administrator.accessToken;

            const environmentAsset: EnvironmentAsset = (
                await assetsApi.environmentApi.createEnvironmentAsset(environment)
            ).body;
            logger.info(
                `environment ${environment.meta.name} created successfully. id: ${environmentAsset.meta.id}`
            );
            return environmentAsset;
        } catch (e) {
            logger.error(`Failed to create environment ${environment.meta.name}: ${e}`);
            throw e;
        }
    }

    async update(
        id: string,
        request: EnvironmentUpdateRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<EnvironmentAsset> {
        assetsApi.environmentApi.accessToken = accessToken;
        const environmentAsset: EnvironmentAsset = (
            await assetsApi.environmentApi.updateEnvironmentAssetById(id, request)
        ).body;

        return environmentAsset;
    }

    async delete(
        id: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        assetsApi.s3Api.accessToken = accessToken;
        await waitUtils.waitForCondition(
            async () => {
                try {
                    assetsApi.environmentApi.accessToken =
                        flowRootState.users.system_administrator.accessToken;
                    await assetsApi.environmentApi.deleteEnvironmentAssetById(id);
                    logger.info(`Deleted environment with id: ${id} successfully`);
                    return true; // Successfully deleted
                } catch (error: any) {
                    if (error?.body?.code === 412) {
                        // Retry only if the error is 412 (asset in use)
                        logger.warn('Retrying delete environment due to 412 error...');
                        return false; // Continue waiting
                    } else {
                        logger.error(`Failed to delete environment with id ${id}, error: [${error}]`);
                        throw error;
                    }
                }
            },
            `Delete environment with ID ${id}`,
            true,
            40,
            2
        );
    }

    async createEnvironmentWithArbitraryFields(
        projectId: number,
        command = 'sleep infinity',
        distFramework = WorkloadSupportedTypes.DistFrameworkEnum.PyTorch,
        image = 'alpine',
        args?: string
    ): Promise<EnvironmentAsset> {
        const environmentCreationRequest: EnvironmentCreationRequest = {
            meta: {
                name: objectGenerator.generateNameByPrefix('environment'),
                scope: Scope.Project,
                projectId: projectId,
                clusterId: flowRootState.cluster.uuid,
                workloadSupportedTypes: {
                    workspace: true,
                    training: true,
                    distributed: true,
                    distFramework: distFramework
                }
            },
            spec: {
                image: image
            }
        };
        if (command) {
            environmentCreationRequest.spec.command = command;
        }
        if (args) {
            environmentCreationRequest.spec.args = args;
        }
        return await this.create(environmentCreationRequest);
    }

    async waitForEnvironmentToBeDeleted(name: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const environmentNames = (await this.list()).map((environment) => environment.meta.name);
                return !environmentNames.includes(name);
            },
            'waiting for environments to be deleted',
            true,
            30
        );
    }

    async waitForEnvironmentToAppear(name: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const environmentNames = (await this.list()).map((environment) => environment.meta.name);
                return environmentNames.includes(name);
            },
            'waiting for environments to appear',
            true,
            30
        );
    }

    async deleteAllProjectScopedEnvironments(
        accessToken = flowRootState.users.system_administrator.accessToken,
        clusterId = flowRootState.cluster.uuid
    ): Promise<void> {
        try {
            assetsApi.environmentApi.accessToken = accessToken;
            const environments: EnvironmentAsset[] = (
                await assetsApi.environmentApi.listEnvironmentAssets(
                    undefined,
                    undefined,
                    undefined,
                    undefined,
                    clusterId
                )
            ).body.entries;
            const projectScopedEnvironments = environments.filter((env) => env.meta.scope === Scope.Project);
            await Promise.all(
                projectScopedEnvironments.map((env) =>
                    assetsApi.environmentApi.deleteEnvironmentAssetById(env.meta.id)
                )
            );
            if (projectScopedEnvironments.length > 0) {
                logger.info('deleted all project scoped environments successfully');
            }
        } catch (error) {
            logger.warn(`failed to delete all project scoped environments: ${error}`);
        }
    }

    async deleteEnvironments(
        prefix: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.environmentApi.accessToken = accessToken;
            const environments: EnvironmentAsset[] = (await assetsApi.environmentApi.listEnvironmentAssets())
                .body.entries;
            const environmentsToDelete = environments.filter((env) => env.meta.name.startsWith(prefix));
            await Promise.all(
                environmentsToDelete.map((env) =>
                    assetsApi.environmentApi.deleteEnvironmentAssetById(env.meta.id)
                )
            );
            if (environmentsToDelete.length > 0) {
                logger.info(`deleted all environments with prefix:"${prefix}" successfully`);
            }
        } catch (error) {
            logger.warn(`failed to delete environmentswith prefix:"${prefix}". error ${error}`);
        }
    }

    async createAsUser(
        environment: EnvironmentCreationRequest,
        accessToken: string
    ): Promise<EnvironmentAsset> {
        try {
            logger.info('creating environment', environment.meta.name);
            assetsApi.environmentApi.accessToken = accessToken;
            const environmentAsset: EnvironmentAsset = (
                await assetsApi.environmentApi.createEnvironmentAsset(environment)
            ).body;
            logger.info(
                `environment ${environment.meta.name} created successfully. id: ${environmentAsset.meta.id}`
            );
            return environmentAsset;
        } catch (e) {
            logger.error(`Failed to create environment ${environment.meta.name}: ${e}`);
            throw e;
        }
    }

    async deleteAsUser(id: string, accessToken: string): Promise<void> {
        try {
            logger.debug(`deleting environment ${id}`);
            assetsApi.environmentApi.accessToken = accessToken;
            await assetsApi.environmentApi.deleteEnvironmentAssetById(id);
            logger.debug(`environment ${id} deleted successfully`);
        } catch (e) {
            logger.error(`Failed to delete environment ${id}: ${e}`);
            throw e;
        }
    }

    /**
     * Deletes an environment asset by its name.
     *
     * This function retrieves a list of environment assets and searches for an asset
     * matching the provided `environmentName`. If found, it deletes the asset based
     * on its ID. If no asset with the given name is found, an error is thrown.
     *
     * @async
     * @function
     * @param {string} environmentName - The name of the environment asset to delete.
     * @returns {Promise<void>} Resolves when the asset is successfully deleted or throws an error if the asset is not found.
     * @throws {Error} If no environment asset with the specified name exists.
     */

    async deleteByName(environmentName: string): Promise<void> {
        const environmentId: string | undefined = await this.getId(environmentName);
        if (!environmentId) {
            throw new Error(`Environment asset with name ${environmentName} not found`);
        }

        await this.delete(environmentId);
    }
}

export function isOCP(): boolean {
    return settings.platform === supportedPlatforms.ocp;
}
