/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    authV2Utils,
    flowRootState,
    hooksUtils,
    logger,
    operations
} from '../src/modules/helpers';
import { config } from '../src/modules/helpers';
import { dirname } from 'path';
import * as fs from 'fs';
import { execSync } from 'child_process';
import { CleanupMethod } from '../src/modules/helpers/utils/AuthUtils';
import { generateSkipsHTMLReport } from '../src/modules/helpers/utils/SkippedTestUtils';
import { stateFile } from './kit-shared-global-setup';

function writeEnvironmentInfo(): void {
    try {
        const envInfo = {
            'Kubernetes Client Version': ((): string => {
                try {
                    const match = execSync('kubectl version --client', { encoding: 'utf8' }).match(/Client Version: v([^\s]+)/);
                    return match ? `v${match[1]}` : 'Not available';
                } catch {
                    return 'Not available';
                }
            })(),
            'Kubernetes Server Version': ((): string => {
                try {
                    const match = execSync('kubectl version', { encoding: 'utf8' }).match(/Server Version: v([^\s]+)/);
                    return match ? `v${match[1]}` : 'Not available';
                } catch {
                    return 'Not available';
                }
            })(),
            'cluster version': flowRootState.cluster?.version || 'Not available',
            'control plane version': flowRootState.controlPlaneVersion || 'Not available',
            'environment name': config.clusterNamePrefix,
            'url': process.env.CONTROL_PLANE_URL || config.clusterUrl,
            'Operating System': process.platform
        };

        const possiblePaths = [
            './allure-results',
            '/app/e2e/allure-results',
            '/app/allure-results',
            process.env.ALLURE_RESULTS_DIR || './allure-results'
        ];

        let writtenTo = '';

        for (const basePath of possiblePaths) {
            if (fs.existsSync(basePath) || basePath.startsWith('./')) {
                const envFile = `${basePath}/environment.properties`;
                const envXmlFile = `${basePath}/environment.xml`;

                try {
                    const dir = dirname(envFile);
                    if (!fs.existsSync(dir)) {
                        fs.mkdirSync(dir, { recursive: true });
                    }

                    const envInfoFormatted = Object.entries(envInfo)
                        .map(([key, value]) => `${key.replace(/ /g, '\\ ')} = ${value}`)
                        .join('\n');
                    fs.writeFileSync(envFile, envInfoFormatted, 'utf8');

                    const envXmlContent = `<?xml version="1.0" encoding="UTF-8"?>
<environment>
${Object.entries(envInfo).map(([key, value]) => `    <parameter key="${key}" value="${value}"/>`).join('\n')}
</environment>`;
                    fs.writeFileSync(envXmlFile, envXmlContent);

                    writtenTo = basePath;
                    break;
                } catch (error) {
                    logger.debug(`Could not write to ${basePath}: ${error.message}`);
                    continue;
                }
            }
        }

        if (writtenTo) {
            logger.info(`Environment info written to Allure results directory: ${writtenTo}`);
        } else {
            logger.warn('Could not write environment info to any location');
        }
    } catch (error) {
        logger.warn(`Could not write environment info: ${error.message}`);
    }
}

export interface KitTeardownOptions {
    kitName: string;
    clearTimerFn: () => void;
}

export async function runKitGlobalTeardown(options: KitTeardownOptions): Promise<void> {
    const { kitName, clearTimerFn } = options;
    logger.debug(`${kitName} global teardown begin`);

    // 1. Handle missing globalState.json (setup failed)
    if (!fs.existsSync(stateFile)) {
        logger.info('Global playwright setup Failed. Please check why');
        logger.info(`Process: ${JSON.stringify(process.argv)}`);

        const isUpgradeSuite = process.argv.some(arg =>
            arg.includes('upgrade-health-check') || arg.includes('pre-upgrade') || arg.includes('post-upgrade')
        );
        if (isUpgradeSuite) {
            logger.info('Detected upgrade test. Skipping teardown.');
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

    // 2. Force cleanup and user cleanup
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
        logger.debug('After all default hook ended successfully');
    } catch (error) {
        logger.warn(`Failed to cleanup: ${error}`);
    }

    // 3. Final cleanup actions
    hooksUtils.forceTheTerminationOfOpenBrowsers();
    generateSkipsHTMLReport();

    // 4. Delete globalState.json
    try {
        if (fs.existsSync(stateFile)) {
            fs.unlinkSync(stateFile);
            logger.debug(`Deleted state file: ${stateFile}`);
        }
    } catch (error) {
        throw new Error(`Failed to delete globalState.json: ${error.message}`);
    }

    // 5. Clear timer and save env tags
    clearTimerFn();
    hooksUtils.saveEnvTags();

    // 6. Write environment info for Allure
    try {
        writeEnvironmentInfo();
    } catch (e) {
        logger.warn(`Could not write environment info: ${e}`);
    }

    logger.info(`${kitName} global teardown completed successfully`);
}
