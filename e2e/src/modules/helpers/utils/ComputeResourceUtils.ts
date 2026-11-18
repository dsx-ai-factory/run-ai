/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../api';
import {
    Scope,
    type ComputeAsset,
    type ComputeCreationRequest,
    GpuRequestType,
    ComputeUpdateRequest
} from '../../api/apis/assetsApi';

import { flowRootState } from '../config/config';

import { logger, objectGenerator, waitUtils } from '..';

export class ComputeResourceUtils {
    set accessToken(accessToken: string) {
        assetsApi.computeApi.accessToken = accessToken;
    }

    async list(): Promise<Array<ComputeAsset>> {
        assetsApi.computeApi.accessToken = flowRootState.users.system_administrator.accessToken;
        return (await assetsApi.computeApi.listComputeAssets(flowRootState.cluster.uuid)).body.entries;
    }

    async getId(computeResourceName: string): Promise<string> {
        assetsApi.computeApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const computeResources = (await assetsApi.computeApi.listComputeAssets()).body.entries;
        return computeResources.find((compute) => compute.meta.name === computeResourceName).meta.id;
    }

    async create(computeResource: ComputeCreationRequest): Promise<ComputeAsset> {
        try {
            logger.info('creating compute resource', computeResource.meta.name);
            assetsApi.computeApi.accessToken = flowRootState.users.system_administrator.accessToken;

            const computeAsset: ComputeAsset = (
                await assetsApi.computeApi.createComputeAsset(computeResource)
            ).body;
            logger.info(
                `compute ${computeResource.meta.name} created successfully. id: ${computeAsset.meta.id}`
            );
            return computeAsset;
        } catch (e) {
            logger.error(`Failed to create compute ${computeResource.meta.name}: ${e}`);
            throw e;
        }
    }

    async update(
        id: string,
        request: ComputeUpdateRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<ComputeAsset> {
        assetsApi.environmentApi.accessToken = accessToken;

        const computeAsset: ComputeAsset = (await assetsApi.computeApi.updateComputeAssetById(id, request))
            .body;

        return computeAsset;
    }

    async createComputeResourceWithArbitraryFields(): Promise<ComputeAsset> {
        const computeCreationRequest: ComputeCreationRequest = {
            meta: {
                name: objectGenerator.generateNameByPrefix('compute'),
                scope: Scope.Tenant
            },
            spec: {
                cpuCoreRequest: 0,
                gpuDevicesRequest: 1,
                gpuRequestType: GpuRequestType.Portion,
                gpuPortionRequest: 1,
                largeShmRequest: false
            }
        };

        return await this.create(computeCreationRequest);
    }

    async delete(id: string): Promise<void> {
        try {
            logger.debug(`deleting compute ${id}`);
            assetsApi.computeApi.accessToken = flowRootState.users.system_administrator.accessToken;

            await assetsApi.computeApi.deleteComputeAssetById(id);
            logger.debug(`compute ${id} deleted successfully`);
        } catch (e) {
            logger.error(`Failed to delete compute ${id}: ${e}`);
            throw e;
        }
    }

    async deleteAllProjectScopedComputeResources(
        accessToken = flowRootState.users.system_administrator.accessToken,
        clusterId = flowRootState.cluster.uuid
    ): Promise<void> {
        try {
            assetsApi.computeApi.accessToken = accessToken;
            const computeResources: ComputeAsset[] = (
                await assetsApi.computeApi.listComputeAssets(
                    undefined,
                    undefined,
                    undefined,
                    undefined,
                    clusterId
                )
            ).body.entries;
            const projectScopedComputeResources = computeResources.filter(
                (cr) => cr.meta.scope === Scope.Project
            );
            await Promise.all(
                projectScopedComputeResources.map((cr) =>
                    assetsApi.computeApi.deleteComputeAssetById(cr.meta.id)
                )
            );
            if (projectScopedComputeResources.length > 0) {
                logger.info('deleted all project scoped environments successfully');
            }
        } catch (error) {
            logger.warn(`failed to delete all project scoped compute resources: ${error}`);
        }
    }

    async deleteComputeResources(
        prefix: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            assetsApi.computeApi.accessToken = accessToken;
            const computeResources: ComputeAsset[] = (await assetsApi.computeApi.listComputeAssets()).body
                .entries;
            const computeResourcesToDelete = computeResources.filter((cr) => cr.meta.name.startsWith(prefix));
            await Promise.all(
                computeResourcesToDelete.map((cr) => assetsApi.computeApi.deleteComputeAssetById(cr.meta.id))
            );
            if (computeResourcesToDelete.length > 0) {
                logger.info(`deleted all compute-resources with prefix:"${prefix}" successfully`);
            }
        } catch (error) {
            logger.warn(`failed to deleted all compute-resources with prefix:"${prefix}". error: ${error}`);
        }
    }

    waitForComputeResourceToBeDeleted(name: string): Promise<void> {
        return waitUtils.waitForCondition(
            async () => {
                const computeResourcesNames = (await this.list()).map((compute) => compute.meta.name);
                return !computeResourcesNames.includes(name);
            },
            `Waiting for compute resource: ${name} to be deleted`,
            true,
            30
        );
    }

    async createAsUser(computeResource: ComputeCreationRequest, accessToken: string): Promise<ComputeAsset> {
        try {
            logger.info('creating compute resource', computeResource.meta.name);
            assetsApi.computeApi.accessToken = accessToken;

            const computeAsset: ComputeAsset = (
                await assetsApi.computeApi.createComputeAsset(computeResource)
            ).body;
            logger.info(
                `compute ${computeResource.meta.name} created successfully. id: ${computeAsset.meta.id}`
            );
            return computeAsset;
        } catch (e) {
            logger.error(`Failed to create compute ${computeResource.meta.name}: ${e}`);
            throw e;
        }
    }

    async deleteAsUser(id: string, accessToken: string): Promise<void> {
        try {
            logger.debug(`deleting compute ${id}`);
            assetsApi.computeApi.accessToken = accessToken;

            await assetsApi.computeApi.deleteComputeAssetById(id);
            logger.debug(`compute ${id} deleted successfully`);
        } catch (e) {
            logger.error(`Failed to delete compute ${id}: ${e}`);
            throw e;
        }
    }

    /**
     * Deletes a compute asset by its name.
     *
     * This function retrieves a list of compute assets and searches for an asset
     * with the specified `computeName`. If found, it deletes the asset based on its ID.
     * If no asset with the given name is found, an error is thrown.
     *
     * @async
     * @function
     * @param {string} computeName - The name of the compute asset to delete.
     * @returns {Promise<void>} Resolves when the asset is successfully deleted or throws an error if the asset is not found.
     * @throws {Error} If no compute asset with the specified name exists.
     */
    async deleteByName(computeName: string): Promise<void> {
        const computeId: string | undefined = await this.getId(computeName);
        if (!computeId) {
            throw new Error(`Compute asset with name ${computeName} not found`);
        }

        await this.delete(computeId);
    }
}
