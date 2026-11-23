/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    authUtils,
    authV2Utils,
    clusterUtils,
    crds,
    flowRootState,
    hooksUtils,
    infoUtils,
    logger,
    reportApiUtils,
    usersManager
} from './src/modules/helpers';
import { preservedUsers, resetFlowRootState } from './src/modules/helpers/config/config';
import { kubeCli, runaiCliV2 } from './src/modules/cli';
import { config } from './src/modules/helpers';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';
import fs from 'fs';
import { setTimer } from './playwright.config';
import { CleanupMethod } from './src/modules/helpers/utils/AuthUtils';
import { resolveNodePoolCrd } from './src/modules/helpers/utils/NodePoolCrdFields';
import { resolveOrgUnitCrds } from './src/modules/helpers/utils/OrgUnitCrdFields';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export const stateFile = resolve(__dirname, './globalState.json');

async function globalSetup(): Promise<void> {
    logger.info('Before all playright suites default hook begin');
    logger.info(`Process: ${JSON.stringify(process.argv)}`);
    const isUpgradeSuite = process.argv.some(
        (arg) =>
            arg.includes('upgrade-health-check') ||
            arg.includes('pre-upgrade') ||
            arg.includes('post-upgrade')
    );
    if (isUpgradeSuite) {
        logger.info('Detected upgrade suite. Skipping');
        return;
    }

    if (fs.existsSync(stateFile)) {
        try {
            fs.unlinkSync(stateFile);
            logger.info('Deleted existing globalState.json file (stale from previous phase).');
        } catch (error) {
            throw new Error(`Failed to delete globalState.json: ${error.message}`);
        }
    }
    resetFlowRootState();

    const adminTokens = await authUtils.authenticateUser(
        config.controlPlaneAdminUsername,
        config.controlPlaneAdminPassword
    );
    logger.debug(`Admin was authenticated , with tokens accessToken: ${JSON.stringify(adminTokens.accessToken)} and idToken: ${JSON.stringify(adminTokens.idToken)}`);
    try {
        if (!preservedUsers.length) {
            throw new Error('E2E_PRESERVED_USERS must be set before user cleanup.');
        }
        await authUtils.cleanUpUsers(
            adminTokens.idToken,
            CleanupMethod.DeleteAllUsersBesidesSpecifiedUsers,
            undefined,
            [config.controlPlaneAdminUsername, ...preservedUsers].filter(Boolean)
        );
    } catch (error) {
        logger.error(`Failed to cleanup users: ${error.message}`);
    }
    flowRootState.cluster = await apiUtils.getClusterByPrefix(config.clusterNamePrefix, adminTokens.idToken);
    logger.debug('Going to create users and authenticate them for tokens');
    await usersManager.initLocalUsers(adminTokens.accessToken);
    if (process.env.INIT_SSO_USERS_AND_IDP === 'true') {
        await usersManager.initSSOFlowUsers(adminTokens.accessToken);
    }
    logger.debug('Users were successfully created');

    // wait until new users' permissions are live before suites use their tokens.
    await authV2Utils.waitForPermissionsPropagation();

    await Promise.all([runaiCliV2.downloadAndInstall(), kubeCli.setFlowRootStateResources(adminTokens)]);
    await hooksUtils.enableFeatureFlags();
    await hooksUtils.disableApprovalRequiredPolicy();
    await kubeCli.initKubeContext(config.kubeconfigClusterName);
    kubeCli.useAdminContext();
    await clusterUtils.deleteDisconnectedClusters();
    crds.initCrds(flowRootState.cluster.version);
    await resolveNodePoolCrd();
    await resolveOrgUnitCrds();
    await infoUtils.printInitSummary();
    await reportApiUtils.generateConsumptionReportsForSuites();

    const sharedFlowRootState = {
        flowRootState: flowRootState
    };
    fs.writeFileSync(stateFile, JSON.stringify(sharedFlowRootState, null, 2));
    // refresh the token periodically
    const timer = setInterval(
        async () => {
            logger.debug('Going to refresh token if needed');
            await authUtils.refreshTokensIfNeeded(stateFile);
        },
        10 * 1000 // 10 seconds
    );
    setTimer(timer); // Store the timer reference
    logger.debug('Global setup terminated successfully');
}

export default globalSetup;