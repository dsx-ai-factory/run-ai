/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    clusterUtils,
    flowRootState,
    logger,
    nodeUtils,
    runaiHealthChecker,
    tags,
    waitUtils,
    authUtils
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { Node } from '../../../../../../modules/api/apis/clusterApi';
import { V1Node } from '@kubernetes/client-node';

test.describe(`Run:AI Installation Validation Suite, tags: ${tags.tool.api} ${tags.resource.cluster} ${tags.teams.phoenix} ${tags.runOn.k8sFlavCertKit}${tags.runOn.k8sFlavCertKitTemp} ${tags.services.clusterService}`,
    { tag: [tags.services.clusterService] }, () => {
    let accessToken: string;

    test.beforeAll(async () => {
        // Get admin access token for API calls
        try {
            const adminTokens = await authUtils.authenticateUser(
                process.env.CONTROL_PLANE_ADMIN_USERNAME || 'admin@example.com',
                process.env.CONTROL_PLANE_ADMIN_PASSWORD || 'admin'
            );
            accessToken = adminTokens.idToken;
        } catch (error) {
            logger.error(`Failed to authenticate: ${error}`);
            throw error;
        }
    });

    test.describe('Core Run:AI Installation Validation', () => {
        test('validate node health', async () => {
            // Source of truth for node health is Kubernetes itself: a node is healthy
            // when its k8s "Ready" condition is True. The Run:AI node API derives its
            // status from the cluster-service node-status sync, which lags behind kubelet
            // readiness for a while after a fresh install. We therefore (1) wait for
            // Kubernetes to report all nodes Ready, then (2) wait for the Run:AI API to
            // converge with Kubernetes, both with a realistic budget, and (3) on failure
            // cross-check the two views so the error tells a genuinely-broken node apart
            // from a cluster-service status-sync lag.
            // Two sequential readiness waits below can each take a few minutes after a
            // fresh install, which exceeds the default per-test timeout (150s/240s); give
            // this validation test enough budget to cover both phases plus the API calls.
            const NODE_READY_TIMEOUT_SEC = 180;
            test.setTimeout((NODE_READY_TIMEOUT_SEC * 2 + 60) * 1000);
            const k8sNodeReady = (node: V1Node): boolean =>
                node.status?.conditions?.some((c) => c.type === 'Ready' && c.status === 'True') ?? false;

            const nodesCount = await nodeUtils.getNodesCount();
            expect(nodesCount).toBeGreaterThan(0);

            // 1) Wait for all Kubernetes nodes to be Ready (handles slow kubelet/CNI startup).
            await waitUtils.waitForCondition(
                async () => {
                    const k8sNodes = (await kubeCli.getNodes()).items;
                    const ready = k8sNodes.filter(k8sNodeReady);
                    if (k8sNodes.length === 0 || ready.length !== k8sNodes.length) {
                        const notReady = k8sNodes.filter((n) => !k8sNodeReady(n)).map((n) => n.metadata?.name);
                        logger.info(
                            `[validate node health] Kubernetes: ${ready.length}/${k8sNodes.length} nodes Ready` +
                                (notReady.length ? ` (waiting on ${notReady.join(', ')})` : '')
                        );
                    }
                    return k8sNodes.length > 0 && ready.length === k8sNodes.length;
                },
                'wait for all Kubernetes nodes to be Ready',
                true,
                NODE_READY_TIMEOUT_SEC
            );

            // 2) Wait for the Run:AI cluster-service node API to converge with Kubernetes:
            //    it must report the *same* node set (cluster-service can lag by omitting a
            //    node entirely, not just by marking it NotReady) and report every node Ready.
            //    Don't abort on timeout here: swallow it and fall through to the cross-check
            //    diagnostics + assertion below, so a failure is reported with actionable
            //    Kubernetes-vs-Run:AI attribution rather than a bare "timed out" message.
            try {
                await waitUtils.waitForCondition(
                    async () => {
                        const k8sNames = new Set(
                            (await kubeCli.getNodes()).items
                                .map((n) => n.metadata?.name)
                                .filter((name): name is string => Boolean(name))
                        );
                        const res = await nodeUtils.getAllNodes(accessToken);
                        const nodes = res.body.nodes;
                        const ready = nodes.filter((n) => n.status === Node.StatusEnum.Ready);
                        const missingInRunai = [...k8sNames].filter((name) => !nodes.some((n) => n.name === name));
                        if (nodes.length !== k8sNames.size || missingInRunai.length > 0 || ready.length !== nodes.length) {
                            logger.info(
                                `[validate node health] Run:AI API: ${ready.length}/${nodes.length} nodes Ready | k8s=${k8sNames.size}` +
                                    (missingInRunai.length ? ` | missing in Run:AI: ${missingInRunai.join(', ')}` : '') +
                                    ' (waiting for cluster-service sync)'
                            );
                        }
                        return (
                            k8sNames.size > 0 &&
                            nodes.length === k8sNames.size &&
                            missingInRunai.length === 0 &&
                            ready.length === nodes.length
                        );
                    },
                    'wait for Run:AI node API to report all nodes Ready',
                    true,
                    NODE_READY_TIMEOUT_SEC
                );
            } catch (err) {
                logger.warn(
                    `[validate node health] Run:AI node API did not converge within ${NODE_READY_TIMEOUT_SEC}s; running cross-check diagnostics: ${err}`
                );
            }

            // Final validation + diagnostics
            const k8sNodes = (await kubeCli.getNodes()).items;
            const k8sNodeNames = new Set(
                k8sNodes.map((n) => n.metadata?.name).filter((name): name is string => Boolean(name))
            );
            const allNodes = await nodeUtils.getAllNodes(accessToken);
            const nodes = allNodes.body.nodes;
            const readyNodes = nodes.filter((node) => node.status === Node.StatusEnum.Ready);
            const missingInRunai = [...k8sNodeNames].filter((name) => !nodes.some((n) => n.name === name));

            logger.info(
                `[validate node health] Total nodes: ${nodes.length}, Ready: ${readyNodes.length}, NotReady/Other: ${nodes.length - readyNodes.length} | Kubernetes nodes: ${k8sNodeNames.size}`
            );
            for (const node of nodes) {
                const isReady = node.status === Node.StatusEnum.Ready;
                logger.info(
                    `[validate node health] Node "${node.name}" | status=${node.status} | nodePool=${node.nodePool}${!isReady && node.conditions ? ` | conditions=${JSON.stringify(node.conditions)}` : ''}`
                );
            }
            if (missingInRunai.length > 0) {
                logger.error(
                    `[validate node health] nodes present in Kubernetes but missing from the Run:AI node API: ${missingInRunai.join(', ')}`
                );
            }
            if (readyNodes.length !== nodes.length) {
                const notReadyNodes = nodes.filter((node) => node.status !== Node.StatusEnum.Ready);
                for (const n of notReadyNodes) {
                    const k8sNode = k8sNodes.find((kn) => kn.metadata?.name === n.name);
                    if (!k8sNode) {
                        logger.error(
                            `[validate node health] Run:AI reports "${n.name}" as ${n.status} | node not found in the Kubernetes node list (possible name mismatch or stale cluster-service entry)`
                        );
                    } else {
                        const k8sReady = k8sNodeReady(k8sNode);
                        const k8sReadyCondition = k8sNode.status?.conditions?.filter((c) => c.type === 'Ready');
                        logger.error(
                            `[validate node health] Run:AI reports "${n.name}" as ${n.status} | kubernetes Ready=${k8sReady} | ` +
                                (k8sReady
                                    ? 'cluster-service node-status sync lag (node is Ready in Kubernetes)'
                                    : `node is not Ready in Kubernetes: ${JSON.stringify(k8sReadyCondition)}`)
                        );
                    }
                }
            }

            // Run:AI must report the same node set as Kubernetes, all Ready.
            expect(missingInRunai).toEqual([]);
            expect(nodes.length).toBe(k8sNodeNames.size);
            expect(readyNodes.length).toBe(nodes.length);
        });

        test('validate Run:AI namespace and core pods are running', async () => {
            //Check if runai namespace exists and has pods
            const runaiPods = await runaiHealthChecker.getPodsByNamespace('runai');
            expect(runaiPods.length).toBeGreaterThan(0);

            //Validate pod health status in runai namespace
            const isRunaiHealthy = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai');
            expect(isRunaiHealthy).toBe(true);
        });

        test('validate Run:AI operator status and conditions', async ({}, testInfo) => {
            // Check if cluster version supports operator
            if (await runaiHealthChecker.isClusterVersionHaveOperator()) {
                await runaiHealthChecker.waitForRetryBackoff(testInfo.retry);
                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            const res = await kubeCli.CoreV1Api.readNamespacedConfigMap({
                                name: 'runai-public',
                                namespace: 'runai'
                            });
                            const runaiPublic = res.data['runai-public'];
                            const types = runaiPublic?.split('conditions:')[1]?.split('optional:')[0]?.split('- type:');
                            if (!types) {
                                logger.info('[validate Run:AI operator] Waiting for conditions in runai-public');
                                return false;
                            }
                            const deployedCondition = types.find((type) => type.includes('Deployed'));
                            const availableCondition = types.find((type) => type.includes('Available'));
                            const ok =
                                deployedCondition?.includes('Deployed') &&
                                deployedCondition?.includes('status: "True"') &&
                                availableCondition?.includes('Available') &&
                                availableCondition?.includes('status: "True"');
                            if (!ok) {
                                logger.info('[validate Run:AI operator] Waiting for Deployed and Available True');
                            }
                            return Boolean(ok);
                        } catch (err) {
                            logger.warn(
                                `[validate Run:AI operator] Transient error polling runai-public, will retry: ${err}`
                            );
                            return false;
                        }
                    },
                    'wait for Run:AI operator Deployed and Available to be True in runai-public',
                    true,
                    120
                );

                await runaiHealthChecker.validateOperatorStatusConditions();
            } else {
                testInfo.skip();
            }
        });

        test('validate Run:AI services and deployments are configured', async () => {
            // Check Run:AI services
            const runaiServices = await kubeCli.CoreV1Api.listNamespacedService({ namespace: 'runai' });
            expect(runaiServices.items.length).toBeGreaterThan(0);

            //Run:AI deployments
            const runaiDeployments = await kubeCli.AppsV1Api.listNamespacedDeployment({ namespace: 'runai' });
            expect(runaiDeployments.items.length).toBeGreaterThan(0);

            // Validate that at least one deployment is ready
            const readyDeployments = runaiDeployments.items.filter(
                (deployment) => deployment.status?.readyReplicas && deployment.status.readyReplicas > 0
            );
            expect(readyDeployments.length).toBeGreaterThan(0);
        });
    });
});
