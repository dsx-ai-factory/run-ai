/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as fs from 'fs';
import * as path from 'path';
import archiver from 'archiver';
import { execSync } from 'child_process';
import { logger } from '../../logger/logger';
import { kubeCli } from '../../../cli';

export async function getPodLogsAndZip(): Promise<void> {
    const namespacesOfRunaiPods = ['runai', 'runai-backend']; // Namespaces to categorize as runai pods
    const logsDir = './pods-logs';
    const runaiPodsDir = path.join(logsDir, 'runai pods'); // Directory for runai pod logs
    const nonRunaiPodsDir = path.join(logsDir, 'non runai pods'); // Directory for non-runai pod logs
    const zipFileName = 'pods-logs.zip';
    const podsStatusesFile = path.join(logsDir, 'pods-statuses.log'); // File for pods statuses

    try {
        [logsDir, runaiPodsDir, nonRunaiPodsDir].forEach((dir) => {
            if (!fs.existsSync(dir)) {
                fs.mkdirSync(dir, { recursive: true });
            }
        });

        logger.debug('Fetching pod statuses...');
        const podsStatuses = execSync('kubectl get pods -A').toString();
        fs.writeFileSync(podsStatusesFile, podsStatuses);
        logger.debug(`Saved pods statuses to: ${podsStatusesFile}`);

        logger.debug('Fetching all pods...');
        const podList = await kubeCli.CoreV1Api.listPodForAllNamespaces();

        for (const pod of podList.items) {
            const podName = pod.metadata?.name;
            const namespace = pod.metadata?.namespace;

            if (!podName || !namespace) continue;

            const targetDir = namespacesOfRunaiPods.includes(namespace) ? runaiPodsDir : nonRunaiPodsDir;

            logger.debug(`Fetching logs for pod: ${namespace}/${podName}`);
            try {
                const logResponse = await kubeCli.CoreV1Api.readNamespacedPodLog({
                    name: podName,
                    namespace
                });
                const logContent = logResponse;

                const logFilePath = path.join(targetDir, `${podName}.log`);
                fs.writeFileSync(logFilePath, logContent);
                logger.debug(`Saved logs to: ${logFilePath}`);
            } catch (err) {
                logger.debug(`Failed to fetch logs for pod: ${namespace}/${podName}`, err);
            }
        }

        logger.debug('Zipping logs...');
        const output = fs.createWriteStream(zipFileName);
        const archive = archiver('zip', { zlib: { level: 9 } });

        output.on('close', () => {
            logger.debug(`Logs zipped into: ${zipFileName} (${archive.pointer()} bytes)`);
        });

        archive.on('error', (err) => {
            throw err;
        });

        archive.pipe(output);
        archive.directory(logsDir, false);
        await archive.finalize();
        logger.debug('Log zipping completed.');

        fs.rmSync(logsDir, { recursive: true, force: true });
        logger.debug('Cleaned up log files and directories.');
    } catch (err) {
        logger.debug('Error occurred:', err);
    }
}
