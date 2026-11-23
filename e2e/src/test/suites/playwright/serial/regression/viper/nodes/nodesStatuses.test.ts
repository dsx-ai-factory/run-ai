/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    clusterUtils,
    config,
    crds,
    flowRootState,
    logger,
    nodePoolUtils,
    nodeUtils,
    GPU_FRACTIONING_READY_CONDITION,
    GPU_PRODUCT_LABEL,
    NVIDIA_GPU_LABEL,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    TrainingWorkload,
    Utils,
    waitUtils,
    workloadUtils
} from '../../../../../../../modules/helpers';
import { bashCli, kubeCli } from '../../../../../../../modules/cli';
import { test, expect } from '@playwright/test';
import { backendClusterApi } from '../../../../../../../modules/api';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import { Node } from '../../../../../../../modules/api/apis/clusterApi';
import { KubernetesObject } from '@kubernetes/client-node';
import type { StagedNodeCondition } from '../../../../../../../modules/helpers';
import StatusEnum = Node.StatusEnum;
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Nodes statuses test, tags: ${tags.resource.nodes} ${tags.tool.api} ${tags.teams.viper}`, () => {
    let foundNodeName: string;

    test.beforeAll(async () => {
        const nodesListFromCli = await kubeCli.getNodes();
        const allNodeNames = nodesListFromCli.items.map((node) => node.metadata.name);
        foundNodeName = allNodeNames[0];
    });

    test('Should get Ready status from worker node', async () => {
        await nodeUtils.getNodeStatusAndValidate(foundNodeName, StatusEnum.Ready);
    });

    test.describe('Cordon worker node', () => {
        test.beforeAll(async () => {
            // cordon the worker node
            const res = bashCli.exec(`kubectl cordon ${foundNodeName}`);
            expect(res).toBe(`node/${foundNodeName} cordoned\n`);
            await nodeUtils.waitForNodeStatus(foundNodeName, StatusEnum.NotReady);
        });

        test.afterAll(() => {
            const res = bashCli.exec(`kubectl uncordon ${foundNodeName}`);
            expect(res).toBe(`node/${foundNodeName} uncordoned\n`);
        });

        test('Should get NotReady status with SchedulingDisabled condition type', async () => {
            await nodeUtils.getNodeStatusAndValidate(
                foundNodeName,
                StatusEnum.NotReady,
                'SchedulingDisabled',
                'node was marked Unschedulable'
            );
        });
    });

    shouldRunNativeWorkloadCrdTests(() =>
    test.describe('Undrained migrated worker node', () => {
        let nodePoolName: string,
            nodePoolId: number,
            nodeName: string,
            jobName: string,
            namespaceName: string,
            projectName: string,
            project;

        test.beforeAll(async () => {
            // set access token
            backendClusterApi.nodePoolsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;

            // label node and create node-pool
            const key = objectGenerator.generateNameByPrefix('key');
            const value = objectGenerator.generateNameByPrefix('value');
            ({ nodePoolName, nodePoolId, nodeName } =
                await nodePoolUtils.createNodePoolAndLabelNodeToAssignToNodePool(key, value));

            // create project
            logger.debug('Creating project');

            projectName = objectGenerator.generateNameByPrefix('p');
            namespaceName = Utils.namespaceFromProjectName(projectName);

            project = await projectResourcesUtils.createProject(
                namespaceName,
                {
                    name: projectName,
                    maxAllowedGpus: 10,
                    deservedGpus: 10
                },
                flowRootState.users.system_administrator.accessToken
            );

            // submit a job to this node-pool
            jobName = objectGenerator.generateNameByPrefix('job');
            const trainingWorkload = new TrainingWorkload(
                config.quickStartImage.ubuntu,
                namespaceName,
                jobName
            );
            trainingWorkload.command = 'sleep infinity';
            trainingWorkload.nodePools = [nodePoolName].join(' ');
            await workloadUtils.submit(trainingWorkload);

            await waitUtils.waitForBackendWorkloadPhase(trainingWorkload.name, project.id, Phase.Running);

            // remove label from node
            await kubeCli.unLabelNode(nodeName, key);
        });

        test.afterAll(async () => {
            // delete the job
            try {
                await kubeCli.deleteCrd(crds.runaiTrainingWorkloadCrd, jobName, namespaceName);
                await waitUtils.waitForPodsDeletion(jobName, namespaceName);
                // 0 gpu for job
            } catch (error) {
                logger.debug(`Could not delete job: ${jobName} in project: ${projectName}`);
            }

            // delete project
            try {
                await operations.deleteProjectAndNamespace(
                    project.id,
                    flowRootState.users.system_administrator.accessToken
                );
            } catch (error) {
                logger.debug(`Could not delete project: ${projectName}`);
            }

            // delete the created node-pool
            const nodePoolToDelete = new NodePool();
            nodePoolToDelete.name = nodePoolName;
            nodePoolToDelete.id = nodePoolId;
            try {
                await nodePoolUtils.deleteNodePool(nodePoolToDelete);
            } catch (error) {
                logger.debug(`Could not delete node-pool: ${nodePoolName}`);
                throw error;
            }
        });

        test('Should get NotReady status with UndrainedMigratedNode condition type', async () => {
            await nodeUtils.waitForNodeStatus(nodeName, StatusEnum.NotReady);
            await nodeUtils.getNodeStatusAndValidate(
                nodeName,
                StatusEnum.NotReady,
                'UndrainedMigratedNode',
                'node is Unschedualble due to nodepool migration while node is undrained'
            );
        });
        })
    );
});

const GPU_FRACTIONING_CONDITION_TYPE = 'MissingGpuFractioningPrerequisite';
const LABEL_OWNER_ANNOTATION = 'run.ai/e2e-owns-nvidia-gpu-label';
const STAGED_DRIVER_REASON = 'GPUDriverVersionUnsupported';
const STAGED_DRIVER_MESSAGE =
    'CUDA driver major version 595 is below minimum supported major version 615';

test.describe(`GPU fractioning node readiness, tags: ${tags.resource.nodes} ${tags.tool.api} ${tags.teams.viper}`, () => {
    let gpuNodeName: string;
    let labelOwnedByTest = false;
    let originalCondition: StagedNodeCondition | undefined;

    test.beforeAll(async () => {
        // cluster-sync suppresses per-node fractioning conditions whenever the cluster
        // itself does not support fractions, so none of these cases can be observed there.
        const supported = await clusterUtils.isGpuFractioningSupported();
        test.skip(!supported, 'cluster does not report gpu-fractioning support');

        // Selected on the GPU product label rather than reported capacity: a node can
        // lose its allocatable GPU - a crash-looping device plugin is enough - while its
        // fractioning condition stays perfectly reportable, and gating on capacity skips
        // these tests silently instead of running them.
        const gpuNodes = (await kubeCli.getNodes(GPU_PRODUCT_LABEL)).items;
        test.skip(gpuNodes.length === 0, `no nodes carrying ${GPU_PRODUCT_LABEL}`);

        // NotReady outranks MissingPrerequisites, so a cordoned or unhealthy node would
        // fail both assertions for reasons unrelated to fractioning.
        const readyNodes = await nodeUtils.getWorkerNodeNames();
        const candidate = gpuNodes.find((node) => readyNodes.includes(node.metadata.name));
        test.skip(candidate === undefined, 'no GPU node is reported Ready');
        gpuNodeName = candidate.metadata.name;

        // The annotation is written before the label, never after: a run killed between
        // the two then leaves a marker and no label, which the next run cleans up. The
        // reverse order would strand the label with nothing recording who owns it.
        const labels = await kubeCli.getNodeLabels(gpuNodeName);
        const node = await kubeCli.getNode(gpuNodeName);
        const ownedAlready = node.metadata.annotations?.[LABEL_OWNER_ANNOTATION] === 'true';
        if (labels[NVIDIA_GPU_LABEL] !== 'true' || ownedAlready) {
            await kubeCli.patchNode(gpuNodeName, [
                {
                    op: 'add',
                    path: `/metadata/annotations/${LABEL_OWNER_ANNOTATION.replace(/\//g, '~1')}`,
                    value: 'true'
                }
            ]);
            labelOwnedByTest = true;
            await kubeCli.labelNode(gpuNodeName, NVIDIA_GPU_LABEL, 'true');
        }

        // Deleting it instead of restoring it would leave the node reporting NotManaged,
        // which puts it in MissingPrerequisites and drops it from the node selections
        // every later suite relies on.
        originalCondition = await nodeUtils.readGpuFractioningCondition(gpuNodeName);
    });

    test.afterAll(async () => {
        if (!gpuNodeName) {
            return;
        }
        try {
            if (originalCondition) {
                await nodeUtils.stageNodeCondition(
                    gpuNodeName,
                    GPU_FRACTIONING_READY_CONDITION,
                    originalCondition
                );
            } else {
                await nodeUtils.clearGpuFractioningCondition(gpuNodeName);
            }
            await nodeUtils.waitForNodeStatus(gpuNodeName, StatusEnum.Ready);
        } catch (error) {
            // Not swallowed: a node left outside Ready drops out of the node selections
            // every later suite depends on, so this has to be loud.
            logger.error(`Could not restore the condition on ${gpuNodeName}: ${error}`);
            throw error;
        }
        if (labelOwnedByTest) {
            try {
                await kubeCli.unLabelNode(gpuNodeName, NVIDIA_GPU_LABEL);
                await kubeCli.patchNode(gpuNodeName, [
                    {
                        op: 'remove',
                        path: `/metadata/annotations/${LABEL_OWNER_ANNOTATION.replace(/\//g, '~1')}`
                    }
                ]);
            } catch (error) {
                logger.debug(`Could not remove ${NVIDIA_GPU_LABEL} from ${gpuNodeName}: ${error}`);
            }
        }
    });

    // On a real GPU cluster the kai-gpu-fractioning operator owns this condition and only
    // fake-gpu-operator has a reason-based exemption, so sustained pod churn on the node
    // can revert the staged value mid-test.
    test.describe('Node reporting a fractioning failure', () => {
        test.beforeAll(async () => {
            test.setTimeout(300_000);
            await nodeUtils.stageGpuFractioningCondition(gpuNodeName, {
                status: 'False',
                reason: STAGED_DRIVER_REASON,
                message: STAGED_DRIVER_MESSAGE
            });
            await nodeUtils.waitForNodeCondition(
                gpuNodeName,
                GPU_FRACTIONING_CONDITION_TYPE,
                true,
                STAGED_DRIVER_REASON
            );
        });

        test('Should surface the operator reason and message verbatim', async () => {
            await nodeUtils.getNodeStatusAndValidate(
                gpuNodeName,
                StatusEnum.MissingPrerequisites,
                GPU_FRACTIONING_CONDITION_TYPE,
                STAGED_DRIVER_MESSAGE,
                STAGED_DRIVER_REASON
            );
        });
    });

    test.describe('Healthy node', () => {
        test.beforeAll(async () => {
            test.setTimeout(300_000);
            await nodeUtils.stageGpuFractioningCondition(gpuNodeName, {
                status: 'True',
                reason: 'AllDaemonsReady',
                message: 'all gpu-fractioning daemons are running'
            });
            await nodeUtils.waitForNodeCondition(gpuNodeName, GPU_FRACTIONING_CONDITION_TYPE, false);
        });

        test('Should report Ready and emit no fractioning condition', async () => {
            await nodeUtils.getNodeStatusAndValidate(gpuNodeName, StatusEnum.Ready);
            await nodeUtils.expectNoNodeCondition(gpuNodeName, GPU_FRACTIONING_CONDITION_TYPE);
        });
    });
});

test.describe(`GPU fractioning cluster readiness, tags: ${tags.resource.nodes} ${tags.tool.api} ${tags.teams.viper}`, () => {
    test('Should select the CUDA-native GPU sharing mode in the KAI config', async () => {
        const kaiConfig = await kubeCli.getCrd<KubernetesObject & {
            spec?: { global?: { gpuSharingMode?: string } };
        }>(crds.kaiConfigCrd, 'kai-config');
        expect(kaiConfig.spec?.global?.gpuSharingMode).toBe('NvFractions');
    });

    // Asserted on the component rather than on gpu-stack itself: gpu-stack ANDs the
    // nvidia and gpu-fractioning sub-checks, so a broken DCGM exporter would fail a test
    // about fractioning, and a broken fractioning chain would skip it instead of failing.
    test('Should report the gpu-fractioning component as available', async () => {
        const gpuStack = await clusterUtils.getRequiredDependency('gpu-stack');
        const fractioning = gpuStack?.components?.['gpu-fractioning'];
        expect(fractioning, 'gpu-stack should carry a gpu-fractioning component').toBeDefined();
        expect(fractioning.available).toBe(true);
    });
});
