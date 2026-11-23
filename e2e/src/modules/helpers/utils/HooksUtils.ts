/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    authUtils,
    config,
    flowRootState,
    infoUtils,
    logger,
    policyUtils,
    settings,
    TokenObject,
    versionUtils
} from '../index';
import { bashCli } from '../../cli';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { CleanupMethod } from './AuthUtils';
import { SupportedVersions } from './SupporetedEnvsUtils';
import { preservedUsers } from '../config/config';

// ESM equivalent of __dirname
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export class HooksUtils {
    async addEnvironmentInfo(): Promise<void> {
        const envInfo = {
            'environment name': config.clusterNamePrefix,
            url: process.env.CONTROL_PLANE_URL || config.clusterUrl,
            'cluster version': flowRootState.cluster.version,
            'control plane version': flowRootState.controlPlaneVersion,
            'automation branch': await infoUtils.getAutomationBranch()
        };
        const envInfoFormatted = Object.entries(envInfo)
            .map(([key, value]) => `${key.replace(/ /g, '\\ ')} = ${value}`)
            .join('\n');

        const projectRoot = path.resolve(__dirname, '..', '..', '..', '..', '..');
        const dirPath = path.join(projectRoot, 'runai', 'allure-results');
        const envInfoPath = path.join(dirPath, 'environment.properties');

        if (!fs.existsSync(dirPath)) {
            fs.mkdirSync(dirPath, { recursive: true });
        }

        fs.writeFileSync(envInfoPath, envInfoFormatted, 'utf8');
    }

    async enableFeatureFlags(): Promise<void> {
        const clusterVersion = flowRootState.cluster.version;
        const isV22 = versionUtils.isVersionSupported(clusterVersion, SupportedVersions.V2_23);


        const baselineFeaturePromises = [
            policyUtils.enableClusterPolicySync(),
            apiUtils.enableHybridSubmission(),
            apiUtils.enableModelCatalog()
        ];
        const versionedTasks: Promise<unknown>[] = [];
        if (isV22) {
            versionedTasks.push(apiUtils.enableTemplatesV2());
        }

        const versionedSoftFail = versionedTasks.map((task) =>
            task.catch((error: unknown) => {
                logger.warn(error);
            })
        );

        await Promise.all([...baselineFeaturePromises, ...versionedSoftFail]);
    }

    async disableApprovalRequiredPolicy(): Promise<void> {
        // Older clusters 2.20 and below are required to send an approval to the policy service per cluster in order to create/modify policies.
        // existing tests are affected by this and therefore are failing,
        // this function disables the approval required for the policy service.
        // once the oldest supported version is 2.23, this function can be removed.
        try {
            const isNotAffectedVersion = versionUtils.isVersionSupported(
                flowRootState.cluster.version,
                SupportedVersions.V2_23
            );
            if (isNotAffectedVersion) return;
            await policyUtils.disableApprovalRequiredPolicy();
        } catch (error) {
            logger.error(
                `Failed to disable approval required policy for cluster ${flowRootState.cluster.uuid}: ${error}`
            );
        }
    }

    async cleanupAllNonDefaultUsers(adminTokens: TokenObject): Promise<void> {
        if (!settings.isSelfHosted && !settings.isLocalRun && process.env.CONTROL_PLANE_URL.includes('e2e')) {
            logger.debug('going to delete all users but the admin from saas-nightly environment');
            try {
                if (!preservedUsers.length) {
                    throw new Error('E2E_PRESERVED_USERS must be set before shared-environment user cleanup.');
                }
                await authUtils.cleanUpUsers(
                    adminTokens.idToken,
                    CleanupMethod.DeleteAllUsersBesidesSpecifiedUsers,
                    undefined,
                    [config.controlPlaneAdminUsername, ...preservedUsers].filter(Boolean)
                );
            } catch (error) {
                logger.warn(`Failed to cleanup users: ${error}`);
            }
        }
    }

    async configureKubeconfigToSkipTLSVerification(): Promise<void> {
        try {
            const setClusterCommand = `kubectl config set-cluster ${flowRootState.cluster.name} --insecure-skip-tls-verify=true`;
            await bashCli.asyncExec(setClusterCommand);
        } catch (e) {
            logger.error(`failed to configure kubeconfig to skip tls verification ${e}`);
        }
    }

    forceTheTerminationOfOpenBrowsers(): void {
        try {
            //forcing the termination of open browsers if there are
            bashCli.exec('pkill -9 -f "ms-playwright"');
        } catch (e) {
            logger.debug('failed to kill processes with open browsers');
        }
    }

    // when SAVE_TAGS is set to true, run staging nightly once per tag
    saveEnvTags(): void {
        if (process.env.SAVE_TAGS) {
            const latestTags = {
                'control-plane-version': flowRootState.controlPlaneVersion,
                'cluster-version': flowRootState.cluster.version
            };
            fs.writeFileSync('tags.json', JSON.stringify(latestTags));
            logger.info('created tags.json file');
        }
    }
}
