/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    flowRootState,
    hooksUtils,
    logger,
    operations,
    settings
} from '../src/modules/helpers';
import { config } from '../src/modules/helpers';
import { resolve, dirname } from 'path';
import * as fs from 'fs';
import { execSync } from 'child_process';
import { clearTimer } from '../playwright.config';
import { CleanupMethod } from '../src/modules/helpers/utils/AuthUtils';
import { generateSkipsHTMLReport } from '../src/modules/helpers/utils/SkippedTestUtils';

// Use the E2E root directory for proper path resolution
const __dirname = '/app/e2e';

export const stateFile = resolve(__dirname, './globalState.json');

// Function to write environment info for Allure reports
function writeEnvironmentInfo(): void {
    try {
        const envInfo = {
            'Kubernetes Client Version': execSync('kubectl version --client', { encoding: 'utf8' }).match(/Client Version: v([^\s]+)/)?.[1] ? `v${execSync('kubectl version --client', { encoding: 'utf8' }).match(/Client Version: v([^\s]+)/)?.[1]}` : 'Not available',
            'Kubernetes Server Version': execSync('kubectl version', { encoding: 'utf8' }).match(/Server Version: v([^\s]+)/)?.[1] ? `v${execSync('kubectl version', { encoding: 'utf8' }).match(/Server Version: v([^\s]+)/)?.[1]}` : 'Not available',
            'cluster version': flowRootState.cluster?.version || 'Not available',
            'control plane version': flowRootState.controlPlaneVersion || 'Not available',
            'environment name': config.clusterNamePrefix,
            'url': process.env.CONTROL_PLANE_URL || config.clusterUrl,
            'Operating System': process.platform
        };
        
        // Try multiple possible locations for allure-results directory
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
                    // Ensure directory exists
                    const dir = dirname(envFile);
                    if (!fs.existsSync(dir)) {
                        fs.mkdirSync(dir, { recursive: true });
                    }
                    
                    // Write environment.properties file (format compatible with existing HooksUtils)
                    const envInfoFormatted = Object.entries(envInfo)
                        .map(([key, value]) => `${key.replace(/ /g, '\\ ')} = ${value}`)
                        .join('\n');
                    fs.writeFileSync(envFile, envInfoFormatted, 'utf8');
                    
                    // Write environment.xml file for better Allure compatibility
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

export default async function (): Promise<void> {
    logger.debug('Certification kit global teardown begin');

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
    
    // Write environment info for Allure reports
    if (!settings.isLocalRun) {
        try {
            writeEnvironmentInfo();
        } catch (e) {
            logger.error(`Failed to add environment info, error: ${e}`);
        }
    } else {
        // Even in local run, try to write environment info for certification kit
        try {
            writeEnvironmentInfo();
        } catch (e) {
            logger.warn(`Could not write environment info in local run: ${e}`);
        }
    }
    
    logger.info('Certification kit global teardown completed successfully');
}
