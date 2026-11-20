/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    authUtils,
    clusterUtils,
    crds,
    flowRootState,
    logger,
    usersManager
} from '../src/modules/helpers';
import { kubeCli, runaiCliV2 } from '../src/modules/cli';
import { config } from '../src/modules/helpers';
import { resolve } from 'path';
import * as fs from 'fs';
import { resolveNodePoolCrd } from '../src/modules/helpers/utils/NodePoolCrdFields';

const STATE_DIR = '/app/e2e';
export const stateFile = resolve(STATE_DIR, './globalState.json');

let isTokenRefreshing = false;

async function refreshTokensIfNeeded(): Promise<void> {
    try {
        if (isTokenRefreshing) {
            return;
        }
        isTokenRefreshing = true;

        const admin = flowRootState.users?.system_administrator;
        const token = admin?.idToken ?? admin?.accessToken;
        const tokenExp = authUtils.getTokenExpMin(token);
        const tokensNeedRefresh = authUtils.isTokenExpiredOrInvalid(token, tokenExp);

        if (tokensNeedRefresh) {
            await usersManager.authenticateAllLocalUsers(flowRootState.users);
            logger.info('Tokens refreshed successfully.');
            if (fs.existsSync(stateFile)) {
                try {
                    const sharedFlowRootState = { flowRootState };
                    fs.writeFileSync(stateFile, JSON.stringify(sharedFlowRootState, null, 2));
                } catch (error) {
                    logger.error(`Failed to write globalState.json during token refresh: ${error.message}`);
                }
            }
        }
    } catch (error) {
        logger.error(`Error during token refresh: ${error}`);
    } finally {
        isTokenRefreshing = false;
    }
}

export interface KitSetupOptions {
    kitName: string;
    setTimerFn: (t: NodeJS.Timeout) => void;
}

export async function runKitGlobalSetup(options: KitSetupOptions): Promise<void> {
    const { kitName, setTimerFn } = options;
    logger.info(`${kitName} Playwright global setup begin`);
    logger.info(`Process: ${JSON.stringify(process.argv)}`);

    // 1. Delete existing globalState.json
    if (fs.existsSync(stateFile)) {
        try {
            fs.unlinkSync(stateFile);
            logger.info('Deleted existing globalState.json file.');
        } catch (error) {
            throw new Error(`Failed to delete globalState.json: ${error.message}`);
        }
    }

    // 2. Authenticate admin user
    logger.debug('Going to authenticate admin user');
    const adminTokens = await authUtils.authenticateUser(
        config.controlPlaneAdminUsername,
        config.controlPlaneAdminPassword
    );
    logger.debug('Admin was authenticated');

    // 3. Get cluster by prefix
    flowRootState.cluster = await apiUtils.getClusterByPrefix(config.clusterNamePrefix, adminTokens.idToken);

    // 4. Initialize users
    logger.debug('Going to create users');
    await usersManager.initLocalUsers(adminTokens.idToken);
    logger.debug('Users were successfully created');

    // 6. Initialize kube context
    if (!config.kubeconfigClusterName) {
        throw new Error('kubeconfigClusterName is required but not set');
    }
    await kubeCli.initKubeContext(config.kubeconfigClusterName);
    kubeCli.useAdminContext();

    // 7. Delete disconnected clusters
    await clusterUtils.deleteDisconnectedClusters();

    // 8. CLI tools setup
    logger.info(`Setting up CLI tools for ${kitName}...`);
    try {
        if (adminTokens && adminTokens.idToken) {
            await Promise.all([kubeCli.setFlowRootStateResources(adminTokens)]);
            logger.info('CLI tools setup completed successfully');
        } else {
            logger.warn('Admin tokens not available, skipping CLI setup');
        }
    } catch (error) {
        logger.warn(`CLI setup warning: ${error.message}`);
    }

    // 8b. V2 CLI: download and install the runai CLI so CLI-tagged tests can run
    // (mirrors the dgxc setup). Guarded so a CLI-download hiccup degrades to
    // CLI-test failures rather than failing the whole run.
    try {
        await runaiCliV2.downloadAndInstall();
        logger.info('V2 CLI setup completed successfully');
    } catch (error) {
        logger.warn(`V2 CLI setup failed, continuing without V2 CLI: ${error.message}`);
    }

    // 9. Initialize CRDs
    if (flowRootState.cluster?.version) {
        crds.initCrds(flowRootState.cluster.version);
        await resolveNodePoolCrd();
    }

    // 10. Write globalState.json
    const sharedFlowRootState = { flowRootState };
    fs.writeFileSync(stateFile, JSON.stringify(sharedFlowRootState, null, 2));

    // 11. Start periodic token refresh
    const timer = setInterval(
        async () => {
            logger.debug('Going to refresh token if needed');
            await refreshTokensIfNeeded();
        },
        10 * 1000 // 10 seconds
    );
    setTimerFn(timer);

    logger.info(`${kitName} global setup terminated successfully`);
}
