/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    clusterService,
    crds,
    flowRootState,
    logger,
    settings,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { bashCli, kubeCli } from '../../../../../../../modules/cli';
import fs from 'fs';
import path from 'path';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

    test.describe(`Helm clusterConfig values patch RunaiConfig: ${tags.teams.phoenix} ${tags.resource.patchRunaiConfig}`, () => {
        shouldRunTests(settings.isInstallRunaiConfigValues)(() => {
        type RunaiConfigLike = { spec: any; status?: { conditions?: any[] } };
        let cluster: any;
        let installInfo: any;

        test.beforeAll(async () => {
            // Verify global setup completed successfully
            expect(flowRootState.cluster).toBeDefined();
            expect(flowRootState.users).toBeDefined();
            expect(flowRootState.users.system_administrator).toBeDefined();

            cluster = flowRootState.cluster;
            logger.info(`Using cluster from global setup: ${cluster.name} (${cluster.uuid})`);

            installInfo = await clusterService.getClusterInfo(
                flowRootState.users.system_administrator.accessToken,
                cluster.uuid,
                '2.23'
            );
            expect(installInfo).toBeTruthy();
            logger.info('Successfully fetched cluster installation info');
        });

        test('install runai-cluster with -f values and verify RunaiConfig', async () => {
            // Helm upgrade operations take time
            test.setTimeout(600000);
            const baseCmd: string = installInfo.installationStr;
            const valuesPath = path.join(process.cwd(), 'runaiconfig.values.yaml');
            const values = `
    clusterConfig:      
      global:
        runtime: containerd
        labeluser:
          enabled: false
        core:
          swap:
            enabled: true
            mode: mapped
    `;
            fs.writeFileSync(valuesPath, values);

            const lines = baseCmd.split('\n');
            const commands: string[] = [];
            let currentCmd = '';

            for (const line of lines) {
                const trimmed = line.trim();
                if (!trimmed) continue;

                if (currentCmd.endsWith('\\')) {
                    currentCmd = currentCmd.slice(0, -1).trim() + ' ' + trimmed;
                } else if (currentCmd) {
                    commands.push(currentCmd);
                    currentCmd = trimmed;
                } else {
                    currentCmd = trimmed;
                }
            }

            if (currentCmd) {
                commands.push(currentCmd);
            }

            // inject -f values into the helm upgrade command
            const commandsWithValues = commands.map((cmd) => {
                if (cmd.match(/helm\s+(install|upgrade).*runai-cluster/)) {
                    // Insert -f after "runai-cluster runai/runai-cluster -n runai"
                    return cmd.replace(
                        /(helm\s+upgrade\s+-i\s+runai-cluster\s+runai\/runai-cluster\s+-n\s+runai)/,
                        `$1 -f ${valuesPath}`
                    );
                }
                return cmd;
            });

            const singleCommand = commandsWithValues.join(' && ');
            logger.info(`Executing helm command: ${singleCommand}`);

            const scriptPath = path.join(process.cwd(), '.helm-cmd.sh');
            try {
                fs.writeFileSync(scriptPath, singleCommand);
                await bashCli.asyncExec(`sh ${scriptPath}`);
            } catch (error) {
                logger.error(`Helm command failed with error: ${error.message}`);
                throw error;
            } finally {
                if (fs.existsSync(scriptPath)) {
                    fs.unlinkSync(scriptPath);
                }
            }

            // Wait for RunaiConfig to become Available and Reconciled
            await waitUtils.waitForCondition(
                async () => {
                    // getCrd returns a single object, not a list
                    const item = (await kubeCli.getCrd(
                        crds.runaiConfigCrd,
                        'runai',
                        'runai'
                    )) as RunaiConfigLike;
                    if (!item) return false;
                    const conditions = item.status?.conditions || [];
                    const available = conditions.some(
                        (c: any) => c.type === 'Available' && (c.status === true || c.status === 'True')
                    );
                    const reconciled = conditions.some(
                        (c: any) => c.type === 'Reconciled' && (c.status === true || c.status === 'True')
                    );
                    return available && reconciled;
                },
                'Wait for RunaiConfig Available & Reconciled',
                true,
                180
            );

            // Validate RunaiConfig spec reflects our overrides
            const rc = (await kubeCli.getCrd(crds.runaiConfigCrd, 'runai', 'runai')) as RunaiConfigLike;
            const spec = rc.spec;
            expect(spec.global.runtime).toBe('containerd');
            expect(spec.global.labeluser.enabled).toBe(false);
            expect(spec.global.core.swap.enabled).toBe(true);
            expect(spec.global.core.swap.mode).toBe('mapped');
        });
    });
});
