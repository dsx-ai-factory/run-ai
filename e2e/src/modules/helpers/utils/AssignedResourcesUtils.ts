/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AssignedResources } from '../../api/apis/projectsDepartmentsApi';
import { NodePoolAssignedResources } from '../../api/apis/projectsDepartmentsApi';
import { nodePoolUtils } from '../index';
import { Resources, Units } from '../../api/apis/orgUnitApi';

export class AssignedResourcesUtils {
    createDepartmentResources(deservedGpus = 0, limit = -1): Resources {
        const resources = new Resources();
        resources.gpu = {
            deserved: deservedGpus,
            limit: limit,
            overQuotaWeight: deservedGpus
        };
        resources.cpu = {
            deserved: null,
            limit: null,
            overQuotaWeight: null
        };
        resources.memory = {
            deserved: null,
            limit: null,
            overQuotaWeight: null,
            units: Units.Mib
        };

        return resources;
    }

    async createNodePoolResourcesForDefault(
        deservedGpus = 0,
        maxAllowed = -1
    ): Promise<Array<NodePoolAssignedResources>> {
        const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();

        const npResources = new NodePoolAssignedResources();
        npResources.nodePool = {
            id: defaultNodePool.id,
            name: defaultNodePool.name
        };

        npResources.gpu = {
            deserved: deservedGpus,
            maxAllowed: maxAllowed,
            overQuotaWeight: null
        };
        npResources.cpu = {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null
        };
        npResources.memory = {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null
        };

        return [npResources];
    }
}
