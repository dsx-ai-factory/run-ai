/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BaseWorkload,
    crds,
    CustomResourceDefinition,
    departmentUtils,
    externalWorkloadUtils,
    flowRootState,
    logger,
    networkTopologiesUtils,
    nodePoolUtils,
    objectGenerator,
    operations,
    orgUnitUtils,
    tags,
    TopologyAnnotationName,
    TopologyRequiredAnnotationName,
    Utils,
    waitUtils,
    workloadsManagerUtils,
    workloadsUtilsV2
} from '../../../../../../../../../modules/helpers';
import {
    DynamoComponentConfig,
    dynamoHealthCommand,
    DynamoSystemPort,
    NodepoolConfig,
    ServiceConfig
} from '../../../../../../../../../modules/helpers/models/externalWorkload';
import { kubeCli } from '../../../../../../../../../modules/cli';
import { Project } from '../../../../../../../../../modules/api/apis/orgUnitApi';
import { test, TestInfo } from '@playwright/test';
import {
    isDynamoInstalled,
    isDynamoV1Alpha1Served,
    isDynamoV1Beta1Served
} from '../../../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { getNodePoolCrdField } from '../../../../../../../../../modules/helpers/utils/NodePoolCrdFields';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { expect } from 'chai';
import { WorkloadUtils } from '../../../../../../../../../modules/helpers/utils/WorkloadUtils';
import { WorkloadTypeConfig } from '../../../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadTypeConfig';
import { WorkloadTypePreemptibility } from '../../../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadTypePreemptibility';
import {
    NetworkTopologyCreateRequest,
    NodepoolUpdateFields
} from '../../../../../../../../../modules/api/apis/clusterApi';
import { DEFAULT_NODE_POOL_NAME } from '../../../../../../../../../modules/ui/utils/nodePoolUtils';
import {
    PodGroupCrd,
    PodGroupTopologyConstraint
} from '../../../../../../../../../modules/helpers/models/podGroups';
import { networkTopologiesApi } from '../../../../../../../../../modules/api';
import { WorkloadElement } from '../../../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadElement';
import { Scope } from '../../../../../../../../../modules/helpers/models/customResourceDefinition';
import {
    isDynamoTopologyBindingInSync,
    resolveDynamoMockRuntimeVersion
} from '../../../../../../../../../modules/helpers/utils/DynamoFixtureUtils';

// Test case configuration interface
export interface TestCase {
    name: string;
    description: string;
    updatePriorityName?: string;
    updatePreemptibility?: WorkloadTypePreemptibility;
    expectedPreemptibility: WorkloadTypePreemptibility;
    expectedPriorityName: string;
    updateTopologyLevels?: string[];
    expectedTopologyConstraint?: PodGroupTopologyConstraint;
    // Expected ExternalWorkload.Spec.Topology as surfaced on the workload API response.
    expectedEWTopology?: { name?: string; preferredPlacement?: string; requiredPlacement?: string };
    addNodepools?: NodepoolConfig[];
    nodepools?: string[];
    dgdServices?: { [serviceName: string]: ServiceConfig[] };
    setAnnotations?: { [key: string]: string };
    setLabels?: { [key: string]: string };
}

test.describe(`DynamoGraphDeployment external workload API tests, tags: ${tags.teams.viper} ${tags.resource.workloads}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let project: Project, namespace: string;
        let originalWorkloadType: WorkloadTypeConfig;
        const workloadUtils = new WorkloadUtils();
        const DGDWorkloadTypeName = 'DynamoGraphDeployment';
        let alphaRuntimeVersionOverride: string | undefined;
        let betaRuntimeVersionOverride: string | undefined;
        const topologyLevel1 = 'cloud.test-provider.com/topology-rack';
        const topologyLevel2 = 'cloud.test-provider.com/hostname';

        let dgdWorkloadType: WorkloadTypeConfig;
        let defaultNodePoolId: string;
        const createdTopologies: string[] = [];
        const labeledNodesForCleanup: Array<{ nodeName: string; labelKeys: string[] }> = [];
        const createdNodePools: string[] = [];
        const labelKey = objectGenerator.generateNameByPrefix('np');
        const labelValue = objectGenerator.generateNameByPrefix('np');
        const topologyName = objectGenerator.generateNameByPrefix('dgd-topo');
        const anotherNodepoolName = objectGenerator.generateNameByPrefix('n-pool');

        // Wait for the nodepool CRD to reflect the expected topology so submitting a workload
        // does not race the reconciler writing spec.networkTopologyName.
        const waitForNodePoolTopology = async (
            nodePoolName: string,
            expectedTopologyName: string
        ): Promise<void> => {
            await waitUtils.waitForCondition(
                async () => {
                    const crd = await nodePoolUtils.getNodePoolCrd(nodePoolName);
                    return (
                        getNodePoolCrdField<string>(crd, 'spec.networkTopologyName') === expectedTopologyName
                    );
                },
                `nodepool ${nodePoolName} CRD network topology name to be ${expectedTopologyName}`,
                true,
                60,
                5
            );
        };

        test.beforeAll(async () => {
            if (!(await isDynamoInstalled())) {
                logger.warn('Dynamo is not installed on the cluster, skipping test');
                test.skip();
            }

            const dgdDefinition = await kubeCli.ApiextensionsV1Api.readCustomResourceDefinition({
                name: 'dynamographdeployments.nvidia.com'
            });
            const deployments = await kubeCli.AppsV1Api.listDeploymentForAllNamespaces();
            const operatorImages = deployments.items.flatMap((deployment) =>
                deployment.spec.template.spec.containers
                    .filter(
                        (container) =>
                            container.image.includes('/ai-dynamo/kubernetes-operator:') ||
                            (deployment.metadata.labels?.['app.kubernetes.io/name'] === 'dynamo-operator' &&
                                container.name === 'manager')
                    )
                    .map((container) => container.image)
            );
            alphaRuntimeVersionOverride = resolveDynamoMockRuntimeVersion(
                dgdDefinition,
                'v1alpha1',
                operatorImages
            );
            betaRuntimeVersionOverride = resolveDynamoMockRuntimeVersion(
                dgdDefinition,
                'v1beta1',
                operatorImages
            );

            originalWorkloadType = await workloadUtils.getWorkloadTypeConfigBy(DGDWorkloadTypeName);
            await workloadUtils.updateWorkloadTypeConfiguration(
                DGDWorkloadTypeName,
                'very-high',
                'Deploy',
                WorkloadTypePreemptibility.NonPreemptible
            );
            workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
            defaultNodePoolId = `${defaultNodePool.id}`;
            const projectName = objectGenerator.generateNameByPrefix('dp');
            namespace = Utils.namespaceFromProjectName(projectName);

            const defaultDepartment = await departmentUtils.getDefaultDepartmentByCluster();

            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                parentId: String(defaultDepartment.id),
                enforceRunaiScheduler: true,
                resources: [
                    {
                        nodePool: { id: `${defaultNodePool.id}`, name: defaultNodePool.name },
                        gpu: { deserved: 8, limit: 8 }
                    }
                ]
            });

            if (!project.id) {
                throw new Error('Project ID is required for the test');
            }

            logger.info(
                `Created project: ${project.name}, namespace: ${namespace} with enforceRunaiScheduler=true`
            );
        });

        test.afterAll(async () => {
            if (!(await isDynamoInstalled())) {
                logger.warn('Dynamo is not installed on the cluster, return');
                return;
            }
            // Cleanup: unlabel nodes
            for (const nodeInfo of labeledNodesForCleanup) {
                for (const labelKey of nodeInfo.labelKeys) {
                    try {
                        await kubeCli.unLabelNode(nodeInfo.nodeName, labelKey);
                        logger.info(`Unlabeled node ${nodeInfo.nodeName} with key: ${labelKey}`);
                    } catch (error) {
                        logger.error(
                            `Failed to unlabel node ${nodeInfo.nodeName} with key ${labelKey}: ${error}`
                        );
                    }
                }
            }

            // Cleanup topologies
            for (const topoId of createdTopologies) {
                try {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(topoId);
                    logger.info(`Deleted topology: ${topoId}`);
                } catch (error) {
                    logger.error(`Failed to delete topology ${topoId}: ${error}`);
                }
            }

            // Cleanup created nodepools
            for (const nodePoolName of createdNodePools) {
                try {
                    const allNodePools = await nodePoolUtils.getAllBackendNodePoolsAndValidate();
                    const nodePool = allNodePools.find((np) => np.name === nodePoolName);
                    if (nodePool) {
                        await nodePoolUtils.deleteNodePoolByIdWithWait(nodePool.id, nodePoolName);
                        logger.info(`Deleted nodepool: ${nodePoolName}`);
                    } else {
                        logger.warn(`Nodepool ${nodePoolName} not found for cleanup`);
                    }
                } catch (error) {
                    logger.error(`Failed to delete nodepool ${nodePoolName}: ${error}`);
                }
            }

            if (project?.id) {
                await operations.deleteProjectAndNamespace(+project.id);
            }

            if (originalWorkloadType?.id) {
                await workloadUtils.updateWorkloadTypeById(originalWorkloadType.id, {
                    categoryId: originalWorkloadType.categoryId,
                    priorityId: originalWorkloadType.priorityId,
                    preemptibility: originalWorkloadType.preemptibility
                });
            }
        });

        const testCases: TestCase[] = [
            {
                name: 'default-priority',
                description: 'DGD with default priority and preemptibility',
                expectedPreemptibility: WorkloadTypePreemptibility.NonPreemptible,
                expectedPriorityName: 'very-high'
            },
            {
                name: 'with not default nodepool',
                description: 'DGD Non default nodepool',
                expectedPreemptibility: WorkloadTypePreemptibility.NonPreemptible,
                expectedPriorityName: 'very-high',
                addNodepools: [
                    {
                        name: anotherNodepoolName,
                        labelKey: labelKey,
                        labelValue: labelValue,
                        setGpuNode: true
                    }
                ],
                nodepools: [anotherNodepoolName]
            },
            {
                name: 'with several nodepools',
                description:
                    'DGD submitted with several nodepools - default nodepool without required resources',
                expectedPreemptibility: WorkloadTypePreemptibility.NonPreemptible,
                expectedPriorityName: 'very-high',
                addNodepools: [{ name: anotherNodepoolName, labelKey: labelKey, labelValue: labelValue }],
                nodepools: [DEFAULT_NODE_POOL_NAME, anotherNodepoolName]
            },
            {
                name: 'with several nodepools, not default first',
                description: 'DGD submitted with several nodepools',
                expectedPreemptibility: WorkloadTypePreemptibility.NonPreemptible,
                expectedPriorityName: 'very-high',
                addNodepools: [{ name: anotherNodepoolName, labelKey: labelKey, labelValue: labelValue }],
                nodepools: [anotherNodepoolName, DEFAULT_NODE_POOL_NAME]
            },
            {
                name: 'updated-priority',
                description: 'DGD with low priority and preemptible',
                updatePriorityName: 'low',
                updatePreemptibility: WorkloadTypePreemptibility.Preemptible,
                expectedPreemptibility: WorkloadTypePreemptibility.Preemptible,
                expectedPriorityName: 'low'
            },
            {
                name: 'with-topology',
                description: 'DGD with network topology on default nodepool',
                updatePriorityName: 'low',
                updatePreemptibility: WorkloadTypePreemptibility.Preemptible,
                expectedPreemptibility: WorkloadTypePreemptibility.Preemptible,
                expectedPriorityName: 'low',
                expectedTopologyConstraint: {
                    topology: topologyName,
                    preferredTopologyLevel: topologyLevel2
                },
                updateTopologyLevels: [topologyLevel1, topologyLevel2]
            },
            {
                name: 'with-topology-set-in-workload',
                description: 'DGD with network topology set on workload',
                updatePriorityName: 'low',
                updatePreemptibility: WorkloadTypePreemptibility.Preemptible,
                expectedPreemptibility: WorkloadTypePreemptibility.Preemptible,
                expectedPriorityName: 'low',
                expectedTopologyConstraint: {
                    topology: topologyName,
                    requiredTopologyLevel: topologyLevel2
                },
                // Topology is set via KAI annotations on the DGD object metadata, which the EWI
                // surfaces to ExternalWorkload.Spec.Topology through the fillTopology fallback.
                expectedEWTopology: {
                    name: topologyName,
                    requiredPlacement: topologyLevel2
                },
                updateTopologyLevels: [topologyLevel1, topologyLevel2],
                setAnnotations: {
                    [TopologyAnnotationName]: topologyName,
                    [TopologyRequiredAnnotationName]: topologyLevel2
                }
            }
        ];

        test.describe('DynamoGraphDeployment external workloads', () => {
            const workloads: BaseWorkload[] = [];

            test.beforeAll(async () => {
                // These cases submit v1alpha1 DGDs, which Dynamo stops serving at some future release.
                if (!(await isDynamoV1Alpha1Served())) {
                    logger.warn('Dynamo does not serve v1alpha1 on the cluster, skipping tests');
                    test.skip();
                }
            });

            test.afterEach(async () => {
                while (workloads.length > 0) {
                    const workload = workloads.pop();
                    if (workload) {
                        try {
                            await kubeCli.deleteCrd(crds.dynamoGraphDeploymentCrd, workload.name, namespace);
                            await kubeCli.waitForCrdToBeDeleted(
                                crds.dynamoGraphDeploymentCrd,
                                workload.name,
                                namespace,
                                10
                            );
                            logger.info(`Cleaned up DGD: ${workload.name}`);
                        } catch (error) {
                            logger.error(`Failed to cleanup DGD ${workload.name}: ${error}`);
                        }
                    }
                }
            });

            for (const testCase of testCases) {
                test(`Should validate DynamoGraphDeployment: ${testCase.description}`, async ({}, testInfo: TestInfo) => {
                    test.setTimeout(7 * 60 * 1000);
                    logger.info(`\n=== Running test case: ${testCase.name} - ${testCase.description} ===\n`);

                    const dgdName = objectGenerator.generateNameByPrefix('dw');
                    const replicas = 2;
                    const host = 'frontend';
                    let workloadId: string;

                    if (testCase.updateTopologyLevels && testCase.updateTopologyLevels.length > 0) {
                        await test.step(`[${testCase.name}] Ensure topology exists and nodepool assignment matches the test case`, async () => {
                            // Check if topology already exists in backend (may have been deleted by an early afterAll)
                            networkTopologiesApi.accessToken =
                                flowRootState.users.system_administrator.accessToken;
                            const topologiesResponse =
                                await networkTopologiesUtils.getNetworkTopologiesViaApi();
                            const existingTopology = topologiesResponse.body?.networkTopologies?.find(
                                (t: any) => t.name === topologyName
                            );

                            let topologyId: string;

                            if (!existingTopology) {
                                const createTopoRequest: NetworkTopologyCreateRequest = {
                                    name: topologyName,
                                    clusterId: flowRootState.cluster.uuid,
                                    levels: testCase.updateTopologyLevels
                                };

                                const topoResponse = await networkTopologiesUtils.createNetworkTopologyViaApi(
                                    createTopoRequest,
                                    flowRootState.users.system_administrator.accessToken
                                );
                                expect(topoResponse.response.statusCode).to.equal(201);
                                topologyId = topoResponse.body.id;
                                createdTopologies.push(topologyId);
                                logger.info(
                                    `Created topology: ${topologyName} with ID: ${topologyId}, levels: ${testCase.updateTopologyLevels.join(', ')}`
                                );

                                const labeledNodesInfo =
                                    await nodePoolUtils.labelNodePoolNodesWithTopologyLevels(
                                        defaultNodePoolId,
                                        testCase.updateTopologyLevels,
                                        flowRootState.users.system_administrator.accessToken
                                    );

                                for (const nodeInfo of labeledNodesInfo) {
                                    labeledNodesForCleanup.push({
                                        nodeName: nodeInfo.nodeName,
                                        labelKeys: nodeInfo.labels.map((l) => l.key)
                                    });
                                }

                                logger.info(
                                    `Labeled ${labeledNodesInfo.length} nodes with ${testCase.updateTopologyLevels.length} topology levels each`
                                );
                            } else {
                                topologyId = existingTopology.id;
                                if (!createdTopologies.includes(topologyId)) {
                                    createdTopologies.push(topologyId);
                                }
                                logger.info(
                                    `Topology ${topologyName} (ID: ${topologyId}) already exists in backend`
                                );
                            }

                            await networkTopologiesUtils.getTopologyCRDByName(topologyName, 60);
                            logger.info(`Topology CRD ${topologyName} exists in cluster`);

                            const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();

                            // Assign to nodepool only if not already assigned
                            if (defaultNodePool['networkTopologyId'] !== topologyId) {
                                await nodePoolUtils.updateClusterNodePoolAndValidate(defaultNodePoolId, {
                                    networkTopologyId: topologyId
                                });
                                // Wait for the nodepool CRD to reflect the assignment before submitting.
                                await waitForNodePoolTopology(defaultNodePool.name, topologyName);
                                logger.info(
                                    `Assigned topology ${topologyName} to default nodepool ${DEFAULT_NODE_POOL_NAME}`
                                );
                            } else {
                                logger.info(
                                    `Topology ${topologyName} already assigned to default nodepool ${DEFAULT_NODE_POOL_NAME}`
                                );
                            }
                        });
                    }

                    if (testCase.addNodepools && testCase.addNodepools.length > 0) {
                        await test.step(`[${testCase.name}] Create and configure nodepools`, async () => {
                            const { nodepoolsNames, nodeLabels } =
                                await nodePoolUtils.createNodepoolsAndValidate(testCase.addNodepools);
                            createdNodePools.push(...nodepoolsNames);
                            labeledNodesForCleanup.push(...nodeLabels);
                        });
                    }

                    if (testCase.updatePriorityName || testCase.updatePreemptibility) {
                        await test.step(`[${testCase.name}] Update workload type for test case`, async () => {
                            dgdWorkloadType = await workloadUtils.updateWorkloadTypeConfiguration(
                                DGDWorkloadTypeName,
                                testCase.updatePriorityName || originalWorkloadType.priorityName,
                                originalWorkloadType.categoryName,
                                testCase.updatePreemptibility || originalWorkloadType.preemptibility
                            );
                            logger.info(
                                `Workload type updated: priority=${dgdWorkloadType.priorityName}, preemptibility=${dgdWorkloadType.preemptibility}`
                            );
                        });
                    } else {
                        await test.step(`[${testCase.name}] Get DGD Workload type`, async () => {
                            dgdWorkloadType =
                                await workloadUtils.getWorkloadTypeConfigBy(DGDWorkloadTypeName);
                            expect(dgdWorkloadType).to.not.be.undefined;

                            if (!testCase.expectedPriorityName) {
                                testCase.expectedPriorityName = dgdWorkloadType.priorityName;
                            }

                            logger.info(
                                `Using default workload type: priority=${dgdWorkloadType.priorityName}, preemptibility=${dgdWorkloadType.preemptibility}`
                            );
                        });
                    }

                    let nodepoolsWithLabels: NodepoolConfig[] | undefined;
                    if (testCase.nodepools && testCase.nodepools.length > 0) {
                        await test.step(`[${testCase.name}] Prepare nodepool information`, async () => {
                            nodepoolsWithLabels = await nodePoolUtils.prepareNodepoolInformation(
                                testCase.nodepools
                            );
                        });
                    }

                    await test.step(`[${testCase.name}] Submit DGD workload and wait for successful state`, async () => {
                        workloadId = await workloadsManagerUtils.submitAndWaitForDGD(
                            dgdName,
                            namespace,
                            project.name,
                            String(project.id),
                            replicas,
                            host,
                            workloads,
                            nodepoolsWithLabels,
                            testCase.dgdServices,
                            testCase.setAnnotations,
                            testCase.setLabels,
                            alphaRuntimeVersionOverride
                        );
                    });

                    await test.step(`[${testCase.name}] Validate workload API response`, async () => {
                        await workloadsUtilsV2.getWorkloadAndValidate(workloadId, {
                            expectedType: 'DynamoGraphDeployment',
                            expectedSource: 'Control-plane',
                            expectedRequestedPods: replicas
                        });
                        logger.info(`Successfully validated DGD workload API response for ${dgdName}`);
                    });

                    if (testCase.expectedEWTopology) {
                        await test.step(`[${testCase.name}] Validate ExternalWorkload.Spec.Topology on workload API`, async () => {
                            const topology = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body
                                .topology;
                            expect(topology, 'workload topology should exist').to.exist;
                            if (testCase.expectedEWTopology.name !== undefined) {
                                expect(topology.name).to.equal(
                                    testCase.expectedEWTopology.name,
                                    'topology name'
                                );
                            }
                            if (testCase.expectedEWTopology.preferredPlacement !== undefined) {
                                expect(topology.preferredPlacement).to.equal(
                                    testCase.expectedEWTopology.preferredPlacement,
                                    'topology preferred placement'
                                );
                            }
                            if (testCase.expectedEWTopology.requiredPlacement !== undefined) {
                                expect(topology.requiredPlacement).to.equal(
                                    testCase.expectedEWTopology.requiredPlacement,
                                    'topology required placement'
                                );
                            }
                            logger.info(
                                `Validated ExternalWorkload.Spec.Topology for ${dgdName}: ${JSON.stringify(topology)}`
                            );
                        });
                    }

                    await test.step(`[${testCase.name}] Verify ExternalWorkload CRD is Running with correct priority/preemptibility`, async () => {
                        await externalWorkloadUtils.validateExternalWorkload(
                            dgdName,
                            namespace,
                            testCase.expectedPriorityName,
                            testCase.expectedPreemptibility
                        );
                    });

                    await test.step(`[${testCase.name}] Verify pod count matches number of replicas`, async () => {
                        const podsResponse = await workloadsUtilsV2.getWorkloadPods(workloadId);
                        const pods = podsResponse.body.pods;
                        logger.info(`DGD ${dgdName} has ${pods.length} pods`);
                        expect(pods.length).to.equal(replicas, `Expected ${replicas} pods`);
                    });

                    await test.step(`[${testCase.name}] Verify PodGroup exists and has correct values`, async () => {
                        const podCrds = await kubeCli.listPodsByNamespaceAndPrefix(namespace, dgdName);
                        expect(podCrds.length).to.be.greaterThan(0, 'At least one pod should exist');

                        const podGroupName = podCrds[0].metadata.annotations['pod-group-name'];
                        expect(podGroupName).to.exist;

                        const podGroup = await kubeCli.getCrd<PodGroupCrd>(
                            crds.podGroupsCrd,
                            podGroupName,
                            namespace
                        );
                        expect(podGroup).to.exist;
                        logger.info(`PodGroup ${podGroupName} exists`);

                        let queueName = project.name;
                        if (
                            testCase.nodepools &&
                            testCase.nodepools.length > 0 &&
                            testCase.nodepools[0] !== DEFAULT_NODE_POOL_NAME
                        ) {
                            queueName = `${queueName}-${testCase.nodepools[0]}`;
                        }
                        expect(podGroup.spec.queue).to.equal(queueName);
                        expect(podGroup.metadata.labels['runai/queue']).to.exist;
                        expect(podGroup.metadata.labels['runai/queue']).to.equal(queueName);
                        // Verify topology constraints if topology was configured
                        if (testCase.expectedTopologyConstraint) {
                            networkTopologiesUtils.assertTopologyConstraint(
                                podGroup.spec.topologyConstraint,
                                testCase.expectedTopologyConstraint
                            );
                        }

                        expect(podGroup.spec.priorityClassName).to.equal(dgdWorkloadType.priorityName);
                        expect(podGroup.spec.preemptibility).to.equal(dgdWorkloadType.preemptibility);
                    });

                    logger.info(`Test case passed: ${testCase.name}\n`);
                });
            }
        });

        // Per-component topology constraints are only exposed on the elements API from 3.0,
        // and only Dynamo 1.2.0+ serves the v1beta1 shape these constraints are declared in.
        supportedFrom(SupportedVersions.V3_0)(function () {
            test.describe('DynamoGraphDeployment per-component topology', () => {
                const workloads: BaseWorkload[] = [];
                const componentTopologyName = objectGenerator.generateNameByPrefix('dgd-ctopo');
                // packDomain has to be a bare token (^[a-z0-9]([a-z0-9-]*[a-z0-9])?$), so the
                // topology levels double as the pack domains.
                const deploymentPackDomain = 'zone';
                const componentPackDomain = 'rack';
                const inheritingComponent = 'Frontend';
                const overridingComponent = 'VllmDecodeWorker';
                let elements: WorkloadElement[];
                let workloadId: string;
                let bindingUid: string | undefined;
                const fixtureLabelValue = componentTopologyName;
                const topologyBindingCrd: CustomResourceDefinition = {
                    group: 'grove.io',
                    version: 'v1alpha1',
                    plural: 'clustertopologybindings',
                    kind: 'ClusterTopologyBinding',
                    scope: Scope.CLUSTER
                };
                const nodeLabelOriginals: Array<{
                    name: string;
                    uid: string;
                    values: Record<string, string | undefined>;
                }> = [];

                const elementsOfComponent = (instanceKey: string): WorkloadElement[] =>
                    elements.filter((element) => element.instanceKey === instanceKey);

                test.beforeAll(async () => {
                    if (!(await isDynamoV1Beta1Served())) {
                        logger.warn('Dynamo v1beta1 is not served on the cluster, skipping test');
                        test.skip();
                    }
                    test.setTimeout(10 * 60 * 1000);

                    networkTopologiesApi.accessToken = flowRootState.users.system_administrator.accessToken;
                    const topologyResponse = await networkTopologiesUtils.createNetworkTopologyViaApi({
                        name: componentTopologyName,
                        clusterId: flowRootState.cluster.uuid,
                        levels: [deploymentPackDomain, componentPackDomain]
                    });
                    expect(topologyResponse.response.statusCode).to.equal(201);
                    createdTopologies.push(topologyResponse.body.id);
                    logger.info(`Created topology ${componentTopologyName} for per-component constraints`);

                    const topology = await networkTopologiesUtils.getTopologyCRDByName(componentTopologyName);
                    // Modern Dynamo references a Grove binding; the API-created KAI topology alone
                    // is insufficient. Older operators keep their original topology contract.
                    if (betaRuntimeVersionOverride !== undefined) {
                        const binding = await kubeCli.createCrd(topologyBindingCrd, {
                            apiVersion: 'grove.io/v1alpha1',
                            kind: 'ClusterTopologyBinding',
                            metadata: { name: componentTopologyName },
                            spec: {
                                levels: [deploymentPackDomain, componentPackDomain].map((domain) => ({
                                    domain,
                                    key: domain
                                })),
                                schedulerTopologyBindings: [
                                    {
                                        schedulerName: 'kai-scheduler',
                                        topologyReference: topology.metadata.name
                                    }
                                ]
                            }
                        });
                        bindingUid = binding.metadata.uid;
                        await waitUtils.waitForCondition(
                            async () => {
                                const currentBinding = await kubeCli.getCrd(
                                    topologyBindingCrd,
                                    componentTopologyName
                                );
                                if (currentBinding.metadata.uid !== bindingUid) {
                                    throw new Error('The fixture topology binding was replaced');
                                }
                                const currentTopology = await kubeCli.getCrd(
                                    crds.topologyCrd,
                                    componentTopologyName
                                );
                                return isDynamoTopologyBindingInSync(currentBinding, currentTopology);
                            },
                            'Grove binding and KAI topology to be in sync',
                            true,
                            120,
                            5
                        );
                    }

                    // A topology constraint needs matching node labels to schedule. Preserve and
                    // restore previous values, rather than removing possibly pre-existing labels.
                    for (const node of (await kubeCli.getNodes()).items.filter((entry) =>
                        entry.status.conditions.some(
                            (condition) => condition.type === 'Ready' && condition.status === 'True'
                        )
                    )) {
                        const domains = [deploymentPackDomain, componentPackDomain];
                        nodeLabelOriginals.push({
                            name: node.metadata.name,
                            uid: node.metadata.uid,
                            values: Object.fromEntries(
                                domains.map((domain) => [domain, node.metadata.labels?.[domain]])
                            )
                        });
                        await kubeCli.patchNode(node.metadata.name, [
                            { op: 'test', path: '/metadata/uid', value: node.metadata.uid },
                            {
                                op: 'test',
                                path: '/metadata/resourceVersion',
                                value: node.metadata.resourceVersion
                            },
                            ...domains.map((domain) => ({
                                op: 'add',
                                path: `/metadata/labels/${domain}`,
                                value: fixtureLabelValue
                            }))
                        ]);
                    }

                    // The frontend is probed on 8000 and every other component on the system port,
                    // so each container serves health on the port its own probes use.
                    const components: DynamoComponentConfig[] = [
                        {
                            name: inheritingComponent,
                            type: 'frontend',
                            runtimeVersionOverride: betaRuntimeVersionOverride,
                            resources: { requests: { cpu: '100m' } }
                        },
                        {
                            name: overridingComponent,
                            type: 'worker',
                            runtimeVersionOverride: betaRuntimeVersionOverride,
                            command: dynamoHealthCommand(DynamoSystemPort),
                            packDomain: componentPackDomain,
                            resources: { requests: { cpu: '100m' } }
                        }
                    ];

                    workloadId = await workloadsManagerUtils.submitAndWaitForDGDV1Beta1({
                        dgdName: objectGenerator.generateNameByPrefix('dw'),
                        namespace,
                        projectId: String(project.id),
                        workloads,
                        components,
                        topologyConstraint: {
                            clusterTopologyName: componentTopologyName,
                            packDomain: deploymentPackDomain
                        }
                    });

                    await waitUtils.waitForCondition(
                        () =>
                            workloadsUtilsV2.elementsHaveTopology(workloadId, [
                                inheritingComponent,
                                overridingComponent
                            ]),
                        `elements of ${workloadId} to report their topology constraints`,
                        true,
                        120,
                        5
                    );
                    elements = (await workloadsUtilsV2.getWorkloadElements(workloadId)).body.elements;
                });

                test.afterAll(async () => {
                    const cleanupErrors: unknown[] = [];
                    while (workloads.length > 0) {
                        const workload = workloads.pop();
                        try {
                            await kubeCli.deleteCrd(
                                crds.dynamoGraphDeploymentV1Beta1Crd,
                                workload.name,
                                namespace
                            );
                            // The DGD carries a finalizer, so deletion regularly needs more
                            // than the 10 seconds the other suites wait.
                            await kubeCli.waitForCrdToBeDeleted(
                                crds.dynamoGraphDeploymentV1Beta1Crd,
                                workload.name,
                                namespace,
                                60
                            );
                        } catch (error) {
                            logger.error(`Failed to cleanup DGD ${workload.name}: ${error}`);
                            cleanupErrors.push(error);
                        }
                    }
                    if (bindingUid) {
                        try {
                            await kubeCli.CustomObjectsApi.deleteClusterCustomObject({
                                group: topologyBindingCrd.group,
                                version: topologyBindingCrd.version,
                                plural: topologyBindingCrd.plural,
                                name: componentTopologyName,
                                body: { preconditions: { uid: bindingUid } }
                            });
                            await kubeCli.waitForCrdToBeDeleted(
                                topologyBindingCrd,
                                componentTopologyName,
                                undefined,
                                60
                            );
                        } catch (error) {
                            cleanupErrors.push(error);
                        }
                    }
                    for (const original of nodeLabelOriginals) {
                        try {
                            const current = await kubeCli.CoreV1Api.readNode({ name: original.name });
                            if (current.metadata.uid !== original.uid) {
                                throw new Error(`Preserve replacement node ${original.name}`);
                            }
                            const changedLabels = Object.entries(original.values).filter(
                                ([domain, value]) => {
                                    const currentValue = current.metadata.labels?.[domain];
                                    if (currentValue === value) return false;
                                    if (currentValue !== fixtureLabelValue) {
                                        throw new Error(
                                            `Preserve externally changed label ${original.name}/${domain}`
                                        );
                                    }
                                    return true;
                                }
                            );
                            if (changedLabels.length > 0) {
                                await kubeCli.patchNode(original.name, [
                                    { op: 'test', path: '/metadata/uid', value: original.uid },
                                    {
                                        op: 'test',
                                        path: '/metadata/resourceVersion',
                                        value: current.metadata.resourceVersion
                                    },
                                    ...changedLabels.map(([domain, value]) =>
                                        value === undefined
                                            ? { op: 'remove', path: `/metadata/labels/${domain}` }
                                            : { op: 'add', path: `/metadata/labels/${domain}`, value }
                                    )
                                ]);
                            }
                        } catch (error) {
                            cleanupErrors.push(error);
                        }
                    }
                    if (cleanupErrors.length > 0) {
                        throw new AggregateError(cleanupErrors, 'Dynamo topology fixture cleanup failed');
                    }
                });

                test('A component without its own constraint inherits the deployment constraint', () => {
                    const componentElements = elementsOfComponent(inheritingComponent);
                    expect(componentElements.length).to.be.greaterThan(
                        0,
                        `no elements found for component ${inheritingComponent}`
                    );
                    for (const element of componentElements) {
                        expect(element.topology.name).to.equal(componentTopologyName, 'topology name');
                        expect(element.topology.requiredPlacement).to.equal(
                            deploymentPackDomain,
                            'inherited required placement'
                        );
                        // packDomain is always a required placement, never a preferred one.
                        expect(element.topology.preferredPlacement || '').to.equal('');
                    }
                });

                test('A component with its own constraint overrides the deployment constraint', () => {
                    const componentElements = elementsOfComponent(overridingComponent);
                    expect(componentElements.length).to.be.greaterThan(
                        0,
                        `no elements found for component ${overridingComponent}`
                    );
                    for (const element of componentElements) {
                        expect(element.topology.name).to.equal(componentTopologyName, 'topology name');
                        expect(element.topology.requiredPlacement).to.equal(
                            componentPackDomain,
                            'overriding required placement'
                        );
                        expect(element.topology.preferredPlacement || '').to.equal('');
                    }
                });

                test('The workload reports the deployment constraint', async () => {
                    const topology = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body
                        .topology;
                    networkTopologiesUtils.validateNetworkTopologyDataInWorkloadsApi(
                        topology,
                        componentTopologyName,
                        undefined,
                        deploymentPackDomain
                    );
                });
            });
        });
    });
});
