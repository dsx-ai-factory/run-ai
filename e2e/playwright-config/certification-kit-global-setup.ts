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
import { resolve, dirname } from 'path';
import * as fs from 'fs';
import { execSync } from 'child_process';
import { setTimer } from '../playwright.config';
import { resolveNodePoolCrd } from '../src/modules/helpers/utils/NodePoolCrdFields';

const __dirname = '/app/e2e';

export const stateFile = resolve(__dirname, './globalState.json');
let isTokenRefreshing = false; // Control variable for token refresh state

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
                    const sharedFlowRootState = {
                        flowRootState: flowRootState
                    };

                    fs.writeFileSync(stateFile, JSON.stringify(sharedFlowRootState, null, 2));
                } catch (error) {
                    logger.error(`Failed to load globalState.json: ${error.message}`);
                }
            }
        }
    } catch (error) {
        logger.error(`Error during token refresh: ${error}`);
    } finally {
        isTokenRefreshing = false;
    }
}

export async function globalSetup(): Promise<void> {
    logger.info('certification kit Playwright suites default hook begin');
    //print the process as string
    logger.info(`Process: ${JSON.stringify(process.argv)}`);

    if (fs.existsSync(stateFile)) {
        try {
            fs.unlinkSync(stateFile);
            logger.info('Deleted existing globalState.json file.');
        } catch (error) {
            throw new Error(`Failed to delete globalState.json: ${error.message}`);
        }
    }

    logger.debug('Going to authenticate admin user');
    const adminTokens = await authUtils.authenticateUser(
        config.controlPlaneAdminUsername,
        config.controlPlaneAdminPassword
    );
    logger.debug('Admin was authenticated');

    flowRootState.cluster = await apiUtils.getClusterByPrefix(config.clusterNamePrefix, adminTokens.idToken);
    logger.debug('Going to create users');
    await usersManager.initLocalUsers(adminTokens.idToken);
    logger.debug('Users were successfully created');

    // await hooksUtils.enableFeatureFlags();
    // Initialize kube context
    console.log('DEBUG: About to call kubeCli.initKubeContext with:', config.kubeconfigClusterName);
    if (!config.kubeconfigClusterName) {
        throw new Error('kubeconfigClusterName is required but not set');
    }
    const clusterName = config.kubeconfigClusterName;
    await kubeCli.initKubeContext(clusterName);
    kubeCli.useAdminContext();
    await clusterUtils.deleteDisconnectedClusters();

    // CLI V1 Setup
    logger.info('Setting up CLI tools for certification kit...');
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

    // V2 CLI: download and install the runai CLI so CLI-tagged cert tests can run
    // (mirrors the dgxc setup). Guarded so a CLI-download hiccup degrades to
    // CLI-test failures rather than failing the whole certification run.
    try {
        await runaiCliV2.downloadAndInstall();
        logger.info('V2 CLI setup completed successfully');
    } catch (error) {
        logger.warn(`V2 CLI setup failed, continuing without V2 CLI: ${error.message}`);
    }

    if (flowRootState.cluster?.version) {
        crds.initCrds(flowRootState.cluster.version);
        await resolveNodePoolCrd();
    }

    const sharedFlowRootState = {
        flowRootState: flowRootState
    };

    fs.writeFileSync(stateFile, JSON.stringify(sharedFlowRootState, null, 2));

    // refresh the token periodically
    const timer = setInterval(
        async () => {
            logger.debug('Going to refresh token if needed');
            await refreshTokensIfNeeded();
        },
        10 * 1000 // 10 seconds
    );
    setTimer(timer); // Store the timer reference
    logger.info('Global certification kit Playwright setup terminated successfully');
}

export default globalSetup;
