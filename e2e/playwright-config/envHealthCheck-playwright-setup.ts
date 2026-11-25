/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { kubeCli } from '../src/modules/cli';
import { config, ensureControlPlaneCredentials, logger } from '../src/modules/helpers';


async function envHealthCheckGlobalSetup(): Promise<void> {
    logger.debug('Before all basic hook begin');
    await kubeCli.initKubeContext(config.kubeconfigClusterName, false);
    kubeCli.useAdminContext();
    await ensureControlPlaneCredentials();
    logger.debug('Before all default hook ended successfully');
}

export default envHealthCheckGlobalSetup; 