/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../../modules/api';
import { ModelAsset, type EnvironmentAsset, type ComputeAsset } from '../../../modules/api/apis/assetsApi';
import { flowRootState } from '../../../modules/helpers';

export class AssetUtils {
    /**
     * @description get the first available model asset (specifying cluster in order to support saas as well)
     *
     * */
    async getModelsAssets(): Promise<ModelAsset[]> {
        assetsApi.modelApi.accessToken = flowRootState.users.editor.accessToken;
        const models: ModelAsset[] = (
            await assetsApi.modelApi.listModelAssets(
                undefined,
                undefined,
                undefined,
                undefined,
                flowRootState.cluster.uuid
            )
        ).body.entries;

        return models;
    }

    /**
     * @description get the first available compute asset (specifying cluster in order to support saas as well)
     *
     * */
    async getFirstComputeAsset(): Promise<ComputeAsset | undefined> {
        assetsApi.computeApi.accessToken = flowRootState.users.editor.accessToken;
        const computes: ComputeAsset[] = (
            await assetsApi.computeApi.listComputeAssets(
                undefined,
                undefined,
                undefined,
                undefined,
                flowRootState.cluster.uuid
            )
        ).body.entries;

        return computes[0];
    }

    /**
     * @description get the first available environment asset that supports inference workloads (specifying cluster in order to support saas as well)
     *
     * */
    async getFirstInferenceEnvironmentAsset(): Promise<EnvironmentAsset | undefined> {
        assetsApi.environmentApi.accessToken = flowRootState.users.editor.accessToken;
        const environments: EnvironmentAsset[] = (
            await assetsApi.environmentApi.listEnvironmentAssets(
                undefined,
                undefined,
                undefined,
                undefined,
                flowRootState.cluster.uuid,
                undefined,
                undefined,
                undefined,
                undefined,
                undefined,
                undefined,
                undefined,
                true
            )
        ).body.entries;

        return environments[0];
    }
}
