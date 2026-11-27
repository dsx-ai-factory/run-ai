/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { networkTopologiesApi } from '../../../../../../modules/api';
import { test } from '@playwright/test';
import { crds, tags } from '../../../../../../modules/helpers';
import {
    flowRootState,
    logger,
    objectGenerator,
    networkTopologiesUtils
} from '../../../../../../modules/helpers';
import { NetworkTopologyCreateRequest, NetworkTopology } from '../../../../../../modules/api/apis/clusterApi';
import { expect } from 'chai';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { StatusCodes } from 'http-status-codes';

test.describe(
    `NetworkTopologies API, tags:  ${tags.resource.networkTopologies} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.services.clusterService, tags.runOn.parallelCI] },
    function () {
        supportedFrom(SupportedVersions.V2_23)(function () {
            let testTopologyName: string;
            let createdTopologyId: string;
            let testClusterId: string;
            let multipleTopologyIds: string[] = []; // Track multiple topologies for cleanup

            test.beforeAll('Setup test environment', function () {
                // Set access token for API calls
                networkTopologiesApi.accessToken = flowRootState.users.system_administrator.accessToken;

                crds.initCrds(flowRootState.cluster.version);

                // Generate unique test topology name
                testTopologyName = `test-topology-${objectGenerator.generateNameByPrefix('')}`;
                testClusterId = flowRootState.cluster.uuid;

                logger.info(
                    `Test setup completed. Topology name: ${testTopologyName}, Cluster ID: ${testClusterId}`
                );
            });

            test.afterAll('Cleanup test resources', async function () {
                const cleanupErrors: string[] = [];

                try {
                    // Cleanup the main test topology if it exists
                    if (createdTopologyId) {
                        try {
                            await networkTopologiesUtils.deleteNetworkTopologyViaApi(createdTopologyId);
                            logger.info(`Test topology ${createdTopologyId} deleted via API`);
                        } catch (error) {
                            cleanupErrors.push(
                                `Failed to cleanup main test topology ${createdTopologyId}: ${error.message}`
                            );
                        }
                        // Reset the variable after cleanup
                        createdTopologyId = undefined;
                    }

                    // Cleanup multiple topologies if they exist
                    if (multipleTopologyIds.length > 0) {
                        for (const topologyId of multipleTopologyIds) {
                            try {
                                await networkTopologiesUtils.deleteNetworkTopologyViaApi(topologyId);
                                logger.info(`Multiple topology ${topologyId} deleted via API`);
                            } catch (error) {
                                cleanupErrors.push(
                                    `Failed to cleanup multiple topology ${topologyId}: ${error.message}`
                                );
                            }
                        }
                        // Reset the array after cleanup
                        multipleTopologyIds = [];
                    }

                    // Cleanup any remaining topologies in the test cluster that might have been created during tests
                    // This handles cases where tests fail and don't get to their individual cleanup steps
                    const remainingTopologies = await networkTopologiesUtils.getNetworkTopologiesViaApi([
                        `clusterId==${testClusterId}`
                    ]);

                    if (remainingTopologies.body.networkTopologies) {
                        for (const topology of remainingTopologies.body.networkTopologies) {
                            // Only cleanup test topologies (those with our test naming patterns)
                            if (
                                topology.name.includes('test-topology-') ||
                                topology.name.includes('multi-topology-') ||
                                topology.name.includes('filter-test-') ||
                                topology.name.includes('source-test-') ||
                                topology.name.includes('cluster-topology-') ||
                                topology.name.includes('api-topology-') ||
                                topology.name.includes('combined-filter-test-')
                            ) {
                                try {
                                    // Check if it's an API-created topology (can be deleted via API)
                                    if (topology.source === 'Run.ai') {
                                        await networkTopologiesUtils.deleteNetworkTopologyViaApi(topology.id);
                                        logger.info(
                                            `Cleaned up remaining test topology ${topology.name} (ID: ${topology.id}) via API`
                                        );
                                    } else if (topology.source === 'Cluster') {
                                        // For cluster-created topologies, we need to delete them directly from cluster
                                        try {
                                            await networkTopologiesUtils.deleteTopologyCRDFromCluster(
                                                topology.name
                                            );
                                            logger.info(
                                                `Cleaned up remaining cluster topology ${topology.name} directly from cluster`
                                            );
                                        } catch (clusterError) {
                                            cleanupErrors.push(
                                                `Failed to cleanup cluster topology ${topology.name}: ${clusterError.message}`
                                            );
                                        }
                                    }
                                } catch (error) {
                                    cleanupErrors.push(
                                        `Failed to cleanup remaining test topology ${topology.name}: ${error.message}`
                                    );
                                }
                            }
                        }
                    }

                    // Validate cleanup by checking that no test topologies remain - getNetworkTopologies by clusterId
                    const finalCheck = await networkTopologiesUtils.getNetworkTopologiesViaApi([
                        `clusterId==${testClusterId}`
                    ]);
                    if (finalCheck.body.networkTopologies) {
                        const remainingTestTopologies = finalCheck.body.networkTopologies.filter(
                            (topology: any) =>
                                topology.name.includes('test-topology-') ||
                                topology.name.includes('multi-topology-') ||
                                topology.name.includes('filter-test-') ||
                                topology.name.includes('source-test-') ||
                                topology.name.includes('cluster-topology-') ||
                                topology.name.includes('api-topology-') ||
                                topology.name.includes('combined-filter-test-')
                        );

                        if (remainingTestTopologies.length > 0) {
                            cleanupErrors.push(
                                `Found ${remainingTestTopologies.length} remaining test topologies in cluster ${testClusterId} after cleanup: ${remainingTestTopologies.map((t: any) => t.name).join(', ')}`
                            );
                        } else {
                            logger.info(
                                `All test topologies successfully cleaned up from cluster ${testClusterId}`
                            );
                        }
                    }
                } catch (error) {
                    cleanupErrors.push(`General cleanup error: ${error.message}`);
                }

                // Log any cleanup errors for debugging
                if (cleanupErrors.length > 0) {
                    logger.warn(`Cleanup completed with ${cleanupErrors.length} errors:`);
                    cleanupErrors.forEach((error) => logger.warn(`  - ${error}`));
                }
            });

            test('should create network topology via API and validate cluster sync', async function () {
                const createRequest: NetworkTopologyCreateRequest = {
                    name: testTopologyName,
                    clusterId: testClusterId,
                    levels: [
                        'cloud.provider.com/topology-block',
                        'cloud.provider.com/topology-rack',
                        'kubernetes.io/hostname'
                    ]
                };

                await test.step('Create network topology via API', async function () {
                    const response = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                    expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                    expect(response.body).to.have.property('id');
                    expect(response.body.name).to.equal(testTopologyName);
                    expect(response.body.clusterId).to.equal(testClusterId);
                    expect(response.body.levels).to.deep.equal(createRequest.levels);
                    expect(response.body.source).to.equal('Run.ai');

                    createdTopologyId = response.body.id;
                    logger.info(`Network topology created via API with ID: ${createdTopologyId}`);
                });

                await test.step('Validate topology CRD is created in cluster and NodeLevels match', async function () {
                    // Create a mock API topology object for validation
                    const mockApiTopology = {
                        levels: createRequest.levels,
                        name: testTopologyName
                    } as any;

                    // Wait for the topology CRD to be created and validate its NodeLevels
                    const topologyCrd = await networkTopologiesUtils.waitForTopologyCRDAndValidateLevels(
                        testTopologyName,
                        mockApiTopology,
                        60 // 60 second timeout
                    );

                    // Additional validation of CRD metadata
                    networkTopologiesUtils.validateTopologyCRDMetadata(topologyCrd, testTopologyName);

                    logger.info(`Topology CRD ${testTopologyName} created and validated successfully`);
                });

                await test.step('Validate topology is synced with control plane', async function () {
                    // Get the topology via API to ensure it's synced
                    const getResponse =
                        await networkTopologiesUtils.getNetworkTopologyByIdViaApi(createdTopologyId);
                    expect(getResponse.response.statusCode).to.equal(StatusCodes.OK);
                    expect(getResponse.body.id).to.equal(createdTopologyId);
                    expect(getResponse.body.name).to.equal(testTopologyName);
                    expect(getResponse.body.levels).to.deep.equal(createRequest.levels);

                    logger.info(`Topology ${createdTopologyId} is synced with control plane`);
                });
            });

            test('should list network topologies and validate filtering', async function () {
                await test.step('List all network topologies', async function () {
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi();
                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body).to.have.property('networkTopologies');
                    expect(Array.isArray(response.body.networkTopologies)).to.be.true;

                    // Verify our test topology is in the list
                    const testTopology = response.body.networkTopologies.find(
                        (topology: any) => topology.id === createdTopologyId
                    );
                    expect(testTopology).to.exist;
                    expect(testTopology.name).to.equal(testTopologyName);

                    logger.info(
                        `Found ${response.body.networkTopologies.length} network topologies, including test topology`
                    );
                });
            });

            test('should count network topologies', async function () {
                await test.step('Get network topology count', async function () {
                    const response = await networkTopologiesApi.getNetworkTopologiesCount(undefined, {
                        headers: {}
                    });
                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body).to.have.property('count');
                    expect(typeof response.body.count).to.equal('number');
                    expect(response.body.count).to.be.greaterThan(0);

                    logger.info(`Network topology count: ${response.body.count}`);
                });
            });

            test('should delete network topology via API and validate cluster cleanup', async function () {
                await test.step('Delete network topology via API', async function () {
                    const response =
                        await networkTopologiesUtils.deleteNetworkTopologyViaApi(createdTopologyId);
                    expect(response.response.statusCode).to.equal(StatusCodes.NO_CONTENT);

                    logger.info(`Network topology ${createdTopologyId} deleted via API`);
                });

                await test.step('Validate topology CRD is removed from cluster', async function () {
                    // Wait for the topology CRD to be removed from the cluster
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(testTopologyName, 60);

                    logger.info(`Topology CRD ${testTopologyName} removed from cluster`);
                });

                await test.step('Validate topology is removed from control plane', async function () {
                    try {
                        await networkTopologiesUtils.getNetworkTopologyByIdViaApi(createdTopologyId);
                        expect.fail('Topology should not exist in control plane');
                    } catch (error) {
                        expect(error.response?.statusCode).to.equal(StatusCodes.NOT_FOUND);
                        logger.info('Correctly returned 404 for non-existent topology');
                    }
                });

                // Reset the createdTopologyId since it's been deleted
                createdTopologyId = undefined;
            });

            test('should create multiple network topologies and validate cluster sync', async function () {
                const topologyNames = [
                    `multi-topology-1-${objectGenerator.generateNameByPrefix('')}`,
                    `multi-topology-2-${objectGenerator.generateNameByPrefix('')}`,
                    `multi-topology-3-${objectGenerator.generateNameByPrefix('')}`
                ];

                const topologyConfigs = [
                    {
                        name: topologyNames[0],
                        levels: ['zone', 'rack', 'host']
                    },
                    {
                        name: topologyNames[1],
                        levels: ['datacenter', 'zone', 'rack', 'host']
                    },
                    {
                        name: topologyNames[2],
                        levels: ['region', 'zone', 'host']
                    }
                ];

                // Store IDs for cleanup in afterAll
                multipleTopologyIds = [];

                await test.step('Create multiple network topologies via API', async function () {
                    for (const config of topologyConfigs) {
                        const createRequest: NetworkTopologyCreateRequest = {
                            name: config.name,
                            clusterId: testClusterId,
                            levels: config.levels
                        };

                        const response =
                            await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                        expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                        expect(response.body.name).to.equal(config.name);
                        expect(response.body.levels).to.deep.equal(config.levels);

                        multipleTopologyIds.push(response.body.id);
                        logger.info(`Created topology ${config.name} with ID: ${response.body.id}`);
                    }
                });

                await test.step('Validate all topology CRDs are created in cluster with correct NodeLevels', async function () {
                    for (let i = 0; i < topologyNames.length; i++) {
                        const topologyName = topologyNames[i];
                        const topologyId = multipleTopologyIds[i];
                        const expectedLevels = topologyConfigs[i].levels;

                        // Get the topology via API to ensure it's created
                        const getResponse =
                            await networkTopologiesUtils.getNetworkTopologyByIdViaApi(topologyId);
                        expect(getResponse.response.statusCode).to.equal(StatusCodes.OK);
                        expect(getResponse.body.name).to.equal(topologyName);
                        expect(getResponse.body.levels).to.deep.equal(expectedLevels);

                        // Wait for the topology CRD to be created and validate its NodeLevels
                        const topologyCrd = await networkTopologiesUtils.waitForTopologyCRDAndValidateLevels(
                            topologyName,
                            getResponse.body,
                            60 // 60 second timeout
                        );

                        logger.info(`Topology CRD ${topologyName} created and validated successfully`);
                    }
                });

                await test.step('Cleanup multiple topologies', async function () {
                    for (const topologyId of multipleTopologyIds) {
                        try {
                            await networkTopologiesUtils.deleteNetworkTopologyViaApi(topologyId);
                            logger.info(`Cleaned up topology ${topologyId}`);
                        } catch (error) {
                            logger.warn(`Failed to cleanup topology ${topologyId}: ${error.message}`);
                        }
                    }

                    // Reset the array since topologies have been cleaned up
                    multipleTopologyIds = [];
                });
            });

            test('should handle invalid network topology creation gracefully', async function () {
                await test.step('Attempt to create topology with invalid name', async function () {
                    const invalidRequest: NetworkTopologyCreateRequest = {
                        name: 'invalid-topology-name-with  space', // Invalid: contains space
                        clusterId: testClusterId,
                        levels: ['zone', 'rack']
                    };

                    try {
                        await networkTopologiesUtils.createNetworkTopologyViaApi(invalidRequest);
                        expect.fail('Should have failed with invalid topology name');
                    } catch (error) {
                        expect(error.response?.statusCode).to.be.oneOf([
                            StatusCodes.BAD_REQUEST,
                            StatusCodes.UNPROCESSABLE_ENTITY
                        ]);
                        logger.info(
                            `Correctly rejected invalid topology name: ${error.response?.statusCode}`
                        );
                    }
                });

                await test.step('Attempt to create topology with empty levels', async function () {
                    const invalidRequest: NetworkTopologyCreateRequest = {
                        name: `test-empty-levels-${objectGenerator.generateNameByPrefix('')}`,
                        clusterId: testClusterId,
                        levels: [] // Invalid: empty levels array
                    };

                    try {
                        await networkTopologiesUtils.createNetworkTopologyViaApi(invalidRequest);
                        expect.fail('Should have failed with empty levels');
                    } catch (error) {
                        expect(error.response?.statusCode).to.be.oneOf([
                            StatusCodes.BAD_REQUEST,
                            StatusCodes.UNPROCESSABLE_ENTITY
                        ]);
                        logger.info(`Correctly rejected empty levels: ${error.response?.statusCode}`);
                    }
                });
            });

            test('should handle network topology operations with non-existent ID gracefully', async function () {
                const nonExistentId = '00000000-0000-0000-0000-000000000000';

                await test.step('Attempt to get non-existent topology', async function () {
                    try {
                        await networkTopologiesUtils.getNetworkTopologyByIdViaApi(nonExistentId);
                        expect.fail('Should have failed with non-existent topology');
                    } catch (error) {
                        expect(error.response?.statusCode).to.equal(StatusCodes.NOT_FOUND);
                        logger.info('Correctly returned 404 for non-existent topology');
                    }
                });

                await test.step('Attempt to delete non-existent topology', async function () {
                    try {
                        await networkTopologiesUtils.deleteNetworkTopologyViaApi(nonExistentId);
                        expect.fail('Should have failed with non-existent topology');
                    } catch (error) {
                        expect(error.response?.statusCode).to.equal(StatusCodes.NOT_FOUND);
                        logger.info('Correctly returned 404 for non-existent topology deletion');
                    }
                });
            });

            test('should validate source: topology created in cluster is visible in Control Plane API but cannot be deleted from it', async function () {
                const clusterTopologyName = `cluster-topology-${objectGenerator.generateNameByPrefix('')}`;
                const clusterTopologyLevels = ['zone', 'rack', 'host'];

                await test.step('Create topology directly in cluster (will be synced to Control Plane API)', async function () {
                    const createdTopology = await networkTopologiesUtils.createTopologyCRDInCluster(
                        clusterTopologyName,
                        clusterTopologyLevels
                    );

                    expect(createdTopology).to.exist;
                    expect(createdTopology.metadata.name).to.equal(clusterTopologyName);
                    expect(createdTopology.spec.levels).to.have.length(clusterTopologyLevels.length);

                    // Verify the topology exists in cluster
                    const topologyExists =
                        await networkTopologiesUtils.topologyCRDExists(clusterTopologyName);
                    expect(topologyExists).to.be.true;

                    logger.info(`Topology ${clusterTopologyName} created directly in cluster`);
                });

                await test.step('Verify topology IS visible in Control Plane API but cannot be deleted from it', async function () {
                    // Wait for the topology to be created in Control Plane API using waitForCriteria
                    const filterBy = [`name==${clusterTopologyName}`];
                    const networkTopologies = await networkTopologiesUtils.waitForNetworkTopologies(filterBy);

                    expect(networkTopologies).to.exist;
                    expect(networkTopologies.length).to.be.equal(1);
                    const clusterTopologyInApi = networkTopologies[0];
                    expect(clusterTopologyInApi.name).to.equal(clusterTopologyName);
                    expect(clusterTopologyInApi.source).to.equal('Cluster');

                    logger.info(
                        `Topology ${clusterTopologyName} is visible in Control Plane API with source 'Cluster'`
                    );
                });

                await test.step('Attempt to delete cluster-created topology via Control Plane API should fail', async function () {
                    // Get the topology ID from the API
                    const getResponse = await networkTopologiesUtils.getNetworkTopologiesViaApi([
                        `name==${clusterTopologyName}`
                    ]);

                    expect(getResponse.body.networkTopologies.length).to.be.greaterThan(0);
                    const topologyId = getResponse.body.networkTopologies[0].id;
                    expect(topologyId).to.exist;

                    try {
                        // Attempt to delete the topology via API
                        await networkTopologiesUtils.deleteNetworkTopologyViaApi(topologyId);
                        expect.fail('deleteNetworkTopology should have failed with cluster topology');
                    } catch (e) {
                        // This tests that API deletion doesn't affect cluster-created topologies
                        await new Promise((resolve) => setTimeout(resolve, 3000));
                        const topologyStillExists =
                            await networkTopologiesUtils.topologyCRDExists(clusterTopologyName);
                        expect(topologyStillExists).to.be.true;
                        logger.info(
                            `Topology ${clusterTopologyName} still exists in cluster after API deletion attempt - API deletion ignored for cluster-created topologies`
                        );
                    }
                });

                await test.step('Cleanup: delete topology directly from cluster', async function () {
                    await networkTopologiesUtils.deleteTopologyCRDFromCluster(clusterTopologyName);

                    // Verify deletion
                    const topologyExists =
                        await networkTopologiesUtils.topologyCRDExists(clusterTopologyName);
                    expect(topologyExists).to.be.false;

                    logger.info(`Topology ${clusterTopologyName} deleted from cluster`);
                });
            });

            test('should validate topology created via API when deleted on cluster will be recreated', async function () {
                const apiTopologyName = `api-topology-${objectGenerator.generateNameByPrefix('')}`;
                const apiTopologyLevels = ['datacenter', 'zone', 'rack', 'host'];
                let apiTopologyId: string;

                await test.step('Create topology via Control Plane API', async function () {
                    const createRequest: NetworkTopologyCreateRequest = {
                        name: apiTopologyName,
                        clusterId: testClusterId,
                        levels: apiTopologyLevels
                    };

                    const response = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                    expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                    expect(response.body.name).to.equal(apiTopologyName);
                    expect(response.body.levels).to.deep.equal(apiTopologyLevels);

                    apiTopologyId = response.body.id;
                    logger.info(`Topology ${apiTopologyName} created via API with ID: ${apiTopologyId}`);
                });

                await test.step('Verify topology CRD is created in cluster', async function () {
                    const runAiNetworkTopology: NetworkTopology = {
                        name: apiTopologyName,
                        levels: apiTopologyLevels,
                        id: apiTopologyId,
                        clusterId: testClusterId,
                        tenantId: flowRootState.currentTenant.id,
                        createdAt: new Date(),
                        source: 'Run.ai'
                    };

                    const topologyCrd = await networkTopologiesUtils.waitForTopologyCRDAndValidateLevels(
                        apiTopologyName,
                        runAiNetworkTopology,
                        60
                    );

                    expect(topologyCrd).to.exist;
                    expect((topologyCrd as any).metadata.name).to.equal(apiTopologyName);

                    logger.info(`Topology CRD ${apiTopologyName} created in cluster`);
                });

                await test.step('Do Delete topology CRD directly from cluster (bypassing Control Plane API)', async function () {
                    await networkTopologiesUtils.deleteTopologyCRDFromCluster(apiTopologyName);

                    await networkTopologiesUtils.waitForTopologyCRDDeletion(apiTopologyName);

                    logger.info(`Topology CRD ${apiTopologyName} deleted directly from cluster`);
                });

                await test.step('Verify topology is recreated in cluster', async function () {
                    // Wait for the topology to be recreated
                    const recreatedTopology = await networkTopologiesUtils.waitForTopologyCRDRecreation(
                        apiTopologyName,
                        40
                    );

                    expect(recreatedTopology).to.exist;
                    expect(recreatedTopology.metadata.name).to.equal(apiTopologyName);
                    expect(recreatedTopology.spec.levels).to.have.length(apiTopologyLevels.length);

                    // Validate the levels match
                    for (let i = 0; i < apiTopologyLevels.length; i++) {
                        expect(recreatedTopology.spec.levels[i].nodeLabel).to.equal(apiTopologyLevels[i]);
                    }

                    logger.info(`Topology CRD ${apiTopologyName} was recreated in cluster by Control Plane`);
                });

                await test.step('Verify topology still exists in Control Plane API', async function () {
                    const getResponse =
                        await networkTopologiesUtils.getNetworkTopologyByIdViaApi(apiTopologyId);
                    expect(getResponse.response.statusCode).to.equal(StatusCodes.OK);
                    expect(getResponse.body.id).to.equal(apiTopologyId);
                    expect(getResponse.body.name).to.equal(apiTopologyName);
                    expect(getResponse.body.levels).to.deep.equal(apiTopologyLevels);

                    logger.info(
                        `Topology ${apiTopologyName} still exists in Control Plane API after cluster deletion`
                    );
                });

                await test.step('Cleanup: delete topology via API', async function () {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(apiTopologyId);

                    // Verify cleanup from cluster
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(apiTopologyName, 60);

                    logger.info(`Topology ${apiTopologyName} cleaned up via API`);
                });
            });

            test('should validate topologies exist in both systems with different sources and deletion restrictions', async function () {
                const clusterTopologyName = `source-test-cluster-${objectGenerator.generateNameByPrefix('')}`;
                const apiTopologyName = `source-test-api-${objectGenerator.generateNameByPrefix('')}`;
                const testLevels = ['region', 'zone', 'host'];
                let apiTopologyId: string;

                await test.step('Create topologies in both systems (cluster and API)', async function () {
                    // Create topology directly in cluster
                    const clusterTopology = await networkTopologiesUtils.createTopologyCRDInCluster(
                        clusterTopologyName,
                        testLevels
                    );
                    expect(clusterTopology).to.exist;

                    // Create topology via API
                    const createRequest: NetworkTopologyCreateRequest = {
                        name: apiTopologyName,
                        clusterId: testClusterId,
                        levels: testLevels
                    };
                    const response = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                    expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                    apiTopologyId = response.body.id;

                    logger.info(
                        `Created topologies: ${clusterTopologyName} in cluster, ${apiTopologyName} via API`
                    );
                });

                await test.step('Verify both topologies exist in their Cluster and Backend', async function () {
                    // Cluster topology should exist in cluster
                    const clusterTopologyExists =
                        await networkTopologiesUtils.topologyCRDExists(clusterTopologyName);
                    expect(clusterTopologyExists).to.be.true;

                    // API topology should exist in cluster (synced) - wait for sync
                    const apiTopology: NetworkTopology = {
                        name: apiTopologyName,
                        levels: testLevels,
                        id: apiTopologyId,
                        clusterId: testClusterId,
                        tenantId: flowRootState.currentTenant.id,
                        createdAt: new Date(),
                        source: 'Run.ai'
                    };

                    const apiTopologyCrd = await networkTopologiesUtils.waitForTopologyCRDAndValidateLevels(
                        apiTopologyName,
                        apiTopology,
                        60 // 60 second timeout for sync
                    );
                    expect(apiTopologyCrd).to.exist;

                    // API topology should exist in Control Plane
                    const getResponse =
                        await networkTopologiesUtils.getNetworkTopologyByIdViaApi(apiTopologyId);
                    expect(getResponse.response.statusCode).to.equal(StatusCodes.OK);

                    // Cluster topology SHOULD exist in Control Plane (synced)
                    const allTopologies = await networkTopologiesUtils.getNetworkTopologiesViaApi();
                    const clusterTopologyInApi = allTopologies.body.networkTopologies.find(
                        (topology: any) => topology.name === clusterTopologyName
                    );
                    expect(clusterTopologyInApi).to.exist;
                    expect(clusterTopologyInApi.source).to.equal('Cluster');

                    logger.info(
                        'Isolation verified: topologies exist in both systems but with different sources'
                    );
                });

                await test.step('Cleanup both topologies', async function () {
                    // Delete cluster topology directly
                    await networkTopologiesUtils.deleteTopologyCRDFromCluster(clusterTopologyName);

                    // Delete API topology via API
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(apiTopologyId);

                    // Verify both are cleaned up
                    const clusterTopologyExists =
                        await networkTopologiesUtils.topologyCRDExists(clusterTopologyName);
                    expect(clusterTopologyExists).to.be.false;

                    await networkTopologiesUtils.waitForTopologyCRDDeletion(apiTopologyName, 60);

                    logger.info('Both topologies cleaned up successfully');
                });
            });

            test('should validate getNetworkTopologies filterBy name parameter', async function () {
                const testTopologyName = `filter-test-name-${objectGenerator.generateNameByPrefix('')}`;
                const testLevels = ['zone', 'rack', 'host'];
                let testTopologyId: string;

                await test.step('Create test topology via API', async function () {
                    const createRequest: NetworkTopologyCreateRequest = {
                        name: testTopologyName,
                        clusterId: testClusterId,
                        levels: testLevels
                    };

                    const response = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                    expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                    testTopologyId = response.body.id;

                    logger.info(`Created test topology ${testTopologyName} with ID: ${testTopologyId}`);
                });

                await test.step('Test filtering by exact name match', async function () {
                    const filterParams = [`name==${testTopologyName}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies).to.have.length(1);
                    expect(response.body.networkTopologies[0].name).to.equal(testTopologyName);
                    expect(response.body.networkTopologies[0].id).to.equal(testTopologyId);

                    logger.info(`Filter by exact name '${testTopologyName}' returned 1 topology`);
                });

                await test.step('Test filtering by partial name match', async function () {
                    const partialName = testTopologyName.substring(0, testTopologyName.length - 5); // Remove last 5 characters
                    const filterParams = [`name=^${partialName}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    // Should find topologies that start with the partial name
                    const matchingTopologies = response.body.networkTopologies.filter((topology: any) =>
                        topology.name.startsWith(partialName)
                    );
                    expect(matchingTopologies.length).to.be.greaterThan(0);

                    logger.info(
                        `Filter by partial name '${partialName}' returned ${matchingTopologies.length} topologies`
                    );
                });

                await test.step('Test filtering by non-existent name', async function () {
                    const nonExistentName = `non-existent-${objectGenerator.generateNameByPrefix('')}`;
                    const filterParams = [`name==${nonExistentName}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies).to.be.undefined;

                    logger.info(`Filter by non-existent name '${nonExistentName}' returned 0 topologies`);
                });

                await test.step('Cleanup test topology', async function () {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(testTopologyId);
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(testTopologyName, 60);
                    logger.info(`Test topology ${testTopologyName} cleaned up`);
                });
            });

            test('should validate getNetworkTopologies filterBy source parameter', async function () {
                const apiTopologyName = `filter-test-api-${objectGenerator.generateNameByPrefix('')}`;
                const clusterTopologyName = `filter-test-cluster-${objectGenerator.generateNameByPrefix('')}`;
                const testLevels = ['zone', 'rack', 'host'];
                let apiTopologyId: string;

                await test.step('Create topologies from different sources', async function () {
                    // Create topology via API (source: Run.ai)
                    const createRequest: NetworkTopologyCreateRequest = {
                        name: apiTopologyName,
                        clusterId: testClusterId,
                        levels: testLevels
                    };

                    const response = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                    expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                    apiTopologyId = response.body.id;

                    // Create topology directly in cluster (source: Cluster)
                    const clusterTopology = await networkTopologiesUtils.createTopologyCRDInCluster(
                        clusterTopologyName,
                        testLevels
                    );
                    expect(clusterTopology).to.exist;

                    const networkTopologies = await networkTopologiesUtils.waitForNetworkTopologies([
                        `name==${clusterTopologyName}`
                    ]);
                    expect(networkTopologies.length).to.be.equal(1);
                    expect(networkTopologies[0].name).to.equal(clusterTopologyName);

                    logger.info(
                        `Created topologies: ${apiTopologyName} via API, ${clusterTopologyName} in cluster`
                    );
                });

                await test.step('Test filtering by source Run.ai', async function () {
                    const filterParams = ['source==Run.ai'];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies.length).to.be.greaterThan(0);

                    // All returned topologies should have source 'Run.ai'
                    const runaiTopologies = response.body.networkTopologies.filter(
                        (topology: any) => topology.source === 'Run.ai'
                    );
                    expect(runaiTopologies.length).to.equal(response.body.networkTopologies.length);

                    // Should include our test topology
                    const testTopology = response.body.networkTopologies.find(
                        (topology: any) => topology.name === apiTopologyName
                    );
                    expect(testTopology).to.exist;
                    expect(testTopology.source).to.equal('Run.ai');

                    logger.info(
                        `Filter by source 'Run.ai' returned ${response.body.networkTopologies.length} topologies`
                    );
                });

                await test.step('Test filtering by source Cluster', async function () {
                    const filterParams = ['source==Cluster'];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    // Cluster-created topologies should not be visible in API
                    expect(response.body.networkTopologies).to.exist;
                    expect(response.body.networkTopologies.length).to.have.greaterThan(0);

                    logger.info('Filter by source Cluster returned 0 topologies (as expected)');
                });

                await test.step('Test filtering by non-existent source', async function () {
                    const nonExistentSource = 'NonExistentSource';
                    const filterParams = [`source==${nonExistentSource}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies).to.be.undefined;

                    logger.info(`Filter by non-existent source '${nonExistentSource}' returned 0 topologies`);
                });

                await test.step('Cleanup test topologies', async function () {
                    // Cleanup API topology
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(apiTopologyId);
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(apiTopologyName, 60);

                    // Cleanup cluster topology
                    await networkTopologiesUtils.deleteTopologyCRDFromCluster(clusterTopologyName);

                    logger.info('Test topologies cleaned up');
                });
            });

            test('should validate getNetworkTopologies filterBy clusterId parameter', async function () {
                const testTopologyName = `filter-test-cluster-${objectGenerator.generateNameByPrefix('')}`;
                const testLevels = ['zone', 'rack', 'host'];
                let testTopologyId: string;

                await test.step('Create test topology via API', async function () {
                    const createRequest: NetworkTopologyCreateRequest = {
                        name: testTopologyName,
                        clusterId: testClusterId,
                        levels: testLevels
                    };

                    const response = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                    expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                    testTopologyId = response.body.id;

                    logger.info(`Created test topology ${testTopologyName} with ID: ${testTopologyId}`);
                });

                await test.step('Test filtering by exact clusterId match', async function () {
                    const filterParams = [`clusterId==${testClusterId}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies.length).to.be.greaterThan(0);

                    // All returned topologies should have the specified clusterId
                    const matchingTopologies = response.body.networkTopologies.filter(
                        (topology: any) => topology.clusterId === testClusterId
                    );
                    expect(matchingTopologies.length).to.equal(response.body.networkTopologies.length);

                    // Should include our test topology
                    const testTopology = response.body.networkTopologies.find(
                        (topology: any) => topology.name === testTopologyName
                    );
                    expect(testTopology).to.exist;
                    expect(testTopology.clusterId).to.equal(testClusterId);

                    logger.info(
                        `Filter by clusterId '${testClusterId}' returned ${response.body.networkTopologies.length} topologies`
                    );
                });

                await test.step('Test filtering by non-existent clusterId', async function () {
                    const nonExistentClusterId = '00000000-0000-0000-0000-000000000000';
                    const filterParams = [`clusterId==${nonExistentClusterId}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies).to.be.undefined;

                    logger.info(
                        `Filter by non-existent clusterId '${nonExistentClusterId}' returned 0 topologies`
                    );
                });

                await test.step('Test filtering by invalid clusterId format', async function () {
                    const invalidClusterId = 'invalid-cluster-id';
                    const filterParams = [`clusterId==${invalidClusterId}`];
                    try {
                        const response =
                            await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                        expect.fail('getNetworkTopologies should fail with invalid clusterId');
                    } catch (error) {
                        expect(error.response?.statusCode).to.equal(StatusCodes.BAD_REQUEST);
                        logger.info('Correctly returned 400 for non-valid clusterId');
                    }
                    logger.info(
                        `Filter by invalid clusterId format '${invalidClusterId}' returned 0 topologies`
                    );
                });

                await test.step('Cleanup test topology', async function () {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(testTopologyId);
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(testTopologyName, 60);
                    logger.info(`Test topology ${testTopologyName} cleaned up`);
                });
            });

            test('should validate getNetworkTopologies with combined filterBy parameters', async function () {
                const testTopologyName = `combined-filter-test-${objectGenerator.generateNameByPrefix('')}`;
                const testLevels = ['zone', 'rack', 'host'];
                let testTopologyId: string;

                await test.step('Create test topology via API', async function () {
                    const createRequest: NetworkTopologyCreateRequest = {
                        name: testTopologyName,
                        clusterId: testClusterId,
                        levels: testLevels
                    };

                    const response = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest);
                    expect(response.response.statusCode).to.equal(StatusCodes.CREATED);
                    testTopologyId = response.body.id;

                    logger.info(`Created test topology ${testTopologyName} with ID: ${testTopologyId}`);
                });

                await test.step('Test filtering by name and source combination', async function () {
                    const filterParams = [`name==${testTopologyName}`, 'source==Run.ai'];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies).to.have.length(1);
                    expect(response.body.networkTopologies[0].name).to.equal(testTopologyName);
                    expect(response.body.networkTopologies[0].source).to.equal('Run.ai');
                    expect(response.body.networkTopologies[0].clusterId).to.equal(testClusterId);

                    logger.info(
                        `Combined filter by name '${testTopologyName}' and source 'Run.ai' returned 1 topology`
                    );
                });

                await test.step('Test filtering by name and clusterId combination', async function () {
                    const filterParams = [`name==${testTopologyName}`, `clusterId==${testClusterId}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies).to.have.length(1);
                    expect(response.body.networkTopologies[0].name).to.equal(testTopologyName);
                    expect(response.body.networkTopologies[0].clusterId).to.equal(testClusterId);

                    logger.info(
                        `Combined filter by name '${testTopologyName}' and clusterId '${testClusterId}' returned 1 topology`
                    );
                });

                await test.step('Test filtering by source and clusterId combination', async function () {
                    const filterParams = ['source==Run.ai', `clusterId==${testClusterId}`];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies.length).to.be.greaterThan(0);

                    // All returned topologies should match both filters
                    const matchingTopologies = response.body.networkTopologies.filter(
                        (topology: any) =>
                            topology.source === 'Run.ai' && topology.clusterId === testClusterId
                    );
                    expect(matchingTopologies.length).to.equal(response.body.networkTopologies.length);

                    // Should include our test topology
                    const testTopology = response.body.networkTopologies.find(
                        (topology: any) => topology.name === testTopologyName
                    );
                    expect(testTopology).to.exist;

                    logger.info(
                        `Combined filter by source 'Run.ai' and clusterId '${testClusterId}' returned ${response.body.networkTopologies.length} topologies`
                    );
                });

                await test.step('Test filtering by all three parameters', async function () {
                    const filterParams = [
                        `name==${testTopologyName}`,
                        'source==Run.ai',
                        `clusterId==${testClusterId}`
                    ];
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi(filterParams);

                    expect(response.response.statusCode).to.equal(StatusCodes.OK);
                    expect(response.body.networkTopologies).to.have.length(1);
                    expect(response.body.networkTopologies[0].name).to.equal(testTopologyName);
                    expect(response.body.networkTopologies[0].source).to.equal('Run.ai');
                    expect(response.body.networkTopologies[0].clusterId).to.equal(testClusterId);

                    logger.info('Combined filter by all three parameters returned 1 topology');
                });

                await test.step('Cleanup test topology', async function () {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(testTopologyId);
                    await networkTopologiesUtils.waitForTopologyCRDDeletion(testTopologyName, 60);
                    logger.info(`Test topology ${testTopologyName} cleaned up`);
                });
            });
        });
    }
);
