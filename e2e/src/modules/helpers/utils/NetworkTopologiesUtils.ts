/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { crds, flowRootState, logger, waitUtils } from '../index';
import { kubeCli } from '../../cli';
import { NetworkTopology, NetworkTopologyCreateRequest } from '../../api/apis/clusterApi';
import { expect } from 'chai';
import { networkTopologiesApi } from '../../api';
import http from 'http';
import { WorkloadTopology } from '../../api/apis/workloadsApi/generated/api';
import { PodGroupTopologyConstraint } from '../models/podGroups';

const NetworkTopologySourceLabel = 'run.ai/network-topology-source';

export class NetworkTopologiesUtils {
    /**
     * Get topology CRD by name from the cluster
     * @param topologyName - Name of the topology to retrieve
     * @param timeout - Timeout in seconds for waiting for CRD to exist
     * @returns Promise<KubernetesObject> - The topology CRD object
     */
    async getTopologyCRDByName(topologyName: string, timeout: number = 60): Promise<any> {
        try {
            // Wait for the topology CRD to exist in the cluster
            await waitUtils.waitForCrd(
                crds.topologyCrd,
                topologyName,
                undefined, // cluster-scoped
                timeout
            );

            // Get the topology CRD from the cluster
            const topologyCrd = await waitUtils.waitAndGetCrd(
                crds.topologyCrd,
                topologyName,
                undefined // cluster-scoped
            );

            logger.info(`Retrieved topology CRD: ${topologyName}`);
            return topologyCrd;
        } catch (error) {
            logger.error(`Failed to get topology CRD by name: ${topologyName}`, error);
            throw error;
        }
    }

    /**
     * Validate that the API model levels match the NodeLevels of the topology CRD
     * @param apiTopology - The network topology from the API
     * @param topologyCrd - The topology CRD from the cluster
     * @throws Error if validation fails
     */
    validateTopologyNodeLevels(apiTopology: NetworkTopology, topologyCrd: object): void {
        try {
            logger.info('validateTopologyNodeLevels', topologyCrd['spec']);
            // Validate that the topology CRD has the required structure

            expect(topologyCrd).to.have.property('spec');
            expect(topologyCrd['spec']).to.have.property('levels');

            const crdNodeLevels = topologyCrd['spec']['levels'];
            const apiLevels = apiTopology.levels;

            // Validate that the levels arrays have the same length
            expect(crdNodeLevels).to.have.length(
                apiLevels.length,
                `Topology ${apiTopology.name}: CRD NodeLevels length (${crdNodeLevels.length}) should match API levels length (${apiLevels.length})`
            );

            // Validate that each level matches exactly (including order)
            for (let i = 0; i < apiLevels.length; i++) {
                expect(crdNodeLevels[i].nodeLabel).to.equal(
                    apiLevels[i],
                    `Topology ${apiTopology.name}: CRD NodeLevels[${i}] "${crdNodeLevels[i]}" should match API levels[${i}] "${apiLevels[i]}"`
                );
            }

            logger.info(
                `Topology ${apiTopology.name}: NodeLevels validation passed. Levels: ${JSON.stringify(crdNodeLevels)}`
            );
        } catch (error) {
            logger.error(`Topology ${apiTopology.name}: NodeLevels validation failed`, {
                apiLevels: apiTopology.levels,
                crdNodeLevels: topologyCrd['spec']?.['levels'],
                error: error.message
            });
            throw error;
        }
    }

    /**
     * Wait for topology CRD to be created and validate its NodeLevels
     * @param topologyName - Name of the topology
     * @param apiTopology - The network topology from the API
     * @param timeout - Timeout in seconds for waiting
     * @returns Promise<any> - The validated topology CRD
     */
    async waitForTopologyCRDAndValidateLevels(
        topologyName: string,
        apiTopology: NetworkTopology,
        timeout: number = 60
    ): Promise<object> {
        const topologyCrd = await this.getTopologyCRDByName(topologyName, timeout);
        this.validateTopologyNodeLevels(apiTopology, topologyCrd);
        return topologyCrd;
    }

    /**
     * Wait for topology CRD to be updated with specific NodeLevels
     * @param topologyName - Name of the topology
     * @param expectedLevels - Expected NodeLevels array
     * @param timeout - Timeout in seconds for waiting
     * @returns Promise<any> - The updated topology CRD
     */
    async waitForTopologyCRDUpdate(
        topologyName: string,
        expectedLevels: string[],
        timeout: number = 60
    ): Promise<object> {
        try {
            // Wait for the topology CRD to be updated with the expected NodeLevels
            await waitUtils.waitForCrdProperty(
                crds.topologyCrd,
                topologyName,
                undefined, // cluster-scoped
                'spec',
                JSON.stringify(expectedLevels),
                'levels', // no sub-property
                timeout
            );

            // Get the updated topology CRD
            const topologyCrd = await this.getTopologyCRDByName(topologyName, 30);

            logger.info(
                `Topology ${topologyName}: CRD updated successfully with NodeLevels: ${JSON.stringify(expectedLevels)}`
            );
            return topologyCrd;
        } catch (error) {
            logger.error(`Failed to wait for topology CRD update: ${topologyName}`, error);
            throw error;
        }
    }

    /**
     * Wait for topology CRD to be deleted
     * @param topologyName - Name of the topology
     * @param timeout - Timeout in seconds for waiting
     */
    async waitForTopologyCRDDeletion(topologyName: string, timeout: number = 60): Promise<void> {
        try {
            // Wait for the topology CRD to be removed from the cluster
            await waitUtils.waitForCrd(
                crds.topologyCrd,
                topologyName,
                undefined, // cluster-scoped
                timeout,
                false // desiredResult = false (waiting for deletion)
            );

            logger.info(`Topology ${topologyName}: CRD deleted successfully`);
        } catch (error) {
            logger.error(`Failed to wait for topology CRD deletion: ${topologyName}`, error);
            throw error;
        }
    }

    /**
     * Get all topology CRDs from the cluster
     * @returns Promise<any[]> - Array of topology CRDs
     */
    async getAllTopologyCRDs(): Promise<any[]> {
        try {
            const topologyCrds = await kubeCli.getCrds(crds.topologyCrd);
            logger.info(`Retrieved ${topologyCrds.length} topology CRDs from cluster`);
            return topologyCrds;
        } catch (error) {
            logger.error('Failed to get all topology CRDs', error);
            throw error;
        }
    }

    /**
     * Check if topology CRD exists in cluster
     * @param topologyName - Name of the topology
     * @returns Promise<boolean> - True if exists, false otherwise
     */
    async topologyCRDExists(topologyName: string): Promise<boolean> {
        try {
            await kubeCli.getCrd(crds.topologyCrd, topologyName, undefined);
            return true;
        } catch (error) {
            return false;
        }
    }

    /**
     * Get topology CRD status information
     * @param topologyName - Name of the topology
     * @returns Promise<any> - Status information from the CRD
     */
    async getTopologyCRDStatus(topologyName: string): Promise<any> {
        try {
            const topologyCrd = await this.getTopologyCRDByName(topologyName, 30);

            if (topologyCrd.status) {
                logger.info(`Topology ${topologyName}: Status: ${JSON.stringify(topologyCrd.status)}`);
                return topologyCrd.status;
            } else {
                logger.info(`Topology ${topologyName}: No status information available`);
                return null;
            }
        } catch (error) {
            logger.error(`Failed to get topology CRD status: ${topologyName}`, error);
            throw error;
        }
    }

    /**
     * Validate topology CRD metadata
     * @param topologyCrd - The topology CRD object
     * @param expectedName - Expected name of the topology
     * @param expectedLabels - Expected labels (optional)
     */
    validateTopologyCRDMetadata(
        topologyCrd: any,
        expectedName: string,
        expectedLabels?: Record<string, string>
    ): void {
        try {
            // Validate metadata exists
            expect(topologyCrd).to.have.property('metadata');

            const metadata = topologyCrd.metadata;

            // Validate name
            expect(metadata.name).to.equal(expectedName, `Topology CRD name should be ${expectedName}`);

            // Validate labels if provided
            if (expectedLabels) {
                expect(metadata).to.have.property('labels');

                for (const [key, value] of Object.entries(expectedLabels)) {
                    expect(metadata.labels[key]).to.equal(
                        value,
                        `Topology CRD should have label ${key}=${value}`
                    );
                }
            }

            logger.info(`Topology CRD metadata validation passed for: ${expectedName}`);
        } catch (error) {
            logger.error(`Topology CRD metadata validation failed for: ${expectedName}`, error);
            throw error;
        }
    }

    /**
     * Create topology CRD directly in the cluster (bypassing Control Plane API)
     * @param topologyName - Name of the topology to create
     * @param levels - Array of topology levels
     * @returns Promise<any> - The created topology CRD
     */
    async createTopologyCRDInCluster(topologyName: string, levels: string[]): Promise<any> {
        try {
            const topologyBody = {
                apiVersion: `${crds.topologyCrd.group}/${crds.topologyCrd.version}`,
                kind: crds.topologyCrd.kind,
                metadata: {
                    name: topologyName,
                    labels: {
                        'run.ai/control-plane-managed': 'false',
                        NetworkTopologySourceLabel: 'Cluster'
                    }
                },
                spec: {
                    levels: levels.map((level) => ({ nodeLabel: level }))
                }
            };

            const createdTopology = await kubeCli.createCrd(
                crds.topologyCrd,
                topologyBody,
                undefined // cluster-scoped
            );

            logger.info(`Topology CRD ${topologyName} created directly in cluster`);
            return createdTopology;
        } catch (error) {
            logger.error(`Failed to create topology CRD directly in cluster: ${topologyName}`, error);
            throw error;
        }
    }

    /**
     * Delete topology CRD directly from the cluster (bypassing Control Plane API)
     * @param topologyName - Name of the topology to delete
     * @returns Promise<any> - The deletion response
     */
    async deleteTopologyCRDFromCluster(topologyName: string): Promise<any> {
        try {
            const deletionResponse = await kubeCli.deleteCrd(
                crds.topologyCrd,
                topologyName,
                undefined // cluster-scoped
            );

            logger.info(`Topology CRD ${topologyName} deleted directly from cluster`);
            return deletionResponse;
        } catch (error) {
            logger.error(`Failed to delete topology CRD directly from cluster: ${topologyName}`, error);
            throw error;
        }
    }

    /**
     * Wait for topology CRD to be recreated in cluster after deletion
     * This is useful for testing that Control Plane recreates topologies that were deleted from cluster
     * @param topologyName - Name of the topology
     * @param timeout - Timeout in seconds for waiting
     * @returns Promise<any> - The recreated topology CRD
     */
    async waitForTopologyCRDRecreation(topologyName: string, timeout: number = 120): Promise<any> {
        try {
            logger.info(`Waiting for topology CRD ${topologyName} to be recreated in cluster...`);

            // Wait for the topology CRD to be recreated
            await waitUtils.waitForCrd(
                crds.topologyCrd,
                topologyName,
                undefined, // cluster-scoped
                timeout
            );

            // Get the recreated topology CRD
            const topologyCrd = await this.getTopologyCRDByName(topologyName, 30);

            logger.info(`Topology CRD ${topologyName} was recreated in cluster`);
            return topologyCrd;
        } catch (error) {
            logger.error(`Failed to wait for topology CRD recreation: ${topologyName}`, error);
            throw error;
        }
    }

    async waitForNetworkTopologies(filterBy: string[], timeout = 30): Promise<NetworkTopology[]> {
        let networkTopologies: NetworkTopology[];

        const isNetworkTopologyExistsInDB = async (): Promise<boolean> => {
            try {
                const response = await networkTopologiesApi.getNetworkTopologies(filterBy);
                networkTopologies = response.body.networkTopologies;

                return networkTopologies !== undefined && networkTopologies.length > 0;
            } catch (e) {
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isNetworkTopologyExistsInDB,
                `waitForNetworkTopologies: filterBy ${filterBy}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(`waitForNetworkTopologies for node pool ${filterBy}`);
            throw e;
        }

        return networkTopologies;
    }

    /**
     * Delete network topology via Control Plane API
     * @param topologyId - ID of the topology to delete
     * @returns Promise<{ response: http.IncomingMessage; body?: any;  }> - The deletion response
     */
    async deleteNetworkTopologyViaApi(
        topologyId: string
    ): Promise<{ response: http.IncomingMessage; body?: any }> {
        try {
            const deletionResponse = await networkTopologiesApi.deleteNetworkTopology(topologyId, {
                headers: {}
            });
            logger.info(`Network topology ${topologyId} deleted via API`);
            return deletionResponse;
        } catch (error) {
            logger.error(`Failed to delete network topology via API: ${topologyId}`, error);
            throw error;
        }
    }

    /**
     * Get network topology by ID via Control Plane API
     * @param topologyId - ID of the topology to retrieve
     * @returns Promise<{ response: http.IncomingMessage; body?: any;  }> - The topology response
     */
    async getNetworkTopologyByIdViaApi(
        topologyId: string
    ): Promise<{ response: http.IncomingMessage; body?: any }> {
        try {
            const topologyResponse = await networkTopologiesApi.getNetworkTopologyById(topologyId, {
                headers: {}
            });
            logger.info(`Network topology ${topologyId} retrieved via API`);
            return topologyResponse;
        } catch (error) {
            logger.error(`Failed to get network topology via API: ${topologyId}`, error);
            throw error;
        }
    }

    /**
     * Get network topologies via Control Plane API
     * @param filterBy - Optional array of filter parameters
     * @returns Promise<{ response: http.IncomingMessage; body?: any;  }> - The topologies response
     */
    async getNetworkTopologiesViaApi(
        filterBy?: string[]
    ): Promise<{ response: http.IncomingMessage; body?: any }> {
        try {
            const topologiesResponse = await networkTopologiesApi.getNetworkTopologies(filterBy, {
                headers: {}
            });
            logger.info(
                `Network topologies retrieved via API${filterBy ? ` with filters: ${filterBy.join(', ')}` : ''}`
            );
            return topologiesResponse;
        } catch (error) {
            logger.error(
                `Failed to get network topologies via API${filterBy ? ` with filters: ${filterBy.join(', ')}` : ''}`,
                error
            );
            throw error;
        }
    }

    /**
     * Create network topology via Control Plane API
     * @param createRequest - The network topology creation request
     * @returns Promise<{ response: http.IncomingMessage; body?: any;  }> - The creation response
     */
    async createNetworkTopologyViaApi(
        createRequest: NetworkTopologyCreateRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body?: NetworkTopology }> {
        networkTopologiesApi.accessToken = accessToken;
        try {
            const creationResponse = await networkTopologiesApi.createNetworkTopology(createRequest, {
                headers: {}
            });
            logger.info(`Network topology ${createRequest.name} created via API`);
            return creationResponse;
        } catch (error) {
            logger.error(`Failed to create network topology via API: ${createRequest.name}`, error);
            throw error;
        }
    }

    assertTopologyConstraint(actual: PodGroupTopologyConstraint, expected: PodGroupTopologyConstraint): void {
        expect(actual, 'topologyConstraint should exist').to.exist;

        expect(actual, 'topologyConstraint should not be empty').to.not.be.empty;
        if (expected.preferredTopologyLevel) {
            expect(actual.preferredTopologyLevel).to.be.equal(expected.preferredTopologyLevel);
        }
        if (expected.requiredTopologyLevel) {
            expect(actual.requiredTopologyLevel).to.be.equal(expected.requiredTopologyLevel);
        }
        expect(actual.topology).to.be.equal(expected.topology, 'topology should match');
        logger.info(`PodGroup has topology constraints: ${JSON.stringify(actual)}`);
    }

    validateNetworkTopologyDataInWorkloadsApi(
        actual_topology: WorkloadTopology,
        exp_topo_name: string,
        exp_topo_preferredPlacement: string | undefined,
        exp_topo_requiredPlacement: string | undefined
    ): void {
        expect(actual_topology.name).to.equal(exp_topo_name);
        expect(actual_topology.preferredPlacement).to.equal(exp_topo_preferredPlacement);
        expect(actual_topology.requiredPlacement).to.equal(exp_topo_requiredPlacement);
    }
}
