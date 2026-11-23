/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import {
    clusterService,
    clusterUtils,
    flowRootState,
    logger,
    objectGenerator,
    tags
} from '../../../../../../modules/helpers';
import { DisplayedCluster } from '../../../../../../modules/api/apis/clusterApi';

test.describe(
    'Installer bootstrap token cluster scope validation',
    {
        tag: [
            tags.teams.phoenix,
            tags.tool.api,
            tags.resource.cluster,
            tags.services.clusterService,
            tags.runOn.parallelCI
        ]
    },
    () => {
        let clusterA: DisplayedCluster;
        let clusterB: DisplayedCluster;
        let bootstrapTokenForClusterA: string;

        test.beforeAll(async (): Promise<void> => {
            const accessToken = flowRootState.users.system_administrator.accessToken;
            clusterA = await clusterUtils.createCluster(
                accessToken,
                objectGenerator.generateNameByPrefix('installer-scope-a'),
                '',
                '1.0.0'
            );
            clusterB = await clusterUtils.createCluster(
                accessToken,
                objectGenerator.generateNameByPrefix('installer-scope-b'),
                '',
                '1.0.0'
            );
            const clusterInfo = await clusterService.getClusterInfo(
                accessToken,
                clusterA.uuid,
                clusterA.version
            );
            bootstrapTokenForClusterA = clusterInfo.bootstrapToken;
        });

        test.afterAll(async (): Promise<void> => {
            const accessToken = flowRootState.users.system_administrator.accessToken;
            try {
                await clusterUtils.deleteCluster(accessToken, clusterA.uuid);
                await clusterUtils.deleteCluster(accessToken, clusterB.uuid);
            } catch (error) {
                logger.error(`failed to delete installer bootstrap scope test clusters: ${error}`);
            }
        });

        test('grants access to its own cluster control plane config', async (): Promise<void> => {
            const config = await clusterService.getClusterControlPlaneConfig(
                bootstrapTokenForClusterA,
                clusterA.uuid
            );
            expect(config.clusterName, 'control plane config should belong to cluster A').to.equal(
                clusterA.name
            );
        });

        test('is rejected when used against a different cluster control plane config', async (): Promise<void> => {
            const error: any = await clusterService
                .getClusterControlPlaneConfig(bootstrapTokenForClusterA, clusterB.uuid)
                .then(
                    () => expect.fail('fetching cluster B config with cluster A bootstrap token should fail'),
                    (e: any) => e
                );
            expect(
                error.response.statusCode,
                'cluster A bootstrap token used against cluster B config should return 403 Forbidden'
            ).to.equal(StatusCodes.FORBIDDEN);
        });

        test('rejects a garbage bootstrap token', async (): Promise<void> => {
            const error: any = await clusterService
                .getClusterControlPlaneConfig('invalidToken', clusterA.uuid)
                .then(
                    () => expect.fail('fetching cluster config with an invalid token should fail'),
                    (e: any) => e
                );
            expect(
                error.response.statusCode,
                'invalid bootstrap token should return 401 Unauthorized'
            ).to.equal(StatusCodes.UNAUTHORIZED);
        });
    }
);
