/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { apiUtils, authUtils, infoUtils, upgradeUtils } from '.';
import { kubeCli } from '../cli';
import { config, flowRootState } from './config/config';
import { logger } from './logger/logger';
import { expect } from 'chai';

export class UpgradesHealthChecker {
    async isClusterVersionHaveOperator(): Promise<boolean> {
        const adminTokens = await upgradeUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        flowRootState.cluster = await apiUtils.getClusterByPrefix(
            config.clusterNamePrefix,
            adminTokens.idToken
        );
        const cluster = await infoUtils.getCluster(config.clusterNamePrefix, adminTokens.idToken);
        logger.debug('cluster version is: ' + cluster.version);
        return true;
    }
}
