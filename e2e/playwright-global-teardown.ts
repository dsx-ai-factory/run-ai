/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    authUtils,
    config,
    flowRootState,
    hooksUtils,
    logger,
    operations,
    settings,
    usersManager
} from './src/modules/helpers';
import { generateSkipsHTMLReport } from './src/modules/helpers/utils/SkippedTestUtils';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

// ESM equivalent of __dirname
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Path to the global state file
const stateFile = path.resolve(__dirname, './globalState.json');
// Load shared configuration if the state file exists
export const sharedConfig = fs.existsSync(stateFile) ? JSON.parse(fs.readFileSync(stateFile, 'utf-8')) : {};
import { clearTimer } from './playwright.config';
import { CleanupMethod } from './src/modules/helpers/utils/AuthUtils';
import { getPodLogsAndZip } from './src/modules/helpers/utils/testing/podLogsUtils';
import { StatusCodes } from 'http-status-codes';

export default async function (): Promise<void> {
    logger.debug('After all default hook begin');

    if (!fs.existsSync(stateFile)) {
        logger.info('Global playwright setup Failed. Please check why');
        logger.info(`Process: ${JSON.stringify(process.argv)}`);
        const isUpgradeSuite = process.argv.some(arg =>
            arg.includes('upgrade-health-check') || arg.includes('pre-upgrade') || arg.includes('post-upgrade')
        );
        if (isUpgradeSuite) {
            logger.info('Detected upgrade test. Running pre-upgrade setup...');
            return;
        }

        logger.debug('Going to authenticate admin user');
        const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        logger.debug('Admin was authenticated');

        await hooksUtils.cleanupAllNonDefaultUsers(adminTokens);
        return;
    }
    try {
        await operations.forceCleanup();
            const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
            await authUtils.cleanUpUsers(
                adminTokens.idToken,
                CleanupMethod.DeleteOnlyUsersWithSubString,
                flowRootState.currentFlowUuid
            );

            if (process.env.INIT_SSO_USERS_AND_IDP === 'true' && flowRootState.usersSSO) {
                logger.debug(`FlowRootState usersSSO: ${JSON.stringify(Object.keys(flowRootState.usersSSO))}`);
                for (const [propertyName, user] of Object.entries(flowRootState.usersSSO)) {
                    try {
                        const username = user.email.split('@')[0]; 
                        await usersManager.deleteSSOUser(
                            adminTokens.idToken,
                            username,
                            user.userId,
                            'automation'
                        );
                        logger.debug(`Deleted SSO user: ${username}, email: ${user.email}`);
                    } catch (error) {
                        if (error.response?.status === StatusCodes.NOT_FOUND || error.statusCode === StatusCodes.NOT_FOUND) {
                            logger.debug(`SSO user ${propertyName} not found 404 error: ${error.message}`);
                        } else {
                            logger.error(`Error deleting SSO user ${propertyName}: ${error}`);
                        }
                    }
                }
                logger.debug('SSO users cleanup completed');
            }
            logger.debug('After all default hook ended successfully');
        
    } catch (error) {
        logger.warn(`Failed to cleanup: ${error}`);
    }

    // Final cleanup actions
    hooksUtils.forceTheTerminationOfOpenBrowsers();
    generateSkipsHTMLReport();

    // Delete the global state file
    try {
        if (fs.existsSync(stateFile)) {
            fs.unlinkSync(stateFile);
            logger.debug(`Deleted state file: ${stateFile}`);
        } else {
            logger.debug(`State file not found, skipping deletion: ${stateFile}`);
        }
    } catch (error) {
        throw new Error(`Failed to delete globalState.json: ${error.message}`);
    }
    clearTimer();
    hooksUtils.saveEnvTags();
    if (!settings.isLocalRun) {
        try {
            await hooksUtils.addEnvironmentInfo();
        } catch (e) {
            logger.error(`Failed to add environment info, error: ${e}`);
        }
    }
    if (process.env.SHOULD_ZIP_PODS_LOGS) {
        try {
            await getPodLogsAndZip();
        } catch (e) {
            logger.error(`Failed to add environment info, error: ${e}`);
        }
    }
    logger.info('Token refresh timer stopped');
}
