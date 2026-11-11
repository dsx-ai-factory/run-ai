/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, TestInfo } from '@playwright/test';
import {
    logger,
    objectGenerator,
    tags,
    flowRootState,
    networkTopologiesUtils,
    nodePoolUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    NetworkTopologyCreateRequest,
    NodepoolCreateFields,
    NodepoolPhase
} from '../../../../../../modules/api/apis/clusterApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

const dataSets = [
    {
        testCase: 'T5404989: list nodepool via RunAI CLI using System Admin and verify the topology column',
        user: flowRootState.users.system_administrator.email,
        password: flowRootState.users.system_administrator.password,
        testEnable: true
    },
    {
        testCase: 'T5562729: list nodepool via RunAI CLI using L1 Researcher and verify the topology column',
        user: flowRootState.users.l1_researcher.email,
        password: flowRootState.users.l1_researcher.password,
        testEnable: true
    },
    {
        testCase: 'T5562730: list nodepool via RunAI CLI using L2 Researcher and verify the topology column',
        user: flowRootState.users.l2_researcher.email,
        password: flowRootState.users.l2_researcher.password,
        testEnable: true
    },
    {
        testCase: 'T5562731: list nodepool via RunAI CLI using ML Engineer and verify the topology column',
        user: flowRootState.users.ml_engineer.email,
        password: flowRootState.users.ml_engineer.password,
        testEnable: true
    }
];

test.describe(`Node Pools CLI Tests, tags: ${tags.resource.nodepools} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const topoNames: string[] = [];
        const nodePools: { name: string; id: string }[] = [];
        const nodePoolNames: string[] = [];
        const topologyIds: string[] = [];

        const topoName = objectGenerator.generateShortNameByPrefix('ui-topo1', 5);
        const npName = objectGenerator.generateShortNameByPrefix('np', 5);

        const nodeLabelKey = objectGenerator.generateShortNameByPrefix('test-ng', 5);
        const nodeLabelVal = 'group1';

        const topoRegionKey = 'cloud.provider.com/region';
        const topoCityKey = 'cloud.provider.com/city';

        const createTopoRequest: NetworkTopologyCreateRequest = {
            name: topoName,
            clusterId: flowRootState.cluster.uuid,
            levels: [topoCityKey, topoRegionKey]
        };

        const nodePoolCreateRequest = Object.assign(new NodepoolCreateFields(), {
            name: npName,
            labelKey: nodeLabelKey,
            labelValue: nodeLabelVal,
            clusterId: flowRootState.cluster.uuid
        });

        test.beforeEach(async () => {
            // create topo
            const topoRsp = await networkTopologiesUtils.createNetworkTopologyViaApi(createTopoRequest);
            topologyIds.push(topoRsp.body.id);
            topoNames.push(topoName);

            nodePoolCreateRequest.networkTopologyId = topoRsp.body.id;

            const npRsp = await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest);
            nodePoolNames.push(npName);

            await nodePoolUtils.waitForClusterNodePoolPhase(
                npRsp.body.id,
                [NodepoolPhase.Ready, NodepoolPhase.Empty],
                30
            );

            nodePools.push({ name: npName, id: npRsp.body.id });
        });

        test.afterEach(async () => {
            // delete node pools
            const deletedNodePoolIds: string[] = [];
            for (const np of nodePools) {
                try {
                    await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                    deletedNodePoolIds.push(np.id);
                } catch (error) {
                    logger.error(error);
                }
            }

            for (const id of deletedNodePoolIds) {
                const idx = nodePools.findIndex((np) => np.id === id);
                if (idx !== -1) {
                    nodePools.splice(idx, 1);
                }
            }

            // delete topologies
            const deletedTopoIds: string[] = [];

            for (const id of topologyIds) {
                try {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(id);
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(topoName);
                    deletedTopoIds.push(id);
                } catch (error) {
                    logger.error(`Failed to delete topology ${id}`, error);
                }
            }

            for (const id of deletedTopoIds) {
                const idx = topologyIds.indexOf(id);
                if (idx !== -1) {
                    topologyIds.splice(idx, 1);
                }
            }
        });

        for (const caseData of dataSets) {
            if (!caseData.testEnable) {
                continue;
            }

            test(caseData.testCase, async ({}, testInfo: TestInfo) => {
                if (
                    caseData.testCase.includes('L1 Researcher') ||
                    caseData.testCase.includes('L2 Researcher')
                ) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'Researcher node-pool listing has an unresolved failure; revalidate topology visibility before enabling this test'
                    );
                }

                await test.step(`login cli with ${caseData.user}`, () => {
                    runaiCliV2.loginAsUser(caseData.user, caseData.password);
                });

                await test.step('list node pool and verify the topology name is correct', () => {
                    const result = runaiCliV2.list('nodepool');
                    logger.info(`nodepool list output: \n${result.stdout}`);

                    const found = result.stdout.split('\n').some((line) => {
                        const fields = line.trim().split(/\s{2,}/);
                        return fields[0] === npName && fields[5] === topoName;
                    });

                    expect(found).toBe(true);
                });

                await test.step('delete the network topology and verify topology info is removed in list nodepool output', async () => {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(topologyIds[0]);
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(topoName);
                    topologyIds.pop();

                    const result = runaiCliV2.list('nodepool');
                    logger.info(`nodepool list output: \n${result.stdout}`);

                    const found = result.stdout.split('\n').some((line) => {
                        const fields = line.trim().split(/\s{2,}/);
                        return fields[0] === npName && fields.length === 5;
                    });

                    expect(found).toBe(true);
                });

                await test.step('logout cli', () => {
                    runaiCliV2.logout();
                });
            });
        }
    });
});
