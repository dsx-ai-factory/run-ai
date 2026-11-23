/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { config, flowRootState } from '../config/config';
import { logger } from '../logger/logger';
import { WaitUtils } from './WaitUtils';
import { WorkloadsUtilsV2 } from './WorkloadsUtilsV2';
import { WorkloadsCrudApiUtils } from './WorkloadsCrudUtils';
import { bashCli, kubeCli } from '../../cli';
import { DistributedInferenceSpecSpec } from '../../api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { Workload } from '../../api/apis/workloadsApi/generated/model/workload';
import { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import { crds } from '../models/customResourceDefinition';

export class DistributedInferenceUtils {
    private waitUtils: WaitUtils;

    private workloadsUtilsV2: WorkloadsUtilsV2;

    private workloadsCrudUtils: WorkloadsCrudApiUtils;

    constructor() {
        this.waitUtils = new WaitUtils();
        this.workloadsUtilsV2 = new WorkloadsUtilsV2();
        this.workloadsCrudUtils = new WorkloadsCrudApiUtils();
    }

    /**
     * Creates a distributed inference workload and returns [workloadId, workloadName]
     */
    async createDistributedInference(
        workloadName: string,
        workloadSpec: DistributedInferenceSpecSpec,
        projectId: string
    ): Promise<string[]> {
        logger.debug(`creating distributed inference workload ${workloadName} in ${projectId}`);
        const result = await this.workloadsCrudUtils.createDistributedInference(
            workloadName,
            true /* add suffix */,
            flowRootState.cluster.uuid,
            projectId,
            workloadSpec
        );

        expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

        const workloadIdReturn = result.body.workloadId;
        const workloadNameReturn = result.body.name;

        return [workloadIdReturn, workloadNameReturn];
    }

    /**
     * Gets the base workload spec for distributed inference
     */
    getBaseWorkloadSpec(): DistributedInferenceSpecSpec {
        return {
            workers: 1,
            replicas: 1,
            servingPort: {
                port: 8080
            },
            leader: {
                image: config.quickStartImage.inferenceHelloWorld
            },
            worker: {
                image: config.quickStartImage.inferenceHelloWorld
            }
        };
    }

    /**
     * Gets the number of workers from a distributed inference spec
     */
    getNumWorkers(workloadSpec: DistributedInferenceSpecSpec): number {
        return workloadSpec?.workers ?? 0;
    }

    /**
     * Gets the number of replicas from a distributed inference spec
     */
    getNumReplicas(workloadSpec: DistributedInferenceSpecSpec): number {
        return workloadSpec?.replicas ?? 1;
    }

    /**
     * Gets the ready replicas count from a workload
     */
    getReadyReplicas(workload: Workload): number {
        return workload?.additionalFields?.readyReplicas ?? 0;
    }

    /**
     * Gets the spec replicas count from a workload
     */
    getNumReplicasForWorkload(workload: Workload): number {
        return workload?.additionalFields?.specReplicas ?? 0;
    }

    /**
     * Waits for a workload to have a specific number of replicas
     */
    async waitForNumReplicas(
        workloadId: string,
        expectedNumReplicas: number,
        timeout: number = 60
    ): Promise<Workload> {
        let workload: Workload;

        const isWorkloadWithCorrectNumReplicas = async (): Promise<boolean> => {
            try {
                workload = (await this.workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
                return this.getNumReplicasForWorkload(workload) === expectedNumReplicas;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        logger.debug(`waiting for workload ${workloadId} to have ${expectedNumReplicas} replicas`);
        await this.waitUtils.waitForCondition(
            isWorkloadWithCorrectNumReplicas,
            `workload: ${workloadId}, expected replicas: ${expectedNumReplicas}`,
            true,
            timeout
        );

        return workload;
    }

    /**
     * Waits for a workload to have a specific number of ready replicas
     */
    async waitForReadyReplicas(
        workloadId: string,
        expectedReadyReplicas: number,
        timeout: number = 60
    ): Promise<Workload> {
        let workload: Workload;

        const isWorkloadWithCorrectReadyReplicas = async (): Promise<boolean> => {
            try {
                workload = (await this.workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
                return this.getReadyReplicas(workload) === expectedReadyReplicas;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        logger.debug(`waiting for workload ${workloadId} to have ${expectedReadyReplicas} ready replicas`);
        await this.waitUtils.waitForCondition(
            isWorkloadWithCorrectReadyReplicas,
            `workload: ${workloadId}, expected ready replicas: ${expectedReadyReplicas}`,
            true,
            timeout
        );

        return workload;
    }

    /**
     * Validates resource allocation for a distributed inference workload
     */
    validateResourceAllocation(workload: Workload, workloadSpec: DistributedInferenceSpecSpec): void {
        const leaderResources =
            workloadSpec.leader?.compute?.gpuDevicesRequest ??
            workloadSpec.leader?.compute?.gpuPortionRequest ??
            0;

        const workerResources =
            workloadSpec.worker?.compute?.gpuDevicesRequest ??
            workloadSpec.worker?.compute?.gpuPortionRequest ??
            0;

        const workers = this.getNumWorkers(workloadSpec);
        const replicas = this.getNumReplicas(workloadSpec);
        const readyReplicas = this.getReadyReplicas(workload);

        const expectedRequestedGpuResources = (leaderResources + workers * workerResources) * replicas;
        const expectedAllocatedGpuResources = (leaderResources + workers * workerResources) * readyReplicas;

        const podsRequestedGpu = workload.podsRequestedResources?.gpu?.request ?? 0;
        expect(podsRequestedGpu).to.be.closeTo(expectedRequestedGpuResources, 0.0001);

        const workloadRequestedGpu = workload.workloadRequestedResources?.gpu?.request ?? 0;
        expect(workloadRequestedGpu).to.be.closeTo(expectedRequestedGpuResources, 0.0001);

        const allocatedGpu = workload.allocatedResources?.gpu ?? 0;
        expect(allocatedGpu).to.be.closeTo(expectedAllocatedGpuResources, 0.0001);
    }

    /**
     * Validates URLs for a distributed inference workload
     */
    validateUrls(workload: Workload, workloadSpec: DistributedInferenceSpecSpec): void {
        const readyReplicas = this.getReadyReplicas(workload);
        const urls = workload.urls || [];

        if (readyReplicas === 0) {
            expect(urls).to.have.lengthOf(0);
            return;
        }

        expect(urls).to.have.length.at.least(1);

        for (const url of urls) {
            const startsWithHttps = url.startsWith('https://');
            const startsWithHttpAndHasClusterLocal =
                url.startsWith('http://') && url.includes('cluster.local');

            expect(startsWithHttps || startsWithHttpAndHasClusterLocal).to.be.true;
        }
    }

    /**
     * Validates a get distributed inference response
     */
    validateGetDistributedInference(
        workload: Workload,
        workloadSpec: DistributedInferenceSpecSpec,
        validateAllocation = true
    ): void {
        logger.debug('validating the get workload response');
        const workers = this.getNumWorkers(workloadSpec);
        const replicas = this.getNumReplicas(workloadSpec);
        const expectedPods = (1 + workers) * replicas;

        const expectedImages = new Set([
            workloadSpec.leader.image,
            ...(workloadSpec.worker && workloadSpec.worker.image ? [workloadSpec.worker.image] : [])
        ]);
        expect(workload.images).to.have.lengthOf(expectedImages.size);
        expect(workload.images).to.have.members(Array.from(expectedImages));
        expect(workload.requestedPods.number).to.equal(expectedPods);

        expect(workload.preemptible).to.equal(false);
        expect(workload.category).to.equal('Deploy');
        expect(workload.priorityClassName).to.equal('very-high');

        expect(workload).to.have.property('additionalFields');
        const additionalFields = workload.additionalFields;

        if (workloadSpec.startupPolicy) {
            expect(additionalFields).to.have.property('startupPolicy');
            expect(additionalFields.startupPolicy).to.equal(workloadSpec.startupPolicy);
        }

        if (workloadSpec.restartPolicy) {
            expect(additionalFields).to.have.property('restartPolicy');
            expect(additionalFields.restartPolicy).to.equal(workloadSpec.restartPolicy);
        }

        if (workload.phase === Phase.Running) {
            expect(workload.runningPods).to.equal(expectedPods);
        }

        if (workload.phase !== Phase.Creating && workload.phase !== Phase.Deleting) {
            expect(workload.phaseMessage).not.to.be.undefined;
            expect(workload.phaseMessage).not.to.be.empty;
            expect(workload.k8sPhase).not.to.be.undefined;
            expect(workload.k8sPhase).not.to.be.empty;
        }

        if (workload.phase !== Phase.Degraded) {
            this.validateUrls(workload, workloadSpec);
        }

        if (validateAllocation) {
            this.validateResourceAllocation(workload, workloadSpec);
        }
    }

    /**
     * Gets pod IDs for a workload
     */
    async getWorkloadPodIds(workloadId: string): Promise<string[]> {
        logger.debug(`Getting pod ids for workload ${workloadId}`);

        try {
            const podsResponse = await this.workloadsUtilsV2.getWorkloadPods(workloadId);
            const podIds = podsResponse.body.pods.map((pod) => pod.id);

            logger.debug(`Found ${podIds.length} pods for workload ${workloadId}: ${podIds.join(', ')}`);
            return podIds;
        } catch (error) {
            logger.error(`Failed to get pod ids for workload ${workloadId}: ${error.message}`);
            throw error;
        }
    }

    /**
     * Logs pod statuses, security context, recent events and container logs (current and
     * previous) for every pod of a workload. Call it when waiting for the workload times out so
     * a crash-looping LeaderWorkerSet group leaves evidence in the test log. Never throws.
     */
    async logWorkloadPodsDiagnostics(workloadName: string, namespace: string, tailLines = 60): Promise<void> {
        try {
            const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadName);
            logger.info(
                `diagnostics for workload ${workloadName}: ${pods.length} pods in namespace ${namespace}`
            );

            for (const pod of pods) {
                const podName = pod.metadata?.name;
                const containers = (pod.status?.containerStatuses ?? [])
                    .map(
                        (c) =>
                            `${c.name}: ready=${c.ready} restarts=${c.restartCount} state=${JSON.stringify(c.state)} lastState=${JSON.stringify(c.lastState)}`
                    )
                    .join('; ');
                const runAsUser = pod.spec?.securityContext?.runAsUser ?? 'unset';
                const scc = pod.metadata?.annotations?.['openshift.io/scc'] ?? 'n/a';
                logger.info(
                    `pod ${podName} phase=${pod.status?.phase} node=${pod.spec?.nodeName} runAsUser=${runAsUser} scc=${scc} containers=[${containers}]`
                );

                for (const previous of [false, true]) {
                    const kind = previous ? 'previous' : 'current';
                    try {
                        const log = await kubeCli.CoreV1Api.readNamespacedPodLog({
                            name: podName,
                            namespace,
                            tailLines,
                            previous
                        });
                        logger.info(
                            `--- ${kind} logs of pod ${podName} (last ${tailLines} lines) ---\n${log}`
                        );
                    } catch (error) {
                        logger.debug(`no ${kind} logs for pod ${podName}: ${error.message ?? error}`);
                    }
                }
            }

            const events = (await kubeCli.CoreV1Api.listNamespacedEvent({ namespace })).items
                .filter((event) => event.involvedObject?.name?.startsWith(workloadName))
                .sort((a, b) => (a.lastTimestamp?.getTime() ?? 0) - (b.lastTimestamp?.getTime() ?? 0))
                .slice(-40);
            const eventLines = events.map(
                (event) =>
                    `${event.lastTimestamp?.toISOString() ?? ''} ${event.type} ${event.reason} ${event.involvedObject?.kind}/${event.involvedObject?.name}: ${event.message}`
            );
            logger.info(`last ${events.length} events for ${workloadName}:\n${eventLines.join('\n')}`);
        } catch (error) {
            logger.warn(
                `failed to collect diagnostics for workload ${workloadName}: ${error.message ?? error}`
            );
        }
    }

    /**
     * Scales a distributed inference workload via API
     */
    async scaleDistributedInferenceViaApi(workloadId: string, newReplicas: number): Promise<void> {
        logger.info(
            `Scaling distributed inference workload ${workloadId} to ${newReplicas} replicas via API`
        );

        const updateRequest = {
            spec: {
                replicas: newReplicas
            }
        };

        await this.workloadsCrudUtils.updateDistributedInferenceSpec(workloadId, updateRequest);

        await this.waitForReadyReplicas(workloadId, newReplicas, 120);

        logger.info(
            `Successfully scaled distributed inference workload ${workloadId} to ${newReplicas} replicas via API`
        );
    }

    /**
     * Scales a distributed inference workload resource via kubectl
     */
    async scaleDistributedInferenceViaKubectl(
        workloadId: string,
        workloadName: string,
        namespace: string,
        newReplicas: number
    ): Promise<void> {
        logger.info(
            `Scaling distributed inference workload ${workloadName} to ${newReplicas} replicas via kubectl`
        );

        const scaleCommand = `kubectl scale distributedinferenceworkload ${workloadName} --replicas=${newReplicas} -n ${namespace}`;

        try {
            bashCli.exec(scaleCommand);
            logger.info(`Successfully executed kubectl scale command: ${scaleCommand}`);

            await this.waitForNumReplicas(workloadId, newReplicas);

            logger.info(
                `Successfully scaled distributed inference workload ${workloadName} to ${newReplicas} replicas via kubectl`
            );
        } catch (error) {
            logger.error(
                `Failed to scale distributed inference workload ${workloadName} via kubectl: ${error}`
            );
            throw error;
        }
    }

    /**
     * Scales a distributed inference workload by patching its WorkloadRunner via kubectl
     */
    async scaleWorkloadRunnerViaKubectl(
        workloadId: string,
        workloadName: string,
        namespace: string,
        newReplicas: number,
        expectedPods: number
    ): Promise<void> {
        logger.info(
            `Scaling WorkloadRunner for distributed inference workload ${workloadName} to ${newReplicas} replicas via kubectl`
        );

        try {
            const workloadRunner = await kubeCli.getCrd(
                crds.workloadRunnerCrd,
                workloadName,
                namespace
            );
            const workloadRunnerName = workloadRunner?.metadata?.name;
            if (!workloadRunnerName) {
                throw new Error(
                    `WorkloadRunner matching workload ${workloadName} was not found in namespace ${namespace}`
                );
            }

            const patch = { spec: { template: { spec: { replicas: newReplicas } } } };
            const patchCommand = `kubectl patch workloadrunners.run.ai ${workloadRunnerName} --type=merge -p '${JSON.stringify(patch)}' -n ${namespace}`;
            bashCli.exec(patchCommand);
            logger.info(`Successfully patched WorkloadRunner ${workloadRunnerName}`);

            await this.waitUtils.waitForCondition(
                async () => {
                    const pods = await kubeCli.listPodsOfJob(namespace, workloadName, true);
                    return pods.length === expectedPods;
                },
                `workload ${workloadName} to have ${expectedPods} pods in namespace ${namespace}`,
                true,
                60,
                2
            );
            logger.info(
                `Validated distributed inference workload ${workloadName} scaled to ${expectedPods} pods in the cluster`
            );

            await this.waitForNumReplicas(workloadId, newReplicas);

            logger.info(
                `Successfully scaled WorkloadRunner ${workloadRunnerName} to ${newReplicas} replicas via kubectl`
            );
        } catch (error) {
            logger.error(
                `Failed to scale WorkloadRunner for distributed inference workload ${workloadName} via kubectl: ${error}`
            );
            throw error;
        }
    }

    /**
     * Validates scaling results for distributed inference workloads
     */
    async validateScalingResults(
        workloadId: string,
        expectedReplicas: number,
        workloadSpec: DistributedInferenceSpecSpec,
        originalPodIds?: string[]
    ): Promise<string[]> {
        logger.info(`Validating scaling results for workload ${workloadId}`);

        const workload = await this.workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

        const workers = this.getNumWorkers(workloadSpec);
        const expectedPods = (1 + workers) * expectedReplicas;

        expect(workload.body.requestedPods.number).to.equal(expectedPods);

        let newPodIds = originalPodIds;
        if (originalPodIds && originalPodIds.length > 0) {
            logger.info(`Validating that ${originalPodIds.length} original pods were not recreated`);

            const currentPodsResponse = await this.workloadsUtilsV2.getWorkloadPods(workloadId);
            const currentPods = currentPodsResponse.body.pods;
            newPodIds = currentPods.map((pod) => pod.id);

            if (originalPodIds.length < expectedPods) {
                // scaled up, all original pods should remain
                this.isSubset(newPodIds, originalPodIds);
            } else {
                // scaled down, current pods should be part of the old ones
                this.isSubset(originalPodIds, newPodIds);
            }
        }

        workloadSpec.replicas = expectedReplicas;

        this.validateGetDistributedInference(workload.body, workloadSpec);

        logger.info(`Successfully validated scaling results for workload ${workloadId}`);

        return newPodIds;
    }

    private isSubset(superset: string[], subset: string[]): void {
        logger.info(
            `Validating subset: checking if ${JSON.stringify(subset)} is contained in ${JSON.stringify(superset)}`
        );

        for (const item of subset) {
            expect(superset).to.include(item, `Expected ${JSON.stringify(superset)} to include ${item}`);
        }
    }
}
